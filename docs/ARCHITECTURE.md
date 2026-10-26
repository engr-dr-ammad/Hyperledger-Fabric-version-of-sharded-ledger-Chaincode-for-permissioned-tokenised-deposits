# Hyperledger Fabric Sharded Ledger - Architecture

## System Overview

This document describes the architecture of the Hyperledger Fabric Sharded Ledger Chaincode for permissioned tokenised deposits.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Client Applications                      │
│  (Banks, Depositories, Regulatory Bodies, Auditors)         │
└──────────────────┬──────────────────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────────────┐
│                  Hyperledger Fabric                          │
│  ┌─────────────┬─────────────┬─────────────┬─────────────┐  │
│  │   Shard 1   │   Shard 2   │   Shard 3   │   Shard N   │  │
│  │  (Channel)  │  (Channel)  │  (Channel)  │  (Channel)  │  │
│  ├─────────────┼─────────────┼─────────────┼─────────────┤  │
│  │   Peers     │   Peers     │   Peers     │   Peers     │  │
│  │  Orderers   │  Orderers   │  Orderers   │  Orderers   │  │
│  │  CouchDB    │  CouchDB    │  CouchDB    │  CouchDB    │  │
│  └─────────────┴─────────────┴─────────────┴─────────────┘  │
│                                                               │
│  Deposits distributed across shards for scalability         │
└─────────────────────────────────────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────────────────────┐
│                 Immutable Ledger History                     │
│            (All transactions recorded on chain)              │
└─────────────────────────────────────────────────────────────┘
```

## Component Architecture

### 1. Chaincode Components

#### Main Chaincode (main.go)
- Entry point for all operations
- Routes requests to appropriate managers
- Implements chaincode interface

#### Ledger Manager (ledger.go)
- Core ledger operations
- CRUD operations for deposits and shards
- Query execution and indexing
- Transaction recording and audit logging

#### Permission Manager (permissions.go)
- Role-based access control (RBAC)
- User authentication and authorization
- Compliance verification
- Shard access validation

#### Sharding Manager (sharding.go)
- Determines shard placement for deposits
- Multiple sharding strategies (hash, currency, issuer, time-based)
- Load balancing calculations
- Consistency verification

#### Models (models.go)
- Data structures for deposits, shards, transfers, redemptions
- Audit logs and transaction records
- Metadata and compliance data

#### Utilities (utils.go)
- ID generation
- Serialization/deserialization
- Validation functions
- Query builders
- Timestamp management

## Data Model

### Deposit
```
┌────────────────────────────────────┐
│          Deposit                   │
├────────────────────────────────────┤
│ DepositID: string                  │
│ IssuerID: string                   │
│ Amount: float64                    │
│ Currency: string (ISO 4217)        │
│ Status: string (active, redeemed)  │
│ ShardID: string                    │
│ CreatedAt: timestamp               │
│ UpdatedAt: timestamp               │
│ ExpiresAt: timestamp (optional)    │
│ Holdings: []Holding                │
│ Metadata: map                      │
└────────────────────────────────────┘
         │
         └─ Holdings
            ├─ UserID
            ├─ Amount
            ├─ AcquiredAt
            └─ Status
```

### Shard
```
┌────────────────────────────────────┐
│           Shard                    │
├────────────────────────────────────┤
│ ShardID: string                    │
│ ChannelName: string                │
│ Description: string                │
│ Issuers: []string                  │
│ Status: string                     │
│ CreatedAt: timestamp               │
│ UpdatedAt: timestamp               │
│ Metadata: map                      │
└────────────────────────────────────┘
```

### Transaction Types

#### Deposit Transfer
- From: User A
- To: User B
- Amount: Specified amount
- Status: Pending/Completed/Failed
- Audit Trail: Complete

#### Deposit Redemption
- Deposit ID: Target deposit
- Amount: Redemption amount
- Status: Pending/Completed/Failed
- Audit Trail: Complete

## Sharding Strategies

### 1. Hash-based Sharding
- Uses SHA256 hash of deposit ID
- Most balanced distribution
- Deterministic placement
- Default strategy

### 2. Currency-based Sharding
- Isolates different currencies to specific shards
- Useful for currency-specific regulation
- Better data locality for currency operations

### 3. Issuer-based Sharding
- Separates deposits by issuer organization
- Improves privacy between issuers
- Easier issuer-specific queries

### 4. Time-based Sharding
- Rotates to new shard on schedule
- Useful for archival and performance optimization
- Historical data isolation

### 5. Custom Sharding
- Application-defined sharding key
- Flexible for specific business logic
- Requires custom implementation

## Access Control

### Role Hierarchy

```
┌─────────────┐
│    Admin    │
│             │
│ (Full Ctrl) │
└─────┬───────┘
      │
      ├─────────────────┬──────────────────┬──────────────────┐
      ▼                 ▼                  ▼                  ▼
┌──────────┐      ┌──────────┐      ┌──────────┐      ┌──────────┐
│  Issuer  │      │Depositor │      │ Auditor  │      │  Others  │
│          │      │          │      │          │      │          │
│ Can:     │      │ Can:     │      │ Can:     │      │ Can:     │
│-Issue    │      │-Transfer │      │-Query    │      │-Query    │
│-Manage   │      │-Redeem   │      │-Audit    │      │-Limited  │
│-Query    │      │-Query    │      │-Report   │      │-Access   │
└──────────┘      └──────────┘      └──────────┘      └──────────┘
```

### Channel-level ACLs

- **IssuancePolicy**: Only admins and authorized issuers
- **TransferPolicy**: Owner + admin authorization required
- **RedemptionPolicy**: Owner + issuer authority required
- **QueryPolicy**: Role-based filtering

## Transaction Flow

### Deposit Issuance Flow

```
Client Request
     │
     ▼
