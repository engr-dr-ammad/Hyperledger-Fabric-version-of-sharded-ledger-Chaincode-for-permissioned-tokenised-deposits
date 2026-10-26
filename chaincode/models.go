package main

import (
	"time"
)

// Deposit represents a tokenised deposit
type Deposit struct {
	DepositID    string      `json:"depositId"`
	IssuerID     string      `json:"issuerId"`
	OwnerID      string      `json:"ownerId"`
	Amount       float64     `json:"amount"`
	Currency     string      `json:"currency"`
	Status       string      `json:"status"` // active, redeemed, suspended
	ShardID      string      `json:"shardId"`
	CreatedAt    string      `json:"createdAt"`
	UpdatedAt    string      `json:"updatedAt"`
	ExpiresAt    string      `json:"expiresAt,omitempty"`
	Metadata     MetadataMap `json:"metadata"`
	Holdings     []Holding   `json:"holdings"`
}

// Holding represents a deposit holding by a specific user
type Holding struct {
	UserID    string    `json:"userId"`
	Amount    float64   `json:"amount"`
	AcquiredAt string   `json:"acquiredAt"`
	Status    string    `json:"status"`
}

// Shard represents a channel/shard in the network
type Shard struct {
	ShardID     string    `json:"shardId"`
	ChannelName string    `json:"channelName"`
	Description string    `json:"description"`
	Issuers     []string  `json:"issuers"`
	Status      string    `json:"status"` // active, inactive
	CreatedAt   string    `json:"createdAt"`
	UpdatedAt   string    `json:"updatedAt"`
	Metadata    MetadataMap `json:"metadata"`
}

// DepositTransfer represents a transfer transaction
type DepositTransfer struct {
	TransferID   string    `json:"transferId"`
	DepositID    string    `json:"depositId"`
	FromUserID   string    `json:"fromUserId"`
	ToUserID     string    `json:"toUserId"`
	Amount       float64   `json:"amount"`
	Status       string    `json:"status"` // pending, completed, failed
	CreatedAt    string    `json:"createdAt"`
	CompletedAt  string    `json:"completedAt,omitempty"`
	Reason       string    `json:"reason,omitempty"`
	Metadata     MetadataMap `json:"metadata"`
}

// DepositRedemption represents a redemption transaction
type DepositRedemption struct {
	RedemptionID string    `json:"redemptionId"`
	DepositID    string    `json:"depositId"`
	UserID       string    `json:"userId"`
	Amount       float64   `json:"amount"`
	Status       string    `json:"status"` // pending, completed, failed
	CreatedAt    string    `json:"createdAt"`
	CompletedAt  string    `json:"completedAt,omitempty"`
	Reason       string    `json:"reason,omitempty"`
	Metadata     MetadataMap `json:"metadata"`
}

// Issuer represents a deposit issuer
type Issuer struct {
	IssuerID    string    `json:"issuerId"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"` // active, suspended, revoked
	CreatedAt   string    `json:"createdAt"`
	UpdatedAt   string    `json:"updatedAt"`
	Metadata    MetadataMap `json:"metadata"`
}

// AuditLog represents a transaction in the audit trail
type AuditLog struct {
	ID            string    `json:"id"`
	TransactionID string    `json:"transactionId"`
	DepositID     string    `json:"depositId"`
	Operation     string    `json:"operation"`
	ActorID       string    `json:"actorId"`
	CreatedAt     string    `json:"createdAt"`
	Details       MetadataMap `json:"details"`
}

// MetadataMap is a flexible metadata container
type MetadataMap map[string]interface{}

// Permissions represents role-based access control
type Permissions struct {
	UserID string   `json:"userId"`
	Roles  []string `json:"roles"` // admin, issuer, depositor, auditor
}

// TransactionRecord records all state changes
type TransactionRecord struct {
	ID            string    `json:"id"`
	EntityID      string    `json:"entityId"`
	EntityType    string    `json:"entityType"` // deposit, shard, issuer, transfer, redemption
	Operation     string    `json:"operation"` // create, update, delete, transfer
	PreviousState MetadataMap `json:"previousState,omitempty"`
	NewState      MetadataMap `json:"newState,omitempty"`
	ActorID       string    `json:"actorId"`
	Timestamp     string    `json:"timestamp"`
	ShardID       string    `json:"shardId"`
}

// QueryResult represents query results with pagination
type QueryResult struct {
	RecordType string        `json:"recordType"`
	Record     interface{}   `json:"record"`
}

// PagedResults represents paginated query results
type PagedResults struct {
	Records      []QueryResult `json:"records"`
	RecordCount  int           `json:"recordCount"`
	Bookmark     string        `json:"bookmark"`
	TotalMatched int           `json:"totalMatched"`
}

// DepositStats represents statistics for a deposit
type DepositStats struct {
	DepositID       string    `json:"depositId"`
	TotalAmount     float64   `json:"totalAmount"`
	RedeemedAmount  float64   `json:"redeemedAmount"`
	ActiveAmount    float64   `json:"activeAmount"`
	HolderCount     int       `json:"holderCount"`
	TransferCount   int       `json:"transferCount"`
	RedemptionCount int       `json:"redemptionCount"`
	LastActivity    string    `json:"lastActivity"`
}

// ShardStats represents statistics for a shard
type ShardStats struct {
	ShardID              string    `json:"shardId"`
	TotalDeposits        int       `json:"totalDeposits"`
	TotalValue           float64   `json:"totalValue"`
	ActiveDeposits       int       `json:"activeDeposits"`
	IssuerCount          int       `json:"issuerCount"`
	UniqueHolders        int       `json:"uniqueHolders"`
	TransactionCount     int       `json:"transactionCount"`
	AverageDepositValue  float64   `json:"averageDepositValue"`
}

// ComplianceReport represents compliance data
type ComplianceReport struct {
	ReportID         string    `json:"reportId"`
	PeriodStart      string    `json:"periodStart"`
	PeriodEnd        string    `json:"periodEnd"`
	TotalTransactions int       `json:"totalTransactions"`
	TotalValue        float64   `json:"totalValue"`
	SuspiciousCount   int       `json:"suspiciousCount"`
	Details          MetadataMap `json:"details"`
	CreatedAt        string    `json:"createdAt"`
}

// Helper function to get current timestamp in RFC3339 format
func getCurrentTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// Helper function to get future timestamp
func getFutureTimestamp(days int) string {
	return time.Now().UTC().AddDate(0, 0, days).Format(time.RFC3339)
}
