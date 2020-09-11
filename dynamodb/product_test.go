package dynamodb

import (
	"fmt"
	"testing"

	"github.com/matryer/is"
)

func TestAddProduct(t *testing.T) {
	is := is.New(t)
	product := Product{
		Name:        "Golf Club",
		Description: "This is a product",
		Category:    "Club",
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
		Category:    "Club",
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

	p, err := tdb.AddProduct(product)
	is.NoErr(err)

	updatedProduct, err := tdb.AddOptionToProduct(p.ID, option)
	is.NoErr(err)

	is.True(len(updatedProduct.Options) == 1)
	is.Equal(updatedProduct.Options[0].Color, option.Color)
	is.Equal(updatedProduct.Options[0].Size, option.Size)
	is.Equal(updatedProduct.Options[0].ShaftStiffness, option.ShaftStiffness)
	is.Equal(updatedProduct.Options[0].Socket, option.Socket)
}

func TestGetProduct(t *testing.T) {
	is := is.New(t)
	product := Product{
		Name:        "Golf Club",
		Category:    "Club",
		Description: "This is a product",
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
				Color:          "green",
				Stock:          2,
				Size:           "Medium",
				ShaftStiffness: 11.5,
				Socket:         "Right",
			},
		},
	}

	tdb, err := NewTestDynamoDB()
	is.NoErr(err)
	defer tdb.Close()

	// Prepare data to get fetched
	p, err := tdb.AddProduct(product)
	is.NoErr(err)

	fetched, err := tdb.GetProduct(p.ID)
	is.NoErr(err)

	is.Equal(p.Name, fetched.Name)
	is.Equal(p.Description, fetched.Description)
	is.Equal(p.Weight, fetched.Weight)
	is.Equal(p.Price, fetched.Price)
	is.Equal(p.Image, fetched.Image)
	is.Equal(p.Category, fetched.Category)
	is.Equal(p.Thumbnail, fetched.Thumbnail)

	is.True(len(fetched.Options) == 2)
	is.Equal(fetched.Options[0].Color, product.Options[0].Color)
	is.Equal(fetched.Options[1].Color, product.Options[1].Color)
}

func TestGetProductsByCategory(t *testing.T) {
	is := is.New(t)
	categoryToFetch := "Clubs"
	products := []Product{
		{
			Name:     "Golf Club",
			Category: "Shoes",
			Price:    1000,
		},
		{
			Name:     "Golf Club",
			Category: categoryToFetch,
			Price:    1000,
		},
		{
			Name:     "Golf Club 2",
			Category: categoryToFetch,
			Price:    500,
		},
	}

	tdb, err := NewTestDynamoDB()
	is.NoErr(err)
	defer tdb.Close()

	// Prepare data to get fetched
	for _, p := range products {
		_, err := tdb.AddProduct(p)
		is.NoErr(err)
	}

	fetched, _, err := tdb.GetProductsByCategory(&GetProductsByCategoryInput{
		Category: categoryToFetch,
	})
	is.NoErr(err)

	is.True(len(fetched) == 2) // should be 2 products with category "Clubs"
}

func TestGetProductsByCategoryAndPrice(t *testing.T) {
	is := is.New(t)
	categoryToFetch := "Clubs"
	products := []Product{
		{
			Name:     "Golf Club",
			Category: "Shoes",
			Price:    1000,
		},
		{
			Name:     "Golf Club",
			Category: categoryToFetch,
			Price:    100,
		},
		{
			Name:     "Golf Club 2",
			Category: categoryToFetch,
			Price:    500,
		},
	}

	tdb, err := NewTestDynamoDB()
	is.NoErr(err)
	defer tdb.Close()

	// Prepare data to get fetched
	for _, p := range products {
		_, err := tdb.AddProduct(p)
		is.NoErr(err)
	}

	fetched, _, err := tdb.GetProductsByCategory(&GetProductsByCategoryInput{
		Category:  categoryToFetch,
		FromPrice: 500,
		ToPrice:   600,
	})
	is.NoErr(err)
	is.True(len(fetched) == 1) // should be 1 products with category "Clubs"
	is.Equal(fetched[0].Price, products[2].Price)
}

func TestGetProductsByCategoryPagination(t *testing.T) {
	is := is.New(t)
	categoryToFetch := "Clubs"

	tdb, err := NewTestDynamoDB()
	is.NoErr(err)
	defer tdb.Close()

	// Prepare data to get fetched
	for i := 9; i != 0; i-- {
		// Add 9 golf clubs to the database.
		_, err := tdb.AddProduct(Product{
			Name:     fmt.Sprintf("Test%d", i),
			Category: categoryToFetch,
		})
		is.NoErr(err)
	}

	fetched, last, err := tdb.GetProductsByCategory(&GetProductsByCategoryInput{
		Category:        categoryToFetch,
		PaginationLimit: 5,
	})
	is.NoErr(err)
	is.True(len(fetched) == 5)
	is.True(last != "")

	fetched, last, err = tdb.GetProductsByCategory(&GetProductsByCategoryInput{
		Category:    categoryToFetch,
		PreviousKey: last,
	})
	is.NoErr(err)
	is.True(len(fetched) == 4)
	is.True(last == "") // No more keys
}
