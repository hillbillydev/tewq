package dynamodb

import (
	"errors"
	"fmt"
	"time"

	//TODO: switch to v2?
	//"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/segmentio/ksuid"
)

//Errors used by order queries
var (
	//ErrOrderItemNotFound is returned when db doesn't find the order item resource
	ErrOrderItemNotFound = errors.New("Order item(s) not found in db")
)

//Order status options represents a customer's order actions
const (
	// StatusPendingOrder   = "PENDING"   //Order pending status
	StatusNewOrder       = "PLACED"    //Order placed status
	StatusShippedOrder   = "SHIPPED"   //Order shipped status
	StatusDeliveredOrder = "DELIVERED" //Order delivered status
	// StatusCancelledOrder = "CANCELLED" //Order was cancelled by customer
	// StatusDisputedOrder  = "DISPUTED"  //Order was disputed by customer
	// StatusRefundedOrder  = "REFUNDED"  //Seller refunded customer order

)

//Address will be marshalled to a dynamodb address map attr
type Address struct {
	AddressType   string
	StreetAddress string
	ZipCode       string
	State         string
	Country       string //TODO:  ISO 3166-1 alpha-2 country code instead?
}

//Order struct represents an order attribute fields for a user's order
type Order struct {
	OrderID         SortableID      `json:"orderId" dynamodbav:"OrderId"`
	UserID          SortableID      `json:"userId" dynamodbav:"UserId"`
	PurchasedDate   time.Time       `json:"createdUtc" dynamodbav:"CreatedUtc"`
	ModifiedDate    time.Time       `json:"modifiedDate" dynamodbav:"ModifiedDate"`
	ShippingAddress Address         `json:"shippingAddress" dynamodbav:"ShippingAddress"`
	Status          string          `json:"status" dynamodbav:"Status"`
	TotalAmount     int             `json:"totalAmount" dynamodbav:"TotalAmount"`
	EstDeliverDate  time.Time       `json:"estDeliverDate" dynamodbav:"EstDeliverDate,omitempty"`
	ActDeliverDate  time.Time       `json:"ActDeliverDate" dynamodbav:"ActDeliverDate,omitempty"`
	OrderLineItems  []OrderLineItem `json:"orderLineItems"  dynamodbav:"-"`
}

//OrderLineItem represents itemized attribute fields for an order
type OrderLineItem struct {
	ItemID          SortableID `json:"orderItemId" dynamodbav:"OrderItemId"`
	OrderID         SortableID `json:"orderId" dynamodbav:"OrderId"`
	ProductID       SortableID `json:"productId" dynamodbav:"ProductId"`
	ProductOptionID SortableID `json:"productOptionId" dynamodbav:"ProductOptionId"`
	UnitPrice       int        `json:"unitPrice" dynamodbav:"UnitPrice"`
	Quantity        int        `json:"quantity" dynamodbav:"Quantity"`
	TotalAmount     int        `json:"totalAmount" dynamodbav:"TotalAmount"`
}

/*
TODO: fn to get the sum (TotalAmount) by iterating all the orderline items

*/

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

/*
For preventing a provisioned throughput exception
for high volume requests should we:
- use redis to handle the temporary spikes in request traffic?
- Amazon SQS and put write requests to SQS and a sepearte job to poll records from SQS at a lim rate and
*/

//TODO: and also clear the basket items when order putItem is completed?

