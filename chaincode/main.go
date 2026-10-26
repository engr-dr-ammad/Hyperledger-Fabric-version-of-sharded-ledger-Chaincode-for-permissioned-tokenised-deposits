package main

import (
	"encoding/json"
	"fmt"
	"log"
)

// DepositChaincode is the main chaincode structure
type DepositChaincode struct {
	ledgerManager   *LedgerManager
	utils           *Utils
	permissions     *PermissionManager
	shardingManager *ShardingManager
}

// InitLedger initializes the ledger with default configuration
func (dc *DepositChaincode) InitLedger(ctx TransactionContextInterface) error {
	log.Println("Initializing deposit chaincode...")

	// Verify caller is admin
	if err := dc.permissions.RequireAdmin(ctx); err != nil {
		return fmt.Errorf("only admin can initialize ledger: %v", err)
	}

	// Create default shards
	shards := []Shard{
		{
			ShardID:     "shard-primary",
			ChannelName: "channel-shard-primary",
			Description: "Primary shard for mixed deposits",
			Issuers:     make([]string, 0),
			Status:      "active",
			CreatedAt:   getCurrentTimestamp(),
			UpdatedAt:   getCurrentTimestamp(),
			Metadata:    MetadataMap{},
		},
		{
			ShardID:     "shard-secondary",
			ChannelName: "channel-shard-secondary",
			Description: "Secondary shard for high-value deposits",
			Issuers:     make([]string, 0),
			Status:      "active",
			CreatedAt:   getCurrentTimestamp(),
			UpdatedAt:   getCurrentTimestamp(),
			Metadata:    MetadataMap{},
		},
	}

	for _, shard := range shards {
		if err := dc.ledgerManager.SaveShard(ctx, &shard); err != nil {
			return fmt.Errorf("failed to save shard: %v", err)
		}
		dc.shardingManager.AddShard(shard.ShardID)
	}

	// Emit initialization event
	if err := dc.ledgerManager.EmitEvent(ctx, "ledgerInitialized", map[string]interface{}{
		"timestamp": getCurrentTimestamp(),
		"shards":    len(shards),
	}); err != nil {
		return fmt.Errorf("failed to emit event: %v", err)
	}

	log.Println("Ledger initialized successfully")
	return nil
}

// CreateShard creates a new shard
func (dc *DepositChaincode) CreateShard(
	ctx TransactionContextInterface,
	shardID string,
	channelName string,
	description string,
) error {
	// Verify caller is admin
	if err := dc.permissions.RequireAdmin(ctx); err != nil {
		return fmt.Errorf("only admin can create shards: %v", err)
	}

	// Check if shard already exists
	existingShard, _ := dc.ledgerManager.GetShard(ctx, shardID)
	if existingShard != nil {
		return fmt.Errorf("shard already exists: %s", shardID)
	}

	// Create new shard
	shard := Shard{
		ShardID:     shardID,
		ChannelName: channelName,
		Description: description,
		Issuers:     make([]string, 0),
		Status:      "active",
		CreatedAt:   getCurrentTimestamp(),
		UpdatedAt:   getCurrentTimestamp(),
		Metadata:    MetadataMap{},
	}

	// Save shard
	if err := dc.ledgerManager.SaveShard(ctx, &shard); err != nil {
		return fmt.Errorf("failed to save shard: %v", err)
	}

	// Add to sharding manager
	dc.shardingManager.AddShard(shardID)

	// Emit event
	return dc.ledgerManager.EmitEvent(ctx, "shardCreated", shard)
}

