package dynamodb

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/segmentio/ksuid"
)

type Option struct {
	ID             SortableID `json:"id" dynamodbav:"Id,omitempty"`
	CreatedDate    string     `json:"createdUtc" dynamodbav:"CreatedUtc,omitempty"`
	Size           string     `json:"size" dynamodbav:"Size,omitempty"`     // TODO enum?
	Socket         string     `json:"socket" dynamodbav:"Socket,omitempty"` // TODO enum?
	Color          string     `json:"color" dynamodbav:"Color,omitempty"`   // TODO enum?
	Stock          int        `json:"stock" dynamodbav:"Stock,omitempty"`
	ShaftStiffness float64    `json:"shaftStiffness" dynamodbav:"ShaftStiffness,omitempty"`
}

type Product struct {
	ID          SortableID `json:"id" dynamodbav:"Id,omitempty"`
	CreatedDate string     `json:"createdUtc" dynamodbav:"CreatedUtc,omitempty"`
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

	p.CreatedDate = time.Now().Format(time.RFC3339)
	p.ID = NewSortableID()

	pk := fmt.Sprintf("PRODUCT#%s", p.ID)
	sort := "METADATA#"
	gs1pk := fmt.Sprintf("PRODUCT#CATEGORY#%s", p.Category)
	gs1sk := strconv.Itoa(p.Price)

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
	option.CreatedDate = time.Now().Format(time.RFC3339)

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

func (db *DynamoDB) GetProductsByCategoryAndPrice(category string, from, to int) ([]Product, error) {
	return db.getProductsByCategoryAndPrice(category, from, to)
}

func (db *DynamoDB) GetProductsByCategory(category string) ([]Product, error) {
	return db.getProductsByCategoryAndPrice(category, 0, math.MaxInt64)
}

func (db *DynamoDB) getProductsByCategoryAndPrice(category string, from, to int) ([]Product, error) {
	var result []Product

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
				S: aws.String(fmt.Sprintf("PRODUCT#CATEGORY#%s", category)),
			},
			":from": {
				S: aws.String(strconv.Itoa(from)),
			},
			":to": {
				S: aws.String(strconv.Itoa(to)),
			},
		},
	})
	if err != nil {
		return nil, err
	}
	if len(res.Items) == 0 {
		// TODO error not found here?
		return nil, nil
	}

	err = dynamodbattribute.UnmarshalListOfMaps(res.Items, &result)
	if err != nil {
		return nil, err
	}

	return result, err
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

func NewSortableID() SortableID         { return SortableID(ksuid.New()) }
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
	sid := SortableID(v)
	id = &sid

	return nil
}
