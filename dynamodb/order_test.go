package dynamodb

import (
	"log"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/matryer/is"
	"github.com/matryer/try"
)

func TestAddNewOrderAndOrderLineItemsAndGetOrdersByUserIDAndUpdateOrder(t *testing.T) {
	is := is.New(t)
	products := getMockProducts()

	tdb, err := NewTestDynamoDB()
	is.NoErr(err)
	//defer tdb.Close()
	//todo: retrieve product item details from /GET request after writing products
	var oLineItems []OrderLineItem
	var subTotal int
	for _, prod := range products {
		p, err := tdb.AddProduct(prod)
		is.NoErr(err)
		for _, opt := range prod.Options {
			o, err := tdb.AddOptionToProduct(p.ID, opt)
			is.NoErr(err)
			total := o.Stock * p.Price
			subTotal = subTotal + total
			item := OrderLineItem{ProductID: p.ID,
				ProductOptionID: o.ID, Quantity: o.Stock,
				UnitPrice: p.Price, TotalAmount: total,
			}
			oLineItems = append(oLineItems, item)

		}

	}

	err = tdb.CreateUserOrderGSI()
	is.NoErr(err)

	indexName := "GSI2KO"
	err = try.Do(func(attempt int) (bool, error) {
		var err error
		_, err = tdb.IsGSIReady(indexName)
		if err != nil {
			time.Sleep(2 * time.Second)
		}
		return attempt < 5, err
	})
	if err != nil {
		log.Fatalln("error:", err)
	}

	userID := NewSortableID()
	shipAddress := Address{
		AddressType:   "Home",
		StreetAddress: "123 Main St",
		ZipCode:       "12345",
		State:         "CA",
		Country:       "US",
	}
	order := Order{
		UserID:          userID,
		ShippingAddress: shipAddress,
		TotalAmount:     subTotal,
		EstDeliverDate:  time.Now().AddDate(0, 0, 7),
		OrderLineItems:  oLineItems,
	}
	// t.Logf("%+v", order)
	err = tdb.AddOrder(order)
	is.NoErr(err)
	arrOfOrders, err := tdb.GetFullOrderItems(userID, 1)
	is.NoErr(err)
	oid := arrOfOrders[0].OrderID
	newStatus := StatusShippedOrder

	output, err := tdb.UpdateOrderStatus(&UpdateOrderStatusInput{OrderID: oid, NewStatus: newStatus})
	is.NoErr(err)
	is.Equal(output.Status, StatusShippedOrder)

	output, err = tdb.UpdateOrderStatus(&UpdateOrderStatusInput{OrderID: oid, NewStatus: StatusDeliveredOrder, ActDeliverDate: time.Now().AddDate(0, 0, 8)})
	is.NoErr(err)
	t.Log(output)
	is.Equal(output.Status, StatusDeliveredOrder)

}

