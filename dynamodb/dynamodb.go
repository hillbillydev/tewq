package dynamodb

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/google/uuid"
)

type Option struct {
	ID             uuid.UUID `json:"id"`
	CreatedDate    time.Time `json:"createdDate"`
	Stock          int       `json:"stock" dynamodbav:",omitempty"`
	ShaftStiffness float64   `json:"shaftStiffness" dynamodbav:",omitempty"`
	Size           string    `json:"size" dynamodbav:",omitempty"`   // TODO enum?
	Socket         string    `json:"socket" dynamodbav:",omitempty"` // TODO enum?
	Color          string    `json:"socket" dynamodbav:",omitempty"` // TODO enum?
}

type Product struct {
	ID          uuid.UUID `json:"id"`
	CreatedDate time.Time `json:"createdDate"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       int       `json:"price"`
	Weight      int       `json:"weight"`
	Image       string    `json:"image"`
	Thumbnail   string    `json:"thumbNail"`
	Options     []Option  `json:"options" dynamodbav:",omitempty"`
}

type Basket struct {
	ID string `json:"id"`
}

type DynamoDB struct {
	db        *dynamodb.DynamoDB
	tableName string
	endpoint  string
}

func New(endpoint, tableName string) (*DynamoDB, error) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := dynamodb.New(sess, &aws.Config{
		Endpoint: aws.String(endpoint),
	})

	return &DynamoDB{
		db:        svc,
		tableName: tableName,
	}, nil
}

// AddProduct take a Product p and attempts to put that item into DynamoDB.
// If the caller provides an ID on the product we will fail straight away.
func (db *DynamoDB) AddProduct(p Product) (uuid.UUID, error) {
	if p.ID != uuid.Nil {
		return uuid.Nil, fmt.Errorf("When adding an product we did not expect the ID to have a value but it got %q", p.ID)
	}
	if !p.CreatedDate.IsZero() {
		return uuid.Nil, fmt.Errorf("When adding an product we did not expect the CreatedDate to already be set but it was set to %q", p.CreatedDate)
	}

	p.CreatedDate = time.Now()
	p.ID = uuid.New()

	pk := fmt.Sprintf("PRODUCT#%s", p.ID)
	sort := "METADATA#"

	item, err := dynamodbattribute.MarshalMap(&p)
	if err != nil {
		return uuid.Nil, err
	}
	item["type"] = &dynamodb.AttributeValue{S: aws.String("product")}
	item["PK"] = &dynamodb.AttributeValue{S: aws.String(pk)}
	item["SK"] = &dynamodb.AttributeValue{S: aws.String(sort)}
	item["id"] = &dynamodb.AttributeValue{S: aws.String(p.ID.String())}
	item["createdUtc"] = &dynamodb.AttributeValue{S: aws.String(p.CreatedDate.Format(time.RFC3339))}

	_, err = db.db.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(db.tableName),
		Item:      item,
	})

	return p.ID, err
}

// AddOptionToProduct adds a single option to a product.
// It returns an error if you try to provide either an ID or CreatedDate.
func (db *DynamoDB) AddOptionToProduct(id uuid.UUID, option Option) (uuid.UUID, error) {
	if option.ID != uuid.Nil {
		return uuid.Nil, fmt.Errorf("When adding an product we did not expect the ID to have a value but it got %q", id)
	}
	if !option.CreatedDate.IsZero() {
		return uuid.Nil, fmt.Errorf("When adding an product we did not expect the CreatedDate to already be set but it was set to %q", option.CreatedDate)
	}
	option.ID = uuid.New()
	option.CreatedDate = time.Now()

	pk := fmt.Sprintf("PRODUCT#%s", id)
	sort := fmt.Sprintf("OPTION#%s", option.CreatedDate)

	item, err := dynamodbattribute.MarshalMap(&option)
	if err != nil {
		return uuid.Nil, err
	}
	item["type"] = &dynamodb.AttributeValue{S: aws.String("product_option")}
	item["PK"] = &dynamodb.AttributeValue{S: aws.String(pk)}
	item["SK"] = &dynamodb.AttributeValue{S: aws.String(sort)}
	item["id"] = &dynamodb.AttributeValue{S: aws.String(option.ID.String())}
	item["createdUtc"] = &dynamodb.AttributeValue{S: aws.String(option.CreatedDate.Format(time.RFC3339))}

	_, err = db.db.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(db.tableName),
		Item:      item,
	})

	return option.ID, err
}
