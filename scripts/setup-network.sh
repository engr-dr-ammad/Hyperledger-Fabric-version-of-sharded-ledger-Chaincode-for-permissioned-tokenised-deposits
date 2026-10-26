#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Setting up Hyperledger Fabric Network${NC}"
echo -e "${GREEN}========================================${NC}"

# Check prerequisites
echo -e "${YELLOW}Checking prerequisites...${NC}"

if ! command -v docker &> /dev/null; then
    echo -e "${RED}Error: Docker is not installed${NC}"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo -e "${RED}Error: Docker Compose is not installed${NC}"
    exit 1
fi

if ! command -v cryptogen &> /dev/null; then
    echo -e "${YELLOW}Installing cryptogen...${NC}"
    mkdir -p bin
    cd bin
    curl -sSL https://github.com/hyperledger/fabric/releases/download/v2.5.0/hyperledger-fabric-linux-amd64-2.5.0.tar.gz | tar xz
    cd ..
    export PATH=$PATH:$(pwd)/bin
fi

echo -e "${GREEN}✓ Prerequisites satisfied${NC}"

# Create directories
echo -e "${YELLOW}Creating directory structure...${NC}"
mkdir -p channel-artifacts crypto-config

echo -e "${GREEN}✓ Directories created${NC}"

# Generate cryptographic materials
echo -e "${YELLOW}Generating certificates...${NC}"
if [ ! -d "crypto-config/ordererOrganizations" ]; then
    cryptogen generate --config=crypto-config.yaml --output=crypto-config
    echo -e "${GREEN}✓ Certificates generated${NC}"
else
    echo -e "${YELLOW}Using existing certificates${NC}"
fi

# Check if configtxgen exists
if ! command -v configtxgen &> /dev/null; then
    echo -e "${YELLOW}Installing configtxgen...${NC}"
    if [ ! -f "bin/configtxgen" ]; then
        mkdir -p bin
        cd bin
        curl -sSL https://github.com/hyperledger/fabric/releases/download/v2.5.0/hyperledger-fabric-linux-amd64-2.5.0.tar.gz | tar xz
        cd ..
    fi
    export PATH=$PATH:$(pwd)/bin
fi

# Generate genesis block
echo -e "${YELLOW}Generating genesis block...${NC}"
if [ ! -f "channel-artifacts/orderer.genesis.block" ]; then
    configtxgen -profile OneOrgOrdererGenesis -channelID system-channel -outputBlock channel-artifacts/orderer.genesis.block -configPath .
    echo -e "${GREEN}✓ Genesis block created${NC}"
else
    echo -e "${YELLOW}Genesis block already exists${NC}"
fi

# Create channels
echo -e "${YELLOW}Creating channels...${NC}"
CHANNELS=("channel-shard-primary" "channel-shard-secondary")

for channel in "${CHANNELS[@]}"; do
    if [ ! -f "channel-artifacts/${channel}.tx" ]; then
        configtxgen -profile OneOrgChannel -channelID $channel -outputCreateChannelTx channel-artifacts/${channel}.tx -configPath .
        echo -e "${GREEN}✓ Channel transaction created: $channel${NC}"
    else
        echo -e "${YELLOW}Channel already exists: $channel${NC}"
    fi
done

# Start network
echo -e "${YELLOW}Starting Docker network...${NC}"
docker-compose down -v 2>/dev/null || true
docker-compose up -d

# Wait for containers to be healthy
echo -e "${YELLOW}Waiting for containers to start...${NC}"
sleep 10

# Check if containers are running
echo -e "${YELLOW}Checking container status...${NC}"
docker-compose ps

# Create channels on the peer
echo -e "${YELLOW}Creating channels...${NC}"
for channel in "${CHANNELS[@]}"; do
    echo -e "${YELLOW}Creating channel: $channel${NC}"

    docker exec cli peer channel create \
        -o orderer.example.com:7050 \
        -c $channel \
        -f /opt/gopath/src/github.com/hyperledger/fabric/peersconfig/channel-artifacts/${channel}.tx \
        --tls \
        --cafile /opt/gopath/src/github.com/hyperledger/fabric/peersconfig/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/tls/ca.crt \
        2>/dev/null || true

    echo -e "${GREEN}✓ Channel created: $channel${NC}"
done

# Join peers to channels
echo -e "${YELLOW}Joining peers to channels...${NC}"
PEERS=("peer0.org1.example.com:7051" "peer1.org1.example.com:8051")

for channel in "${CHANNELS[@]}"; do
    for peer in "${PEERS[@]}"; do
        PEER_ADDR=$(echo $peer | cut -d: -f1)
        echo -e "${YELLOW}Joining $PEER_ADDR to $channel...${NC}"

        if [ "$PEER_ADDR" = "peer1.org1.example.com" ]; then
            docker exec -e CORE_PEER_ADDRESS=peer1.org1.example.com:8051 cli peer channel join \
                -b /opt/gopath/src/github.com/hyperledger/fabric/peersconfig/${channel}.block \
                --tls \
                --cafile /opt/gopath/src/github.com/hyperledger/fabric/peersconfig/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/tls/ca.crt \
                2>/dev/null || true
        else
            docker exec cli peer channel join \
                -b /opt/gopath/src/github.com/hyperledger/fabric/peersconfig/${channel}.block \
                --tls \
                --cafile /opt/gopath/src/github.com/hyperledger/fabric/peersconfig/crypto-config/ordererOrganizations/example.com/orderers/orderer.example.com/tls/ca.crt \
                2>/dev/null || true
        fi

        echo -e "${GREEN}✓ $PEER_ADDR joined $channel${NC}"
    done
done

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Network setup complete!${NC}"
echo -e "${GREEN}========================================${NC}"
echo -e "${YELLOW}Next steps:${NC}"
echo "1. Run ./scripts/deploy-chaincode.sh to deploy the chaincode"
echo "2. Run ./scripts/init-ledger.sh to initialize the ledger"
