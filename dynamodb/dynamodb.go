package dynamodb

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/segmentio/ksuid"
)

type Option struct {
	ID             SortableID `json:"id" dynamodbav:"Id,omitempty"`
	CreatedDate    time.Time  `json:"createdUtc" dynamodbav:"CreatedUtc,omitempty"`
	Size           string     `json:"size" dynamodbav:"Size,omitempty"`     // TODO enum?
	Socket         string     `json:"socket" dynamodbav:"Socket,omitempty"` // TODO enum?
	Color          string     `json:"color" dynamodbav:"Color,omitempty"`   // TODO enum?
	Stock          int        `json:"stock" dynamodbav:"Stock,omitempty"`
	ShaftStiffness float64    `json:"shaftStiffness" dynamodbav:"ShaftStiffness,omitempty"`
}

type Product struct {
	ID          SortableID `json:"id" dynamodbav:"Id,omitempty"`
	CreatedDate time.Time  `json:"createdUtc" dynamodbav:"CreatedUtc,omitempty"`
	Category    string     `json:"category" dynamodbav:"Category,omitempty"`
	Name        string     `json:"name" dynamodbav:"Name,omitempty"`
	Description string     `json:"description" dynamodbav:"Description,omitempty"`
	Image       string     `json:"image" dynamodbav:"Image,omitempty"`
	Thumbnail   string     `json:"thumbNail" dynamodbav:"ThumbNail,omitempty"`
	Price       int        `json:"price" dynamodbav:"Price,omitempty"`
	Weight      int        `json:"weight" dynamodbav:"Weight,omitempty"`
	Sale        int        `json:"sale" dynamodbav:"Sale,omitempty"`
	Options     []Option   `json:"options" dynamodbav:"-"`
}

type Basket struct {
	Products []Product `json:"products"`
}

type BasketItem struct {
	CustomerID      SortableID `json:"customerId" dynamodbav:"CustomerId"`
	ProductID       SortableID `json:"productId" dynamodbav:"ProductId"`
	ProductOptionID SortableID `json:"productOptionId" dynamodbav:"ProductOptionId"`
}

type DynamoDB struct {
	db        *dynamodb.DynamoDB
	tableName string
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
func (db *DynamoDB) AddProduct(p Product) (Product, error) {

	p.CreatedDate = time.Now()
	p.ID = NewSortableID()

	pk := fmt.Sprintf("PRODUCT#%s", p.ID)
	sort := "METADATA#"
	gs1pk := fmt.Sprintf("PRODUCT#CATEGORY#%s", p.Category)
	gs1sk := zerosPricePadding(p.Price)

	item, err := dynamodbattribute.MarshalMap(&p)
	if err != nil {
		return Product{}, err
	}

	item["Type"] = &dynamodb.AttributeValue{S: aws.String("product")}
	item["PK"] = &dynamodb.AttributeValue{S: aws.String(pk)}
	item["SK"] = &dynamodb.AttributeValue{S: aws.String(sort)}
	item["GSI1PK"] = &dynamodb.AttributeValue{S: aws.String(gs1pk)}
	item["GSI1SK"] = &dynamodb.AttributeValue{S: aws.String(gs1sk)}

	_, err = db.db.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(db.tableName),
		Item:      item,
	})

	return p, err
}

// AddOptionToProduct adds a single option to a product.
func (db *DynamoDB) AddOptionToProduct(id SortableID, option Option) (Option, error) {
	option.ID = NewSortableID()
	option.CreatedDate = time.Now()

	pk := fmt.Sprintf("PRODUCT#%s", id)
	sort := fmt.Sprintf("OPTION#%s", option.ID)

	item, err := dynamodbattribute.MarshalMap(&option)
	if err != nil {
		return Option{}, err
	}
	item["Type"] = &dynamodb.AttributeValue{S: aws.String("product_option")}
	item["PK"] = &dynamodb.AttributeValue{S: aws.String(pk)}
	item["SK"] = &dynamodb.AttributeValue{S: aws.String(sort)}

	_, err = db.db.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(db.tableName),
		Item:      item,
	})

	return option, err
}

