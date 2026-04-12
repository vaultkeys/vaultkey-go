package vaultkey

// ErrorResponse is returned by the API on non-2xx responses.
type ErrorResponse struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

func (e *ErrorResponse) Error() string {
	return e.Code + ": " + e.Message
}

// ChainType is either "evm" or "solana".
type ChainType string

const (
	ChainTypeEVM    ChainType = "evm"
	ChainTypeSolana ChainType = "solana"
	ChainTypeTron   ChainType = "tron"
)

// TransferSpeed controls transaction priority for stablecoin transfers.
type TransferSpeed string

const (
	SpeedSlow   TransferSpeed = "slow"
	SpeedNormal TransferSpeed = "normal"
	SpeedFast   TransferSpeed = "fast"
)

// JobStatus is the current state of an async job.
type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
)

// ── Chains ────────────────────────────────────────────────────────────────────

// Chain represents a supported EVM chain.
type Chain struct {
	Name         string `json:"name"`
	ChainID      string `json:"chain_id"`
	NativeSymbol string `json:"native_symbol"`
	LegacySymbol string `json:"legacy_symbol,omitempty"`
	Testnet      bool   `json:"testnet"`
}

// ── Wallets ───────────────────────────────────────────────────────────────────

// Wallet represents a custodial wallet.
type Wallet struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	ChainType ChainType `json:"chain_type"`
	Address   string    `json:"address"`
	Label     string    `json:"label,omitempty"`
	CreatedAt string    `json:"created_at"`
}

// CreateWalletPayload is the request body for wallet creation.
type CreateWalletPayload struct {
	UserID    string    `json:"user_id"`
	ChainType ChainType `json:"chain_type"`
	Label     string    `json:"label,omitempty"`
}

// WalletList is the paginated response from listing wallets.
type WalletList struct {
	Wallets    []Wallet `json:"wallets"`
	NextCursor string   `json:"next_cursor,omitempty"`
	HasMore    bool     `json:"has_more"`
}

// ── Signing ───────────────────────────────────────────────────────────────────

// SigningJob is returned by async signing operations.
type SigningJob struct {
	JobID  string    `json:"job_id"`
	Status JobStatus `json:"status"`
}

// SignMessagePayload is the request body for signing a message.
type SignMessagePayload struct {
	Payload        map[string]any `json:"payload"`
	IdempotencyKey string         `json:"idempotency_key,omitempty"`
}

// ── Balance ───────────────────────────────────────────────────────────────────

// EVMBalance is the response from an EVM balance lookup.
type EVMBalance struct {
	Address    string `json:"address"`
	Balance    string `json:"balance"`
	RawBalance string `json:"raw_balance"`
	Symbol     string `json:"symbol"`
	ChainName  string `json:"chain_name"`
	ChainID    string `json:"chain_id"`
}

// SolanaBalance is the response from a Solana balance lookup.
type SolanaBalance struct {
	Address    string `json:"address"`
	Balance    string `json:"balance"`
	RawBalance string `json:"raw_balance"`
	Symbol     string `json:"symbol"`
}

// ── Broadcast ─────────────────────────────────────────────────────────────────

// BroadcastPayload is the request body for broadcasting a signed transaction.
type BroadcastPayload struct {
	SignedTx  string `json:"signed_tx"`
	ChainName string `json:"chain_name,omitempty"`
	ChainID   string `json:"chain_id,omitempty"`
}

// BroadcastEVMResult is the response from an EVM broadcast.
type BroadcastEVMResult struct {
	TxHash    string `json:"tx_hash"`
	ChainName string `json:"chain_name"`
	ChainID   string `json:"chain_id"`
}

// BroadcastSolanaResult is the response from a Solana broadcast.
type BroadcastSolanaResult struct {
	Signature string `json:"signature"`
}

// ── Sweep ─────────────────────────────────────────────────────────────────────

// SweepPayload is the request body for triggering a sweep.
type SweepPayload struct {
	ChainType ChainType `json:"chain_type"`
	ChainName string    `json:"chain_name,omitempty"`
	ChainID   string    `json:"chain_id,omitempty"`
}

// ── Jobs ──────────────────────────────────────────────────────────────────────

// Job is the full state of an async operation.
type Job struct {
	ID        string         `json:"id"`
	Status    JobStatus      `json:"status"`
	Operation string         `json:"operation"`
	Result    map[string]any `json:"result,omitempty"`
	Error     string         `json:"error,omitempty"`
	CreatedAt string         `json:"created_at"`
	UpdatedAt string         `json:"updated_at"`
}

// ── Stablecoin ────────────────────────────────────────────────────────────────

// StablecoinTransferPayload is the request body for a stablecoin transfer.
type StablecoinTransferPayload struct {
	Token          string        `json:"token"`
	To             string        `json:"to"`
	Amount         string        `json:"amount"`
	ChainName      string        `json:"chain_name,omitempty"`
	ChainID        string        `json:"chain_id,omitempty"`
	Gasless        bool          `json:"gasless,omitempty"`
	Speed          TransferSpeed `json:"speed,omitempty"`
	IdempotencyKey string        `json:"idempotency_key,omitempty"`
}

// StablecoinTransferResult is the response from a stablecoin transfer.
type StablecoinTransferResult struct {
	JobID  string `json:"job_id"`
	Status string `json:"status"`
}

// StablecoinBalanceResult is the response from a stablecoin balance lookup.
type StablecoinBalanceResult struct {
	Address    string `json:"address"`
	Token      string `json:"token"`
	Symbol     string `json:"symbol"`
	Balance    string `json:"balance"`
	RawBalance string `json:"raw_balance"`
	ChainID    string `json:"chain_id,omitempty"`
}