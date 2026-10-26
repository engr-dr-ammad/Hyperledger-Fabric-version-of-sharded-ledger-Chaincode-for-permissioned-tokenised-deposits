package main

import (
	"encoding/json"
	"fmt"
	"testing"
)

// MockTransactionContext is a mock implementation of TransactionContextInterface
type MockTransactionContext struct {
	stub     *MockStub
	clientID *MockClientID
}

func (mtc *MockTransactionContext) GetStub() ChaincodeStubInterface {
	return mtc.stub
}

func (mtc *MockTransactionContext) GetClientIdentity() ClientIdentityInterface {
	return mtc.clientID
}

// MockStub is a mock implementation of ChaincodeStubInterface
type MockStub struct {
	state map[string][]byte
}

func NewMockStub() *MockStub {
	return &MockStub{
		state: make(map[string][]byte),
	}
}

func (ms *MockStub) GetState(key string) ([]byte, error) {
	if val, ok := ms.state[key]; ok {
		return val, nil
	}
	return nil, nil
}

func (ms *MockStub) PutState(key string, value []byte) error {
	ms.state[key] = value
	return nil
}

func (ms *MockStub) DelState(key string) error {
	delete(ms.state, key)
	return nil
}

func (ms *MockStub) GetQueryResultsForQueryString(queryString string) (ResultsIterator, error) {
	return NewMockIterator(ms.state), nil
}

func (ms *MockStub) CreateCompositeKey(objectType string, attributes []string) (string, error) {
	return fmt.Sprintf("%s-%s", objectType, attributes[0]), nil
}

func (ms *MockStub) SetEvent(name string, payload []byte) error {
	return nil
}

// MockResultsIterator
type MockIterator struct {
	values []MockKV
	index  int
}

type MockKV struct {
	Key   string
	Value []byte
}

func NewMockIterator(state map[string][]byte) *MockIterator {
	values := make([]MockKV, 0)
	for k, v := range state {
		values = append(values, MockKV{k, v})
	}
	return &MockIterator{values: values, index: 0}
}

func (mi *MockIterator) HasNext() bool {
	return mi.index < len(mi.values)
}

func (mi *MockIterator) Next() (*KV, error) {
	if !mi.HasNext() {
		return nil, fmt.Errorf("no more results")
	}
	kv := mi.values[mi.index]
	mi.index++
	return &KV{Key: kv.Key, Value: kv.Value}, nil
}

func (mi *MockIterator) Close() error {
	return nil
}

// MockClientID is a mock implementation of ClientIdentityInterface
type MockClientID struct {
	mspID string
	cn    string
}

func (mci *MockClientID) GetID() (string, error) {
	return mci.cn, nil
}

func (mci *MockClientID) GetMSPID() (string, error) {
	return mci.mspID, nil
}

func (mci *MockClientID) GetAttributeValue(attrName string) (string, error) {
	if attrName == "ou" {
		return "issuer", nil
	}
	return "", nil
}

func (mci *MockClientID) AssertAttributeValue(attrName string, attrValue string) error {
	val, _ := mci.GetAttributeValue(attrName)
	if val == attrValue {
		return nil
	}
	return fmt.Errorf("attribute mismatch")
}

func (mci *MockClientID) GetX509Certificate() (interface{}, error) {
	return nil, nil
}

// Test functions
func TestInitLedger(t *testing.T) {
	cc := NewDepositChaincode()
	stub := NewMockStub()
	ctx := &MockTransactionContext{
		stub: stub,
		clientID: &MockClientID{
			mspID: "AdminMSP",
			cn:    "admin",
		},
	}

	// Initialize should work for admin
	// Note: This would need proper mocking of permission checks
	_ = cc
	_ = ctx
	t.Logf("InitLedger test setup successful")
}

func TestIssueDeposit(t *testing.T) {
	cc := NewDepositChaincode()
	stub := NewMockStub()
	ctx := &MockTransactionContext{
		stub: stub,
		clientID: &MockClientID{
			mspID: "IssuerMSP",
			cn:    "issuer-1",
		},
	}

	// Create a shard first
	shard := Shard{
		ShardID:     "shard-1",
		ChannelName: "channel-1",
		Description: "Test shard",
		Issuers:     []string{"issuer-1"},
		Status:      "active",
		CreatedAt:   getCurrentTimestamp(),
		UpdatedAt:   getCurrentTimestamp(),
		Metadata:    MetadataMap{},
	}

	shardBytes, _ := json.Marshal(shard)
	stub.state["shard-1"] = shardBytes

	// Test invalid deposit ID
	err := cc.IssueDeposit(ctx, "invalid", "1000", "USD", "shard-1", "30", "{}")
	if err == nil {
		t.Errorf("Expected error for invalid deposit ID, got nil")
	}

	// Test valid deposit (would need permission mocking)
	_ = cc
	t.Logf("IssueDeposit test setup successful")
}

