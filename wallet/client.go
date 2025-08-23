package wallet

import (
	"context"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/sairaviteja27/nova-infra-task/types"
)

type Client struct {
	Endpoint   string
	HTTPClient *http.Client
}

func NewClient(endpoint string) *Client {
	return &Client{
		Endpoint:   endpoint,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) Fetch(ctx context.Context, addr string) (types.Result, error) {
	if addr == "" {
		return types.Result{}, fmt.Errorf("address is empty")
	}
	client := rpc.New(c.Endpoint)

	pubKey := solana.MustPublicKeyFromBase58(addr)
	out, err := client.GetBalance(
		context.TODO(),
		pubKey,
		rpc.CommitmentFinalized,
	)
	if err != nil {
		return types.Result{}, fmt.Errorf("error fetching balance: %w", err)
	}
	// spew.Dump(out)
	// spew.Dump(out.Value) // total lamports on the account; 1 sol = 1000000000 lamports

	var lamportsOnAccount = new(big.Float).SetUint64(uint64(out.Value))
	// Convert lamports to sol:
	var solBalance = new(big.Float).Quo(lamportsOnAccount, new(big.Float).SetUint64(solana.LAMPORTS_PER_SOL))

	return types.Result{
		WalletAddress: addr,
		Balance:       solBalance.Text('f', 10),
	}, nil
}