// IssueDeposit creates a new tokenised deposit
func (dc *DepositChaincode) IssueDeposit(
	ctx TransactionContextInterface,
	depositID string,
	amount string,
	currency string,
	shardID string,
	expiryDays string,
	metadataJSON string,
) error {
	// Verify caller is issuer
	if err := dc.permissions.RequireIssuer(ctx); err != nil {
		return fmt.Errorf("only issuers can issue deposits: %v", err)
	}

	// Validate inputs
	if !dc.utils.ValidateDepositID(depositID) {
		return fmt.Errorf("invalid deposit ID: %s", depositID)
	}

	parsedAmount, err := dc.utils.ParseAmount(amount)
	if err != nil {
		return fmt.Errorf("invalid amount: %v", err)
	}

	if !dc.utils.ValidateCurrency(currency) {
		return fmt.Errorf("invalid currency: %s", currency)
	}

	// Get issuer
	issuerID, err := dc.permissions.GetCallerID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get issuer ID: %v", err)
	}

	// Get shard
	shard, err := dc.ledgerManager.GetShard(ctx, shardID)
	if err != nil {
		return fmt.Errorf("failed to get shard: %v", err)
	}

	// Verify issuer is authorized for this shard
	if err := dc.permissions.ValidateShardAccess(ctx, shard); err != nil {
		return fmt.Errorf("issuer not authorized for shard: %v", err)
	}

	// Parse metadata
	metadata := MetadataMap{}
	if metadataJSON != "" && dc.utils.IsJSONValid(metadataJSON) {
		if err := json.Unmarshal([]byte(metadataJSON), &metadata); err != nil {
			return fmt.Errorf("invalid metadata JSON: %v", err)
		}
	}

	// Create deposit
	now := getCurrentTimestamp()
	var expiryTime string
	if expiryDays != "" {
		days, err := dc.utils.ParseAmount(expiryDays)
		if err == nil && days > 0 {
			expiryTime = getFutureTimestamp(int(days))
		}
	}

	deposit := Deposit{
		DepositID:  depositID,
		IssuerID:   issuerID,
		OwnerID:    issuerID,
		Amount:     parsedAmount,
		Currency:   currency,
		Status:     "active",
		ShardID:    shardID,
		CreatedAt:  now,
		UpdatedAt:  now,
		ExpiresAt:  expiryTime,
		Metadata:   metadata,
		Holdings: []Holding{
			{
				UserID:     issuerID,
				Amount:     parsedAmount,
				AcquiredAt: now,
				Status:     "active",
			},
		},
	}

	// Save deposit
	if err := dc.ledgerManager.SaveDeposit(ctx, &deposit); err != nil {
		return fmt.Errorf("failed to save deposit: %v", err)
	}

	// Create audit log
	auditLog := dc.utils.AddAuditLog(
		depositID,
		"issue",
		issuerID,
		MetadataMap{
			"amount":   parsedAmount,
			"currency": currency,
			"shardId":  shardID,
		},
	)

	if err := dc.ledgerManager.SaveAuditLog(ctx, auditLog); err != nil {
		return fmt.Errorf("failed to save audit log: %v", err)
	}

	// Emit event
	return dc.ledgerManager.EmitEvent(ctx, "depositIssued", deposit)
}

