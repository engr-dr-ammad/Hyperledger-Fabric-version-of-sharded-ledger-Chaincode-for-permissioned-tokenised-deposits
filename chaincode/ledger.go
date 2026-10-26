package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
)

// LedgerManager handles core ledger operations
type LedgerManager struct {
	utils              *Utils
	permissions        *PermissionManager
	shardingManager    *ShardingManager
}

// NewLedgerManager creates a new ledger manager
func NewLedgerManager(utils *Utils, permissions *PermissionManager, shardingManager *ShardingManager) *LedgerManager {
	return &LedgerManager{
		utils:           utils,
		permissions:     permissions,
		shardingManager: shardingManager,
	}
}

// GetDeposit retrieves a deposit by ID
func (lm *LedgerManager) GetDeposit(ctx contractapi.TransactionContextInterface, depositID string) (*Deposit, error) {
	// Validate input
	if !lm.utils.ValidateDepositID(depositID) {
		return nil, fmt.Errorf("invalid deposit ID format: %s", depositID)
	}

	// Get from state
	depositBytes, err := ctx.GetStub().GetState(depositID)
	if err != nil {
		return nil, fmt.Errorf("failed to read deposit from state: %v", err)
	}

	if depositBytes == nil {
		return nil, fmt.Errorf("deposit not found: %s", depositID)
	}

	// Unmarshal and return
	deposit, err := lm.utils.UnmarshalDeposit(depositBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal deposit: %v", err)
	}

	// Check access permissions
	canAccess, err := lm.permissions.CanAccessDeposit(ctx, deposit)
	if err != nil {
		return nil, fmt.Errorf("permission check failed: %v", err)
	}
	if !canAccess {
		return nil, fmt.Errorf("access denied to deposit %s", depositID)
	}

	return deposit, nil
}

// GetShard retrieves a shard by ID
func (lm *LedgerManager) GetShard(ctx contractapi.TransactionContextInterface, shardID string) (*Shard, error) {
	// Get from state
	shardBytes, err := ctx.GetStub().GetState(shardID)
	if err != nil {
		return nil, fmt.Errorf("failed to read shard from state: %v", err)
	}

	if shardBytes == nil {
		return nil, fmt.Errorf("shard not found: %s", shardID)
	}

	// Unmarshal and return
	shard, err := lm.utils.UnmarshalShard(shardBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal shard: %v", err)
	}

	return shard, nil
}

// SaveDeposit saves a deposit to state
func (lm *LedgerManager) SaveDeposit(ctx contractapi.TransactionContextInterface, deposit *Deposit) error {
	if deposit == nil {
		return fmt.Errorf("deposit cannot be nil")
	}

	// Marshal the deposit
	depositBytes, err := lm.utils.MarshalDeposit(deposit)
	if err != nil {
		return fmt.Errorf("failed to marshal deposit: %v", err)
	}

	// Save to state
	err = ctx.GetStub().PutState(deposit.DepositID, depositBytes)
	if err != nil {
		return fmt.Errorf("failed to save deposit to state: %v", err)
	}

	return nil
}

// SaveShard saves a shard to state
func (lm *LedgerManager) SaveShard(ctx contractapi.TransactionContextInterface, shard *Shard) error {
	if shard == nil {
		return fmt.Errorf("shard cannot be nil")
	}

	// Marshal the shard
	shardBytes, err := lm.utils.MarshalShard(shard)
	if err != nil {
		return fmt.Errorf("failed to marshal shard: %v", err)
	}

	// Save to state
	err = ctx.GetStub().PutState(shard.ShardID, shardBytes)
	if err != nil {
		return fmt.Errorf("failed to save shard to state: %v", err)
	}

	return nil
}

// SaveTransfer saves a transfer transaction to state
func (lm *LedgerManager) SaveTransfer(ctx contractapi.TransactionContextInterface, transfer *DepositTransfer) error {
	if transfer == nil {
		return fmt.Errorf("transfer cannot be nil")
	}

	transferBytes, err := lm.utils.MarshalTransfer(transfer)
	if err != nil {
		return fmt.Errorf("failed to marshal transfer: %v", err)
	}

	err = ctx.GetStub().PutState(transfer.TransferID, transferBytes)
	if err != nil {
		return fmt.Errorf("failed to save transfer to state: %v", err)
	}

	return nil
}

