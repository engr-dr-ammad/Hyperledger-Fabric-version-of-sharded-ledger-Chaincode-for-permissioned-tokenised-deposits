package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
)

// Utils provides utility functions for the chaincode
type Utils struct{}

// GenerateID generates a unique ID with a prefix
func (u *Utils) GenerateID(prefix string) string {
	timestamp := time.Now().UnixNano()
	randomNum := rand.Intn(100000)
	return fmt.Sprintf("%s-%d-%d", prefix, timestamp, randomNum)
}

// GenerateDepositID generates a unique deposit ID
func (u *Utils) GenerateDepositID() string {
	return u.GenerateID("DEP")
}

// GenerateTransferID generates a unique transfer ID
func (u *Utils) GenerateTransferID() string {
	return u.GenerateID("TXF")
}

// GenerateRedemptionID generates a unique redemption ID
func (u *Utils) GenerateRedemptionID() string {
	return u.GenerateID("RED")
}

// GenerateAuditLogID generates a unique audit log ID
func (u *Utils) GenerateAuditLogID() string {
	return u.GenerateID("AUD")
}

// MarshalDeposit converts a Deposit to JSON
func (u *Utils) MarshalDeposit(deposit *Deposit) ([]byte, error) {
	return json.Marshal(deposit)
}

// UnmarshalDeposit converts JSON to a Deposit
func (u *Utils) UnmarshalDeposit(data []byte) (*Deposit, error) {
	deposit := &Deposit{}
	err := json.Unmarshal(data, deposit)
	return deposit, err
}

// MarshalShard converts a Shard to JSON
func (u *Utils) MarshalShard(shard *Shard) ([]byte, error) {
	return json.Marshal(shard)
}

// UnmarshalShard converts JSON to a Shard
func (u *Utils) UnmarshalShard(data []byte) (*Shard, error) {
	shard := &Shard{}
	err := json.Unmarshal(data, shard)
	return shard, err
}

// MarshalTransfer converts a DepositTransfer to JSON
func (u *Utils) MarshalTransfer(transfer *DepositTransfer) ([]byte, error) {
	return json.Marshal(transfer)
}

// UnmarshalTransfer converts JSON to a DepositTransfer
func (u *Utils) UnmarshalTransfer(data []byte) (*DepositTransfer, error) {
	transfer := &DepositTransfer{}
	err := json.Unmarshal(data, transfer)
	return transfer, err
}

// MarshalRedemption converts a DepositRedemption to JSON
func (u *Utils) MarshalRedemption(redemption *DepositRedemption) ([]byte, error) {
	return json.Marshal(redemption)
}

// UnmarshalRedemption converts JSON to a DepositRedemption
func (u *Utils) UnmarshalRedemption(data []byte) (*DepositRedemption, error) {
	redemption := &DepositRedemption{}
	err := json.Unmarshal(data, redemption)
	return redemption, err
}

// ValidateDepositID validates deposit ID format
func (u *Utils) ValidateDepositID(depositID string) bool {
	return strings.HasPrefix(depositID, "DEP-") && len(depositID) > 4
}

// ValidateAmount validates that an amount is positive
func (u *Utils) ValidateAmount(amount float64) bool {
	return amount > 0
}

// ValidateCurrency validates currency code (ISO 4217)
func (u *Utils) ValidateCurrency(currency string) bool {
	// Basic validation - should be 3 uppercase letters
	if len(currency) != 3 {
		return false
	}
	for _, ch := range currency {
		if ch < 'A' || ch > 'Z' {
			return false
		}
	}
	return true
}

// ParseAmount parses a string to float64 amount
func (u *Utils) ParseAmount(amountStr string) (float64, error) {
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid amount format: %v", err)
	}
	if !u.ValidateAmount(amount) {
		return 0, fmt.Errorf("amount must be positive")
	}
	return amount, nil
}

// FormatAmount formats a float64 amount as a string with 2 decimal places
func (u *Utils) FormatAmount(amount float64) string {
	return fmt.Sprintf("%.2f", amount)
}

// CompareAmounts compares two amounts with precision
func (u *Utils) CompareAmounts(a, b float64) int {
	epsilon := 0.01
	if a-b > epsilon {
		return 1
	}
	if b-a > epsilon {
		return -1
	}
	return 0
}

// ValidateQuery ensures query string is safe
func (u *Utils) ValidateQuery(queryString string) bool {
	// Prevent injection attacks by checking for suspicious patterns
	suspiciousPatterns := []string{
		"selector",
		"sort",
		"fields",
		"skip",
		"limit",
	}

	lowerQuery := strings.ToLower(queryString)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(lowerQuery, pattern) {
			// These are allowed in structured queries, but not in raw strings
			return false
		}
	}
	return true
}

