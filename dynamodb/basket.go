package dynamodb

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

// Basket contains the Products an customer wants to buy in the future.
type Basket struct {
	Products []Product `json:"products"`
}

// BasketItem contains the pointers to which customer
// wants which product within the basket.
type BasketItem struct {
	CustomerID      SortableID `json:"customerId" dynamodbav:"CustomerId"`
	ProductID       SortableID `json:"productId" dynamodbav:"ProductId"`
	ProductOptionID SortableID `json:"productOptionId" dynamodbav:"ProductOptionId"`
}

// AddBasketItem adds an BasketItem
func (db *DynamoDB) AddBasketItem(item BasketItem) error {
	pk := fmt.Sprintf("BASKET#%s", item.CustomerID)
	sort := fmt.Sprintf("PRODUCT#%s", NewSortableID())

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

func (db *DynamoDB) GetBasketProducts(customerID SortableID) ([]Product, error) {
	pk := fmt.Sprintf("BASKET#%s", customerID)

	res, err := db.db.Query(&dynamodb.QueryInput{
		TableName:              aws.String(db.tableName),
		KeyConditionExpression: aws.String("#PK = :pk"),
		ExpressionAttributeNames: map[string]*string{
			"#PK": aws.String("PK"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":pk": {
				S: aws.String(pk),
			},
		},
	})
	if err != nil {
		return nil, err
	}
	if len(res.Items) == 0 {
		return nil, nil
	}

	attrs := []map[string]*dynamodb.AttributeValue{}
	for _, att := range res.Items {
		attrs = append(attrs, map[string]*dynamodb.AttributeValue{
			"PK": {
				S: aws.String(fmt.Sprintf("PRODUCT#%s", *att["ProductId"].S)),
			},
			"SK": {
				S: aws.String("METADATA#"),
			},
		})
	}
	batch, err := db.db.BatchGetItem(&dynamodb.BatchGetItemInput{
		RequestItems: map[string]*dynamodb.KeysAndAttributes{
			db.tableName: {
				Keys: attrs,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	prods, ok := batch.Responses[db.tableName]
	if !ok {
		return nil, nil
	}

	var products []Product
	err = dynamodbattribute.UnmarshalListOfMaps(prods, &products)
	if err != nil {
		return nil, err
	}

	return products, nil
}

