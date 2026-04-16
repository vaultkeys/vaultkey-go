package vaultkey

import "context"

// StablecoinService handles stablecoin transfers and balance lookups.
type StablecoinService struct {
	client *Client
}

// Transfer sends a stablecoin (USDC or USDT) from a wallet.
//
// For EVM, set ChainName (preferred) or ChainID in the payload.
// ChainName takes precedence if both are set.
// For Solana, leave both ChainName and ChainID empty — the server rejects them.
//
// The operation is async. Poll via Jobs.Get(result.JobID).
// Use IdempotencyKey to safely retry without double-sending.
//
//	// EVM — gasless transfer on Base
//	result, apiErr, err := client.Stablecoin.Transfer(ctx, "wallet_id",
//	    vaultkey.ChainTypeEVM,
//	    vaultkey.StablecoinTransferPayload{
//	        Token:     "usdc",
//	        To:        "0xRecipient",
//	        Amount:    "50.00",
//	        ChainName: "base",
//	        Gasless:   true,
//	        Speed:     vaultkey.SpeedFast,
//	    },
//	)
//
//	// Solana
//	result, apiErr, err := client.Stablecoin.Transfer(ctx, "wallet_id",
//	    vaultkey.ChainTypeSolana,
//	    vaultkey.StablecoinTransferPayload{
//	        Token:  "usdc",
//	        To:     "RecipientBase58...",
//	        Amount: "50.00",
//	    },
//	)
func (s *StablecoinService) Transfer(ctx context.Context, walletID string, chainType ChainType, payload StablecoinTransferPayload) (StablecoinTransferResult, *ErrorResponse, error) {
	// For Solana, chain fields must not be sent — server returns 400.
	if chainType == ChainTypeSolana {
		payload.ChainName = ""
		payload.ChainID = ""
	} else if chainType == ChainTypeEVM {
		// chainName takes precedence — mirrors server-side resolveEVMChain logic.
		if payload.ChainName != "" {
			payload.ChainID = ""
		}
	}

	var resp StablecoinTransferResult
	apiErr, err := s.client.post(ctx, "/wallets/"+walletID+"/stablecoin/transfer/"+string(chainType), payload, &resp)
	return resp, apiErr, err
}

// Balance returns the stablecoin balance for a wallet.
//
// For EVM, provide chainName (preferred) or chainID.
// For Solana, leave both empty.
//
//	bal, apiErr, err := client.Stablecoin.Balance(ctx, "wallet_id",
//	    vaultkey.ChainTypeEVM, "usdc", "polygon", "",
//	)
//	fmt.Println(bal.Balance) // "50.00"
func (s *StablecoinService) Balance(ctx context.Context, walletID string, chainType ChainType, token, chainName, chainID string) (StablecoinBalanceResult, *ErrorResponse, error) {
	path := "/wallets/" + walletID + "/stablecoin/balance/" + string(chainType) + "?token=" + token

	if chainType == ChainTypeEVM {
		// chainName takes precedence
		if chainName != "" {
			path += "&chain_name=" + chainName
		} else if chainID != "" {
			path += "&chain_id=" + chainID
		}
	}

	var resp StablecoinBalanceResult
	apiErr, err := s.client.get(ctx, path, &resp)
	return resp, apiErr, err
}

// MasterWalletBalance returns the stablecoin balance of the project's configured
// master wallet for a given chain and token. Useful for reconciliation.
//
// For EVM, provide chainName (preferred) or chainID.
// For Solana, leave both empty.
//
//	bal, apiErr, err := client.Stablecoin.MasterWalletBalance(ctx,
//	    vaultkey.ChainTypeEVM, "usdc", "base", "",
//	)
//	fmt.Println(bal.Balance) // "1000.00"
func (s *StablecoinService) MasterWalletBalance(ctx context.Context, chainType ChainType, token, chainName, chainID string) (StablecoinBalanceResult, *ErrorResponse, error) {
	path := "/master-wallets/balance?chain_type=" + string(chainType) + "&token=" + token

	if chainType == ChainTypeEVM {
		if chainName != "" {
			path += "&chain_name=" + chainName
		} else if chainID != "" {
			path += "&chain_id=" + chainID
		}
	}

	var resp StablecoinBalanceResult
	apiErr, err := s.client.get(ctx, path, &resp)
	return resp, apiErr, err
}