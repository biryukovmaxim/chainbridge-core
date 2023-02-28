package bridge

import (
	"os"
	"testing"

	"github.com/ChainSafe/chainbridge-core/chains/tvm/calls/tvmclient"
	"github.com/ChainSafe/chainbridge-core/chains/tvm/executor/proposal"
	secp256k12 "github.com/ChainSafe/chainbridge-core/crypto/secp256k1"
	"github.com/ChainSafe/chainbridge-core/relayer/message"
	"github.com/ChainSafe/chainbridge-core/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	secp256k1 "github.com/ethereum/go-ethereum/crypto"
	common "github.com/fbsobreira/gotron-sdk/pkg/address"
	tronClient "github.com/fbsobreira/gotron-sdk/pkg/client"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

var (
	bridge  *BridgeContract
	signer  tvmclient.Signer
	handler common.Address
)

const (
	bridgeHex     = "TPH9cWgafMHhGmzL3ccaWX5gF7e8kbicZr"
	network       = "grpc.shasta.trongrid.io:50051"
	handlerBase58 = "TBq9Rc5mPtq7tLHBxnHUXGkuaEDxrKX3ya"
	resourceHex   = "0x6d792d746f6b656e000000000000000000000000000000000000000000000000"
	private       = "96d3b1ee6ddead968abeb3dd064fd38c1f4cc08cf1aca5c1c9e6321fe8ab9207"
)

func TestMain(m *testing.M) {
	client := tronClient.NewGrpcClient(network)

	err := client.Start(grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer client.Stop()

	bridgeAddres, err := common.Base58ToAddress(bridgeHex)
	if err != nil {
		panic(err)
	}
	privateKey, err := secp256k1.HexToECDSA(private)
	if err != nil {
		panic(err)
	}
	handler, err = common.Base58ToAddress(handlerBase58)
	if err != nil {
		panic(err)
	}
	signer = tvmclient.NewAdapter(secp256k12.NewKeypair(*privateKey))

	bridge = NewBridgeContract(client, signer, bridgeAddres)
	os.Exit(m.Run())
}

func TestBridgeContract_SimulateVoteProposal(t *testing.T) {
	err := bridge.SimulateVoteProposal(&proposal.Proposal{
		Source:         0,
		Destination:    0,
		DepositNonce:   0,
		ResourceId:     types.ResourceID{},
		Metadata:       message.Metadata{},
		Data:           nil,
		HandlerAddress: handler,
		BridgeAddress:  *bridge.ContractAddress(),
	})
	require.NoError(t, err)
}

func TestBridgeContract_ProposalStatus(t *testing.T) {
	_, err := bridge.ProposalStatus(&proposal.Proposal{
		Source:         0,
		Destination:    0,
		DepositNonce:   0,
		ResourceId:     types.ResourceID{},
		Metadata:       message.Metadata{},
		Data:           nil,
		HandlerAddress: handler,
		BridgeAddress:  *bridge.ContractAddress(),
	})
	require.NoError(t, err)

}

func TestBridgeContract_GetThreshold(t *testing.T) {
	count, err := bridge.GetThreshold()
	require.NoError(t, err)
	require.Equal(t, count, uint8(1))
}

func TestGetHandlerForResourceID(t *testing.T) {
	resource, err := hexutil.Decode(resourceHex)
	require.NoError(t, err)

	actualAddress, err := bridge.GetHandlerAddressForResourceID(toFixed(resource))
	require.NoError(t, err)
	require.Equal(t, handler.String(), actualAddress.String())
}

func TestBridgeContract_IsProposalVotedBy(t *testing.T) {
	voted, err := bridge.IsProposalVotedBy(signer.CommonAddress(), &proposal.Proposal{
		Source:         0,
		Destination:    0,
		DepositNonce:   0,
		ResourceId:     types.ResourceID{},
		Metadata:       message.Metadata{},
		Data:           nil,
		HandlerAddress: handler,
		BridgeAddress:  *bridge.ContractAddress(),
	})
	require.NoError(t, err)
	require.False(t, voted)
}

func toFixed(in []byte) [32]byte {
	res := (*[32]byte)(in)
	return *res
}
