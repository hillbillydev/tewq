# Tewq

<p align="center">
<img src="https://github.com/Tinee/tewq/workflows/Go/badge.svg" alt="Tests Status" />
</p>

Me trying to apply what I have learned from [this book](https://www.dynamodbbook.com/), and also to learn more about "Serverless".

## DynamoDB Access Patterns

| Access Pattern                     | Index      | Parameters                                        |
|:-----------------------------------|:-----------|:--------------------------------------------------|
| Add Product to Basket              | Main Table | * CustomerID <br /> * ProductID <br /> * OptionID |
| Get Baskets Products               | Main Table | * CustomerID                                      |
| Create Product                     | Main Table | * Product                                         |
| Get Product                        | Main Table | * ProductID                                       |
| Create an Option for an Product    | Main Table | * Option                                          |
| Get Products by Category and Price | GSI1       | * Category <br /> * Price                         |
| Get Products by Category           | GSI1       | * Category                                        |
| Get Products Reviews               | Main Table | * ProductID                                       |
| Get Users latest Orders            | Main Table | * UserID                                          |
| Get Users latest Reviews           | GSI1       | * UserID <br /> * Date                            |
| Get Order                          | GSI1       | * OrderID                                         |
| Get Order Information              | GSI1       | * OrderID                                         |

## Entity Chart

### Main Table

| Entity             | PK                  | SK                |
| :----------------- | ----------------:   | ----------------: |
| Basket             | Basket#[CustomerID] | PRODUCT#[Date]    |
| Product            | Product#[ProductID] | METADATA#         |
| Option             | Product#[ProductID] | OPTION#[OptionID] |
| Review             | Product#[ProductID] | REVIEW#[ReviewID] |
| Order              | USER#[UserID]       | ORDER#[OrderId]   |
| OrderLineItem      | ORDERITEM#[ItemID]  | Order#[OrderID]   |
| Category           | N/A                 | N/A               |

### GSI1

| Entity             | GSI1PK                      | GSI1SK             |
| :----------------- | -------------------:        | -------:           |
| Product            | PRODUCT#CATEGORY#[Category] | [Price]            |
| Review             | USER#[UserID]               | REVIEW#[Date]      |
| Order              | ORDER#[OrderId]             | METADATA#          |
| OrderLineItem      | ORDER#[OrderID]             | ORDERITEM#[ItemId] |


## Entity Relationship Diagram

![ERD](https://github.com/Tinee/tewq/blob/assets/erd.png)


## Testing

### Integration

1. Install [docker](https://www.docker.com/get-started).
2. Run `docker-compose up` at the root of the project, this will run DynamoDB locally.
3. You can now run `go test ./...` this will run all the tests.

**Test Tip**

* Download [NoSQL Workbench](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/workbench.html)

```go
  func TestSomethingInDynamoDB(t *testing.T) {
    tdb, _ := NewTestDynamoDB() // creates a table in your DynamoDB instance.

    // If you comment out the Close method it will not delete the test database that got created.
    // This give you an opportunity to peek into the instance with NoSQL Workbench.
    // defer tdb.Close()
  }
```
