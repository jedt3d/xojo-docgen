# EddiesElectronics data dictionary

- Dialect: `sqlite`
- Source: `EddiesElectronics.sqlite`
- Tables: 4
- Views: 0
- Relationships: 3

## Relationships

| Origin | From | References | Evidence |
|---|---|---|---|
| suggested | `InvoiceItems.InvoiceNo` | `Invoices.InvoiceNo` | Column InvoiceNo uniquely matches table Invoices and primary key InvoiceNo by name and type. |
| suggested | `InvoiceItems.ProductCode` | `Products.Code` | Column ProductCode uniquely matches table Products and primary key Code by name and type. |
| suggested | `Invoices.CustomerID` | `Customers.ID` | Column CustomerID uniquely matches table Customers and primary key ID by name and type. |

## Customers

| Column | Type | Nullable | Key | Default | Generated |
|---|---|---|---|---|---|
| `ID` | `INTEGER` | No | PK 1 | `` |  |
| `FirstName` | `TEXT` | Yes |  | `` |  |
| `LastName` | `TEXT` | Yes |  | `` |  |
| `Address` | `TEXT` | Yes |  | `` |  |
| `City` | `TEXT` | Yes |  | `` |  |
| `State` | `TEXT` | Yes |  | `` |  |
| `Zip` | `TEXT` | Yes |  | `` |  |
| `Phone` | `TEXT` | Yes |  | `` |  |
| `Email` | `TEXT` | Yes |  | `` |  |
| `Photo` | `BLOB` | Yes |  | `` |  |
| `Taxable` | `INTEGER` | Yes |  | `` |  |

## InvoiceItems

| Column | Type | Nullable | Key | Default | Generated |
|---|---|---|---|---|---|
| `ID` | `INTEGER` | No | PK 1 | `` |  |
| `InvoiceNo` | `INTEGER` | Yes |  | `` |  |
| `ProductCode` | `TEXT` | Yes |  | `` |  |
| `Quantity` | `INTEGER` | Yes |  | `` |  |

## Invoices

| Column | Type | Nullable | Key | Default | Generated |
|---|---|---|---|---|---|
| `InvoiceNo` | `INTEGER` | No | PK 1 | `` |  |
| `CustomerID` | `INTEGER` | Yes |  | `` |  |
| `InvoiceDate` | `TEXT` | Yes |  | `` |  |
| `InvoiceAmount` | `FLOAT` | Yes |  | `` |  |

## Products

| Column | Type | Nullable | Key | Default | Generated |
|---|---|---|---|---|---|
| `Code` | `TEXT` | No | PK 1 | `` |  |
| `Name` | `TEXT` | Yes |  | `` |  |
| `Price` | `FLOAT` | Yes |  | `` |  |

