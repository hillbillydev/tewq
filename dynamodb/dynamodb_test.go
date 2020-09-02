package dynamodb_test

import (
	"testing"

	"github.com/Tinee/tewq/dynamodb"
	"github.com/google/uuid"
	"github.com/matryer/is"
)

type TestDynamoDB struct {
	*dynamodb.DynamoDB
}

func New() (*TestDynamoDB, error) {
	db, err := dynamodb.New()
	if err != nil {
		return nil, err
	}

	return &TestDynamoDB{db}, nil
}

func TestSomething(t *testing.T) {
	is := is.New(t)

	tdb, err := New()
	is.NoErr(err)

	err = tdb.AddProduct(dynamodb.Product{
		ID:   uuid.New(),
		Name: "Test",
	})
	is.NoErr(err)
}
