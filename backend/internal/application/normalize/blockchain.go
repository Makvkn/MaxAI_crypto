// Package normalize maps provider observations into canonical domain records.
package normalize

import (
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/maxaicrypto/backend/internal/domain/asset"
	"github.com/maxaicrypto/backend/internal/domain/chain"
	"github.com/maxaicrypto/backend/internal/domain/provider"
	"github.com/maxaicrypto/backend/internal/domain/shared"
	"github.com/maxaicrypto/backend/internal/domain/transaction"
)

// AssetFromBalance builds a domain asset from a normalized balance.
func AssetFromBalance(balance provider.NormalizedBalance) asset.Asset {
	contract := normalizeContract(balance.AssetIdentity.ContractAddress)
	return asset.Asset{
		ChainID:         balance.ChainID,
		ContractAddress: contract,
		Symbol:          balance.Metadata.Symbol,
		Name:            balance.Metadata.Name,
		Decimals:        balance.Metadata.Decimals,
		Type:            balance.Metadata.Type,
		IconURL:         balance.Metadata.IconURL,
	}
}

// AssetFromTransfer builds a domain asset from a transfer observation.
func AssetFromTransfer(chainID chain.ID, metadata provider.AssetMetadata, identity asset.Identity) asset.Asset {
	contract := normalizeContract(identity.ContractAddress)
	return asset.Asset{
		ChainID:         chainID,
		ContractAddress: contract,
		Symbol:          metadata.Symbol,
		Name:            metadata.Name,
		Decimals:        metadata.Decimals,
		Type:            metadata.Type,
		IconURL:         metadata.IconURL,
	}
}

// Transaction maps one normalized transaction movement into the canonical model.
func Transaction(
	walletID uuid.UUID,
	walletAddress string,
	normalized provider.NormalizedTransaction,
	logIndex int,
	assetInID *uuid.UUID,
	amountIn shared.NullDecimal,
	assetOutID *uuid.UUID,
	amountOut shared.NullDecimal,
	feeAssetID *uuid.UUID,
	feeAmount shared.NullDecimal,
) transaction.Transaction {
	status := transaction.StatusSuccess
	if !normalized.Successful {
		status = transaction.StatusFailed
	}
	return transaction.Transaction{
		WalletID:     walletID,
		ChainID:      normalized.ChainID,
		TxHash:       normalized.TxHash,
		BlockNumber:  normalized.BlockNumber,
		Timestamp:    normalized.Timestamp,
		Status:       status,
		Type:         transaction.TypeUnknown,
		FromAddress:  normalized.FromAddress,
		ToAddress:    normalized.ToAddress,
		AssetInID:    assetInID,
		AmountIn:     amountIn,
		AssetOutID:   assetOutID,
		AmountOut:    amountOut,
		FeeAssetID:   feeAssetID,
		FeeAmount:    feeAmount,
		Protocol:     normalized.Protocol,
		Counterparty: normalized.Counterparty,
		RawReference: normalized.ProviderRef,
	}
}

func normalizeContract(contract *string) *string {
	if contract == nil {
		return nil
	}
	trimmed := strings.TrimSpace(strings.ToLower(*contract))
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

// TransferLogIndex returns the stable log index for a transfer inside a tx.
func TransferLogIndex(index int) int { return index }

// NowUTC returns the current UTC timestamp.
func NowUTC() time.Time { return time.Now().UTC() }
