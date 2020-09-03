# Tewq

## DynamoDB Access Patterns

|  Access Patterns                                | By                     | Table/Index  | Key Condition                         | Filter Condition                                          |
| :---------------------------------------------: | :---------------       | :----------: | :-----------------------------------: | :-------------------------------------------------------: |
|  Get Users Orders                               |                        |              |                                       |                                                           |
|                                                 | by date and email      | Table        | GSIPK = emails                        | duartion > 0                                              |
|  Get Basket                                     |                        |              |                                       |                                                           |
|                                                 | by date and email      | Table        | GSIPK = emails                        | duartion > 0                                              |
|  Get Item                                       |                        |              |                                       |                                                           |
|                                                 | by id                  | Table        | PK = emails                           | duartion > 0                                              |
|                                                 | by category and price  | GSI1PK       | GSIPK = category, GI1SK price > 0     | stock > 0                                                 |


## Entity Relationship Diagram

![ERD](./erd.png)
