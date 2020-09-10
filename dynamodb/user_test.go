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
		UserName:  "jdoe123",
		Email:     "johnDoe@gmail.com",
	}
	tdb, err := NewTestDynamoDB()
	is.NoErr(err)
	defer tdb.Close()

	u, err := tdb.AddUser(user)
	t.Log(u)
	is.NoErr(err)

	fetched, err := tdb.GetUser(u.ID)

	is.NoErr(err)
	t.Log(fetched)

	is.Equal(u.FirstName, fetched.FirstName)
	is.Equal(u.LastName, fetched.LastName)
	is.Equal(u.UserName, fetched.UserName)
	is.Equal(u.Email, fetched.Email)

}

func TestGetUserByEmail(t *testing.T) {
	is := is.New(t)
	user := User{
		FirstName: "John",
		LastName:  "Smith",
		UserName:  "jd_gunner",
		Email:     "jd.smith@gmail.com",
	}
	tdb, err := NewTestDynamoDB()
	is.NoErr(err)
	defer tdb.Close()

	u, err := tdb.AddUser(user)
	t.Log(u)
	is.NoErr(err)
	fetched, err := tdb.GetUserByEmail(u.Email)
	is.NoErr(err)
	t.Log(fetched)
	is.Equal(u.ID, fetched.ID)
	is.Equal(u.UserName, fetched.UserName)
	is.Equal(u.FirstName, fetched.FirstName)
	is.Equal(u.LastName, fetched.LastName)
	is.Equal(u.Email, fetched.Email)

}

func TestAddNewOrdersAndGetOrdersByOrderID(t *testing.T) {
	is := is.New(t)
	user := User{
		FirstName: "John",
		LastName:  "Doe",
		UserName:  "jdoeman",
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
			UserID: u.ID,
			ShippingAddress: Address{
				AddressType:   "Home",
				StreetAddress: "123 Main St",
				ZipCode:       "12345",
				State:         "CA",
				Country:       "US",
			},
			TotalAmount: 5000,
		},
		{
			UserID: u.ID,
			ShippingAddress: Address{
				AddressType:   "Work",
				StreetAddress: "123 Wall St",
				ZipCode:       "543322",
				State:         "NY",
				Country:       "USA",
			},
			TotalAmount: 6700,
		},
		{
			UserID: u.ID,
			ShippingAddress: Address{
				AddressType:   "Home",
				StreetAddress: "Roslagsgatan 10",
				ZipCode:       "111 28",
				State:         "STHLM",
				Country:       "SE",
			},
			TotalAmount: 6700,
		},
	}
	orderIDs := []SortableID{}
	for _, o := range orders {
		order, err := tdb.AddOrder(o)
		is.NoErr(err)
		orderIDs = append(orderIDs, order.OrderID)
	}
	for _, oid := range orderIDs {
		fetchedOrder, err := tdb.GetOrderByOrderID(oid)
		is.NoErr(err)
		t.Logf(" %+v", fetchedOrder)
		is.Equal(fetchedOrder.OrderID, oid)
		// is.Equal(fetchedOrder.)

	}

}

func TestUpdateStatusAttrForOrders(t *testing.T) {
	is := is.New(t)
	user := User{
		FirstName: "John",
		LastName:  "Doe",
		UserName:  "jdoe123",
		Email:     "johnDoe@gmail.com",
	}

	tdb, err := NewTestDynamoDB()
	is.NoErr(err)
	defer tdb.Close()

	u, err := tdb.AddUser(user)
	t.Log(u)
	is.NoErr(err)
	fetchedUser, err := tdb.GetUser(u.ID)

	is.NoErr(err)
	t.Log(fetchedUser)

	is.Equal(u.FirstName, fetchedUser.FirstName)
	is.Equal(u.LastName, fetchedUser.LastName)
	is.Equal(u.UserName, fetchedUser.UserName)
	is.Equal(u.Email, fetchedUser.Email)

	orders := []Order{
		{
			UserID: u.ID,
			ShippingAddress: Address{
				AddressType:   "Home",
				StreetAddress: "123 Main St",
				ZipCode:       "12345",
				State:         "CA",
				Country:       "US",
			},
			TotalAmount: 5000,
		},
		{
			UserID: u.ID,
			ShippingAddress: Address{
				AddressType:   "Work",
				StreetAddress: "123 Wall St",
				ZipCode:       "12345",
				State:         "NY",
				Country:       "USA",
			},
			TotalAmount: 6700,
		},
	}
	statuses := []string{StatusShippedOrder, StatusShippedOrder}
	orderIDs := []SortableID{}
	for _, o := range orders {
		order, err := tdb.AddOrder(o)
		is.NoErr(err)
		orderIDs = append(orderIDs, order.OrderID)
	}
	for i, oid := range orderIDs {
		err = tdb.UpdateOrderStatus(fetchedUser.ID, oid, statuses[i])
		is.NoErr(err)
	}

}

// func TestAddUserAlreadyExists(t *testing.T) {

// }

// func TestAddOrderLineItemsToUser(t *testing.T) {

// }

func TestAddOrderLineItemsAndGetOrderLineItems(t *testing.T) {
	is := is.New(t)
	user := User{
		FirstName: "John",
		LastName:  "Doe",
		UserName:  "jdoe123",
		Email:     "johnDoe@gmail.com",
	}

	tdb, err := NewTestDynamoDB()
	is.NoErr(err)
	// defer tdb.Close()

	u, err := tdb.AddUser(user)
	t.Log(u)
	is.NoErr(err)
	fetchedUser, err := tdb.GetUser(u.ID)

	is.NoErr(err)
	t.Log(fetchedUser)

	is.Equal(u.FirstName, fetchedUser.FirstName)
	is.Equal(u.LastName, fetchedUser.LastName)
	is.Equal(u.UserName, fetchedUser.UserName)
	is.Equal(u.Email, fetchedUser.Email)
	o := Order{
		UserID: fetchedUser.ID,
		ShippingAddress: Address{
			AddressType:   "Home",
			StreetAddress: "123 Main St",
			ZipCode:       "12345",
			State:         "CA",
			Country:       "US",
		},
		TotalAmount: 5000,
	}

	oAddedToDB, err := tdb.AddOrder(o)
	t.Log(oAddedToDB)
	is.NoErr(err)

	oLineItems := []OrderLineItem{
		{
			OrderID: oAddedToDB.OrderID,
			//todo: any way to link productID w/o a getproduct req?
			ProductID:   NewSortableID(),
			Price:       500,
			Quantity:    5,
			TotalAmount: 2500,
		},
		{
			OrderID: oAddedToDB.OrderID,
			//todo: any way to link productID w/o a getproduct req?
			ProductID:   NewSortableID(),
			Price:       500,
			Quantity:    5,
			TotalAmount: 2500,
		},
	}

	for _, i := range oLineItems {
		_, err := tdb.AddOrderLineItem(i)
		is.NoErr(err)
	}
	fetchedItems, err := tdb.GetOrderLineItemsByOrderID(oAddedToDB.OrderID)
	is.NoErr(err)
	is.Equal(len(fetchedItems), len(oLineItems))
	t.Log(fetchedItems)
	t.Log(len(fetchedItems))

}
