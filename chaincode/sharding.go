package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"strconv"
	"strings"
)

// ShardingStrategy defines how deposits are distributed across shards
type ShardingStrategy string

const (
	ShardByCurrency   ShardingStrategy = "currency"
	ShardByIssuer     ShardingStrategy = "issuer"
	ShardByTime       ShardingStrategy = "time"
	ShardByHash       ShardingStrategy = "hash"
	ShardByCustom     ShardingStrategy = "custom"
)

// ShardingManager handles deposit distribution across shards
type ShardingManager struct {
	strategy ShardingStrategy
	shards   []string // list of available shard IDs
}

// NewShardingManager creates a new sharding manager
func NewShardingManager(strategy ShardingStrategy) *ShardingManager {
	return &ShardingManager{
		strategy: strategy,
		shards:   make([]string, 0),
	}
}

// AddShard adds a shard to the manager
func (sm *ShardingManager) AddShard(shardID string) {
	sm.shards = append(sm.shards, shardID)
}

// RemoveShard removes a shard from the manager
func (sm *ShardingManager) RemoveShard(shardID string) {
	for i, s := range sm.shards {
		if s == shardID {
			sm.shards = append(sm.shards[:i], sm.shards[i+1:]...)
			break
		}
	}
}

// GetAvailableShards returns all available shards
func (sm *ShardingManager) GetAvailableShards() []string {
	return sm.shards
}

// DetermineShardForDeposit determines which shard a deposit should be placed in
func (sm *ShardingManager) DetermineShardForDeposit(
	depositID string,
	currency string,
	issuerID string,
	issuanceTime string,
	customKey string,
) (string, error) {
	if len(sm.shards) == 0 {
		return "", fmt.Errorf("no shards available")
	}

	switch sm.strategy {
	case ShardByCurrency:
		return sm.shardByCurrency(currency)
	case ShardByIssuer:
		return sm.shardByIssuer(issuerID)
	case ShardByTime:
		return sm.shardByTime(issuanceTime)
	case ShardByHash:
		return sm.shardByHash(depositID)
	case ShardByCustom:
		return sm.shardByCustomKey(customKey)
	default:
		return sm.shards[0], nil
	}
}

// shardByCurrency assigns shards based on currency
func (sm *ShardingManager) shardByCurrency(currency string) (string, error) {
	// Normalize currency
	currency = strings.ToUpper(currency)

	// Create a hash of the currency
	hash := fnv.New32a()
	hash.Write([]byte(currency))
	hashValue := hash.Sum32()

	// Select shard based on hash
	shardIndex := int(hashValue) % len(sm.shards)
	return sm.shards[shardIndex], nil
}

// shardByIssuer assigns shards based on issuer
func (sm *ShardingManager) shardByIssuer(issuerID string) (string, error) {
	// Create a hash of the issuer ID
	hash := fnv.New32a()
	hash.Write([]byte(issuerID))
	hashValue := hash.Sum32()

	// Select shard based on hash
	shardIndex := int(hashValue) % len(sm.shards)
	return sm.shards[shardIndex], nil
}

// shardByTime assigns shards based on issuance time (for rotation)
func (sm *ShardingManager) shardByTime(issuanceTime string) (string, error) {
	// Extract date from timestamp (YYYY-MM-DD)
	parts := strings.Split(issuanceTime, "T")
	if len(parts) == 0 {
		return "", fmt.Errorf("invalid timestamp format")
	}

	dateStr := parts[0]
	// Convert date to day number for rotation
	hash := fnv.New32a()
	hash.Write([]byte(dateStr))
	hashValue := hash.Sum32()

	// Select shard based on hash
	shardIndex := int(hashValue) % len(sm.shards)
	return sm.shards[shardIndex], nil
}

// shardByHash assigns shards using SHA256 hash of deposit ID
func (sm *ShardingManager) shardByHash(depositID string) (string, error) {
	// Create SHA256 hash of deposit ID
	hasher := sha256.New()
	hasher.Write([]byte(depositID))
	hash := hasher.Sum(nil)
	hashHex := hex.EncodeToString(hash)

	// Convert first 8 characters of hex to number
	hashNum, err := strconv.ParseUint(hashHex[:8], 16, 32)
	if err != nil {
		return "", fmt.Errorf("failed to parse hash: %v", err)
	}

	// Select shard based on hash
	shardIndex := int(hashNum) % len(sm.shards)
	return sm.shards[shardIndex], nil
}

// shardByCustomKey assigns shards based on custom key
func (sm *ShardingManager) shardByCustomKey(customKey string) (string, error) {
	if customKey == "" {
		return "", fmt.Errorf("custom key cannot be empty")
	}

	// Create a hash of the custom key
	hash := fnv.New32a()
	hash.Write([]byte(customKey))
	hashValue := hash.Sum32()

	// Select shard based on hash
	shardIndex := int(hashValue) % len(sm.shards)
	return sm.shards[shardIndex], nil
}

