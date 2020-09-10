# Tewq

Me trying to apply what I have learned from [this book](https://www.dynamodbbook.com/), and also to learn more about "Serverless".

<p align="center">
<img src="https://github.com/Tinee/tewq/workflows/Go/badge.svg" alt="Tests Status" />
</p>

# DynamoDB

## Access Patterns

|      Access Pattern     | Index |                   Key Condition                   | Filter Condition |
|:-----------------------:|:-----:|:-------------------------------------------------:|------------------|
|     **Get Products**    |       |                                                   |                  |
|       by productID      | Table |                   PK = productID                  |                  |
|       by category       |  GSI1 |                 GSI1PK = category                 |                  |
|  by category and price  |  GSI1 | GSI1PK = category, GSI1PK between(price1, price2) |                  |
| **Get Product Reviews** |       |                                                   |                  |
|       by reviewID       | Table |           PK = productID, SK = reviewID           |                  |
| **Get Basket Products** |       |                                                   |                  |
|        by userID        | Table |                    PK = userID                    |                  |
| **Get Users Dashboard** |       |                                                   |                  |
|       get reviews       | Table |                  GSI1PK = userID                  |                  |
|        get orders       | Table |                  GSI1PK = userID                  |                  |
|     **Get Reviews**     |       |                                                   |                  |
|       by productID      | Table |     PK = productID, SK begins_with("REVIEW#")     |                  |
|        by userID        |  GSI1 |    GSI1PK = USER, GSI2SK begins_with("REVIEW#")   |                  |
|      **Get Orders**     |       |                                                   |                  |
|        by userID        | Table |       PK = userID, SK begins_with("ORDER#")       |                  |
|  **Get Orders Details** |       |                                                   |                  |
|        by orderID       |  GSI1 |                  GSI1PK = orderID                 |                  |

## Entity Charts

**Main Table**

| Entity             | PK                  | SK                |
| :----------------- | ----------------:   | ----------------: |
| Basket             | Basket#[CustomerID] | PRODUCT#[Date]    |
| Product            | Product#[ProductID] | METADATA#         |
| Option             | Product#[ProductID] | OPTION#[OptionID] |
| Review             | Product#[ProductID] | REVIEW#[ReviewID] |
| Order              | USER#[UserID]       | ORDER#[OrderId]   |
| OrderLineItem      | ORDERITEM#[ItemID]  | Order#[OrderID]   |
| Category           | N/A                 | N/A               |

**GSI1**

| Entity             | GSI1PK                      | GSI1SK             |
| :----------------- | -------------------:        | -------:           |
| Product            | PRODUCT#CATEGORY#[Category] | [Price]            |
| Review             | USER#[UserID]               | REVIEW#[Date]      |
| Order              | ORDER#[OrderId]             | METADATA#          |
| OrderLineItem      | ORDER#[OrderID]             | ORDERITEM#[ItemId] |


## Entity Relationship Diagram

![ERD](https://github.com/Tinee/tewq/blob/assets/erd.png)

# Testing

## Integration

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

