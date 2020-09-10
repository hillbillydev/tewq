package dynamodb

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type TestDynamoDB struct {
	*DynamoDB
}

// NewTestDynamoDB connects to http://localhost:8000, and this should not change.
// It then create a temporary test tables, with name looking like this Tewq-Test_2020-09-04_09-21-37.
// To clean up your test resources you will have to call the Close() method.
func NewTestDynamoDB() (*TestDynamoDB, error) {
	tableName := fmt.Sprintf("%s_%s", "Tewq-Test", time.Now().Format("2006-01-02_15-04-05.000000"))

	db, err := New("http://localhost:8000", tableName)
	if err != nil {
		return nil, err
	}
	tdb := &TestDynamoDB{db}

	return tdb, tdb.createTestTable()
}

// Close closes the test resources.
func (t *TestDynamoDB) Close() error {
	if !strings.Contains(t.db.Endpoint, "localhost") {
		return fmt.Errorf("Tried to run against %s, but can only run against an local instance", t.db.Endpoint)
	}

	_, err := t.db.DeleteTable(&dynamodb.DeleteTableInput{
		TableName: aws.String(t.tableName),
	})

	return err
}

func (t *TestDynamoDB) createTestTable() error {
	if !strings.Contains(t.db.Endpoint, "localhost") {
		return fmt.Errorf("Tried to run against %s, but can only run against an local instance", t.db.Endpoint)
	}

	_, err := t.db.CreateTable(&dynamodb.CreateTableInput{
		TableName: aws.String(t.tableName),
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("PK"),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("SK"),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("GSI1PK"),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("GSI1SK"),
				AttributeType: aws.String("S"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("PK"),
				KeyType:       aws.String("HASH"),
			},
			{
				AttributeName: aws.String("SK"),
				KeyType:       aws.String("RANGE"),
			},
		},
		GlobalSecondaryIndexes: []*dynamodb.GlobalSecondaryIndex{
			{
				IndexName: aws.String("GSI1"),
				KeySchema: []*dynamodb.KeySchemaElement{
					{
						AttributeName: aws.String("GSI1PK"),
						KeyType:       aws.String("HASH"),
					},
					{
						AttributeName: aws.String("GSI1SK"),
						KeyType:       aws.String("RANGE"),
					},
				},
				Projection: &dynamodb.Projection{
					ProjectionType: aws.String(dynamodb.ProjectionTypeAll),
				},
				ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(10),
					WriteCapacityUnits: aws.Int64(10),
				},
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(10),
			WriteCapacityUnits: aws.Int64(10),
		},
	})

	return err
}