//AddOrder starts with doing a transactionwrite to update the stock value of the product options
//then a batchwriteitem logic to write the user order items to ddb
func (db *DynamoDB) AddOrder(order Order) error {
	//batch and transactions both have a cap of 25 items in one request
	chunkSize := 25

	order.OrderID = NewSortableID()
	order.PurchasedDate = time.Now()
	order.ModifiedDate = order.PurchasedDate
	order.Status = StatusNewOrder

	orderItem, err := marshallOrderMetaItem(order)
	if err != nil {
		return err
	}

	oLineItems := order.OrderLineItems

	sizeOrderItemsToInput := len(oLineItems) + 1
	batchInput := make([]*dynamodb.WriteRequest, 0, sizeOrderItemsToInput)
	batchInput = append(batchInput, &dynamodb.WriteRequest{PutRequest: &dynamodb.PutRequest{Item: orderItem}})

	transactCap := len(oLineItems)
	transactUpdateInput := make([]*dynamodb.TransactWriteItem, 0, transactCap)

	for _, oLine := range oLineItems {
		transItem := &dynamodb.TransactWriteItem{Update: &dynamodb.Update{
			TableName: aws.String(db.tableName),
			Key: map[string]*dynamodb.AttributeValue{
				"PK": {
					S: aws.String(fmt.Sprintf("PRODUCT#%s", oLine.ProductID)),
				},
				"SK": {
					S: aws.String(fmt.Sprintf("OPTION#%s", oLine.ProductOptionID)),
				},
			},
			UpdateExpression:    aws.String("SET #Stock = #Stock - :decr"),
			ConditionExpression: aws.String("#Stock > :zero AND #Stock >= :decr"),
			ExpressionAttributeNames: map[string]*string{
				"#Stock": aws.String("Stock"),
			},
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":decr": {
					N: aws.String(fmt.Sprintf("%d", oLine.Quantity)),
				},
				":zero": {
					N: aws.String("0"),
				},
			},
		}}
		transactUpdateInput = append(transactUpdateInput, transItem)
		oLine.OrderID = order.OrderID
		item, err := marshallOrderLineItem(oLine)
		if err != nil {
			return err
		}
		wr := &dynamodb.WriteRequest{
			PutRequest: &dynamodb.PutRequest{
				Item: item,
			},
		}
		batchInput = append(batchInput, wr)
	}

	for i := 0; i < len(transactUpdateInput); i += chunkSize {
		_, err = db.db.TransactWriteItems(&dynamodb.TransactWriteItemsInput{
			TransactItems: transactUpdateInput[i:min(i+chunkSize, len(transactUpdateInput))],
		})
		if err != nil {
			return err
		}
	}
	fmt.Printf("len of batchwrite: %d \n", len(batchInput))
	//can comprise as much as 25 put item requests in one batch; each item can be 400kb's max
	for i := 0; i < len(batchInput); i += chunkSize {
		res, err := db.db.BatchWriteItem(&dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]*dynamodb.WriteRequest{
				db.tableName: batchInput[i:min(i+chunkSize, len(batchInput))],
			},
		})
		if err != nil {
			return err
		}
		unprocessed := res.UnprocessedItems
		sizeUnprocessed := len(unprocessed[db.tableName])
		//todo: check unprocessed items
		fmt.Printf("unprocessed len: %d \n ", len(unprocessed[db.tableName]))

		//https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Programming.Errors.html#Programming.Errors.BatchOperations
		//throttling is the most likely reason for batchwriteitem failure
		//todo:
		//if there's unprocessed items retry with an exponential backoff algorithm & delay our attempt
		//maxAttempts := 5
		for sizeUnprocessed != 0 {
			time.Sleep(time.Second * 3)
			res, err := db.db.BatchWriteItem(&dynamodb.BatchWriteItemInput{
				RequestItems: map[string][]*dynamodb.WriteRequest{db.tableName: unprocessed[db.tableName]}})
			if err != nil {
				return err
			}
			sizeUnprocessed = len(res.UnprocessedItems[db.tableName])
		}

	}
	return nil

}

//TODO: pagination
//orderlimit; offset

