# Tewq

Me trying to apply what I have learned from [this book](https://www.dynamodbbook.com/), and also to learn more about "Serverless".

## DynamoDB Access Patterns

| Access Pattern                     | Index      | Parameters                | Notes |
|:-----------------------------------|:-----------|:--------------------------|:------|
| Get Baskets Products               | Main Table | * CustomerID              | TODO  |
| Get Products by Category           | GSI1       | * Category <br /> * Price | TODO  |
| Get Products by Category and Price | GSI1       | * Category                | TODO  |
| Get Products Reviews               |            | *                         |       |
| Get Users 10 latest Reviews        |            | *                         |       |
| Get Users 10 latest Orders         |            | *                         |       |
| Get Featured Products              |            | *                         |       |
| Create Product                     |            | *                         |       |
| Create an Option for an Product    |            | *                         |       |
| Get a Product                      |            | *                         |       |
| Add Product to Basket              |            | *                         |       |

## Entity Chart


### Main Table

| Entity             | PK                  | SK                |
| :----------------- | ----------------:   | ----------------: |
| Basket             | Basket#[CustomerID] | PRODUCT#[Date]    |
| Product            | Product#[ProductID] | METADATA#         |
| Option             | Product#[ProductID] | OPTION#[OptionID] |
| Review             |                     |                   |
| Order              |                     |                   |
| Category           | N/A                 | N/A               |
| OrderDetails       |                     |                   |


### GSI1

| Entity             | GSI1K                       | GSI2K    |
| :----------------- | -------------------:        | -------: |
| Product            | PRODUCT#CATEGORY#[Category] | [Price]  |



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