func TestAddNewOrderAndOrderLineItemsAndisTransError(t *testing.T) {
	is := is.New(t)
	products := []Product{
		{
			Name:     "Super Duper",
			Category: "Clubs",
			Price:    1000,
			Options: []Option{
				{
					Color: "Green",
					Stock: 3,
				},
			},
		},
		{
			Name:     "Adidas",
			Category: "Shoes",
			Price:    1500,
			Options: []Option{
				{
					Color: "Red",
					Stock: 3,
				},
			},
		},
	}

	tdb, err := NewTestDynamoDB()
	is.NoErr(err)
	//defer tdb.Close()
	var oLineItems []OrderLineItem
	var subTotal int
	for _, p := range products {
		p, err := tdb.AddProduct(p)
		is.NoErr(err)
		o, err := tdb.AddOptionToProduct(p.ID, p.Options[0])
		is.NoErr(err)
		total := o.Stock * p.Price
		subTotal = subTotal + total
		item := OrderLineItem{ProductID: p.ID,
			ProductOptionID: o.ID, Quantity: o.Stock + 1,
			UnitPrice: p.Price, TotalAmount: total,
		}
		oLineItems = append(oLineItems, item)
	}

	err = tdb.CreateUserOrderGSI()
	is.NoErr(err)

	indexName := "GSI2KO"
	err = try.Do(func(attempt int) (bool, error) {
		var err error
		_, err = tdb.IsGSIReady(indexName)
		if err != nil {
			time.Sleep(2 * time.Second)
		}
		return attempt < 5, err
	})
	if err != nil {
		log.Fatalln("error:", err)
	}

	userID := NewSortableID()
	shipAddress := Address{
		AddressType:   "Home",
		StreetAddress: "123 Main St",
		ZipCode:       "12345",
		State:         "CA",
		Country:       "US",
	}
	order := Order{
		UserID:          userID,
		ShippingAddress: shipAddress,
		TotalAmount:     subTotal,
		EstDeliverDate:  time.Now().AddDate(0, 0, 7),
		OrderLineItems:  oLineItems,
	}
	/*
		in this case we expect a TransactCanceledException b/c the condition expression is not met;
		we're attempting to add userorderitems which would lead to a negative stock value. so an err is expected
	*/
	//https://docs.aws.amazon.com/sdk-for-go/api/service/dynamodb/#example_DynamoDB_TransactWriteItems_transactionCanceledException
	t.Logf("%+v", order)
	err = tdb.AddOrder(order)
	if err == nil {
		t.Errorf("EXPECTED ERROR %v", err)

	} else {
		switch tErr := err.(type) {
		case *dynamodb.TransactionCanceledException:
			t.Log("Request failed", err.Error())
			for _, tErrItem := range tErr.CancellationReasons {
				is.Equal(*tErrItem.Code, "ConditionalCheckFailed")
			}

		default:
			t.Error("We were expecting a TransactionCanceledException but got:", err)
		}

	}

}

func getMockProducts() []Product {
	return []Product{
		{
			Name:     "Super Duper",
			Category: "Clubs",
			Price:    1000,
			Options: []Option{
				{
					Color: "Green",
					Stock: 5,
				},
				{
					Color: "Teal",
					Stock: 13,
				},
				{
					Color: "Diamond",
					Stock: 10,
				},
				{
					Color: "Brown",
					Stock: 12,
				},
				{
					Color: "Yellow",
					Stock: 12,
				},
				{
					Color: "Orange",
					Stock: 12,
				},
				{
					Color: "Red",
					Stock: 12,
				},
				{
					Color: "Dark Red",
					Stock: 12,
				},
			},
		},
		{
			Name:     "Adidas",
			Category: "Shoes",
			Price:    1500,
			Options: []Option{
				{
					Color: "Red",
					Stock: 3,
				},
				{
					Color: "Blue",
					Stock: 10,
				},
				{
					Color: "Orange",
					Stock: 13,
				},
				{
					Color: "Purple",
					Stock: 31,
				},
				{
					Color: "Violet",
					Stock: 3,
				},
				{
					Color: "Teal",
					Stock: 3,
				},
			},
		},
		{
			Name:     "Ripper",
			Category: "Gloves",
			Price:    60,
			Options: []Option{
				{
					Color: "Red",
					Stock: 3,
				},
				{
					Color: "Blue",
					Stock: 10,
				},
				{
					Color: "Orange",
					Stock: 13,
				},
				{
					Color: "Purple",
					Stock: 31,
				},
			},
		},
		{
			Name:     "Eyecandy",
			Category: "Sunglasses",
			Price:    120,
			Options: []Option{
				{
					Color: "Red",
					Stock: 3,
				},
				{
					Color: "Blue",
					Stock: 10,
				},
				{
					Color: "Orange",
					Stock: 13,
				},
				{
					Color: "Purple",
					Stock: 31,
				},
				{
					Color: "Violet",
					Stock: 3,
				},
				{
					Color: "Teal",
					Stock: 3,
				},
			},
		},
		{
			Name:        "Golf Club",
			Description: "This is a product",
			Category:    "Club",
			Price:       1000,
			Weight:      1500,
			Image:       "s3://images/image.png",
			Thumbnail:   "s3://images/thumbnail.png",
			Options: []Option{
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
					Size:           "Small",
					ShaftStiffness: 8.5,
					Socket:         "Right",
				},
				{
					Color:          "red",
					Stock:          1,
					Size:           "Large",
					ShaftStiffness: 13.5,
					Socket:         "Right",
				},
				{
					Color:          "red",
					Stock:          1,
					Size:           "Large",
					ShaftStiffness: 11.5,
					Socket:         "Left",
				},
			},
		},
	}

}