// SaveRedemption saves a redemption transaction to state
func (lm *LedgerManager) SaveRedemption(ctx contractapi.TransactionContextInterface, redemption *DepositRedemption) error {
	if redemption == nil {
		return fmt.Errorf("redemption cannot be nil")
	}

	redemptionBytes, err := lm.utils.MarshalRedemption(redemption)
	if err != nil {
		return fmt.Errorf("failed to marshal redemption: %v", err)
	}

	err = ctx.GetStub().PutState(redemption.RedemptionID, redemptionBytes)
	if err != nil {
		return fmt.Errorf("failed to save redemption to state: %v", err)
	}

	return nil
}

// SaveAuditLog saves an audit log entry to state
func (lm *LedgerManager) SaveAuditLog(ctx contractapi.TransactionContextInterface, auditLog AuditLog) error {
	auditBytes, err := json.Marshal(auditLog)
	if err != nil {
		return fmt.Errorf("failed to marshal audit log: %v", err)
	}

	err = ctx.GetStub().PutState(auditLog.ID, auditBytes)
	if err != nil {
		return fmt.Errorf("failed to save audit log to state: %v", err)
	}

	return nil
}

// SaveTransactionRecord saves a transaction record to state
func (lm *LedgerManager) SaveTransactionRecord(ctx contractapi.TransactionContextInterface, record TransactionRecord) error {
	recordBytes, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction record: %v", err)
	}

	err = ctx.GetStub().PutState(record.ID, recordBytes)
	if err != nil {
		return fmt.Errorf("failed to save transaction record to state: %v", err)
	}

	return nil
}

// QueryDepositsByRange queries deposits within a value range
func (lm *LedgerManager) QueryDepositsByRange(ctx contractapi.TransactionContextInterface, minAmount, maxAmount float64) ([]*Deposit, error) {
	queryString := fmt.Sprintf(
		`{"selector":{"docType":"deposit","amount":{"$gte":%f,"$lte":%f}}}`,
		minAmount,
		maxAmount,
	)

	results, err := ctx.GetStub().GetQueryResultsForQueryString(queryString)
	if err != nil {
		return nil, fmt.Errorf("query failed: %v", err)
	}
	defer results.Close()

	deposits := make([]*Deposit, 0)
	for results.HasNext() {
		result, err := results.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to get next result: %v", err)
		}

		deposit, err := lm.utils.UnmarshalDeposit(result.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal deposit: %v", err)
		}

		deposits = append(deposits, deposit)
	}

	return deposits, nil
}

// QueryDepositsByStatus queries deposits by status
func (lm *LedgerManager) QueryDepositsByStatus(ctx contractapi.TransactionContextInterface, status string) ([]*Deposit, error) {
	queryString := fmt.Sprintf(
		`{"selector":{"docType":"deposit","status":"%s"}}`,
		status,
	)

	results, err := ctx.GetStub().GetQueryResultsForQueryString(queryString)
	if err != nil {
		return nil, fmt.Errorf("query failed: %v", err)
	}
	defer results.Close()

	deposits := make([]*Deposit, 0)
	for results.HasNext() {
		result, err := results.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to get next result: %v", err)
		}

		deposit, err := lm.utils.UnmarshalDeposit(result.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal deposit: %v", err)
		}

		deposits = append(deposits, deposit)
	}

	return deposits, nil
}

// GetUserDeposits retrieves all deposits held by a user in a specific shard
func (lm *LedgerManager) GetUserDeposits(ctx contractapi.TransactionContextInterface, userID, shardID string) ([]*Deposit, error) {
	queryString := fmt.Sprintf(
		`{"selector":{"docType":"deposit","shardId":"%s"}}`,
		shardID,
	)

	results, err := ctx.GetStub().GetQueryResultsForQueryString(queryString)
	if err != nil {
		return nil, fmt.Errorf("query failed: %v", err)
	}
	defer results.Close()

	userDeposits := make([]*Deposit, 0)
	for results.HasNext() {
		result, err := results.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to get next result: %v", err)
		}

		deposit, err := lm.utils.UnmarshalDeposit(result.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal deposit: %v", err)
		}

		// Check if user holds this deposit
		for _, holding := range deposit.Holdings {
			if holding.UserID == userID {
				userDeposits = append(userDeposits, deposit)
				break
			}
		}
	}

	return userDeposits, nil
}

