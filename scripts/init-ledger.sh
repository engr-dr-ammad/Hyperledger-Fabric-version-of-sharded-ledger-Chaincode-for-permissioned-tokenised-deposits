#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Initializing Ledger${NC}"
echo -e "${GREEN}========================================${NC}"

# Initialize ledger
echo -e "${YELLOW}Running InitLedger...${NC}"
docker exec cli peer chaincode invoke \
    -C channel-shard-primary \
    -n deposit-chaincode \
    -c '{"function":"InitLedger","Args":[]}' \
    --tls \
    --cafile /opt/gopath/src/github.com/hyperledger/fabric/peersconfig/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/tls/ca.crt \
    > /dev/null 2>&1

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Ledger initialized${NC}"
else
    echo -e "${RED}Error: Failed to initialize ledger${NC}"
    exit 1
fi

# Issue test deposits
echo -e "${YELLOW}Issuing test deposits...${NC}"

docker exec cli peer chaincode invoke \
    -C channel-shard-primary \
    -n deposit-chaincode \
    -c '{
        "function":"IssueDeposit",
        "Args":["DEP-2024-001","1000000","USD","shard-primary","365","{}"]
    }' \
    --tls \
    --cafile /opt/gopath/src/github.com/hyperledger/fabric/peersconfig/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/tls/ca.crt \
    > /dev/null 2>&1

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Test deposit 1 issued (DEP-2024-001: 1,000,000 USD)${NC}"
else
    echo -e "${RED}Error: Failed to issue test deposit 1${NC}"
fi

docker exec cli peer chaincode invoke \
    -C channel-shard-primary \
    -n deposit-chaincode \
    -c '{
        "function":"IssueDeposit",
        "Args":["DEP-2024-002","5000000","EUR","shard-primary","365","{}"]
    }' \
    --tls \
    --cafile /opt/gopath/src/github.com/hyperledger/fabric/peersconfig/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/tls/ca.crt \
    > /dev/null 2>&1

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Test deposit 2 issued (DEP-2024-002: 5,000,000 EUR)${NC}"
else
    echo -e "${RED}Error: Failed to issue test deposit 2${NC}"
fi

docker exec cli peer chaincode invoke \
    -C channel-shard-primary \
    -n deposit-chaincode \
    -c '{
        "function":"IssueDeposit",
        "Args":["DEP-2024-003","250000000","GBP","shard-primary","730","{}"]
    }' \
    --tls \
    --cafile /opt/gopath/src/github.com/hyperledger/fabric/peersconfig/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/tls/ca.crt \
    > /dev/null 2>&1

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Test deposit 3 issued (DEP-2024-003: 250,000,000 GBP)${NC}"
else
    echo -e "${RED}Error: Failed to issue test deposit 3${NC}"
fi

# Verify test deposits
echo -e "${YELLOW}Verifying test deposits...${NC}"

echo -e "${YELLOW}Deposit 1:${NC}"
docker exec cli peer chaincode query \
    -C channel-shard-primary \
    -n deposit-chaincode \
    -c '{"function":"GetDeposit","Args":["DEP-2024-001"]}' \
    --tls \
    --cafile /opt/gopath/src/github.com/hyperledger/fabric/peersconfig/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/tls/ca.crt \
    2>/dev/null | jq -r '.depositId, .amount, .currency' | head -3

echo -e "${YELLOW}Deposit 2:${NC}"
docker exec cli peer chaincode query \
    -C channel-shard-primary \
    -n deposit-chaincode \
    -c '{"function":"GetDeposit","Args":["DEP-2024-002"]}' \
    --tls \
    --cafile /opt/gopath/src/github.com/hyperledger/fabric/peersconfig/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/tls/ca.crt \
    2>/dev/null | jq -r '.depositId, .amount, .currency' | head -3

echo -e "${YELLOW}Deposit 3:${NC}"
docker exec cli peer chaincode query \
    -C channel-shard-primary \
    -n deposit-chaincode \
    -c '{"function":"GetDeposit","Args":["DEP-2024-003"]}' \
    --tls \
    --cafile /opt/gopath/src/github.com/hyperledger/fabric/peersconfig/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/tls/ca.crt \
    2>/dev/null | jq -r '.depositId, .amount, .currency' | head -3

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Ledger initialization complete!${NC}"
echo -e "${GREEN}========================================${NC}"
echo -e "${YELLOW}Test data summary:${NC}"
echo "- DEP-2024-001: 1,000,000 USD (primary shard)"
echo "- DEP-2024-002: 5,000,000 EUR (primary shard)"
echo "- DEP-2024-003: 250,000,000 GBP (primary shard)"
