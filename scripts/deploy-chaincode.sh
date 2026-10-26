#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Deploying Deposit Chaincode${NC}"
echo -e "${GREEN}========================================${NC}"

# Set environment variables
export PATH=$PATH:$(pwd)/bin
export FABRIC_CFG_PATH=$(pwd)/config

# Package the chaincode
echo -e "${YELLOW}Packaging chaincode...${NC}"
peer lifecycle chaincode package deposit-chaincode.tar.gz \
    --path ./chaincode \
    --lang golang \
    --label deposit-chaincode-1.0 \
    > /dev/null 2>&1

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Chaincode packaged${NC}"
else
    echo -e "${RED}Error: Failed to package chaincode${NC}"
    exit 1
fi

# Install the chaincode on peer0
echo -e "${YELLOW}Installing chaincode on peer0...${NC}"
docker exec cli peer lifecycle chaincode install deposit-chaincode.tar.gz \
    > /dev/null 2>&1

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Chaincode installed on peer0${NC}"
else
    echo -e "${RED}Error: Failed to install on peer0${NC}"
    exit 1
fi

# Install the chaincode on peer1
echo -e "${YELLOW}Installing chaincode on peer1...${NC}"
docker exec -e CORE_PEER_ADDRESS=peer1.org1.example.com:8051 cli peer lifecycle chaincode install deposit-chaincode.tar.gz \
    > /dev/null 2>&1

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Chaincode installed on peer1${NC}"
else
    echo -e "${RED}Error: Failed to install on peer1${NC}"
    exit 1
fi

# Get the package ID
echo -e "${YELLOW}Getting package ID...${NC}"
PACKAGE_ID=$(docker exec cli peer lifecycle chaincode queryinstalled | grep deposit-chaincode | awk '{print $3}' | sed 's/,$//')

if [ -z "$PACKAGE_ID" ]; then
    echo -e "${RED}Error: Could not get package ID${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Package ID: $PACKAGE_ID${NC}"

# Approve for organization
echo -e "${YELLOW}Approving chaincode for Org1...${NC}"
docker exec cli peer lifecycle chaincode approveformyorg \
    --channelID channel-shard-primary \
    --name deposit-chaincode \
    --version 1.0 \
    --package-id $PACKAGE_ID \
    --sequence 1 \
    --tls \
    --cafile /opt/gopath/src/github.com/hyperledger/fabric/peersconfig/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/tls/ca.crt \
    > /dev/null 2>&1

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Chaincode approved for Org1${NC}"
else
    echo -e "${RED}Error: Failed to approve chaincode${NC}"
    exit 1
fi

# Commit the chaincode definition to the channel
echo -e "${YELLOW}Committing chaincode definition...${NC}"
docker exec cli peer lifecycle chaincode commit \
    --channelID channel-shard-primary \
    --name deposit-chaincode \
    --version 1.0 \
    --sequence 1 \
    --tls \
    --cafile /opt/gopath/src/github.com/hyperledger/fabric/peersconfig/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/tls/ca.crt \
    --peerAddresses peer0.org1.example.com:7051 \
    > /dev/null 2>&1

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Chaincode committed${NC}"
else
    echo -e "${RED}Error: Failed to commit chaincode${NC}"
    exit 1
fi

# Deploy to secondary shard
echo -e "${YELLOW}Deploying to secondary shard...${NC}"
docker exec cli peer lifecycle chaincode approveformyorg \
    --channelID channel-shard-secondary \
    --name deposit-chaincode \
    --version 1.0 \
    --package-id $PACKAGE_ID \
    --sequence 1 \
    --tls \
    --cafile /opt/gopath/src/github.com/hyperledger/fabric/peersconfig/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/tls/ca.crt \
    > /dev/null 2>&1

docker exec cli peer lifecycle chaincode commit \
    --channelID channel-shard-secondary \
    --name deposit-chaincode \
    --version 1.0 \
    --sequence 1 \
    --tls \
    --cafile /opt/gopath/src/github.com/hyperledger/fabric/peersconfig/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/tls/ca.crt \
    --peerAddresses peer0.org1.example.com:7051 \
    > /dev/null 2>&1

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Chaincode deployed to secondary shard${NC}"
else
    echo -e "${YELLOW}Note: Secondary shard deployment optional${NC}"
fi

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Chaincode deployment complete!${NC}"
echo -e "${GREEN}========================================${NC}"
