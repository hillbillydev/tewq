package dynamodb

import (
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/google/uuid"
)

type User struct {
	ID          string `json:"id"`
	CreatedDate string `json:"createdDate"`
	Email       string `json:"email"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
}

type Status int

const (
	OrderNew Status = iota + 1
	OrderShipped
	OrderDelivered
)

type Order struct {
	ID              string `json:"id"`
	PurchaseDate    string `json:"purchaseDate"`
	ShippingAddress string `json:"shippingAddress"`
	Status          Status `json:"status"`
	TotalAmount     int    `json:"totalAmount"`
	DeliverDate     string `json:"deliverDate"`
}

func (db *DynamoDB) AddUser(u User) (User, error) {

	u.ID = uuid.New().String()
	u.CreatedDate = time.Now().Format(time.RFC3339)
	pk := fmt.Sprintf("USER#%s", u.ID)
	sort := "METADATA#"
	item, err := dynamodbattribute.MarshalMap(&u)
	if err != nil {
		return User{}, err
	}

	item["type"] = &dynamodb.AttributeValue{S: aws.String("person")}
	item["PK"] = &dynamodb.AttributeValue{S: aws.String(pk)}
	item["SK"] = &dynamodb.AttributeValue{S: aws.String(sort)}

	//TODO:
	//if we update the email field; we'd need to update the GSIPK=EMAIL<email>
	// gs1pk := fmt.Sprintf("EMAIL#%s", u.Email)
	// gs1sk := sort
	// item["GSI1PK"] = &dynamodb.AttributeValue{S: aws.String(gs1pk)}
	// item["GSI1SK"] = &dynamodb.AttributeValue{S: aws.String(gs1sk)}

	//so we don't overwrite an existing user metadata item
	checkCond := "attribute_not_exists(PK) AND attribute_not_exists(SK)"
	_, err = db.db.PutItem(&dynamodb.PutItemInput{
		TableName:           aws.String(db.tableName),
		Item:                item,
		ConditionExpression: aws.String(checkCond)})

	return u, err

}

func (db *DynamoDB) GetUser(id string) (User, error) {
	//TODO: refactor with getItem request instead for speed
	// temp comment
	var result User
	pk := fmt.Sprintf("USER#%s", id)
	sort := "METADATA#"

	res, err := db.db.Query(&dynamodb.QueryInput{
		TableName:              aws.String(db.tableName),
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

func (db *DynamoDB) AddNewOrderToUser(uid string, order Order) (Order, error) {
	/*
		TODO: gotta think about below some more
		PK=USER#<id> ; SK=ORDER<id>
		GSI1PK=ORDER<id> ; GSISK=STATUS#<status>
		GS12PK=EMAIL#<id> ; GSI2SK=ORDER<id>

	*/
	order.ID = uuid.New().String()
	order.PurchaseDate = time.Now().Format(time.RFC3339)

	pk := fmt.Sprintf("USER#%s", uid)
	sort := fmt.Sprintf("ORDER#%s", order.ID)

	item, err := dynamodbattribute.MarshalMap(&order)
	if err != nil {
		return Order{}, err
	}
	item["Type"] = &dynamodb.AttributeValue{S: aws.String("user_order")}
	item["PK"] = &dynamodb.AttributeValue{S: aws.String(pk)}
	item["SK"] = &dynamodb.AttributeValue{S: aws.String(sort)}
	_, err = db.db.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(db.tableName),
		Item:      item,
	})

	return order, err

}
