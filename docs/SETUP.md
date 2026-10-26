# Setup and Deployment Guide

## Prerequisites

### System Requirements
- Linux or macOS (Windows with WSL2)
- 8GB RAM minimum
- 50GB disk space
- Docker 20.10+
- Docker Compose 1.29+
- Go 1.20+

### Software Dependencies
- Hyperledger Fabric 2.5.0+
- CouchDB 3.0+
- Node.js 16+ (for client applications)

### Network Setup
- Firewall configured to allow inter-container communication
- Port ranges: 7050-7060, 5984 (CouchDB) available

## Installation Steps

### 1. Clone the Repository

```bash
git clone https://github.com/engr-dr-ammad/hyperledger-fabric-sharded-ledger.git
cd hyperledger-fabric-sharded-ledger
```

### 2. Install Hyperledger Fabric Binaries

```bash
# Download Fabric binaries
curl -sSL https://bit.ly/2ysbOFE | bash -s -- 2.5.0

# This creates a bin directory with fabric-ca-client, fabric-ca-server, etc.
export PATH=$PATH:$(pwd)/bin
```

### 3. Build the Chaincode

```bash
cd chaincode
go mod download
go build -o deposit-chaincode
cd ..
```

### 4. Start the Network

```bash
# Using docker-compose
docker-compose up -d

# Wait for all containers to be ready (about 30 seconds)
sleep 30

# Check container status
docker-compose ps
```

### 5. Create Channels

```bash
cd scripts
./setup-network.sh
```

### 6. Deploy the Chaincode

```bash
./deploy-chaincode.sh
```

### 7. Initialize the Ledger

```bash
./init-ledger.sh
```

## Configuration

### Environment Variables

Create a `.env` file in the root directory:

```env
# Network Configuration
FABRIC_PEER_HOST=localhost
FABRIC_PEER_PORT=7051
FABRIC_ORDERER_HOST=localhost
FABRIC_ORDERER_PORT=7050

# Database Configuration
COUCHDB_URL=http://localhost:5984
COUCHDB_USER=admin
COUCHDB_PASS=password

# TLS Configuration
TLS_ENABLED=true
TLS_CERT_PATH=/path/to/cert.pem
TLS_KEY_PATH=/path/to/key.pem

# Logging
LOG_LEVEL=INFO
```

### Fabric Configuration Files

#### configtx.yaml
Update with your organization details:
```yaml
Organizations:
  - &OrdererOrg
    Name: OrdererOrg
    ID: OrdererMSP
    MSPDir: crypto-config/ordererOrganizations/example.com/msp
    
  - &Org1
    Name: Org1
    ID: Org1MSP
    MSPDir: crypto-config/peerOrganizations/org1.example.com/msp
    AdminPrincipal: Role.ADMIN
    Policies:
      Readers:
        Type: Signature
        Rule: "OR('Org1MSP.admin', 'Org1MSP.peer', 'Org1MSP.client')"
      Writers:
        Type: Signature
        Rule: "OR('Org1MSP.admin', 'Org1MSP.client')"
      Admins:
        Type: Signature
        Rule: "OR('Org1MSP.admin')"
```

#### crypto-config.yaml
Generate cryptographic materials:
```yaml
OrdererOrgs:
  - Name: Orderer
    Domain: example.com
    Specs:
      - Hostname: orderer

PeerOrgs:
  - Name: Org1
    Domain: org1.example.com
    Template:
      Count: 2
    Users:
      Count: 1
```

## Running the Network

### Start Services

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f peer0.org1

# Stop services
docker-compose down

# Clean up volumes
docker-compose down -v
```

### Verify Installation

```bash
# Check all containers are running
docker-compose ps

# Expected output:
# peer0.org1.example.com    Up
# orderer.example.com       Up
# couchdb0                  Up
```

## Invoking Chaincode Functions

### Set Environment Variables

```bash
export FABRIC_CFG_PATH=$(pwd)/config
export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID=Org1MSP
export CORE_PEER_TLS_ROOTCERT_FILE=$(pwd)/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=$(pwd)/crypto-config/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=localhost:7051
export ORDERER_CA=$(pwd)/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/tls/ca.crt
```

### Initialize Ledger

```bash
peer chaincode invoke -C channel-shard-primary \
  -n deposit-chaincode \
  -c '{"function":"InitLedger","Args":[]}' \
  --tls --cafile $ORDERER_CA
