
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

func TestAddNewOrderToUser(t *testing.T) {
	is := is.New(t)
	user := User{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "johnDoe@gmail.com",
	}
	order := Order{
		ShippingAddress: "123 Main Street NY, NY 12345",
		Status:          OrderNew,
		TotalAmount:     5000,
	}
	tdb, err := NewTestDynamoDB()
	is.NoErr(err)
	// defer tdb.Close()

	u, err := tdb.AddUser(user)
	t.Log(u)
	is.NoErr(err)
	_, err = tdb.AddNewOrderToUser(u.ID, order)
	is.NoErr(err)

}

// func TestAddUserAlreadyExists(t *testing.T) {

// }
