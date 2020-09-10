package dynamodb

import (
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

//Custom err types
// ------------------------------------------------

//ErrUserItemNotFound is thrown when the db doesn't find the user item resource
var ErrUserItemNotFound = errors.New("User metadata item not found in db")

//ErrOrderItemNotFound is thrown when db doesn't find the order item resource
var ErrOrderItemNotFound = errors.New("Order item(s) not found in db")

//ErrOrderLineItemNotFound is thrown when db doesn't find the order line item resource
var ErrOrderLineItemNotFound = errors.New("Order line item not found in db")

//Order status options represents a customer's order actions
// ------------------------------------------------
const (
	// StatusPendingOrder   = "PENDING"   //Order pending status
	StatusNewOrder       = "PLACED"    //Order placed status
	StatusShippedOrder   = "SHIPPED"   //Order shipped status
	StatusDeliveredOrder = "DELIVERED" //Order delivered status
	// StatusCancelledOrder = "CANCELLED" //Order was cancelled by customer
	// StatusDisputedOrder  = "DISPUTED"  //Order was disputed by customer
	// StatusRefundedOrder  = "REFUNDED"  //Seller refunded customer order

)

//Custom Struct types for Users, Orders, and OrderLineItem entities
// ------------------------------------------------

//User struct represents the user's metadata item attribute fields
type User struct {
	ID          SortableID `json:"userId" dynamodbav:"UserId"`
	CreatedDate time.Time  `json:"createdUtc" dynamodbav:"CreatedUtc,omitempty"`
	UserName    string     `json:"userName" dynamodbav:"UserName"`
	Email       string     `json:"email" dynamodbav:"Email"`
	FirstName   string     `json:"firstName" dynamodbav:"FirstName"`
	LastName    string     `json:"lastName" dynamodbav:"LastName"`
}

//Address will be marshalled into a dynamodb map attr
//in the order item
type Address struct {
	AddressType   string
	StreetAddress string
	ZipCode       string
	State         string
	Country       string //TODO:  ISO 3166-1 alpha-2 country code instead?
}

//Order struct represents an order attribute fields for a user's order
type Order struct {
	OrderID         SortableID `json:"orderId" dynamodbav:"OrderId"`
	UserID          SortableID `json:"userId" dynamodbav:"UserId"`
	PurchasedDate   time.Time  `json:"createdUtc" dynamodbav:"CreatedUtc"`
	ModifiedDate    time.Time  `json:"modifiedDate" dynamodbav:"ModifiedDate"`
	ShippingAddress Address    `json:"shippingAddress" dynamodbav:"ShippingAddress"`
	Status          string     `json:"status" dynamodbav:"Status"`
	TotalAmount     int        `json:"totalAmount" dynamodbav:"TotalAmount"`
	EstDeliverDate  time.Time  `json:"estDeliverDate" dynamodbav:"EstDeliverDate,omitempty"`
	ActDeliverDate  time.Time  `json:"ActDeliverDate" dynamodbav:"ActDeliverDate,omitempty"`
	// Currency int  //TODO: add currency types
}

//OrderLineItem represents itemized attribute fields for an order
type OrderLineItem struct {
	ItemID      SortableID `json:"orderItemId" dynamodbav:"OrderItemId"`
	OrderID     SortableID `json:"orderId" dynamodbav:"OrderId"`
	ProductID   SortableID `json:"productId" dynamodbav:"ProductId"`
	Price       int        `json:"price" dynamodbav:"Price"`
	Quantity    int        `json:"quantity" dynamodbav:"Quantity"`
	TotalAmount int        `json:"totalAmount" dynamodbav: "TotalAmount"`
	// Currency int  //TODO: add currency types
}

// ------------------------------------------------

//TODO: check db to see if that email is in already in the db

//AddUser takes a User struct and marshalls it to a ddb item on the db
func (db *DynamoDB) AddUser(u User) (User, error) {

	u.ID = NewSortableID()
	u.CreatedDate = time.Now()

	pk := fmt.Sprintf("USER#%s", u.ID)
	sort := "METADATA#"
	item, err := dynamodbattribute.MarshalMap(&u)
	if err != nil {
		return User{}, err
	}

	item["Type"] = &dynamodb.AttributeValue{S: aws.String("Person")}
	item["PK"] = &dynamodb.AttributeValue{S: aws.String(pk)}
	item["SK"] = &dynamodb.AttributeValue{S: aws.String(sort)}

	//TODO:
	//if we update the email field; we'd need to update the GSIPK=EMAIL<email>
	gs1pk := fmt.Sprintf("EMAIL#%s", u.Email)
	gs1sk := sort
	item["GSI1PK"] = &dynamodb.AttributeValue{S: aws.String(gs1pk)}
	item["GSI1SK"] = &dynamodb.AttributeValue{S: aws.String(gs1sk)}

	//TODO: checkCond deprecated since uuid is generated in local scope
	//so we don't overwrite an existing user metadata item
	// checkCond := "attribute_not_exists(PK) AND attribute_not_exists(SK)"
	_, err = db.db.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(db.tableName),
		Item:      item,
		// ConditionExpression: aws.String(checkCond)
	})

	return u, err

}