// CanReshardDeposit checks if a deposit can be resharded
func (sm *ShardingManager) CanReshardDeposit(currentShard string, newShard string) bool {
	// Check if both shards exist
	currentExists := false
	newExists := false

	for _, s := range sm.shards {
		if s == currentShard {
			currentExists = true
		}
		if s == newShard {
			newExists = true
		}
	}

	return currentExists && newExists && currentShard != newShard
}

// ValidateShardConfiguration validates the sharding configuration
func (sm *ShardingManager) ValidateShardConfiguration() error {
	if sm.strategy == "" {
		return fmt.Errorf("sharding strategy not set")
	}

	if len(sm.shards) == 0 {
		return fmt.Errorf("no shards configured")
	}

	return nil
}

// GetShardMetrics returns metrics for a specific shard
type ShardMetrics struct {
	ShardID          string
	DepositCount     int
	ActiveDeposits   int
	TotalValue       float64
	UniqueHolders    int
	TransactionCount int
}

// CalculateLoadBalance calculates how balanced the shards are
type LoadBalance struct {
	ShardID     string
	Deposits    int
	Value       float64
	LoadPercent float64
}

// GetLoadBalance returns load balance information for all shards
func (sm *ShardingManager) GetLoadBalance(shardMetrics map[string]*ShardMetrics) []LoadBalance {
	totalDeposits := 0
	totalValue := 0.0

	for _, metrics := range shardMetrics {
		totalDeposits += metrics.DepositCount
		totalValue += metrics.TotalValue
	}

	balance := make([]LoadBalance, 0)

	for shardID, metrics := range shardMetrics {
		depositPercent := 0.0
		if totalDeposits > 0 {
			depositPercent = (float64(metrics.DepositCount) / float64(totalDeposits)) * 100
		}

		balance = append(balance, LoadBalance{
			ShardID:     shardID,
			Deposits:    metrics.DepositCount,
			Value:       metrics.TotalValue,
			LoadPercent: depositPercent,
		})
	}

	return balance
}

// RecommendRebalancing recommends which deposits should be resharded for load balancing
func (sm *ShardingManager) RecommendRebalancing(shardMetrics map[string]*ShardMetrics) map[string][]string {
	recommendations := make(map[string][]string)

	loadBalance := sm.GetLoadBalance(shardMetrics)

	// Find average load
	totalLoad := 0.0
	for _, lb := range loadBalance {
		totalLoad += lb.LoadPercent
	}
	avgLoad := totalLoad / float64(len(loadBalance))

	// Overloaded shards are those with > 1.5x average load
	overloadThreshold := avgLoad * 1.5
	underloadThreshold := avgLoad * 0.5

	for _, lb := range loadBalance {
		if lb.LoadPercent > overloadThreshold {
			// This shard is overloaded
			recommendations["overloaded:"+lb.ShardID] = []string{}

			// Find underloaded shards to move deposits to
			for _, targetLB := range loadBalance {
				if targetLB.LoadPercent < underloadThreshold && targetLB.ShardID != lb.ShardID {
					recommendations["move_to:"+targetLB.ShardID] = append(
						recommendations["move_to:"+targetLB.ShardID],
						lb.ShardID,
					)
				}
			}
		}
	}

	return recommendations
}

// ShardConsistencyCheck checks data consistency across shards
type ConsistencyCheckResult struct {
	ShardID        string
	DepositCount   int
	TotalValue     float64
	Issues         []string
	IsConsistent   bool
}

// VerifyShardConsistency verifies consistency for a shard
func (sm *ShardingManager) VerifyShardConsistency(
	deposits []Deposit,
	expectedDepositCount int,
) ConsistencyCheckResult {
	result := ConsistencyCheckResult{
		Issues:       make([]string, 0),
		IsConsistent: true,
	}

	if len(deposits) != expectedDepositCount {
		result.Issues = append(result.Issues,
			fmt.Sprintf("deposit count mismatch: expected %d, got %d", expectedDepositCount, len(deposits)))
		result.IsConsistent = false
	}

	// Verify each deposit
	for _, deposit := range deposits {
		// Check that holdings sum doesn't exceed total
		holdingsTotal := 0.0
		for _, holding := range deposit.Holdings {
			if holding.Status == "active" {
				holdingsTotal += holding.Amount
			}
		}

		if holdingsTotal > deposit.Amount {
			result.Issues = append(result.Issues,
				fmt.Sprintf("deposit %s: holdings total (%.2f) exceeds deposit amount (%.2f)",
					deposit.DepositID, holdingsTotal, deposit.Amount))
			result.IsConsistent = false
		}

		result.TotalValue += deposit.Amount
	}

	result.DepositCount = len(deposits)
	return result
}
