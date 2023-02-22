package bridge

import (
	"encoding/hex"
	"fmt"

	"github.com/ChainSafe/chainbridge-core/chains/tvm/calls/tvmclient"
	"github.com/ChainSafe/chainbridge-core/types"

	"github.com/fbsobreira/gotron-sdk/pkg/address"
	"github.com/fbsobreira/gotron-sdk/pkg/client"
	"github.com/fbsobreira/gotron-sdk/pkg/common"
	"github.com/rs/zerolog/log"
)

type BridgeContract struct {
	grpc            *client.GrpcClient
	signer          tvmclient.Signer
	contractAddress address.Address
}

func NewBridgeContract(grpc *client.GrpcClient, signer tvmclient.Signer, contractAddress address.Address) *BridgeContract {
	return &BridgeContract{grpc: grpc, signer: signer, contractAddress: contractAddress}
}

func (c *BridgeContract) GetHandlerAddressForResourceID(resourceID types.ResourceID) (address.Address, error) {
	resourceInHex := hex.EncodeToString(resourceID[:])
	log.Debug().Msgf("Getting handler address for resource %s", resourceInHex)

	jsonString := fmt.Sprintf(`[{"bytes32":%q}]`, resourceInHex)

	tx, err := c.grpc.TriggerContract(
		c.signer.CommonAddress().String(),
		c.contractAddress.String(),
		"_resourceIDToHandlerAddress(bytes32)",
		jsonString,
		0,
		0,
		"",
		0,
	)
	if err != nil {
		return nil, err
	}
	return append([]byte{address.TronBytePrefix}, common.TrimLeftZeroes(tx.ConstantResult[0])...), nil
}