```

### Issue a Deposit

```bash
peer chaincode invoke -C channel-shard-primary \
  -n deposit-chaincode \
  -c '{
    "function":"IssueDeposit",
    "Args":[
      "DEP-2024-001",
      "1000000",
      "USD",
      "shard-primary",
      "365",
      "{}"
    ]
  }' \
  --tls --cafile $ORDERER_CA
```

### Query a Deposit

```bash
peer chaincode query -C channel-shard-primary \
  -n deposit-chaincode \
  -c '{"function":"GetDeposit","Args":["DEP-2024-001"]}' \
  --tls --cafile $ORDERER_CA
```

### Transfer a Deposit

```bash
peer chaincode invoke -C channel-shard-primary \
  -n deposit-chaincode \
  -c '{
    "function":"TransferDeposit",
    "Args":[
      "DEP-2024-001",
      "user-2",
      "100000",
      "{\"reason\":\"Payment\"}"
    ]
  }' \
  --tls --cafile $ORDERER_CA
```

## Testing

### Run Unit Tests

```bash
cd chaincode
go test -v

# With coverage
go test -v -cover
```

### Integration Testing

```bash
# Create test script
cat > test-integration.sh << 'EOF'
#!/bin/bash

# Set environment variables
export FABRIC_CFG_PATH=$(pwd)/config
export CORE_PEER_TLS_ENABLED=true
# ... more env vars ...

# Initialize ledger
peer chaincode invoke -C channel-shard-primary \
  -n deposit-chaincode \
  -c '{"function":"InitLedger","Args":[]}' \
  --tls --cafile $ORDERER_CA

# Issue deposit
peer chaincode invoke -C channel-shard-primary \
  -n deposit-chaincode \
  -c '{"function":"IssueDeposit","Args":["DEP-2024-001","1000000","USD","shard-primary","365","{}"]}' \
  --tls --cafile $ORDERER_CA

# Verify deposit created
peer chaincode query -C channel-shard-primary \
  -n deposit-chaincode \
  -c '{"function":"GetDeposit","Args":["DEP-2024-001"]}' \
  --tls --cafile $ORDERER_CA

# Transfer deposit
peer chaincode invoke -C channel-shard-primary \
  -n deposit-chaincode \
  -c '{"function":"TransferDeposit","Args":["DEP-2024-001","user-2","100000","{}"]}' \
  --tls --cafile $ORDERER_CA

echo "All tests passed!"
EOF

chmod +x test-integration.sh
./test-integration.sh
```

## Monitoring

### View Logs

```bash
# Peer logs
docker logs peer0.org1.example.com -f

# Orderer logs
docker logs orderer.example.com -f

# CouchDB logs
docker logs couchdb0 -f
```

### Monitor Chaincode

```bash
# Get chaincode container logs
docker logs dev-peer0.org1.example.com-deposit-chaincode-1.0-f5e0de5e0ec8f1a4c8d5b9e7c4a1f8d2 -f
```

### Database Monitoring

```bash
# Access CouchDB Fauxton interface
# Open browser: http://localhost:5984/_utils

# Query database stats
curl http://localhost:5984/channel-shard-primary

# List all documents
curl http://localhost:5984/channel-shard-primary/_all_docs
```

## Troubleshooting

### Issue: Container fails to start

```bash
# Check container logs
docker-compose logs {container_name}

# Rebuild images
docker-compose build --no-cache

# Restart
docker-compose down
docker-compose up -d
```

### Issue: Chaincode deployment fails

```bash
# Check chaincode package
peer lifecycle chaincode queryinstalled \
  --peerAddresses localhost:7051 \
  --tlsRootCertFiles $CORE_PEER_TLS_ROOTCERT_FILE