// CreateIndex creates a composite index for efficient queries
func (lm *LedgerManager) CreateIndex(ctx contractapi.TransactionContextInterface, indexName, docType string, attributes []string) error {
	// Create index definition
	indexDef := map[string]interface{}{
		"index": map[string]interface{}{
			"fields": append([]string{"docType"}, attributes...),
		},
		"name": indexName,
		"type": "json",
	}

	indexDefJSON, err := json.Marshal(indexDef)
	if err != nil {
		return fmt.Errorf("failed to marshal index definition: %v", err)
	}

	// Create the index
	err = ctx.GetStub().CreateCompositeKey(indexName, []string{docType})
	if err != nil {
		return fmt.Errorf("failed to create index: %v", err)
	}

	return nil
}

// DeleteDeposit deletes a deposit from state
func (lm *LedgerManager) DeleteDeposit(ctx contractapi.TransactionContextInterface, depositID string) error {
	err := ctx.GetStub().DelState(depositID)
	if err != nil {
		return fmt.Errorf("failed to delete deposit: %v", err)
	}
	return nil
}

// GetDepositHistory retrieves the history of changes to a deposit
func (lm *LedgerManager) GetDepositHistory(ctx contractapi.TransactionContextInterface, depositID string) ([]*TransactionRecord, error) {
	queryString := fmt.Sprintf(
		`{"selector":{"docType":"transactionRecord","entityId":"%s"}}`,
		depositID,
	)

	results, err := ctx.GetStub().GetQueryResultsForQueryString(queryString)
	if err != nil {
		return nil, fmt.Errorf("query failed: %v", err)
	}
	defer results.Close()

	records := make([]*TransactionRecord, 0)
	for results.HasNext() {
		result, err := results.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to get next result: %v", err)
		}

		var record TransactionRecord
		err = json.Unmarshal(result.Value, &record)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal record: %v", err)
		}

		records = append(records, &record)
	}

	return records, nil
}

// EmitEvent emits a chaincode event
func (lm *LedgerManager) EmitEvent(ctx contractapi.TransactionContextInterface, eventName string, data interface{}) error {
	eventPayload, err := lm.utils.CreateEventPayload(eventName, data)
	if err != nil {
		return fmt.Errorf("failed to create event payload: %v", err)
	}

	err = ctx.GetStub().SetEvent(eventName, eventPayload)
	if err != nil {
		return fmt.Errorf("failed to emit event: %v", err)
	}

	return nil
}

// ReshardDeposit moves a deposit from one shard to another
func (lm *LedgerManager) ReshardDeposit(
	ctx contractapi.TransactionContextInterface,
	deposit *Deposit,
	newShardID string,
) error {
	// Verify shard transition is valid
	if !lm.shardingManager.CanReshardDeposit(deposit.ShardID, newShardID) {
		return fmt.Errorf("cannot reshard from %s to %s", deposit.ShardID, newShardID)
	}

	// Get old shard
	oldShard, err := lm.GetShard(ctx, deposit.ShardID)
	if err != nil {
		return fmt.Errorf("failed to get old shard: %v", err)
	}

	// Get new shard
	newShard, err := lm.GetShard(ctx, newShardID)
	if err != nil {
		return fmt.Errorf("failed to get new shard: %v", err)
	}

	// Store previous state for audit
	previousState := map[string]interface{}{
		"shardId": deposit.ShardID,
	}

	// Update deposit shard
	deposit.ShardID = newShardID
	deposit.UpdatedAt = getCurrentTimestamp()

	// Save updated deposit
	if err := lm.SaveDeposit(ctx, deposit); err != nil {
		return fmt.Errorf("failed to save resharded deposit: %v", err)
	}

	// Create audit entry
	auditLog := AuditLog{
		ID:            lm.utils.GenerateAuditLogID(),
		DepositID:     deposit.DepositID,
		Operation:     "reshard",
		ActorID:       "system",
		CreatedAt:     getCurrentTimestamp(),
		Details: MetadataMap{
			"from_shard": oldShard.ShardID,
			"to_shard":   newShard.ShardID,
		},
	}

	if err := lm.SaveAuditLog(ctx, auditLog); err != nil {
		return fmt.Errorf("failed to save audit log: %v", err)
	}

	// Emit event
	return lm.EmitEvent(ctx, "depositResharded", map[string]interface{}{
		"depositId":  deposit.DepositID,
		"fromShard":  oldShard.ShardID,
		"toShard":    newShard.ShardID,
		"timestamp":  getCurrentTimestamp(),
	})
}
