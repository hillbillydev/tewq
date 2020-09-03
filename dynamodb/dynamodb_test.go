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

    t.Run("Should Add Product.", func(t *testing.T) {
        err = tdb.AddProduct(dynamodb.Product{
            ID:   uuid.New(),
            Name: "Test",
        })
        is.NoErr(err)
    })

    //t.Run("Should Add and Get Product.", func(t *testing.T) {
    //    id := uuid.New()

    //    err = tdb.AddProduct(dynamodb.Product{
    //        ID:   id,
    //        Name: "Add&Get",
    //    })
    //    is.NoErr(err)
    //    p, err := tdb.GetProduct(id)
    //    is.NoErr(err)

    //    is.Equal(p.ID.String(), id.String())
    //})
}
