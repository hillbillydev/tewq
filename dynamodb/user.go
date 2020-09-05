package dynamodb

import (
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type User struct {
	ID          SortableID `json:"id" dynamodbav:"Id"`
	CreatedDate string     `json:"createdUtc" dynamodbav:"CreatedUtc,omitempty"`
	Email       string     `json:"email" dynamodbav:"Email"`
	FirstName   string     `json:"firstName" dynamodbav:"Email"`
	LastName    string     `json:"lastName" dynamodbav:"LastName"`
}

type Status int

/*
status ["PLACED", "SHIPPED","DELIVERED", "RETURNED"]
*/

const (
	OrderNew Status = iota + 1
	OrderShipped
	OrderDelivered
)

/*
TODO:
- make shipping address into a document type
- address -----> {"StreetAddress":"123 Main St", "State": "NY", "Country":"USA"}


*/

type Order struct {
	OrderID         SortableID `json:"orderId" dynamodbav:"OrderId"`
	UserID          SortableID `json:"userId" dynamodbav:"UserId"`
	PurchaseDate    string     `json:"purchaseDate" dynamodbav:"PurchaseDate"`
	ShippingAddress string     `json:"shippingAddress" dynamodbav:"ShippingAddress"`
	Status          string     `json:"status" dynamodbav:"Status"`
	// Status          Status     `json:"status" dynamodbav:"Status"`
	TotalAmount int    `json:"totalAmount" dynamodbav:"TotalAmount"`
	DeliverDate string `json:"deliverDate" dynamodbav:"DeliverDate"`
}

/*
TODO:
- turn price into float type
*/

type OrderItem struct {
	ItemID    SortableID `json:"orderItemId" dynamodbav:"OrderItemId"`
	OrderID   SortableID `json:"orderId" dynamodbav:"OrderId"`
	ProductID SortableID `json:"productId" dynamodbav:"ProductId"`
	Price     int        `json:"price" dynamodbav:"Price"`
	Quantity  int        `json:"quantity" dynamodbav:"Quantity"`
}

//AddUser takes a User struct and marshalls it to a ddb item on the db
func (db *DynamoDB) AddUser(u User) (User, error) {

	u.ID = NewSortableID()
	u.CreatedDate = time.Now().Format(time.RFC3339)
	pk := fmt.Sprintf("USER#%s", u.ID)
	sort := "METADATA#"
	item, err := dynamodbattribute.MarshalMap(&u)
	if err != nil {
		return User{}, err
	}

	item["Type"] = &dynamodb.AttributeValue{S: aws.String("person")}
	item["PK"] = &dynamodb.AttributeValue{S: aws.String(pk)}
	item["SK"] = &dynamodb.AttributeValue{S: aws.String(sort)}

	//TODO:
	//if we update the email field; we'd need to update the GSIPK=EMAIL<email>
	// gs1pk := fmt.Sprintf("EMAIL#%s", u.Email)
	// gs1sk := sort
	// item["GSI1PK"] = &dynamodb.AttributeValue{S: aws.String(gs1pk)}
	// item["GSI1SK"] = &dynamodb.AttributeValue{S: aws.String(gs1sk)}

	//TODO: checkCond deprecated since uuid is generated in local scope
	//so we don't overwrite an existing user metadata item
	checkCond := "attribute_not_exists(PK) AND attribute_not_exists(SK)"
	_, err = db.db.PutItem(&dynamodb.PutItemInput{
		TableName:           aws.String(db.tableName),
		Item:                item,
		ConditionExpression: aws.String(checkCond)})

	return u, err

}

//GetUser takes a user_id and returns a single user item with the metadata info
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
		}})

	if err != nil {
		return User{}, err

	}
	if len(res.Items) == 0 {
		return User{}, errors.New("No user item in db with PK given")
	}
	item := res.Items[0]

	err = dynamodbattribute.UnmarshalMap(item, &result)
	if err != nil {
		return User{}, err

	}
	// fmt.Println("kets see")
	// log.Printf("%+v", result)

	return result, nil

}

//AddNewOrderToUser takes a user_id and attempts to put a new item withe the User's Order
func (db *DynamoDB) AddNewOrderToUser(id SortableID, order Order) (Order, error) {
	/*
		TODO: gotta think about below some more
		PK=USER#<id> ; SK=ORDER<id>
		GSI1PK=ORDER<id> ; GSISK=USER#<id>
		GS12PK=EMAIL#<id> ; GSI2SK=ORDER<id>

	*/
	order.OrderID = NewSortableID()
	order.PurchaseDate = time.Now().Format(time.RFC3339)
	order.Status = "PLACED"

	pk := fmt.Sprintf("USER#%s", id)
	sort := fmt.Sprintf("ORDER#%s", order.OrderID)

	//using an inverted index but with the new GSI1 attrs
	gs1pk := sort
	gs1sk := pk

	item, err := dynamodbattribute.MarshalMap(&order)
	if err != nil {
		return Order{}, err
	}
	item["Type"] = &dynamodb.AttributeValue{S: aws.String("UserOrder")}
	item["PK"] = &dynamodb.AttributeValue{S: aws.String(pk)}
	item["SK"] = &dynamodb.AttributeValue{S: aws.String(sort)}
	item["GSI1PK"] = &dynamodb.AttributeValue{S: aws.String(gs1pk)}
	item["GSI1SK"] = &dynamodb.AttributeValue{S: aws.String(gs1sk)}

	_, err = db.db.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(db.tableName),
		Item:      item,
	})

	return order, err

}

func (db *DynamoDB) GetUserOrdersByUser() {

}

// func (db *DynamoDB) GetUserOrder(id SortableID) (Order, error) {

// }

func (db *DynamoDB) AddNewOrderItem(item OrderItem) error {
	pk := fmt.Sprintf("ITEM#%s", item.ItemID)
	sort := fmt.Sprintf("ORDER#%s", item.OrderID)
	//using an inverted index but with the new GSI1 attrs
	gs1pk := sort
	gs1sk := pk

	i, err := dynamodbattribute.MarshalMap(&item)
	if err != nil {
		return err
	}
	i["Type"] = &dynamodb.AttributeValue{S: aws.String("OrderItem")}
	i["PK"] = &dynamodb.AttributeValue{S: aws.String(pk)}
	i["SK"] = &dynamodb.AttributeValue{S: aws.String(sort)}
	i["GSI1PK"] = &dynamodb.AttributeValue{S: aws.String(gs1pk)}
	i["GSI1SK"] = &dynamodb.AttributeValue{S: aws.String(gs1sk)}
	_, err = db.db.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(db.tableName),
		Item:      i,
	})

	return err

}

// func (db *DynamoDB) GetUserOrderItem(id SortableID) (Order, error) {

// }