// TransferDeposit transfers a deposit between users
func (dc *DepositChaincode) TransferDeposit(
	ctx TransactionContextInterface,
	depositID string,
	toUserID string,
	amountStr string,
	reasonJSON string,
) error {
	// Verify caller is depositor
	if err := dc.permissions.RequireDepositor(ctx); err != nil {
		return fmt.Errorf("only depositors can transfer: %v", err)
	}

	// Parse amount
	amount, err := dc.utils.ParseAmount(amountStr)
	if err != nil {
		return fmt.Errorf("invalid amount: %v", err)
	}

	// Get deposit
	deposit, err := dc.ledgerManager.GetDeposit(ctx, depositID)
	if err != nil {
		return fmt.Errorf("failed to get deposit: %v", err)
	}

	// Get caller
	callerID, err := dc.permissions.GetCallerID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get caller ID: %v", err)
	}

	// Check permissions
	if err := dc.permissions.CanTransferDeposit(ctx, deposit, callerID); err != nil {
		return fmt.Errorf("transfer not authorized: %v", err)
	}

	// Check compliance
	if err := dc.permissions.CheckCompliance(ctx, deposit, amount); err != nil {
		return fmt.Errorf("compliance check failed: %v", err)
	}

	// Parse reason
	reason := ""
	if reasonJSON != "" {
		var reasonObj map[string]interface{}
		if err := json.Unmarshal([]byte(reasonJSON), &reasonObj); err == nil {
			if r, ok := reasonObj["reason"].(string); ok {
				reason = r
			}
		}
	}

	// Create transfer record
	transfer := DepositTransfer{
		TransferID:  dc.utils.GenerateTransferID(),
		DepositID:   depositID,
		FromUserID:  callerID,
		ToUserID:    toUserID,
		Amount:      amount,
		Status:      "completed",
		CreatedAt:   getCurrentTimestamp(),
		CompletedAt: getCurrentTimestamp(),
		Reason:      reason,
		Metadata:    MetadataMap{},
	}

	// Update holdings
	senderHolding := dc.utils.FindHolding(deposit.Holdings, callerID)
	if senderHolding != nil {
		senderHolding.Amount -= amount
		if senderHolding.Amount <= 0 {
			senderHolding.Status = "transferred"
		}
	}

	// Add receiver holding
	receiverHolding := dc.utils.FindHolding(deposit.Holdings, toUserID)
	if receiverHolding != nil {
		receiverHolding.Amount += amount
		receiverHolding.Status = "active"
	} else {
		deposit.Holdings = append(deposit.Holdings, Holding{
			UserID:     toUserID,
			Amount:     amount,
			AcquiredAt: getCurrentTimestamp(),
			Status:     "active",
		})
	}

	deposit.UpdatedAt = getCurrentTimestamp()

	// Save updated deposit
	if err := dc.ledgerManager.SaveDeposit(ctx, deposit); err != nil {
		return fmt.Errorf("failed to save deposit: %v", err)
	}

	// Save transfer record
	if err := dc.ledgerManager.SaveTransfer(ctx, &transfer); err != nil {
		return fmt.Errorf("failed to save transfer: %v", err)
	}

	// Create audit log
	auditLog := dc.utils.AddAuditLog(
		depositID,
		"transfer",
		callerID,
		MetadataMap{
			"fromUserId": callerID,
			"toUserId":   toUserID,
			"amount":     amount,
		},
	)

	if err := dc.ledgerManager.SaveAuditLog(ctx, auditLog); err != nil {
		return fmt.Errorf("failed to save audit log: %v", err)
	}

	// Emit event
	return dc.ledgerManager.EmitEvent(ctx, "depositTransferred", transfer)
}

// RedeemDeposit redeems a deposit
func (dc *DepositChaincode) RedeemDeposit(
	ctx TransactionContextInterface,
	depositID string,
	amountStr string,
) error {
	// Verify caller is depositor
	if err := dc.permissions.RequireDepositor(ctx); err != nil {
		return fmt.Errorf("only depositors can redeem: %v", err)
	}

	// Parse amount
	amount, err := dc.utils.ParseAmount(amountStr)
	if err != nil {
		return fmt.Errorf("invalid amount: %v", err)
	}

	// Get deposit
	deposit, err := dc.ledgerManager.GetDeposit(ctx, depositID)
	if err != nil {
		return fmt.Errorf("failed to get deposit: %v", err)
	}

	// Get caller
	callerID, err := dc.permissions.GetCallerID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get caller ID: %v", err)
	}

	// Check permissions
	if err := dc.permissions.CanRedeemDeposit(ctx, deposit, callerID); err != nil {
		return fmt.Errorf("redemption not authorized: %v", err)
	}

	// Check compliance
	if err := dc.permissions.CheckCompliance(ctx, deposit, amount); err != nil {
		return fmt.Errorf("compliance check failed: %v", err)
	}

	// Create redemption record
	redemption := DepositRedemption{
		RedemptionID: dc.utils.GenerateRedemptionID(),
		DepositID:    depositID,
		UserID:       callerID,
		Amount:       amount,
		Status:       "completed",
		CreatedAt:    getCurrentTimestamp(),
		CompletedAt:  getCurrentTimestamp(),
		Metadata:     MetadataMap{},
	}

	// Update holdings
	holding := dc.utils.FindHolding(deposit.Holdings, callerID)
	if holding != nil {
		holding.Amount -= amount
		if holding.Amount <= 0 {
			holding.Status = "redeemed"
		}
	}

	// Update deposit status if fully redeemed
	totalRemaining := 0.0
	for _, h := range deposit.Holdings {
		if h.Status == "active" {
			totalRemaining += h.Amount
		}
	}

	if totalRemaining <= 0 {
		deposit.Status = "redeemed"
	}

	deposit.UpdatedAt = getCurrentTimestamp()

	// Save updated deposit
	if err := dc.ledgerManager.SaveDeposit(ctx, deposit); err != nil {
		return fmt.Errorf("failed to save deposit: %v", err)
	}

	// Save redemption record
	if err := dc.ledgerManager.SaveRedemption(ctx, &redemption); err != nil {
		return fmt.Errorf("failed to save redemption: %v", err)
	}

	// Create audit log
	auditLog := dc.utils.AddAuditLog(
		depositID,
		"redeem",
		callerID,
		MetadataMap{
			"amount": amount,
		},
	)

	if err := dc.ledgerManager.SaveAuditLog(ctx, auditLog); err != nil {
		return fmt.Errorf("failed to save audit log: %v", err)
	}

	// Emit event
	return dc.ledgerManager.EmitEvent(ctx, "depositRedeemed", redemption)
}

