package dynamodb

import (
	"testing"

	"github.com/matryer/is"
)

func TestGetUser(t *testing.T) {
	is := is.New(t)
	user := User{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "johnDoe@gmail.com",
	}
	tdb, err := NewTestDynamoDB()
	is.NoErr(err)
	// defer tdb.Close()

	u, err := tdb.AddUser(user)
	t.Log(u)
	is.NoErr(err)

	fetched, err := tdb.GetUser(u.ID)

	is.NoErr(err)
	t.Log(fetched)

	is.Equal(u.FirstName, fetched.FirstName)
	is.Equal(u.LastName, fetched.LastName)
	is.Equal(u.Email, fetched.Email)

}

func TestAddNewOrdersToUserAndGetOrdersByID(t *testing.T) {
	is := is.New(t)
	user := User{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "johnDoe@gmail.com",
	}

	tdb, err := NewTestDynamoDB()
	is.NoErr(err)
	// defer tdb.Close()

	u, err := tdb.AddUser(user)
	t.Log(u)
	is.NoErr(err)

	orders := []Order{
		{
			UserID:          u.ID,
			ShippingAddress: "123 Main Street NY, NY 12345",
			TotalAmount:     5000,
		},
		{
			UserID:          u.ID,
			ShippingAddress: "123 Main Street NY, NY 12345",
			TotalAmount:     6700,
		},
	}
	orderIDs := []SortableID{}
	for _, op := range orders {
		order, err := tdb.AddNewOrderToUser(u.ID, op)
		is.NoErr(err)
		orderIDs = append(orderIDs, order.OrderID)
	}
	for _, oid := range orderIDs {
		fetchedOrder, err := tdb.GetUserOrderByOrderID(oid)
		is.NoErr(err)
		t.Log(fetchedOrder)
		// is.Equal(fetchedOrder.OrderID, oid)

	}
	// fetched, err := tdb.GetU

}

func TestUpdateUserOrderStatus(t *testing.T) {

}

// func TestAddUserAlreadyExists(t *testing.T) {

// }

// func TestAddNewOrderItemToUser(t *testing.T) {

// }