//GetUser fetches a customer by their userId
//takes a user_id and returns a single user item with the metadata info
func (db *DynamoDB) GetUser(id SortableID) (User, error) {
	//TODO: refactor with getItem request instead for speed

	var result User
	pk := fmt.Sprintf("USER#%s", id)
	sort := "METADATA#"

	res, err := db.db.Query(&dynamodb.QueryInput{
		TableName:              aws.String(db.tableName),
		ScanIndexForward:       aws.Bool(false),
		KeyConditionExpression: aws.String("#PK = :pk AND #SK = :sk"),
		ExpressionAttributeNames: map[string]*string{
			"#PK": aws.String("PK"),
			"#SK": aws.String("SK"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":pk": {
				S: aws.String(pk),
			},
			":sk": {
				S: aws.String(sort),
			},
		},
	})

	if err != nil {
		return User{}, err

	}
	if len(res.Items) == 0 {
		return User{}, ErrUserItemNotFound
	}
	item := res.Items[0]

	err = dynamodbattribute.UnmarshalMap(item, &result)
	if err != nil {
		return User{}, err

	}

	// log.Printf("%+v", result)

	return result, nil

}

//GetUserByEmail fetches a single user by their email using the gsi1
func (db *DynamoDB) GetUserByEmail(email string) (User, error) {
	var result User
	gsi1pk := fmt.Sprintf("EMAIL#%s", email)
	gsi1sk := "METADATA#"

	res, err := db.db.Query(&dynamodb.QueryInput{
		TableName:              aws.String(db.tableName),
		IndexName:              aws.String("GSI1"),
		KeyConditionExpression: aws.String("#GSI1PK = :gsi1pk And #GSI1SK = :gsi1sk "),
		ExpressionAttributeNames: map[string]*string{
			"#GSI1PK": aws.String("GSI1PK"),
			"#GSI1SK": aws.String("GSI1SK"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":gsi1pk": {
				S: aws.String(gsi1pk),
			},
			":gsi1sk": {
				S: aws.String(gsi1sk),
			},
		},
	})
	if err != nil {
		return User{}, err
	}
	if len(res.Items) == 0 {
		return User{}, ErrUserItemNotFound
	}
	item := res.Items[0]
	err = dynamodbattribute.UnmarshalMap(item, &result)
	if err != nil {
		return User{}, err
	}
	return result, nil

}

//AddOrder takes a userId and attempts to put a new order item for the customer
func (db *DynamoDB) AddOrder(order Order) (Order, error) {

	order.OrderID = NewSortableID()
	order.PurchasedDate = time.Now()
	order.ModifiedDate = order.PurchasedDate
	order.Status = StatusNewOrder

	pk := fmt.Sprintf("USER#%s", order.UserID)
	sort := fmt.Sprintf("ORDER#%s", order.OrderID)

	//using an inverted index but with the new GSI1 attrs
	// For this GSI:
	// if we query GSI1PK=ORDER#<id> AND GSI1SK=METADATA#
	// we'll get the user that the order belongs to
	gs1pk := sort
	gs1sk := "METADATA#"

	//NOTE: gsi2 deprecated for now
	//setting GSI2 attrs
	//below
	//NOTE: since we'd need to update the status at least 2 more times
	// this may not be most optimal from a WCU pov, but we'll review this soon
	//using a composite sort key of OrderStatus & MODIFIEDDATE
	//GS2PK=USER#<id> GSI2SK=<STATUS>#<MODIFIEDDATE>
	// gs2pk := pk
	// gs2sk := fmt.Sprintf("%s#%s", order.Status, order.PurchasedDate)

	item, err := dynamodbattribute.MarshalMap(&order)
	if err != nil {
		return Order{}, err
	}
	item["Type"] = &dynamodb.AttributeValue{S: aws.String("UserOrder")}
	item["PK"] = &dynamodb.AttributeValue{S: aws.String(pk)}
	item["SK"] = &dynamodb.AttributeValue{S: aws.String(sort)}

	item["GSI1PK"] = &dynamodb.AttributeValue{S: aws.String(gs1pk)}
	item["GSI1SK"] = &dynamodb.AttributeValue{S: aws.String(gs1sk)}

	// item["GSI2PK"] = &dynamodb.AttributeValue{S: aws.String(gs2pk)}
	// item["GSI2SK"] = &dynamodb.AttributeValue{S: aws.String(gs2sk)}

	_, err = db.db.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(db.tableName),
		Item:      item,
	})

	return order, err

}