# Approve chaincode
peer lifecycle chaincode approveformyorg \
  --channelID channel-shard-primary \
  --name deposit-chaincode \
  --version 1.0 \
  --package-id {PACKAGE_ID}
```

### Issue: TLS certificate errors

```bash
# Verify TLS certs exist
ls -la crypto-config/peerOrganizations/*/peers/*/tls/

# Regenerate certificates
cryptogen generate --config=crypto-config.yaml --output=crypto-config
```

### Issue: CouchDB connection refused

```bash
# Check CouchDB is running
docker-compose ps couchdb0

# Check CouchDB logs
docker logs couchdb0

# Wait for CouchDB to be ready
sleep 10
docker-compose up -d
```

## Performance Tuning

### CouchDB Configuration

Edit docker-compose.yml:
```yaml
environment:
  # Increase memory allocation
  COUCHDB_MEMORY: "2048"
  # Reduce rebalancing delay
  COUCHDB_COMPACTION_THRESHOLD: "50"
  # Index optimization
  COUCHDB_BATCH_SIZE: "500"
```

### Peer Configuration

Edit config/core.yaml:
```yaml
peer:
  chaincode:
    executetimeout: "30s"
  ledger:
    state:
      couchDBConfig:
        maxRetries: 3
        maxRetriesOnEmptyCommit: 3
        requestTimeout: 35s
```

### Increase Shard Count

```bash
# Create additional shards
peer chaincode invoke -C channel-shard-1 \
  -n deposit-chaincode \
  -c '{
    "function":"CreateShard",
    "Args":["shard-2","channel-shard-2","Secondary shard"]
  }' \
  --tls --cafile $ORDERER_CA
```

## Security Hardening

### Enable Additional TLS Verification

```bash
export CORE_PEER_TLS_CLIENTAUTHREQUIRED=true
export CORE_PEER_TLS_CLIENTCERT_FILE=$(pwd)/crypto-config/peerOrganizations/org1.example.com/users/User1@org1.example.com/tls/client.crt
export CORE_PEER_TLS_CLIENTKEY_FILE=$(pwd)/crypto-config/peerOrganizations/org1.example.com/users/User1@org1.example.com/tls/client.key
```

### Restrict Channel Access

```bash
# Set channel access policy in configtx.yaml
Policies:
  Readers:
    Type: Signature
    Rule: "OR('Org1MSP.member')"
  Writers:
    Type: Signature
    Rule: "OR('Org1MSP.member')"
  Admins:
    Type: Signature
    Rule: "OR('Org1MSP.admin')"
```

## Upgrade Chaincode

```bash
# Build new version
cd chaincode
go build -o deposit-chaincode

# Package new version
peer lifecycle chaincode package deposit-chaincode-v1.1.tar.gz \
  --path ./ \
  --label deposit-chaincode-v1.1

# Install on all peers
peer lifecycle chaincode install deposit-chaincode-v1.1.tar.gz

# Approve new version
peer lifecycle chaincode approveformyorg \
  --channelID channel-shard-primary \
  --name deposit-chaincode \
  --version 1.1 \
  --package-id {NEW_PACKAGE_ID}

# Check readiness
peer lifecycle chaincode checkcommitreadiness \
  --channelID channel-shard-primary \
  --name deposit-chaincode \
  --version 1.1

# Commit upgrade
peer lifecycle chaincode commit \
  --channelID channel-shard-primary \
  --name deposit-chaincode \
  --version 1.1
```

## Backup and Recovery

### Backup State

```bash
# Backup CouchDB
docker exec couchdb0 curl -X GET http://localhost:5984/_all_dbs

# Export channel state
peer channel fetch oldest blockchain.block -c channel-shard-primary

# Backup crypto materials
tar -czf crypto-backup.tar.gz crypto-config/
```

### Restore State

```bash
# Restore from backup
tar -xzf crypto-backup.tar.gz

# Restart network
docker-compose down -v
docker-compose up -d

# Reinitialize ledger
./scripts/init-ledger.sh
```

## Production Deployment

See [PRODUCTION.md](./PRODUCTION.md) for production deployment guidelines.
