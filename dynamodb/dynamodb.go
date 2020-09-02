package dynamodb

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/google/uuid"
)

type Product struct {
	ID   uuid.UUID
	Name string
}

type DynamoDB struct {
	db *dynamodb.DynamoDB
}

func New() (*DynamoDB, error) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	return &DynamoDB{
		db: dynamodb.New(sess, &aws.Config{
			Endpoint: aws.String("http://localhost:8000"),
		}),
	}, nil
}

func (db *DynamoDB) AddProduct(p Product) error {
	_, err := db.db.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("Tewq"),
		Item: map[string]*dynamodb.AttributeValue{
			"PK": {
				S: aws.String("Test"),
			},
			"SK": {
				S: aws.String("Test"),
			},
		},
	})

	return err
}