// GetOrdersByUserID fetches a customer's orders by their userId
//if forward is false, ScanIndexForward will be false (descending query)
func (db *DynamoDB) GetOrdersByUserID(uid SortableID, limit int64, forward bool) ([]Order, error) {
	var result []Order
	pk := fmt.Sprintf("USER#%s", uid)
	sort := "ORDER#"
	res, err := db.db.Query(&dynamodb.QueryInput{
		TableName:              aws.String(db.tableName),
		KeyConditionExpression: aws.String("#PK = :pk AND begins_with(#SK,:sk)"),
		ExpressionAttributeNames: map[string]*string{
			"#PK": aws.String("PK"),
			"SK":  aws.String("SK"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":pk": {
				S: aws.String(pk),
			},
			":sk": {
				S: aws.String(sort),
			},
		},
		Limit:            aws.Int64(limit),
		ScanIndexForward: aws.Bool(forward),
	})
	if err != nil {
		return nil, err
	}
	if len(res.Items) == 0 {
		return nil, ErrOrderItemNotFound
	}
	err = dynamodbattribute.UnmarshalListOfMaps(res.Items, &result)
	if err != nil {
		return nil, err
	}

	return result, err

}

/*
TODO:
-maybe modify with querybuilder api
https://docs.aws.amazon.com/sdk-for-go/api/service/dynamodb/expression/#KeyBuilder.BeginsWith

*/

//:one

//GetOrderByOrderID uses the orderID as an input
//and returns an order item using the gsi key
//entrypoint: fetch a user's order given an orderID
func (db *DynamoDB) GetOrderByOrderID(oid SortableID) (Order, error) {
	var result Order
	gsi1pk := fmt.Sprintf("ORDER#%s", oid)
	gsi1sk := "METADATA#"
	res, err := db.db.Query(&dynamodb.QueryInput{
		TableName:              aws.String(db.tableName),
		IndexName:              aws.String("GSI1"),
		KeyConditionExpression: aws.String("#GSI1PK = :gsi1pk And #GSI1SK = :gsi1sk"),
		ExpressionAttributeNames: map[string]*string{
			"#GSI1PK": aws.String("GSI1PK"),
			"#GSI1SK": aws.String("GSI1SK"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":gsi1pk": {
				S: aws.String(gsi1pk),
			},
			":gsi1sk": {
				S: aws.String(gsi1sk),
			},
		},
	})
	if err != nil {
		return Order{}, err
	}
	if len(res.Items) == 0 {
		return Order{}, ErrOrderItemNotFound
	}
	item := res.Items[0]

	err = dynamodbattribute.UnmarshalMap(item, &result)
	if err != nil {
		return Order{}, err
	}

	return result, err

}

func validStatus(status string) bool {
	items := []string{StatusNewOrder, StatusShippedOrder, StatusDeliveredOrder}
	for _, item := range items {
		if item == status {
			return true
		}
	}
	return false
}

func getPriorStatus(newStatus string) (string, error) {
	switch newStatus {
	case StatusShippedOrder:
		return StatusNewOrder, nil
	case StatusDeliveredOrder:
		return StatusShippedOrder, nil
	default:
		return "", errors.New("Order status is delivered. No need to update ddb item")
	}
}

// An alternative for order progression status updating
// and ensure

// UpdateOrderStatus updates a user's order status
// simple order progression goes from PLACED --> SHIPPED --> DELIVERED
func (db *DynamoDB) UpdateOrderStatus(uid, oid SortableID, newStatus string) error {

	modifiedDate := time.Now().Format(time.RFC3339)

	if !validStatus(newStatus) {
		return errors.New("invalid status value; check your status parameter")
	}
	oldStatus, err := getPriorStatus(newStatus)
	if err != nil {
		return err
	}

	pk := fmt.Sprintf("USER#%s", uid)
	sort := fmt.Sprintf("ORDER#%s", oid)

	_, err = db.db.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String(db.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"PK": {
				S: aws.String(pk),
			},
			"SK": {
				S: aws.String(sort),
			},
		},
		ReturnValues: aws.String("UPDATED_NEW"),
		// UpdateExpression: aws.String(
		// 	"SET #ModifiedDate = :m, #Status = :s, #GSI2SK = :gsi2sk"),
		UpdateExpression: aws.String(
			"SET #ModifiedDate = :m, #Status = :s"),
		ConditionExpression: aws.String("#Status = :old"),
		ExpressionAttributeNames: map[string]*string{
			"#ModifiedDate": aws.String("ModifiedDate"),
			"#Status":       aws.String("Status"),
			// "#GSI2SK":       aws.String("GSI2SK"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":m": {
				S: aws.String(modifiedDate),
			},
			":s": {
				S: aws.String(newStatus),
			},
			":old": {
				S: aws.String(oldStatus),
			},
			// ":gsi2sk": {
			// 	S: aws.String(fmt.Sprintf("%s#%s", newStatus, modifiedDate)),
			// },
		},
	})

	return err
}

