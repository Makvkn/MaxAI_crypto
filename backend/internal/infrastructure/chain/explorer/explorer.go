// Package explorer builds canonical block explorer links per chain.
package explorer

import (
	"github.com/maxaicrypto/backend/internal/domain/chain"
)

var txBases = map[chain.ID]string{
	chain.Ethereum:  "https://etherscan.io/tx/",
	chain.Bitcoin:   "https://mempool.space/tx/",
	chain.BNBChain:  "https://bscscan.com/tx/",
	chain.Solana:    "https://solscan.io/tx/",
	chain.Litecoin:  "https://blockchair.com/litecoin/transaction/",
	chain.XRPLedger: "https://livenet.xrpl.org/transactions/",
	chain.Tron:      "https://tronscan.org/#/transaction/",
	chain.Dogecoin:  "https://blockchair.com/dogecoin/transaction/",
}

// TransactionURL returns a canonical explorer link for a transaction hash.
func TransactionURL(id chain.ID, txHash string) *string {
	base, ok := txBases[id]
	if !ok || txHash == "" {
		return nil
	}
	url := base + txHash
	return &url
}
