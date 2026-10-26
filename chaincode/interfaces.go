package main

// Local interface definitions for testing without external dependencies
// In production, these would be provided by hyperledger/fabric-contract-api-go

// TransactionContextInterface represents the transaction context
type TransactionContextInterface interface {
	GetStub() ChaincodeStubInterface
	GetClientIdentity() ClientIdentityInterface
}

// ChaincodeStubInterface represents the chaincode stub
type ChaincodeStubInterface interface {
	GetState(key string) ([]byte, error)
	PutState(key string, value []byte) error
	DelState(key string) error
	GetQueryResultsForQueryString(queryString string) (ResultsIterator, error)
	CreateCompositeKey(objectType string, attributes []string) (string, error)
	SetEvent(name string, payload []byte) error
}

// ClientIdentityInterface represents the client identity
type ClientIdentityInterface interface {
	GetID() (string, error)
	GetMSPID() (string, error)
	GetAttributeValue(attrName string) (string, error)
	AssertAttributeValue(attrName string, attrValue string) error
	GetX509Certificate() (interface{}, error)
}

// ResultsIterator represents query results
type ResultsIterator interface {
	HasNext() bool
	Next() (*KV, error)
	Close() error
}

// KV represents a key-value pair
type KV struct {
	Key   string
	Value []byte
}

// ChaincodeInterface represents the chaincode
type ChaincodeInterface interface {
	SmartContractInit(ctx TransactionContextInterface) error
}
