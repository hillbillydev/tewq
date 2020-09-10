package dynamodb

import (
	"testing"

	"github.com/matryer/is"
)

func TestAddBasketItem(t *testing.T) {
	customerID := NewSortableID()
	is := is.New(t)
	product := Product{
		Name:     "Golf Club",
		Category: "Shoes",
		Options: []Option{
			{
				Color: "Red",
				Stock: 2,
			},
		},
	}

	tdb, err := NewTestDynamoDB()
	is.NoErr(err)
	defer tdb.Close()

	// Prepare data to get fetched
	p, err := tdb.AddProduct(product)
	is.NoErr(err)
	o, err := tdb.AddOptionToProduct(p.ID, product.Options[0])
	is.NoErr(err)

	err = tdb.AddBasketItem(BasketItem{
		CustomerID:      customerID,
		ProductID:       p.ID,
		ProductOptionID: o.ID,
	})
	is.NoErr(err)
}

func TestGetBasketProducts(t *testing.T) {
	is := is.New(t)
	products := []Product{
		{
			Name:     "Super Duper",
			Category: "Clubs",
			Options: []Option{
				{
					Color: "Green",
					Stock: 2,
				},
			},
		},
		{
			Name:     "A Shoe",
			Category: "Shoes",
			Options: []Option{
				{
					Color: "Brown",
					Stock: 2,
				},
			},
		},
		{
			Name:     "Adidas",
			Category: "Shoes",
			Options: []Option{
				{
					Color: "Red",
					Stock: 3,
				},
			},
		},
	}

	customerID := NewSortableID()
	tdb, err := NewTestDynamoDB()
	is.NoErr(err)
	//defer tdb.Close()

	for _, p := range products {
		p, err := tdb.AddProduct(p)
		is.NoErr(err)

		o, err := tdb.AddOptionToProduct(p.ID, p.Options[0])
		is.NoErr(err)

		if p.Name == "A Shoe" {
			continue
		}

		err = tdb.AddBasketItem(BasketItem{
			CustomerID:      customerID,
			ProductID:       p.ID,
			ProductOptionID: o.ID,
		})
		is.NoErr(err)
	}

	p, err := tdb.GetBasketProducts(customerID)
	is.NoErr(err)

	is.True(len(p) == 2) // Only put 2 products in the basket..
	is.True(p[0].Name != "A Shoe")
	is.True(p[1].Name != "A Shoe")
}