Permission Check (Issuer Role)
     │
     ▼
Validate Input (Amount, Currency, Shard)
     │
     ▼
Determine Shard Assignment
     │
     ▼
Create Deposit Object
     │
     ▼
Save to Ledger
     │
     ▼
Create Audit Log Entry
     │
     ▼
Emit Event
     │
     ▼
Return Confirmation
```

### Deposit Transfer Flow

```
Client Request
     │
     ▼
Permission Check (Depositor Role)
     │
     ▼
Validate Deposit Exists
     │
     ▼
Permission Check (Ownership)
     │
     ▼
Compliance Check (Amount, Status, Expiry)
     │
     ▼
Update Holdings
     │
     ▼
Create Transfer Record
     │
     ▼
Update Deposit State
     │
     ▼
Save All Changes
     │
     ▼
Create Audit Log
     │
     ▼
Emit Event
     │
     ▼
Return Confirmation
```

## Performance Characteristics

### Throughput

- **Per Shard**: 1000+ TPS
- **Across 4 Shards**: 4000+ TPS
- **Linear Scaling**: Additional shards increase throughput proportionally

### Latency

- **Average Transaction**: < 500ms
- **Consensus**: < 1s (depends on network configuration)
- **Query Operations**: < 100ms

### Storage

- **Per Deposit**: ~500 bytes (base) + metadata
- **Per Transfer**: ~300 bytes
- **Per Shard**: Scales based on deposit count and history

## State Database

### CouchDB Configuration

- **Database**: Per shard (separate database per channel)
- **Indexing**: Composite indices on (docType, shardId, status, currency)
- **Partitioning**: By shard ID for query optimization

### State Queries

```json
{
  "selector": {
    "docType": "deposit",
    "status": "active",
    "currency": "USD"
  },
  "sort": [{"amount": "desc"}],
  "limit": 100
}
```

## Audit and Compliance

### Audit Trail Structure

```
┌─────────────────────────────────────┐
│       Audit Log Entry               │
├─────────────────────────────────────┤
│ ID: Unique identifier               │
│ DepositID: Target deposit           │
│ Operation: Type of operation        │
│ ActorID: Who performed operation    │
│ CreatedAt: Timestamp                │
│ Details: Operation details          │
│   - From/To users                   │
│   - Amounts                         │
│   - Previous/New state              │
└─────────────────────────────────────┘
```

### Compliance Checks

1. **KYC/AML Integration Points**
   - User identity verification
   - Sanction list checking
   - Anomaly detection

2. **Regulatory Reporting**
   - Transaction reporting
   - Deposit inventory reports
   - Compliance audit trails

3. **Data Retention**
   - Configurable retention policies
   - Historical data archival
   - Compliance holds

## Security Architecture

### Network Security

- **TLS Encryption**: All peer-to-peer communication
- **Certificate-based Auth**: X.509 certificates
- **Channel Isolation**: Separate channels per shard

### Chaincode Security

- **Input Validation**: All inputs validated
- **SQL Injection Prevention**: Parameterized queries
- **Access Control**: RBAC enforcement at every operation
- **Audit Logging**: Complete operation audit trail

### Data Security

- **State Encryption**: CouchDB at-rest encryption
- **Key Management**: HSM integration ready
- **Sensitive Data Masking**: PII/sensitive field handling

## Monitoring and Logging

### Chaincode Logging

```
[TIMESTAMP] [LEVEL] [COMPONENT] Message
[2024-01-15T10:30:45Z] [INFO] [Deposit] Issue deposit DEP-2024-001 from issuer-1
[2024-01-15T10:30:46Z] [INFO] [Transfer] Transfer 100 USD from user-1 to user-2
[2024-01-15T10:30:47Z] [ERROR] [Redemption] Insufficient balance for user-3
```

### Event Monitoring

```
Event: depositIssued
Data: {
  depositId: "DEP-2024-001",
  amount: 1000000,
  currency: "USD",
  shardId: "shard-1"
}

Event: depositTransferred
Data: {
  transferId: "TXF-2024-001",
  depositId: "DEP-2024-001",
  fromUserId: "user-1",
  toUserId: "user-2",
  amount: 100000
}
```

## Integration Points

### External Systems

1. **Banking Systems**: Deposit source verification
2. **KYC/AML Systems**: User compliance checking
3. **Regulatory Reporting**: Compliance data extraction
4. **Audit Systems**: Full audit trail export
5. **Settlement Systems**: Redemption processing

## Disaster Recovery

### Backup Strategy

- Channel state snapshots
- Ledger backup points
- Off-chain audit log storage

### Recovery Procedures

1. Restore from latest backup
2. Replay transactions from audit logs
3. Verify state consistency
4. Resume operations

## Future Enhancements

- Cross-shard atomic transactions
- Advanced sharding strategies (machine learning based)
- Interoperability with other DLT platforms
- Enhanced privacy (zero-knowledge proofs)
- Smart contract upgrades without restart