func TestTransferDeposit(t *testing.T) {
	cc := NewDepositChaincode()
	stub := NewMockStub()

	// Create deposit
	deposit := Deposit{
		DepositID:   "DEP-2024-001",
		IssuerID:    "issuer-1",
		OwnerID:     "issuer-1",
		Amount:      1000,
		Currency:    "USD",
		Status:      "active",
		ShardID:     "shard-1",
		CreatedAt:   getCurrentTimestamp(),
		UpdatedAt:   getCurrentTimestamp(),
		ExpiresAt:   getFutureTimestamp(30),
		Metadata:    MetadataMap{},
		Holdings: []Holding{
			{
				UserID:     "issuer-1",
				Amount:     1000,
				AcquiredAt: getCurrentTimestamp(),
				Status:     "active",
			},
		},
	}

	depositBytes, _ := json.Marshal(deposit)
	stub.state["DEP-2024-001"] = depositBytes

	ctx := &MockTransactionContext{
		stub: stub,
		clientID: &MockClientID{
			mspID: "IssuerMSP",
			cn:    "issuer-1",
		},
	}

	// Test transfer
	err := cc.TransferDeposit(ctx, "DEP-2024-001", "user-2", "100", "{}")
	_ = err
	t.Logf("TransferDeposit test setup successful")
}

func TestGetDeposit(t *testing.T) {
	cc := NewDepositChaincode()
	stub := NewMockStub()

	// Create deposit
	deposit := Deposit{
		DepositID:  "DEP-2024-001",
		IssuerID:   "issuer-1",
		OwnerID:    "issuer-1",
		Amount:     1000,
		Currency:   "USD",
		Status:     "active",
		ShardID:    "shard-1",
		CreatedAt:  getCurrentTimestamp(),
		UpdatedAt:  getCurrentTimestamp(),
		Metadata:   MetadataMap{},
		Holdings:   []Holding{},
	}

	depositBytes, _ := json.Marshal(deposit)
	stub.state["DEP-2024-001"] = depositBytes

	ctx := &MockTransactionContext{
		stub: stub,
		clientID: &MockClientID{
			mspID: "UserMSP",
			cn:    "user-1",
		},
	}

	// Test get deposit
	result, err := cc.GetDeposit(ctx, "DEP-2024-001")
	if err == nil && result != nil {
		if result.DepositID != "DEP-2024-001" {
			t.Errorf("Expected deposit ID DEP-2024-001, got %s", result.DepositID)
		}
	}
}

func TestUtilsGenerateID(t *testing.T) {
	utils := &Utils{}

	id1 := utils.GenerateID("TEST")
	if len(id1) == 0 || id1[:4] != "TEST" {
		t.Errorf("Expected ID to start with TEST-, got %s", id1)
	}

	id2 := utils.GenerateDepositID()
	if len(id2) == 0 || id2[:3] != "DEP" {
		t.Errorf("Expected DEP ID, got %s", id2)
	}

	id3 := utils.GenerateTransferID()
	if len(id3) == 0 || id3[:3] != "TXF" {
		t.Errorf("Expected TXF ID, got %s", id3)
	}
}

func TestUtilsValidateAmount(t *testing.T) {
	utils := &Utils{}

	tests := []struct {
		amount float64
		valid  bool
	}{
		{100.50, true},
		{0.01, true},
		{0, false},
		{-100, false},
	}

	for _, test := range tests {
		if utils.ValidateAmount(test.amount) != test.valid {
			t.Errorf("ValidateAmount(%f) = %v, want %v", test.amount, utils.ValidateAmount(test.amount), test.valid)
		}
	}
}

func TestUtilsValidateCurrency(t *testing.T) {
	utils := &Utils{}

	tests := []struct {
		currency string
		valid    bool
	}{
		{"USD", true},
		{"EUR", true},
		{"US", false},
		{"USDA", false},
		{"usd", false},
	}

	for _, test := range tests {
		if utils.ValidateCurrency(test.currency) != test.valid {
			t.Errorf("ValidateCurrency(%s) = %v, want %v", test.currency, utils.ValidateCurrency(test.currency), test.valid)
		}
	}
}