//TODO: validate TotalAmount = Price * Quantity

//AddOrderLineItem adds a new order line item to a user's order
func (db *DynamoDB) AddOrderLineItem(item OrderLineItem) (OrderLineItem, error) {
	item.ItemID = NewSortableID()

	pk := fmt.Sprintf("ORDERITEM#%s", item.ItemID)
	sort := fmt.Sprintf("ORDER#%s", item.OrderID)

	//using an inverted index but with the new GSI1 attrs
	gs1pk := sort
	gs1sk := pk

	i, err := dynamodbattribute.MarshalMap(&item)
	if err != nil {
		return OrderLineItem{}, err
	}
	i["Type"] = &dynamodb.AttributeValue{S: aws.String("OrderLineItem")}
	i["PK"] = &dynamodb.AttributeValue{S: aws.String(pk)}
	i["SK"] = &dynamodb.AttributeValue{S: aws.String(sort)}
	i["GSI1PK"] = &dynamodb.AttributeValue{S: aws.String(gs1pk)}
	i["GSI1SK"] = &dynamodb.AttributeValue{S: aws.String(gs1sk)}
	_, err = db.db.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(db.tableName),
		Item:      i,
	})

	return item, err

}

//GetOrderLineItemsByOrderID fetches many order items
func (db *DynamoDB) GetOrderLineItemsByOrderID(oid SortableID) ([]OrderLineItem, error) {
	var result []OrderLineItem
	gsi1pk := fmt.Sprintf("ORDER#%s", oid)
	gsi1sk := "ORDERITEM#"
	res, err := db.db.Query(&dynamodb.QueryInput{
		TableName:              aws.String(db.tableName),
		ScanIndexForward:       aws.Bool(true),
		IndexName:              aws.String("GSI1"),
		KeyConditionExpression: aws.String("#GSI1PK = :gsi1pk AND begins_with(#GSI1SK,:gsi1sk)"),
		ExpressionAttributeNames: map[string]*string{
			"#GSI1PK": aws.String("GSI1PK"),
			"#GSI1SK": aws.String("GSI1SK"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":gsi1pk": {
				S: aws.String(gsi1pk),
			},
			":gsi1sk": {
				S: aws.String(gsi1sk),
			},
		},
	})
	if err != nil {
		return nil, err
	}
	if len(res.Items) == 0 {
		return nil, ErrOrderLineItemNotFound
	}
	err = dynamodbattribute.UnmarshalListOfMaps(res.Items, &result)
	if err != nil {
		return nil, err
	}
	return result, err

}

//deprecated for now
//entrypoint: get all user orders based on an orderID and status option
//GetUserOrdersByOrderIDAndStatus
// func (db *DynamoDB) GetUserOrdersByOrderIDAndStatus(oid SortableID) ([]Order, error) {

// }

//GetOpenOrders retrieves multiple users's open orders
// func (db *DynamoDB) GetOpenOrders(uid, oid SortableID) {

// }
