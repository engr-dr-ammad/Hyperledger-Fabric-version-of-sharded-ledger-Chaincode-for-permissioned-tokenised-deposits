# Chaincode API Reference

## Overview

This document provides detailed API documentation for all chaincode functions in the deposit ledger system.

## Admin Functions

### InitLedger

Initializes the ledger with default configuration and shards.

**Permission**: Admin only

**Function Signature**:
```go
func (dc *DepositChaincode) InitLedger(ctx contractapi.TransactionContextInterface) error
```

**Parameters**: None

**Returns**: Error if initialization fails

**Example**:
```bash
peer chaincode invoke -C channel-shard-1 \
  -n deposit-chaincode \
  -c '{"function":"InitLedger","Args":[]}'
```

**Side Effects**:
- Creates default shards (shard-primary, shard-secondary)
- Emits "ledgerInitialized" event
- Initializes sharding manager

---

### CreateShard

Creates a new shard (channel partition).

**Permission**: Admin only

**Function Signature**:
```go
func (dc *DepositChaincode) CreateShard(
  ctx contractapi.TransactionContextInterface,
  shardID string,
  channelName string,
  description string,
) error
```

**Parameters**:
- `shardID`: Unique identifier for the shard (e.g., "shard-usd")
- `channelName`: Fabric channel name for this shard
- `description`: Human-readable description of the shard

**Returns**: Error if creation fails

**Validation**:
- shardID must be unique
- channelName must be valid
- description must not be empty

**Example**:
```bash
peer chaincode invoke -C channel-shard-1 \
  -n deposit-chaincode \
  -c '{"function":"CreateShard","Args":["shard-eur","channel-shard-eur","EUR deposit shard"]}'
```

**Side Effects**:
- Creates new Shard object
- Saves to ledger state
- Adds to sharding manager
- Emits "shardCreated" event

---

## Deposit Functions

### IssueDeposit

Issues a new tokenised deposit.

**Permission**: Issuer, Admin

**Function Signature**:
```go
func (dc *DepositChaincode) IssueDeposit(
  ctx contractapi.TransactionContextInterface,
  depositID string,
  amount string,
  currency string,
  shardID string,
  expiryDays string,
  metadataJSON string,
) error
```

**Parameters**:
- `depositID`: Unique deposit identifier (format: "DEP-XXXX-XXX")
- `amount`: Deposit amount as string (e.g., "1000000")
- `currency`: ISO 4217 currency code (e.g., "USD", "EUR")
- `shardID`: Target shard ID for this deposit
- `expiryDays`: Number of days until expiry (empty = no expiry)
- `metadataJSON`: Optional JSON metadata

**Returns**: Error if issuance fails

**Validation**:
- depositID must match pattern "DEP-*"
- amount must be positive
- currency must be valid ISO code
- shard must exist and be active
- issuer must be authorized for shard

**Example**:
```bash
peer chaincode invoke -C channel-shard-1 \
  -n deposit-chaincode \
  -c '{"function":"IssueDeposit","Args":["DEP-2024-001","1000000","USD","shard-primary","365","{}"]}'
```

**Side Effects**:
- Creates Deposit with initial holding
- Saves to ledger state
- Creates audit log entry
- Emits "depositIssued" event

**Response Format**:
```json
{
  "depositId": "DEP-2024-001",
  "issuerId": "issuer-org-1",
  "amount": 1000000,
  "currency": "USD",
  "status": "active",
  "shardId": "shard-primary",
  "createdAt": "2024-01-15T10:30:45Z",
  "holdings": [
    {
      "userId": "issuer-org-1",
      "amount": 1000000,
      "status": "active"
    }
  ]
}
```

---

### TransferDeposit

Transfers deposit between users.

**Permission**: Depositor, Admin

**Function Signature**:
```go
func (dc *DepositChaincode) TransferDeposit(
  ctx contractapi.TransactionContextInterface,
  depositID string,
  toUserID string,
  amountStr string,
  reasonJSON string,
) error
```

**Parameters**:
- `depositID`: ID of deposit to transfer
- `toUserID`: Recipient user ID
- `amountStr`: Transfer amount as string
- `reasonJSON`: Optional JSON with transfer reason

**Returns**: Error if transfer fails

**Validation**:
- Caller must own holdings in deposit
- Amount must not exceed available balance
- Deposit must be active and not expired
- Recipient cannot be same as caller

**Example**:
```bash
peer chaincode invoke -C channel-shard-1 \
  -n deposit-chaincode \
  -c '{"function":"TransferDeposit","Args":["DEP-2024-001","user-2","100000","{\"reason\":\"Payment for services\"}"]}'
```

**Side Effects**:
- Updates deposit holdings
- Creates transfer record
- Creates audit log entry
- Emits "depositTransferred" event