// GetDeposit returns a deposit
func (dc *DepositChaincode) GetDeposit(
	ctx TransactionContextInterface,
	depositID string,
) (*Deposit, error) {
	return dc.ledgerManager.GetDeposit(ctx, depositID)
}

// GetUserDeposits returns all deposits held by a user in a shard
func (dc *DepositChaincode) GetUserDeposits(
	ctx TransactionContextInterface,
	userID string,
	shardID string,
) ([]*Deposit, error) {
	return dc.ledgerManager.GetUserDeposits(ctx, userID, shardID)
}

// QueryDepositsByAmount queries deposits by amount range
func (dc *DepositChaincode) QueryDepositsByAmount(
	ctx TransactionContextInterface,
	minAmount string,
	maxAmount string,
) ([]*Deposit, error) {
	minVal, err := dc.utils.ParseAmount(minAmount)
	if err != nil {
		return nil, fmt.Errorf("invalid minimum amount: %v", err)
	}

	maxVal, err := dc.utils.ParseAmount(maxAmount)
	if err != nil {
		return nil, fmt.Errorf("invalid maximum amount: %v", err)
	}

	return dc.ledgerManager.QueryDepositsByRange(ctx, minVal, maxVal)
}

// GetShardInfo returns information about a shard
func (dc *DepositChaincode) GetShardInfo(
	ctx TransactionContextInterface,
	shardID string,
) (*Shard, error) {
	return dc.ledgerManager.GetShard(ctx, shardID)
}

// GetTransactionHistory returns the transaction history for a deposit
func (dc *DepositChaincode) GetTransactionHistory(
	ctx TransactionContextInterface,
	depositID string,
) ([]*TransactionRecord, error) {
	return dc.ledgerManager.GetDepositHistory(ctx, depositID)
}

// SmartContractInit initializes the smart contract
func (sc *DepositChaincode) SmartContractInit(ctx TransactionContextInterface) error {
	return nil
}

// NewDepositChaincode creates a new deposit chaincode instance
func NewDepositChaincode() *DepositChaincode {
	utils := &Utils{}
	permissions := &PermissionManager{}
	shardingManager := NewShardingManager(ShardByHash)
	ledgerManager := NewLedgerManager(utils, permissions, shardingManager)

	return &DepositChaincode{
		ledgerManager:   ledgerManager,
		utils:           utils,
		permissions:     permissions,
		shardingManager: shardingManager,
	}
}

func main() {
	log.Println("Deposit Chaincode initialized")
	cc := NewDepositChaincode()
	log.Printf("Chaincode ready: %+v", cc)
	select {}
}