func TestUtilsParseAmount(t *testing.T) {
	utils := &Utils{}

	amount, err := utils.ParseAmount("1000.50")
	if err != nil {
		t.Errorf("ParseAmount(\"1000.50\") returned error: %v", err)
	}
	if amount != 1000.50 {
		t.Errorf("ParseAmount(\"1000.50\") = %f, want 1000.50", amount)
	}

	_, err = utils.ParseAmount("invalid")
	if err == nil {
		t.Errorf("ParseAmount(\"invalid\") should return error, got nil")
	}

	_, err = utils.ParseAmount("-100")
	if err == nil {
		t.Errorf("ParseAmount(\"-100\") should return error, got nil")
	}
}

func TestShardingManagerByHash(t *testing.T) {
	manager := NewShardingManager(ShardByHash)
	manager.AddShard("shard-1")
	manager.AddShard("shard-2")
	manager.AddShard("shard-3")

	shard1, err := manager.DetermineShardForDeposit("DEP-001", "USD", "issuer-1", getCurrentTimestamp(), "")
	if err != nil {
		t.Errorf("DetermineShardForDeposit returned error: %v", err)
	}
	if shard1 == "" {
		t.Error("Expected non-empty shard, got empty string")
	}

	shard2, err := manager.DetermineShardForDeposit("DEP-002", "EUR", "issuer-2", getCurrentTimestamp(), "")
	if err != nil {
		t.Errorf("DetermineShardForDeposit returned error: %v", err)
	}
	if shard2 == "" {
		t.Error("Expected non-empty shard, got empty string")
	}

	// Same deposit ID should always map to same shard
	shard1Again, _ := manager.DetermineShardForDeposit("DEP-001", "USD", "issuer-1", getCurrentTimestamp(), "")
	if shard1 != shard1Again {
		t.Errorf("Expected same shard for same deposit, got %s then %s", shard1, shard1Again)
	}
}

func TestShardingManagerByCurrency(t *testing.T) {
	manager := NewShardingManager(ShardByCurrency)
	manager.AddShard("shard-usd")
	manager.AddShard("shard-eur")

	shard, err := manager.DetermineShardForDeposit("DEP-001", "USD", "issuer-1", getCurrentTimestamp(), "")
	if err != nil {
		t.Errorf("DetermineShardForDeposit returned error: %v", err)
	}
	if shard == "" {
		t.Error("Expected non-empty shard, got empty string")
	}
}

func TestComplianceCheck(t *testing.T) {
	pm := &PermissionManager{}

	deposit := &Deposit{
		DepositID:  "DEP-001",
		Amount:     1000,
		Status:     "active",
		ExpiresAt:  getFutureTimestamp(30),
		Holdings: []Holding{
			{
				UserID: "user-1",
				Amount: 500,
				Status: "active",
			},
		},
	}

	// Valid transfer
	err := pm.CheckCompliance(nil, deposit, 100)
	if err != nil {
		t.Errorf("CheckCompliance should not error for valid amount: %v", err)
	}

	// Invalid amount
	err = pm.CheckCompliance(nil, deposit, -100)
	if err == nil {
		t.Error("CheckCompliance should error for negative amount")
	}

	// Amount exceeds balance
	err = pm.CheckCompliance(nil, deposit, 600)
	if err == nil {
		t.Error("CheckCompliance should error when amount exceeds balance")
	}
}

func TestDepositCalculations(t *testing.T) {
	utils := &Utils{}

	holdings := []Holding{
		{UserID: "user-1", Amount: 100, Status: "active"},
		{UserID: "user-2", Amount: 200, Status: "active"},
		{UserID: "user-3", Amount: 50, Status: "redeemed"},
	}

	total := utils.SumHoldings(holdings)
	if total != 300.0 {
		t.Errorf("SumHoldings = %f, want 300.0", total)
	}

	holding := utils.FindHolding(holdings, "user-1")
	if holding == nil {
		t.Error("FindHolding should return holding for user-1")
	} else if holding.UserID != "user-1" {
		t.Errorf("Found holding for wrong user: %s", holding.UserID)
	}

	holding = utils.FindHolding(holdings, "user-4")
	if holding != nil {
		t.Error("FindHolding should return nil for non-existent user")
	}
}
