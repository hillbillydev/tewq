package dynamodb

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

//https://docs.aws.amazon.com/sdk-for-go/api/service/dynamodb/#CreateGlobalSecondaryIndexAction

//CreateUserOrderGSI creates a new GSI2KO with KeysOnly projection; will be used as a UserAndOrderIdx
//but made the name and keyschema attrs generic for now for testing
func (t *TestDynamoDB) CreateUserOrderGSI() error {

	c := &dynamodb.GlobalSecondaryIndexUpdate{
		Create: &dynamodb.CreateGlobalSecondaryIndexAction{
			IndexName: aws.String("GSI2KO"),
			KeySchema: []*dynamodb.KeySchemaElement{
				{
					AttributeName: aws.String("GSI2PK"),
					KeyType:       aws.String("HASH"),
				},
				{
					AttributeName: aws.String("GSI2SK"),
					KeyType:       aws.String("RANGE"),
				},
			},
			Projection: &dynamodb.Projection{
				ProjectionType: aws.String(dynamodb.ProjectionTypeKeysOnly),
			},
			ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
				ReadCapacityUnits:  aws.Int64(1),
				WriteCapacityUnits: aws.Int64(1),
			},
		},
	}

	_, err := t.db.UpdateTable(&dynamodb.UpdateTableInput{
		TableName:                   aws.String(t.tableName),
		GlobalSecondaryIndexUpdates: []*dynamodb.GlobalSecondaryIndexUpdate{c},
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("GSI2PK"),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String("GSI2SK"),
				AttributeType: aws.String("S"),
			},
		},
	})
	return err

}

//Part of this operation involves backfilling data from the table into the new index.
// During backfilling, the table remains available. However, the index is not ready until its
//Backfilling attribute changes from true to false. You can use the DescribeTable action to view this attribute.

/*
UpdateTable is async so we need to check if the GSI is done backfilling
index is not ready until its Backfilling attribute changes from true to false.
With the go-sdk this is readily captured in the IndexStatus attr of the GSI Description struct

*/

func (t *TestDynamoDB) IsGSIReady(indexName string) (bool, error) {
	res, err := t.db.DescribeTable(&dynamodb.DescribeTableInput{
		TableName: aws.String(t.tableName),
	})
	if err != nil {
		return false, err
	}

	var gsiProps []*dynamodb.GlobalSecondaryIndexDescription
	gsiProps = res.Table.GlobalSecondaryIndexes

	var status string
	for _, gsi := range gsiProps {
		if *gsi.IndexName == indexName {
			status = *gsi.IndexStatus
		}
	}
	if status == "" {
		return false, fmt.Errorf("IndexName=%v not in db", indexName)
	}
	if status != dynamodb.IndexStatusActive {
		return false, fmt.Errorf("indexName=%v Status=%v; not ACTIVE. Try Again", indexName, status)
	}

	return true, nil
}