// GetProduct fetches the product will all their options included.
func (db *DynamoDB) GetProduct(id SortableID) (Product, error) {
	var result Product

	res, err := db.db.Query(&dynamodb.QueryInput{
		TableName:              aws.String(db.tableName),
		KeyConditionExpression: aws.String("#PK = :pk"),
		ExpressionAttributeNames: map[string]*string{
			"#PK": aws.String("PK"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":pk": {
				S: aws.String(fmt.Sprintf("PRODUCT#%s", id)),
			},
		},
		ScanIndexForward: aws.Bool(true),
	})
	if err != nil {
		return Product{}, err
	}
	if len(res.Items) == 0 {
		// TODO error not found here?
		return Product{}, nil
	}

	metadata, options := res.Items[0], res.Items[1:]

	err = dynamodbattribute.UnmarshalMap(metadata, &result)
	if err != nil {
		return Product{}, err
	}

	err = dynamodbattribute.UnmarshalListOfMaps(options, &result.Options)
	if err != nil {
		return Product{}, err
	}

	return result, err
}

type GetProductsByCategoryInput struct {
	Category        string // required
	FromPrice       int
	ToPrice         int
	PaginationLimit int
	PreviousKey     ProductCategoryPaginationKey
}

func (in *GetProductsByCategoryInput) validate() error {
	// TODO use enums here .
	if in.Category == "" {
		return errors.New("Expected Category to have a value.")
	}

	if in.ToPrice < in.FromPrice {
		return fmt.Errorf("PriceRange.To (%d) is smaller then PriceRange.From (%d).", in.ToPrice, in.FromPrice)
	}

	if in.ToPrice == 0 {
		in.ToPrice = math.MaxInt64
	}

	if in.PaginationLimit == 0 {
		in.PaginationLimit = 20
	}

	return nil
}

// GetProductsByCategory fetches all products with a specific Category and price range.
func (db *DynamoDB) GetProductsByCategory(input *GetProductsByCategoryInput) ([]Product, ProductCategoryPaginationKey, error) {
	if err := input.validate(); err != nil {
		return nil, "", err
	}

	var result []Product
	var lastKey ProductCategoryPaginationKey

	res, err := db.db.Query(&dynamodb.QueryInput{
		TableName:              aws.String(db.tableName),
		IndexName:              aws.String("GSI1"),
		KeyConditionExpression: aws.String("#GSI1PK = :gsi1pk And #GSI1SK BETWEEN :from AND :to"),
		ExpressionAttributeNames: map[string]*string{
			"#GSI1PK": aws.String("GSI1PK"),
			"#GSI1SK": aws.String("GSI1SK"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":gsi1pk": {
				S: aws.String(fmt.Sprintf("PRODUCT#CATEGORY#%s", input.Category)),
			},
			":from": {
				S: aws.String(zerosPricePadding(input.FromPrice)),
			},
			":to": {
				S: aws.String(zerosPricePadding(input.ToPrice)),
			},
		},
		Limit:             aws.Int64(int64(input.PaginationLimit)),
		ExclusiveStartKey: decodePaginationKey(input.PreviousKey),
	})
	if err != nil {
		return nil, "", err
	}
	if len(res.Items) == 0 {
		// TODO error not found here?
		return nil, "", nil
	}

	err = dynamodbattribute.UnmarshalMap(res.LastEvaluatedKey, &lastKey)
	if err != nil {
		return nil, "", err
	}

	err = dynamodbattribute.UnmarshalListOfMaps(res.Items, &result)
	if err != nil {
		return nil, "", err
	}

	return result, lastKey, err
}

func (db *DynamoDB) AddBasketItem(item BasketItem) error {
	pk := fmt.Sprintf("BASKET#%s", item.CustomerID)
	sort := fmt.Sprintf("PRODUCT#%s", time.Now().Format(time.RFC3339))

	i, err := dynamodbattribute.MarshalMap(&item)
	if err != nil {
		return err
	}
	i["Type"] = &dynamodb.AttributeValue{S: aws.String("BasketItem")}
	i["PK"] = &dynamodb.AttributeValue{S: aws.String(pk)}
	i["SK"] = &dynamodb.AttributeValue{S: aws.String(sort)}

	_, err = db.db.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(db.tableName),
		Item:      i,
	})

	return err
}

type SortableID ksuid.KSUID

// NewSortableID creates a new sortable id.
func NewSortableID() SortableID { return SortableID(ksuid.New()) }

// String satisfies the Stringer interface.
func (id SortableID) String() string { return ksuid.KSUID(id).String() }

func (id *SortableID) MarshalDynamoDBAttributeValue(av *dynamodb.AttributeValue) error {
	v := fmt.Sprintf("%s", id)
	av.S = &v
	return nil
}

func (id *SortableID) UnmarshalDynamoDBAttributeValue(av *dynamodb.AttributeValue) error {
	if av.S == nil {
		return nil
	}

	v, err := ksuid.Parse(*av.S)
	if err != nil {
		return err
	}
	*id = SortableID(v)

	return nil
}

type ProductCategoryPaginationKey string

func (k *ProductCategoryPaginationKey) UnmarshalDynamoDBAttributeValue(av *dynamodb.AttributeValue) error {
	if av.M == nil {
		return nil
	}
	key := fmt.Sprintf("%s_%s_%s_%s", *av.M["PK"].S, *av.M["SK"].S, *av.M["GSI1PK"].S, *av.M["GSI1SK"].S)
	key = base64.StdEncoding.EncodeToString([]byte(key))

	*k = ProductCategoryPaginationKey(key)

	return nil
}

func decodePaginationKey(pkey ProductCategoryPaginationKey) map[string]*dynamodb.AttributeValue {
	if pkey == "" {
		return nil
	}
	key, err := base64.StdEncoding.DecodeString(string(pkey))
	if err != nil {
		// TODO return error here instead?
		return nil
	}
	s := strings.Split(string(key), "_")
	if len(s) != 4 {
		// TODO error
		return nil
	}

	pk, sk, gsi, gsisk := s[0], s[1], s[2], s[3]

	return map[string]*dynamodb.AttributeValue{
		"PK": {
			S: aws.String(pk),
		},
		"SK": {
			S: aws.String(sk),
		},
		"GSI1PK": {
			S: aws.String(gsi),
		},
		"GSI1SK": {
			S: aws.String(gsisk),
		},
	}

}

func zerosPricePadding(i int) string {
	return fmt.Sprintf("%015d", i)
}