**Transfer Record**:
```json
{
  "transferId": "TXF-2024-00001",
  "depositId": "DEP-2024-001",
  "fromUserId": "user-1",
  "toUserId": "user-2",
  "amount": 100000,
  "status": "completed",
  "createdAt": "2024-01-15T10:31:00Z"
}
```

---

### RedeemDeposit

Redeems a deposit (claims the underlying asset).

**Permission**: Depositor, Admin

**Function Signature**:
```go
func (dc *DepositChaincode) RedeemDeposit(
  ctx contractapi.TransactionContextInterface,
  depositID string,
  amountStr string,
) error
```

**Parameters**:
- `depositID`: ID of deposit to redeem
- `amountStr`: Redemption amount as string

**Returns**: Error if redemption fails

**Validation**:
- Caller must own holdings in deposit
- Amount must not exceed available balance
- Deposit must be active

**Example**:
```bash
peer chaincode invoke -C channel-shard-1 \
  -n deposit-chaincode \
  -c '{"function":"RedeemDeposit","Args":["DEP-2024-001","500000"]}'
```

**Side Effects**:
- Updates deposit holdings
- Creates redemption record
- Updates deposit status if fully redeemed
- Creates audit log entry
- Emits "depositRedeemed" event

**Redemption Record**:
```json
{
  "redemptionId": "RED-2024-00001",
  "depositId": "DEP-2024-001",
  "userId": "user-1",
  "amount": 500000,
  "status": "completed",
  "createdAt": "2024-01-15T10:32:00Z"
}
```

---

## Query Functions

### GetDeposit

Retrieves a specific deposit.

**Permission**: Deposit owner, Issuer, Admin, Auditor

**Function Signature**:
```go
func (dc *DepositChaincode) GetDeposit(
  ctx contractapi.TransactionContextInterface,
  depositID string,
) (*Deposit, error)
```

**Parameters**:
- `depositID`: Deposit ID to retrieve

**Returns**: Deposit object or error

**Example**:
```bash
peer chaincode query -C channel-shard-1 \
  -n deposit-chaincode \
  -c '{"function":"GetDeposit","Args":["DEP-2024-001"]}'
```

**Response**:
```json
{
  "depositId": "DEP-2024-001",
  "issuerId": "issuer-org-1",
  "ownerId": "issuer-org-1",
  "amount": 1000000,
  "currency": "USD",
  "status": "active",
  "shardId": "shard-primary",
  "createdAt": "2024-01-15T10:30:45Z",
  "updatedAt": "2024-01-15T10:32:00Z",
  "holdings": [
    {
      "userId": "issuer-org-1",
      "amount": 500000,
      "status": "active"
    },
    {
      "userId": "user-2",
      "amount": 500000,
      "status": "active"
    }
  ]
}
```

---

### GetUserDeposits

Retrieves all deposits held by a user in a specific shard.

**Permission**: Depositor (own records), Admin, Auditor

**Function Signature**:
```go
func (dc *DepositChaincode) GetUserDeposits(
  ctx contractapi.TransactionContextInterface,
  userID string,
  shardID string,
) ([]*Deposit, error)
```

**Parameters**:
- `userID`: User ID to query
- `shardID`: Shard to search within

**Returns**: Array of Deposit objects

**Example**:
```bash
peer chaincode query -C channel-shard-1 \
  -n deposit-chaincode \
  -c '{"function":"GetUserDeposits","Args":["user-1","shard-primary"]}'
```

**Response**:
```json
[
  {
    "depositId": "DEP-2024-001",
    "amount": 500000,
    "currency": "USD",
    "status": "active",
    "holdings": [
      {
        "userId": "user-1",
        "amount": 500000
      }
    ]
  },
  {
    "depositId": "DEP-2024-002",
    "amount": 250000,
    "currency": "USD",
    "status": "active",
    "holdings": [
      {
        "userId": "user-1",
        "amount": 250000
      }
    ]
  }
]
```

---

### QueryDepositsByAmount

Queries deposits within a value range.

**Permission**: Admin, Auditor (full results), Others (filtered results)

**Function Signature**:
```go
func (dc *DepositChaincode) QueryDepositsByAmount(
  ctx contractapi.TransactionContextInterface,
  minAmount string,
  maxAmount string,
) ([]*Deposit, error)
```

**Parameters**:
- `minAmount`: Minimum deposit amount (inclusive)
- `maxAmount`: Maximum deposit amount (inclusive)

**Returns**: Array of matching Deposit objects

**Example**:
```bash
peer chaincode query -C channel-shard-1 \
  -n deposit-chaincode \
  -c '{"function":"QueryDepositsByAmount","Args":["100000","1000000"]}'
```

---

### GetShardInfo

Retrieves information about a specific shard.

**Permission**: Public (read-only)

