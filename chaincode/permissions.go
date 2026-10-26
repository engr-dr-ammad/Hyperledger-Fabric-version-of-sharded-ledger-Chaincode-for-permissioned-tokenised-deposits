package main

import (
	"fmt"
	"strings"
)

// PermissionManager handles access control
type PermissionManager struct{}

// UserRole represents a user's role in the system
type UserRole string

const (
	RoleAdmin    UserRole = "admin"
	RoleIssuer   UserRole = "issuer"
	RoleDepositor UserRole = "depositor"
	RoleAuditor  UserRole = "auditor"
)

// RequireRole checks if the caller has the required role
func (pm *PermissionManager) RequireRole(ctx TransactionContextInterface, requiredRoles ...string) error {
	clientID := ctx.GetClientIdentity()

	// Get MSPID for organization
	mspID, err := clientID.GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get MSP ID: %v", err)
	}

	// Get organization unit attributes
	attrs, err := clientID.GetAttributeValue("ou")
	if err != nil {
		return fmt.Errorf("failed to get organization unit: %v", err)
	}

	// Check if user's role matches required roles
	for _, requiredRole := range requiredRoles {
		if strings.Contains(attrs, requiredRole) {
			return nil
		}
	}

	return fmt.Errorf("caller (MSP: %s, OU: %s) is not authorized to perform this action", mspID, attrs)
}

// RequireIssuer checks if the caller is an issuer
func (pm *PermissionManager) RequireIssuer(ctx TransactionContextInterface) error {
	return pm.RequireRole(ctx, string(RoleIssuer), string(RoleAdmin))
}

// RequireAdmin checks if the caller is an admin
func (pm *PermissionManager) RequireAdmin(ctx TransactionContextInterface) error {
	return pm.RequireRole(ctx, string(RoleAdmin))
}

// RequireDepositor checks if the caller is a depositor
func (pm *PermissionManager) RequireDepositor(ctx TransactionContextInterface) error {
	return pm.RequireRole(ctx, string(RoleDepositor), string(RoleAdmin))
}

// RequireAuditor checks if the caller is an auditor
func (pm *PermissionManager) RequireAuditor(ctx TransactionContextInterface) error {
	return pm.RequireRole(ctx, string(RoleAuditor), string(RoleAdmin))
}

// GetCallerOrgID returns the organization ID of the caller
func (pm *PermissionManager) GetCallerOrgID(ctx TransactionContextInterface) (string, error) {
	clientID := ctx.GetClientIdentity()
	mspID, err := clientID.GetMSPID()
	if err != nil {
		return "", fmt.Errorf("failed to get MSP ID: %v", err)
	}
	return mspID, nil
}

// GetCallerID returns the unique identifier of the caller
func (pm *PermissionManager) GetCallerID(ctx TransactionContextInterface) (string, error) {
	clientID := ctx.GetClientIdentity()

	// Try to get caller ID from the client identity
	id, err := clientID.GetID()
	if err == nil && id != "" {
		return id, nil
	}

	// Fallback: use MSP ID
	mspID, err := clientID.GetMSPID()
	if err != nil {
		return "", fmt.Errorf("failed to get MSP ID: %v", err)
	}
	return mspID, nil
}

// CanAccessDeposit checks if the caller can access a specific deposit
func (pm *PermissionManager) CanAccessDeposit(ctx TransactionContextInterface, deposit *Deposit) (bool, error) {
	callerID, err := pm.GetCallerID(ctx)
	if err != nil {
		return false, err
	}

	// Check if caller is admin
	if err := pm.RequireAdmin(ctx); err == nil {
		return true, nil
	}

	// Check if caller is auditor (read-only)
	if err := pm.RequireAuditor(ctx); err == nil {
		return true, nil
	}

	// Check if caller is the issuer
	if deposit.IssuerID == callerID {
		return true, nil
	}

	// Check if caller is an owner
	for _, holding := range deposit.Holdings {
		if holding.UserID == callerID {
			return true, nil
		}
	}

	return false, nil
}

// CanTransferDeposit checks if the caller can transfer a deposit
func (pm *PermissionManager) CanTransferDeposit(ctx TransactionContextInterface, deposit *Deposit, fromUserID string) error {
	callerID, err := pm.GetCallerID(ctx)
	if err != nil {
		return err
	}

	// Admin can always transfer
	if err := pm.RequireAdmin(ctx); err == nil {
		return nil
	}

	// User can only transfer their own deposits
	if callerID != fromUserID {
		return fmt.Errorf("caller is not authorized to transfer deposits from %s", fromUserID)
	}

	// Check if user actually holds the deposit
	for _, holding := range deposit.Holdings {
		if holding.UserID == fromUserID && holding.Status == "active" {
			return nil
		}
	}

	return fmt.Errorf("caller does not have an active holding in this deposit")
}

// CanRedeemDeposit checks if the caller can redeem a deposit
func (pm *PermissionManager) CanRedeemDeposit(ctx TransactionContextInterface, deposit *Deposit, userID string) error {
	callerID, err := pm.GetCallerID(ctx)
	if err != nil {
		return err
	}

	// Admin can always redeem
	if err := pm.RequireAdmin(ctx); err == nil {
		return nil
	}

	// User can only redeem their own deposits
	if callerID != userID {
		return fmt.Errorf("caller is not authorized to redeem deposits for %s", userID)
	}

	// Check if user actually holds the deposit
	for _, holding := range deposit.Holdings {
		if holding.UserID == userID && holding.Status == "active" {
			return nil
		}
	}

	return fmt.Errorf("caller does not have an active holding in this deposit")
}

// ValidateShardAccess checks if the caller can access a shard
func (pm *PermissionManager) ValidateShardAccess(ctx TransactionContextInterface, shard *Shard) error {
	orgID, err := pm.GetCallerOrgID(ctx)
	if err != nil {
		return err
	}

	// Admin can access any shard
	if err := pm.RequireAdmin(ctx); err == nil {
		return nil
	}

	// Issuers can access shards they are authorized for
	for _, issuerID := range shard.Issuers {
		if issuerID == orgID {
			return nil
		}
	}

	return fmt.Errorf("organization %s is not authorized to access shard %s", orgID, shard.ShardID)
}

// CheckCompliance verifies compliance requirements
func (pm *PermissionManager) CheckCompliance(ctx TransactionContextInterface, deposit *Deposit, amount float64) error {
	// Verify deposit is not expired
	if deposit.ExpiresAt != "" {
		expiresAt := deposit.ExpiresAt
		now := getCurrentTimestamp()
		if now > expiresAt {
			return fmt.Errorf("deposit %s has expired", deposit.DepositID)
		}
	}

	// Verify deposit status
	if deposit.Status != "active" {
		return fmt.Errorf("deposit %s is not active (status: %s)", deposit.DepositID, deposit.Status)
	}

	// Verify amount is valid
	if amount <= 0 {
		return fmt.Errorf("amount must be greater than 0")
	}

	// Verify amount doesn't exceed deposit balance
	totalAmount := 0.0
	for _, holding := range deposit.Holdings {
		totalAmount += holding.Amount
	}

	if amount > totalAmount {
		return fmt.Errorf("requested amount %.2f exceeds available balance %.2f", amount, totalAmount)
	}

	return nil
}
