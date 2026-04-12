package vaultkey

import "context"

// ChainsService lists supported chains for the current environment.
type ChainsService struct {
	client *Client
}

// List returns all supported EVM chains for the current environment.
// Testnet keys return testnet chains; live keys return mainnet chains.
//
//	chains, apiErr, err := client.Chains.List(ctx)
//	for _, c := range chains {
//	    fmt.Println(c.Name, c.ChainID)
//	}
func (c *ChainsService) List(ctx context.Context) ([]Chain, *ErrorResponse, error) {
	var resp []Chain
	apiErr, err := c.client.get(ctx, "/chains", &resp)
	return resp, apiErr, err
}