//GetFullOrderItems retrieves the orders with the respective orderline items
func (db *DynamoDB) GetFullOrderItems(userID SortableID, limit int64) ([]Order, error) {
	//we need to query on GSI first
	gsipk := fmt.Sprintf("USER#%s", userID)
	res, err := db.db.Query(&dynamodb.QueryInput{
		TableName:              aws.String(db.tableName),
		IndexName:              aws.String("GSI2KO"),
		KeyConditionExpression: aws.String("#GSI2PK = :gsi2pk"),
		ExpressionAttributeNames: map[string]*string{
			"#GSI2PK": aws.String("GSI2PK"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":gsi2pk": {
				S: aws.String(gsipk),
			},
		},
		Limit:            aws.Int64(limit),
		ScanIndexForward: aws.Bool(false),
	})
	if err != nil {
		return nil, err
	}
	if len(res.Items) == 0 {
		return nil, ErrOrderItemNotFound
	}
	attrs := []map[string]*dynamodb.AttributeValue{}
	for _, attr := range res.Items {
		attrs = append(attrs, map[string]*dynamodb.AttributeValue{
			":pk": {
				S: aws.String(*attr["PK"].S),
			},
		})
	}

	var result []Order
	for _, attr := range attrs {
		var order Order
		res, err := db.db.Query(&dynamodb.QueryInput{
			TableName:              aws.String(db.tableName),
			KeyConditionExpression: aws.String("#PK = :pk"),
			ExpressionAttributeNames: map[string]*string{
				"#PK": aws.String("PK"),
			},
			ExpressionAttributeValues: attr,
			ScanIndexForward:          aws.Bool(true),
		})
		if err != nil {
			return nil, err
		}
		orderMetaItem, orderLineItems := res.Items[0], res.Items[1:]
		err = dynamodbattribute.UnmarshalMap(orderMetaItem, &order)
		if err != nil {
			return nil, err
		}
		err = dynamodbattribute.UnmarshalListOfMaps(orderLineItems, &order.OrderLineItems)
		if err != nil {
			return nil, err
		}
		result = append(result, order)

	}
	// fmt.Printf("%+v", result)
	return result, nil

}

//UpdateOrderStatusInput is used for the UpdateOrderStatus fn
type UpdateOrderStatusInput struct {
	OrderID        SortableID //required
	NewStatus      string     //required
	ActDeliverDate time.Time
}

func (in *UpdateOrderStatusInput) validate() error {
	if ksuid.KSUID(in.OrderID).IsNil() {
		return errors.New("Expected input.OrderID attr to have a value")
	}
	if !validStatus(in.NewStatus) {
		return fmt.Errorf("input.NewStatus %q is invalid, check your status parameter", in.NewStatus)
	}
	if in.NewStatus == StatusDeliveredOrder && in.ActDeliverDate.IsZero() {
		return fmt.Errorf("Expected a non-zero value for input.ActDeliverDate attr when NewStatus is DELIVERED")
	}
	return nil
}

//UpdateOrderStatusOutput represents the ddb's return attrs when applying the UpdateOrderStatus fn; we're mainly interested in verifying the Status
type UpdateOrderStatusOutput struct {
	// ModifiedDate time.Time
	Status string
}

//TODO: updateItem only succeed if there is a primary key that exists?
// updateItem conditionexpression

//UpdateOrderStatus  updates the status attr for the order meta item
func (db *DynamoDB) UpdateOrderStatus(input *UpdateOrderStatusInput) (UpdateOrderStatusOutput, error) {
	if err := input.validate(); err != nil {
		return UpdateOrderStatusOutput{}, err
	}
	modifiedDate := time.Now().Format(time.RFC3339)
	currentStatus, err := getCurrStatus(input.NewStatus)
	if err != nil {
		return UpdateOrderStatusOutput{}, err
	}
	pk := fmt.Sprintf("ORDER#%s", input.OrderID)
	sort := "METADATA#"

	attrVals := map[string]*dynamodb.AttributeValue{
		":m": {
			S: aws.String(modifiedDate),
		},
		":curr": {
			S: aws.String(currentStatus),
		},
		":new": {
			S: aws.String(input.NewStatus),
		},
	}
	exprAttrNames := map[string]*string{"#ModifiedDate": aws.String("ModifiedDate"), "#Status": aws.String("Status")}
	condExprStr := "#Status = :curr"
	updateExprStr := "SET #ModifiedDate = :m, #Status = :new"

	if input.NewStatus == StatusDeliveredOrder && !input.ActDeliverDate.IsZero() {
		exprAttrNames["#Deliver"] = aws.String("ActDeliverDate")
		updateExprStr = updateExprStr + ", #Deliver = :ddate"
		attrVals[":ddate"] = &dynamodb.AttributeValue{S: aws.String(input.ActDeliverDate.Format(time.RFC3339))}
	}
	res, err := db.db.UpdateItem(&dynamodb.UpdateItemInput{
		TableName: aws.String(db.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"PK": {
				S: aws.String(pk),
			},
			"SK": {
				S: aws.String(sort),
			},
		},
		ReturnValues:              aws.String(dynamodb.ReturnValueUpdatedNew),
		UpdateExpression:          aws.String(updateExprStr),
		ConditionExpression:       aws.String(condExprStr),
		ExpressionAttributeNames:  exprAttrNames,
		ExpressionAttributeValues: attrVals,
	})

	if err != nil {
		return UpdateOrderStatusOutput{}, err
	}
	var result UpdateOrderStatusOutput
	err = dynamodbattribute.UnmarshalMap(res.Attributes, &result)

	return result, err

}

func marshallOrderMetaItem(order Order) (map[string]*dynamodb.AttributeValue, error) {
	pk := fmt.Sprintf("ORDER#%s", order.OrderID)
	sort := "METADATA#"
	gsipk := fmt.Sprintf("USER#%s", order.UserID)
	gsisk := fmt.Sprintf("ORDER#%s", order.OrderID)

	orderItem, err := dynamodbattribute.MarshalMap(&order)
	if err != nil {
		return orderItem, err
	}

	orderItem["PK"] = &dynamodb.AttributeValue{S: aws.String(pk)}
	orderItem["SK"] = &dynamodb.AttributeValue{S: aws.String(sort)}
	orderItem["Type"] = &dynamodb.AttributeValue{S: aws.String("UserOrder")}
	orderItem["GSI2PK"] = &dynamodb.AttributeValue{S: aws.String(gsipk)}
	orderItem["GSI2SK"] = &dynamodb.AttributeValue{S: aws.String(gsisk)}

	return orderItem, nil

}

func marshallOrderLineItem(item OrderLineItem) (map[string]*dynamodb.AttributeValue, error) {
	item.ItemID = NewSortableID()

	pk := fmt.Sprintf("ORDER#%s", item.OrderID)
	sort := fmt.Sprintf("ORDERITEM#%s", item.ItemID)

	i, err := dynamodbattribute.MarshalMap(&item)
	if err != nil {
		return i, err
	}
	i["Type"] = &dynamodb.AttributeValue{S: aws.String("OrderLineItem")}
	i["PK"] = &dynamodb.AttributeValue{S: aws.String(pk)}
	i["SK"] = &dynamodb.AttributeValue{S: aws.String(sort)}

	return i, nil

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

//for our conditionexpr when using the UpdateItem API, only update if ddb item Status attr currently equals this
func getCurrStatus(newStatus string) (string, error) {
	switch newStatus {
	case StatusShippedOrder:
		return StatusNewOrder, nil
	case StatusDeliveredOrder:
		return StatusShippedOrder, nil
	default:
		return "", errors.New("Order status is delivered. No need to update ddb item")
	}
}