**Function Signature**:
```go
func (dc *DepositChaincode) GetShardInfo(
  ctx contractapi.TransactionContextInterface,
  shardID string,
) (*Shard, error)
```

**Parameters**:
- `shardID`: Shard ID to retrieve

**Returns**: Shard object

**Example**:
```bash
peer chaincode query -C channel-shard-1 \
  -n deposit-chaincode \
  -c '{"function":"GetShardInfo","Args":["shard-primary"]}'
```

**Response**:
```json
{
  "shardId": "shard-primary",
  "channelName": "channel-shard-primary",
  "description": "Primary shard for mixed deposits",
  "issuers": ["issuer-org-1", "issuer-org-2"],
  "status": "active",
  "createdAt": "2024-01-01T00:00:00Z"
}
```

---

### GetTransactionHistory

Retrieves complete transaction history for a deposit.

**Permission**: Deposit owner, Auditor, Admin

**Function Signature**:
```go
func (dc *DepositChaincode) GetTransactionHistory(
  ctx contractapi.TransactionContextInterface,
  depositID string,
) ([]*TransactionRecord, error)
```

**Parameters**:
- `depositID`: Deposit ID to get history for

**Returns**: Array of transaction records

**Example**:
```bash
peer chaincode query -C channel-shard-1 \
  -n deposit-chaincode \
  -c '{"function":"GetTransactionHistory","Args":["DEP-2024-001"]}'
```

**Response**:
```json
[
  {
    "id": "TXN-2024-00001",
    "entityId": "DEP-2024-001",
    "entityType": "deposit",
    "operation": "create",
    "actorId": "issuer-org-1",
    "timestamp": "2024-01-15T10:30:45Z",
    "newState": {
      "amount": 1000000,
      "currency": "USD",
      "status": "active"
    }
  },
  {
    "id": "TXN-2024-00002",
    "entityId": "DEP-2024-001",
    "entityType": "deposit",
    "operation": "transfer",
    "actorId": "user-1",
    "timestamp": "2024-01-15T10:31:00Z",
    "previousState": {
      "holdings": [{"userId": "issuer-org-1", "amount": 1000000}]
    },
    "newState": {
      "holdings": [
        {"userId": "issuer-org-1", "amount": 500000},
        {"userId": "user-2", "amount": 500000}
      ]
    }
  }
]
```

---

## Error Responses

All functions return errors in the following format:

```
Error [ComponentName]: [ErrorDescription]
```

### Common Error Codes

| Error | Meaning | HTTP Equivalent |
|-------|---------|-----------------|
| `access denied` | Permission check failed | 403 |
| `not found` | Resource does not exist | 404 |
| `invalid {field}` | Input validation failed | 400 |
| `compliance check failed` | Business rule violation | 422 |
| `already exists` | Duplicate resource | 409 |
| `insufficient balance` | Not enough funds | 400 |

### Example Error Response

```json
{
  "error": "access denied to deposit DEP-2024-001",
  "code": "PERMISSION_DENIED",
  "timestamp": "2024-01-15T10:30:45Z"
}
```

---

## Chaincode Events

The chaincode emits events for important state changes:

### ledgerInitialized
```json
{
  "type": "ledgerInitialized",
  "timestamp": "2024-01-15T10:30:45Z",
  "data": {
    "shards": 2
  }
}
```

### depositIssued
```json
{
  "type": "depositIssued",
  "timestamp": "2024-01-15T10:30:45Z",
  "data": {
    "depositId": "DEP-2024-001",
    "amount": 1000000,
    "currency": "USD"
  }
}
```

### depositTransferred
```json
{
  "type": "depositTransferred",
  "timestamp": "2024-01-15T10:31:00Z",
  "data": {
    "transferId": "TXF-2024-00001",
    "depositId": "DEP-2024-001",
    "fromUserId": "user-1",
    "toUserId": "user-2",
    "amount": 100000
  }
}
```

### depositRedeemed
```json
{
  "type": "depositRedeemed",
  "timestamp": "2024-01-15T10:32:00Z",
  "data": {
    "redemptionId": "RED-2024-00001",
    "depositId": "DEP-2024-001",
    "userId": "user-1",
    "amount": 500000
  }
}
```

---

## Best Practices

### Input Validation
- Always provide amounts as strings to preserve precision
- Validate currency codes before invoking
- Ensure IDs follow the specified format

### Error Handling
- Check error responses before processing results
- Implement retry logic for transient failures
- Log all API errors for audit trails

### Performance
- Use GetUserDeposits for filtered queries instead of QueryDepositsByAmount
- Cache Shard information rather than querying repeatedly
- Use pagination for large result sets

### Security
- Never expose user IDs in client logs
- Validate all external input
- Use TLS for all API communications
- Implement rate limiting at the client level
