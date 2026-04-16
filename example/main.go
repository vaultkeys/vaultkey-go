package main

import (
	"context"
	"fmt"
	"log"
	"time"

	vaultkey "github.com/vaultkeys/vaultkey-go"
)

func main() {
	client, err := vaultkey.NewClient(
		"vk_live_your_api_key",
		"your_api_secret",
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// ── 1. Create wallets ─────────────────────────────────────────────────────

	evmWallet, apiErr, err := client.Wallets.Create(ctx, vaultkey.CreateWalletPayload{
		UserID:    "user_123",
		ChainType: vaultkey.ChainTypeEVM,
		Label:     "Primary EVM wallet",
	})
	mustOK("create EVM wallet", apiErr, err)
	fmt.Printf("EVM wallet: %s  %s\n", evmWallet.ID, evmWallet.Address)

	solWallet, apiErr, err := client.Wallets.Create(ctx, vaultkey.CreateWalletPayload{
		UserID:    "user_123",
		ChainType: vaultkey.ChainTypeSolana,
	})
	mustOK("create Solana wallet", apiErr, err)
	fmt.Printf("Solana wallet: %s  %s\n", solWallet.ID, solWallet.Address)

	// ── 2. List wallets ───────────────────────────────────────────────────────

	result, apiErr, err := client.Wallets.ListByUser(ctx, "user_123", "", 20)
	mustOK("list wallets", apiErr, err)
	fmt.Printf("Found %d wallet(s)\n", len(result.Wallets))
	if result.HasMore {
		fmt.Println("  (more pages available)")
	}

	// ── 3. Balances ───────────────────────────────────────────────────────────

	evmBal, apiErr, err := client.Wallets.EVMBalance(ctx, evmWallet.ID, "base-sepolia", "")
	mustOK("evm balance", apiErr, err)
	fmt.Printf("EVM balance: %s %s on %s\n", evmBal.Balance, evmBal.Symbol, evmBal.ChainName)

	solBal, apiErr, err := client.Wallets.SolanaBalance(ctx, solWallet.ID)
	mustOK("solana balance", apiErr, err)
	fmt.Printf("Solana balance: %s %s\n", solBal.Balance, solBal.Symbol)

	// ── 4. Sign messages ──────────────────────────────────────────────────────

	signing := client.Wallets.Signing(evmWallet.ID)

	evmJob, apiErr, err := signing.EVMMessage(ctx, vaultkey.SignMessagePayload{
		Payload:        map[string]any{"message": "Hello from VaultKey"},
		IdempotencyKey: "sign-evm-001",
	})
	mustOK("sign EVM message", apiErr, err)
	fmt.Printf("EVM sign job: %s\n", evmJob.JobID)
	pollJob(ctx, client, evmJob.JobID)

	solJob, apiErr, err := client.Wallets.Signing(solWallet.ID).SolanaMessage(ctx, vaultkey.SignMessagePayload{
		Payload: map[string]any{"data": "SGVsbG8gZnJvbSBWYXVsdEtleQ=="},
	})
	mustOK("sign Solana message", apiErr, err)
	fmt.Printf("Solana sign job: %s\n", solJob.JobID)
	pollJob(ctx, client, solJob.JobID)

	// ── 5. Stablecoin transfer ────────────────────────────────────────────────

	transfer, apiErr, err := client.Stablecoin.Transfer(ctx, evmWallet.ID,
		vaultkey.ChainTypeEVM,
		vaultkey.StablecoinTransferPayload{
			Token:          "usdc",
			To:             "0xRecipientAddress",
			Amount:         "10.00",
			ChainName:      "base-sepolia",
			Gasless:        true,
			Speed:          vaultkey.SpeedNormal,
			IdempotencyKey: "transfer-usdc-001",
		},
	)
	mustOK("stablecoin transfer", apiErr, err)
	fmt.Printf("Transfer job: %s\n", transfer.JobID)
	pollJob(ctx, client, transfer.JobID)

	// Balance after transfer
	bal, apiErr, err := client.Stablecoin.Balance(ctx, evmWallet.ID,
		vaultkey.ChainTypeEVM, "usdc", "base-sepolia", "",
	)
	mustOK("stablecoin balance", apiErr, err)
	fmt.Printf("USDC balance after transfer: %s %s\n", bal.Balance, bal.Symbol)

	// Master wallet stablecoin balance (reconciliation)
	masterBal, apiErr, err := client.Stablecoin.MasterWalletBalance(ctx,
		vaultkey.ChainTypeEVM, "usdc", "base-sepolia", "",
	)
	mustOK("master wallet balance", apiErr, err)
	fmt.Printf("Master wallet USDC balance: %s %s\n", masterBal.Balance, masterBal.Symbol)

	// ── 6. Sweep ──────────────────────────────────────────────────────────────

	sweepJob, apiErr, err := client.Wallets.Sweep(ctx, evmWallet.ID, vaultkey.SweepPayload{
		ChainType: vaultkey.ChainTypeEVM,
		ChainName: "base-sepolia",
	})
	mustOK("EVM sweep", apiErr, err)
	fmt.Printf("EVM sweep job: %s\n", sweepJob.JobID)
	pollJob(ctx, client, sweepJob.JobID)

	solSweep, apiErr, err := client.Wallets.Sweep(ctx, solWallet.ID, vaultkey.SweepPayload{
		ChainType: vaultkey.ChainTypeSolana,
	})
	mustOK("Solana sweep", apiErr, err)
	fmt.Printf("Solana sweep job: %s\n", solSweep.JobID)
	pollJob(ctx, client, solSweep.JobID)

	// ── 7. Chains ─────────────────────────────────────────────────────────────

	chains, apiErr, err := client.Chains.List(ctx)
	mustOK("list chains", apiErr, err)
	fmt.Printf("\nSupported chains (%d):\n", len(chains))
	for _, c := range chains {
		env := "mainnet"
		if c.Testnet {
			env = "testnet"
		}
		fmt.Printf("  %-20s chain_id=%-10s %s  (%s)\n", c.Name, c.ChainID, c.NativeSymbol, env)
	}

	fmt.Println("\n✓ All examples completed.")
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func mustOK(op string, apiErr *vaultkey.ErrorResponse, err error) {
	if err != nil {
		log.Fatalf("%s: transport error: %v", op, err)
	}
	if apiErr != nil {
		log.Fatalf("%s: api error %s: %s", op, apiErr.Code, apiErr.Message)
	}
}

func pollJob(ctx context.Context, client *vaultkey.Client, jobID string) vaultkey.Job {
	fmt.Printf("  Polling job %s...\n", jobID)
	for {
		result, apiErr, err := client.Jobs.Get(ctx, jobID)
		if err != nil {
			log.Fatalf("poll job: %v", err)
		}
		if apiErr != nil {
			log.Fatalf("poll job: %s: %s", apiErr.Code, apiErr.Message)
		}
		fmt.Printf("  status: %s\n", result.Status)
		if result.Status == vaultkey.JobStatusCompleted {
			return result
		}
		if result.Status == vaultkey.JobStatusFailed {
			log.Fatalf("job %s failed: %s", jobID, result.Error)
		}
		time.Sleep(time.Second)
	}
}