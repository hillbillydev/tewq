package dynamodb

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/matryer/is"
)

func TestAddProduct(t *testing.T) {
	is := is.New(t)
	product := Product{
		Name:        "Golf Club",
		Description: "This is a product",
		Price:       1000,
		Weight:      1500,
		Image:       "s3://images/image.png",
		Thumbnail:   "s3://images/thumbnail.png",
	}
	tdb, err := NewTestDynamoDB()
	is.NoErr(err)
	defer tdb.Close()

	_, err = tdb.AddProduct(product)
	is.NoErr(err)
}

func TestAddOptionToProduct(t *testing.T) {
	is := is.New(t)
	product := Product{
		Name:        "Golf Club",
		Description: "This is a product",
		Price:       1000,
		Weight:      1500,
		Image:       "s3://images/image.png",
		Thumbnail:   "s3://images/thumbnail.png",
	}
	option := Option{
		Color:          "red",
		Stock:          1,
		Size:           "Medium",
		ShaftStiffness: 11.5,
		Socket:         "Right",
	}

	tdb, err := NewTestDynamoDB()
	is.NoErr(err)
	defer tdb.Close()

	id, err := tdb.AddProduct(product)
	is.NoErr(err)

	_, err = tdb.AddOptionToProduct(id, option)
	is.NoErr(err)
}

func TestGetProduct(t *testing.T) {
	is := is.New(t)
	product := Product{
		Name:        "Golf Club",
		Description: "This is a product",
		Price:       1000,
		Weight:      1500,
		Image:       "s3://images/image.png",
		Thumbnail:   "s3://images/thumbnail.png",
	}
	options := []Option{
		{
			Color:          "red",
			Stock:          1,
			Size:           "Medium",
			ShaftStiffness: 11.5,
			Socket:         "Right",
		},
		{
			Color:          "red",
			Stock:          1,
			Size:           "Medium",
			ShaftStiffness: 11.5,
			Socket:         "Right",
		},
	}

	tdb, err := NewTestDynamoDB()
	is.NoErr(err)
	//defer tdb.Close()

	// Prepare data to get fetched
	id, err := tdb.AddProduct(product)
	is.NoErr(err)
	for _, op := range options {
		_, err = tdb.AddOptionToProduct(id, op)
		is.NoErr(err)
	}

	p, err := tdb.GetProduct(id)
	is.NoErr(err)

	is.True(len(p.Options) == 2) // We provided 2 options, so why is it not there?
	fmt.Println(p)
	is.Equal(product.Name, p.Name)
}

type TestDynamoDB struct {
	*DynamoDB
}

// NewTestDynamoDB connects to http://localhost:8000, and this should not change.
// It then create a temporary test tables, with name looking like this Tewq-Test_2020-09-04_09-21-37.
// To clean up your test resources you will have to call the Close() method.
func NewTestDynamoDB() (*TestDynamoDB, error) {
	tableName := fmt.Sprintf("%s_%s", "Tewq-Test", time.Now().Format("2006-01-02_15-04-05"))

	db, err := New("http://localhost:8000", tableName)
	if err != nil {
		return nil, err
	}
	tdb := &TestDynamoDB{db}

	return tdb, tdb.createTestTable()
}

// Close closes the test resources.
func (t *TestDynamoDB) Close() error {
	if !strings.Contains(t.db.Endpoint, "localhost") {
		return fmt.Errorf("Tried to run against %s, but can only run against an local instance", t.db.Endpoint)
	}

	_, err := t.db.DeleteTable(&dynamodb.DeleteTableInput{
		TableName: aws.String(t.tableName),
	})

	return err
}

func (t *TestDynamoDB) createTestTable() error {
	if !strings.Contains(t.db.Endpoint, "localhost") {
		return fmt.Errorf("Tried to run against %s, but can only run against an local instance", t.db.Endpoint)
	}

	_, err := t.db.CreateTable(&dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("PK"),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("SK"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("PK"),
				KeyType:       aws.String("HASH"),
			},
			{
				AttributeName: aws.String("SK"),
				KeyType:       aws.String("RANGE"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(10),
			WriteCapacityUnits: aws.Int64(10),
		},
		TableName: aws.String(t.tableName),
	})

	return err
}
