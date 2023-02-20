package tvmclient

import (
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/fbsobreira/gotron-sdk/pkg/client"
	"google.golang.org/grpc"
)

type TronClient struct {
	grpc      *client.GrpcClient
	signer    Signer
	nonce     *big.Int
	nonceLock sync.Mutex
}

func (c *TronClient) Start() error {
	return c.grpc.Start(grpc.WithInsecure())
}
func (c *TronClient) Stop() {
	c.grpc.Stop()
}

func NewTronClient(network string) *TronClient {
	return &TronClient{grpc: client.NewGrpcClient("grpc.shasta.trongrid.io:50051")}
}

type Signer interface {
	CommonAddress() common.Address

	// Sign calculates an ECDSA signature.
	// The produced signature must be in the [R || S || V] format where V is 0 or 1.
	Sign(digestHash []byte) ([]byte, error)
}

type CommonTransaction interface {
	// Hash returns the transaction hash.
	Hash() common.Hash

	// RawWithSignature Returns signed transaction by provided signer
	RawWithSignature(signer Signer, domainID *big.Int) ([]byte, error)
}

// LatestBlock returns the latest block timestamp from the current chain
func (c *TronClient) LatestBlock() (number *big.Int, timestamp time.Time, err error) {
	b, err := c.grpc.GetNowBlock()
	if err != nil {
		return
	}
	number = big.NewInt(b.BlockHeader.RawData.Number)
	timestamp = time.UnixMilli(b.BlockHeader.RawData.Timestamp)
	return
}

// BlockTsByNum returns the timestamp from the specified block number
func (c *TronClient) BlockTsByNum(num int64) (time.Time, error) {
	b, err := c.grpc.GetBlockByNum(num)
	if err != nil {
		return time.Time{}, err
	}
	return time.UnixMilli(b.BlockHeader.RawData.Timestamp), nil
}
