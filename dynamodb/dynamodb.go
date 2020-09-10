package dynamodb

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/segmentio/ksuid"
)

// DynamoDB wraps AWS dynamodb.DynamoDB
// This is to add domain logic.
type DynamoDB struct {
	db        *dynamodb.DynamoDB
	tableName string
}

// New creates a DynamoDB wrapper.
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

// SortableID makes the ID sortable.
type SortableID ksuid.KSUID

// NewSortableID creates a new sortable id.
func NewSortableID() SortableID { return SortableID(ksuid.New()) }

// String satisfies the Stringer interface.
func (id SortableID) String() string { return ksuid.KSUID(id).String() }

// MarshalDynamoDBAttributeValue satisfy the dynamodbattribute.Marshaler interface.
// By doing that I can tell DynamoDB how to handle my SortableID.
func (id *SortableID) MarshalDynamoDBAttributeValue(av *dynamodb.AttributeValue) error {
	v := fmt.Sprintf("%s", id)
	av.S = &v
	return nil
}

// UnmarshalDynamoDBAttributeValue satisfy the dynamodbattribute.Unmarshaler interface.
// By doing that I can tell DynamoDB how to handle my SortableID.
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

func zerosPricePadding(i int) string {
	return fmt.Sprintf("%015d", i)
}
