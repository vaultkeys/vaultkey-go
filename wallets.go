package vaultkey

import (
	"context"
	"fmt"
)

// WalletsService handles wallet management and operations.
type WalletsService struct {
	client *Client
}

// Signing returns a SigningService scoped to the given wallet ID.
//
//	job, apiErr, err := client.Wallets.Signing("wallet_id").EVMMessage(ctx, payload)
func (w *WalletsService) Signing(walletID string) *SigningService {
	return &SigningService{client: w.client, walletID: walletID}
}

// Create creates a new custodial wallet for a user.
//
//	wallet, apiErr, err := client.Wallets.Create(ctx, vaultkey.CreateWalletPayload{
//	    UserID:    "user_123",
//	    ChainType: vaultkey.ChainTypeEVM,
//	    Label:     "Primary",
//	})
func (w *WalletsService) Create(ctx context.Context, payload CreateWalletPayload) (Wallet, *ErrorResponse, error) {
	var resp Wallet
	apiErr, err := w.client.post(ctx, "/wallets", payload, &resp)
	return resp, apiErr, err
}

// Get retrieves a wallet by its ID.
func (w *WalletsService) Get(ctx context.Context, walletID string) (Wallet, *ErrorResponse, error) {
	var resp Wallet
	apiErr, err := w.client.get(ctx, "/wallets/"+walletID, &resp)
	return resp, apiErr, err
}

// ListByUser returns a paginated list of wallets for a user.
// Pass after="" and limit=0 to use server defaults.
//
//	result, apiErr, err := client.Wallets.ListByUser(ctx, "user_123", "", 20)
//	if result.HasMore {
//	    page2, _, _ := client.Wallets.ListByUser(ctx, "user_123", result.NextCursor, 20)
//	}
func (w *WalletsService) ListByUser(ctx context.Context, userID, after string, limit int) (WalletList, *ErrorResponse, error) {
	path := "/users/" + userID + "/wallets"
	sep := "?"
	if after != "" {
		path += sep + "after=" + after
		sep = "&"
	}
	if limit > 0 {
		path += fmt.Sprintf("%slimit=%d", sep, limit)
	}

	var resp WalletList
	apiErr, err := w.client.get(ctx, path, &resp)
	return resp, apiErr, err
}

// EVMBalance returns the native token balance for an EVM wallet.
// Provide chainName (preferred) or chainID — chainName takes precedence.
//
//	bal, apiErr, err := client.Wallets.EVMBalance(ctx, "wallet_id", "base", "")
func (w *WalletsService) EVMBalance(ctx context.Context, walletID, chainName, chainID string) (EVMBalance, *ErrorResponse, error) {
	path := "/wallets/" + walletID + "/balance/evm?"
	// chainName takes precedence — mirrors server-side resolveEVMChain logic
	if chainName != "" {
		path += "chain_name=" + chainName
	} else if chainID != "" {
		path += "chain_id=" + chainID
	}

	var resp EVMBalance
	apiErr, err := w.client.get(ctx, path, &resp)
	return resp, apiErr, err
}

// SolanaBalance returns the SOL balance for a Solana wallet.
//
//	bal, apiErr, err := client.Wallets.SolanaBalance(ctx, "wallet_id")
func (w *WalletsService) SolanaBalance(ctx context.Context, walletID string) (SolanaBalance, *ErrorResponse, error) {
	var resp SolanaBalance
	apiErr, err := w.client.get(ctx, "/wallets/"+walletID+"/balance/solana", &resp)
	return resp, apiErr, err
}

// BroadcastEVM broadcasts a pre-signed EVM transaction.
// Provide chainName (preferred) or chainID.
//
//	result, apiErr, err := client.Wallets.BroadcastEVM(ctx, "wallet_id", "0x...", "base", "")
func (w *WalletsService) BroadcastEVM(ctx context.Context, walletID, signedTx, chainName, chainID string) (BroadcastEVMResult, *ErrorResponse, error) {
	payload := BroadcastPayload{SignedTx: signedTx}
	// chainName takes precedence
	if chainName != "" {
		payload.ChainName = chainName
	} else if chainID != "" {
		payload.ChainID = chainID
	}

	var resp BroadcastEVMResult
	apiErr, err := w.client.post(ctx, "/wallets/"+walletID+"/broadcast", payload, &resp)
	return resp, apiErr, err
}

// BroadcastSolana broadcasts a pre-signed Solana transaction.
//
//	result, apiErr, err := client.Wallets.BroadcastSolana(ctx, "wallet_id", "base58tx...")
func (w *WalletsService) BroadcastSolana(ctx context.Context, walletID, signedTx string) (BroadcastSolanaResult, *ErrorResponse, error) {
	payload := BroadcastPayload{SignedTx: signedTx}
	var resp BroadcastSolanaResult
	apiErr, err := w.client.post(ctx, "/wallets/"+walletID+"/broadcast", payload, &resp)
	return resp, apiErr, err
}

// Sweep triggers a sweep — moves all funds from the wallet to the configured
// master wallet. The operation is async; poll via Jobs.Get(jobID).
//
// For EVM, provide chainName (preferred) or chainID.
// For Solana, leave both chainName and chainID empty.
//
//	// EVM sweep
//	job, apiErr, err := client.Wallets.Sweep(ctx, "wallet_id", vaultkey.SweepPayload{
//	    ChainType: vaultkey.ChainTypeEVM,
//	    ChainName: "base",
//	})
//
//	// Solana sweep
//	job, apiErr, err := client.Wallets.Sweep(ctx, "wallet_id", vaultkey.SweepPayload{
//	    ChainType: vaultkey.ChainTypeSolana,
//	})
func (w *WalletsService) Sweep(ctx context.Context, walletID string, payload SweepPayload) (SigningJob, *ErrorResponse, error) {
	var resp SigningJob
	apiErr, err := w.client.post(ctx, "/wallets/"+walletID+"/sweep", payload, &resp)
	return resp, apiErr, err
}

// ── Signing ───────────────────────────────────────────────────────────────────

// SigningService scopes signing operations to a specific wallet.
// Obtain via client.Wallets.Signing("wallet_id").
type SigningService struct {
	client   *Client
	walletID string
}

// EVMMessage signs an EVM message or typed data (EIP-712).
// Returns a job — poll via Jobs.Get(job.JobID) until completed or failed.
//
//	job, apiErr, err := client.Wallets.Signing("wallet_id").EVMMessage(ctx,
//	    vaultkey.SignMessagePayload{
//	        Payload:        map[string]any{"message": "Hello from VaultKey"},
//	        IdempotencyKey: "sign-001",
//	    },
//	)
func (s *SigningService) EVMMessage(ctx context.Context, payload SignMessagePayload) (SigningJob, *ErrorResponse, error) {
	var resp SigningJob
	apiErr, err := s.client.post(ctx, "/wallets/"+s.walletID+"/sign/message/evm", payload, &resp)
	return resp, apiErr, err
}

// SolanaMessage signs a Solana message.
// Returns a job — poll via Jobs.Get(job.JobID) until completed or failed.
//
//	job, apiErr, err := client.Wallets.Signing("wallet_id").SolanaMessage(ctx,
//	    vaultkey.SignMessagePayload{
//	        Payload: map[string]any{"data": "SGVsbG8="},
//	    },
//	)
func (s *SigningService) SolanaMessage(ctx context.Context, payload SignMessagePayload) (SigningJob, *ErrorResponse, error) {
	var resp SigningJob
	apiErr, err := s.client.post(ctx, "/wallets/"+s.walletID+"/sign/message/solana", payload, &resp)
	return resp, apiErr, err
}