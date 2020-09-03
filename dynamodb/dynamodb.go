package dynamodb

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/google/uuid"
)

type Product struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type Basket struct {
	ID uuid.UUID `json:"id"`
}

type DynamoDB struct {
	db        *dynamodb.DynamoDB
	tableName string
}

func New() (*DynamoDB, error) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := dynamodb.New(sess, &aws.Config{
		Endpoint: aws.String("http://localhost:8000"),
	})

	return &DynamoDB{
		db:        svc,
		tableName: "Tewq", // TODO pass tablename as argument.
	}, nil
}

func (db *DynamoDB) AddProduct(p Product) error {
	pk := p.ID.String()
	sort := p.Name

	item, err := dynamodbattribute.MarshalMap(&p)
	if err != nil {
		return err
	}
	item["type"] = &dynamodb.AttributeValue{S: aws.String("Product")}
	item["PK"] = &dynamodb.AttributeValue{S: aws.String(pk)}
	item["SK"] = &dynamodb.AttributeValue{S: aws.String(sort)}
	item["id"] = &dynamodb.AttributeValue{S: aws.String(pk)}

	_, err = db.db.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(db.tableName),
		Item:      item,
	})

	return err
}

func (db *DynamoDB) GetProduct(id uuid.UUID) (*Product, error) {
	res, err := db.db.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(db.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"PK": {
				S: aws.String(id.String()),
			},
			"SK": {
				S: aws.String("Add&Get"),
			},
		},
	})
	if err != nil {
		return nil, err
	}
	if res.Item == nil {
		return nil, nil
	}

	return nil, err
}