// GetPagedResultsQueryString creates a paged query string for CouchDB
func (u *Utils) GetPagedResultsQueryString(docType string, bookmark string, pageSize int) string {
	query := fmt.Sprintf(
		`{"selector":{"docType":"%s"},"pageSize":%d`,
		docType,
		pageSize,
	)
	if bookmark != "" {
		query += fmt.Sprintf(`,"bookmark":"%s"`, bookmark)
	}
	query += "}"
	return query
}

// SumHoldings calculates total holding amount for a deposit
func (u *Utils) SumHoldings(holdings []Holding) float64 {
	total := 0.0
	for _, holding := range holdings {
		if holding.Status == "active" {
			total += holding.Amount
		}
	}
	return total
}

// FindHolding finds a holding by user ID
func (u *Utils) FindHolding(holdings []Holding, userID string) *Holding {
	for _, h := range holdings {
		if h.UserID == userID {
			return &h
		}
	}
	return nil
}

// RemoveHolding removes a holding by user ID
func (u *Utils) RemoveHolding(holdings []Holding, userID string) []Holding {
	result := make([]Holding, 0)
	for _, h := range holdings {
		if h.UserID != userID {
			result = append(result, h)
		}
	}
	return result
}

// AddAuditLog creates an audit log entry
func (u *Utils) AddAuditLog(
	depositID string,
	operation string,
	actorID string,
	details MetadataMap,
) AuditLog {
	return AuditLog{
		ID:            u.GenerateAuditLogID(),
		DepositID:     depositID,
		Operation:     operation,
		ActorID:       actorID,
		CreatedAt:     getCurrentTimestamp(),
		Details:       details,
	}
}

// AddTransactionRecord creates a transaction record
func (u *Utils) AddTransactionRecord(
	entityID string,
	entityType string,
	operation string,
	previousState MetadataMap,
	newState MetadataMap,
	actorID string,
	shardID string,
) TransactionRecord {
	return TransactionRecord{
		ID:            u.GenerateID("TXN"),
		EntityID:      entityID,
		EntityType:    entityType,
		Operation:     operation,
		PreviousState: previousState,
		NewState:      newState,
		ActorID:       actorID,
		Timestamp:     getCurrentTimestamp(),
		ShardID:       shardID,
	}
}

// IsJSONValid checks if a string is valid JSON
func (u *Utils) IsJSONValid(jsonStr string) bool {
	var js interface{}
	return json.Unmarshal([]byte(jsonStr), &js) == nil
}

// GetClientPrincipal extracts client principal from context
func (u *Utils) GetClientPrincipal(ctx contractapi.TransactionContextInterface) string {
	clientID := ctx.GetClientIdentity()
	principal, _ := clientID.GetID()
	return principal
}

// RoundTwoDecimals rounds a float to 2 decimal places
func (u *Utils) RoundTwoDecimals(value float64) float64 {
	return float64(int(value*100)) / 100
}

// GetRangeQueryString creates a range query for CouchDB
func (u *Utils) GetRangeQueryString(startKey, endKey string) string {
	return fmt.Sprintf(
		`{"selector":{"key":{"$gte":"%s","$lte":"%s"}}}`,
		startKey,
		endKey,
	)
}

// CalculateChecksum creates a checksum for data verification
func (u *Utils) CalculateChecksum(data string) string {
	// Simple checksum - in production use a proper hash
	checksum := 0
	for _, char := range data {
		checksum += int(char)
	}
	return fmt.Sprintf("%d", checksum)
}

// TimeStringToUnix converts RFC3339 time string to Unix timestamp
func (u *Utils) TimeStringToUnix(timeStr string) (int64, error) {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return 0, err
	}
	return t.Unix(), nil
}

// UnixToTimeString converts Unix timestamp to RFC3339 time string
func (u *Utils) UnixToTimeString(unix int64) string {
	return time.Unix(unix, 0).UTC().Format(time.RFC3339)
}

// IsExpired checks if a timestamp has passed
func (u *Utils) IsExpired(expiryTimeStr string) bool {
	now := getCurrentTimestamp()
	return now > expiryTimeStr
}

// DaysUntilExpiry calculates days until expiry
func (u *Utils) DaysUntilExpiry(expiryTimeStr string) (int, error) {
	expiryTime, err := time.Parse(time.RFC3339, expiryTimeStr)
	if err != nil {
		return 0, err
	}
	now := time.Now().UTC()
	days := int(expiryTime.Sub(now).Hours() / 24)
	return days, nil
}

// CreateEventPayload creates a chaincode event payload
func (u *Utils) CreateEventPayload(eventType string, data interface{}) ([]byte, error) {
	event := map[string]interface{}{
		"type":      eventType,
		"timestamp": getCurrentTimestamp(),
		"data":      data,
	}
	return json.Marshal(event)
}

// BatchOperation represents a batch operation
type BatchOperation struct {
	OperationType string
	Entity        interface{}
	Action        string
}

// ValidateBatchSize checks if batch size is within limits
func (u *Utils) ValidateBatchSize(size int, maxSize int) bool {
	return size > 0 && size <= maxSize
}
