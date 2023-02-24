package bridge

import "C"
import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/ChainSafe/chainbridge-core/chains/tvm/calls/tvmclient"
	"github.com/ChainSafe/chainbridge-core/chains/tvm/executor/proposal"
	"github.com/ChainSafe/chainbridge-core/relayer/message"
	"github.com/ChainSafe/chainbridge-core/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/fbsobreira/gotron-sdk/pkg/abi"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
	"github.com/fbsobreira/gotron-sdk/pkg/client"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/api"
	"github.com/golang/protobuf/proto"
	"github.com/mitchellh/mapstructure"
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

func (c *BridgeContract) IsProposalVotedBy(by address.Address, p *proposal.Proposal) (bool, error) {
	log.Debug().
		Str("depositNonce", strconv.FormatUint(p.DepositNonce, 10)).
		Str("resourceID", hexutil.Encode(p.ResourceId[:])).
		Str("handler", p.HandlerAddress.String()).
		Msgf("Getting is proposal voted by %s", by.String())

	params := []abi.Param{
		{"uint72": idAndNonce(p.Source, p.DepositNonce)},
		{"bytes32": hex.EncodeToString(p.GetDataHash().Bytes())},
		{"address": by.String()},
	}
	tx, err := c.grpc.TriggerContractV2(
		c.signer.CommonAddress().String(),
		c.contractAddress.String(),
		"_hasVotedOnProposal(uint72,bytes32,address)",
		params,
		0,
		0,
		"",
		0,
	)
	if len(tx.Result.Message) != 0 {
		log.Error().Bytes("message", tx.Result.Message)
	}
	if err != nil {
		return false, err
	}
	return len(common.TrimLeftZeroes(tx.ConstantResult[0])) != 0, nil
}

func (c *BridgeContract) VoteProposal(proposal *proposal.Proposal) (*common.Hash, error) {
	log.Debug().
		Str("depositNonce", strconv.FormatUint(proposal.DepositNonce, 10)).
		Str("resourceID", hexutil.Encode(proposal.ResourceId[:])).
		Str("handler", proposal.HandlerAddress.String()).
		Msgf("Vote proposal")

	exTx, err := c.voteProposal(proposal)
	if err != nil {
		return nil, err
	}
	if len(exTx.Result.Message) != 0 {
		log.Error().Bytes("message", exTx.Result.Message)
	}
	tx := exTx.Transaction
	data, err := proto.Marshal(tx.GetRawData())
	if err != nil {
		return nil, err
	}
	h256h := sha256.New()
	h256h.Write(data)
	txHash := common.BytesToHash(h256h.Sum(nil))
	signature, err := c.signer.Sign(txHash.Bytes())
	if err != nil {
		return nil, err
	}
	tx.Signature = append(tx.Signature, signature)
	result, err := c.grpc.Broadcast(tx)
	if err != nil {
		return nil, err
	}
	if result.Code != 0 {
		return nil, fmt.Errorf("bad transaction: %v", string(result.GetMessage()))
	}

	return &txHash, c.txConfirmation(txHash.Hex())
}

func (c *BridgeContract) txConfirmation(txHash string) error {
	for i := 0; i < 50; i++ {
		if _, err := c.grpc.GetTransactionInfoByID(txHash); err == nil {
			return nil
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("tx did not appear")
}

func (c *BridgeContract) voteProposal(p *proposal.Proposal) (*api.TransactionExtention, error) {
	params := []abi.Param{
		{"uint8": p.Source},
		{"uint64": p.DepositNonce},
		{"bytes32": hex.EncodeToString(p.ResourceId[:])},
		{"bytes": hex.EncodeToString(p.Data)},
	}
	tx, err := c.grpc.TriggerContractV2(
		c.signer.CommonAddress().String(),
		c.contractAddress.String(),
		"voteProposal(uint8,uint64,bytes32,bytes)",
		params,
		4000000000,
		0,
		"",
		0,
	)
	if len(tx.Result.Message) != 0 {
		log.Error().Bytes("message", tx.Result.Message)
	}
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (c *BridgeContract) SimulateVoteProposal(p *proposal.Proposal) error {
	log.Debug().
		Str("depositNonce", strconv.FormatUint(p.DepositNonce, 10)).
		Str("resourceID", hexutil.Encode(p.ResourceId[:])).
		Str("handler", p.HandlerAddress.String()).
		Msgf("Simulate vote proposal")

	_, err := c.voteProposal(p)

	return err
}

func (c *BridgeContract) ProposalStatus(p *proposal.Proposal) (message.ProposalStatus, error) {
	log.Debug().
		Str("depositNonce", strconv.FormatUint(p.DepositNonce, 10)).
		Str("resourceID", hexutil.Encode(p.ResourceId[:])).
		Str("handler", p.HandlerAddress.String()).
		Msg("Getting proposal status")

	params := []abi.Param{
		{"uint8": p.Source},
		{"uint64": p.DepositNonce},
		{"bytes32": hex.EncodeToString(p.GetDataHash().Bytes())},
	}
	tx, err := c.grpc.TriggerContractV2(
		c.signer.CommonAddress().String(),
		c.contractAddress.String(),
		"getProposal(uint8,uint64,bytes32)",
		params,
		0,
		0,
		"",
		0,
	)
	if err != nil {
		return message.ProposalStatus{}, err
	}
	a, err := c.grpc.GetContractABI(c.contractAddress.String())
	if err != nil {
		return message.ProposalStatus{}, err
	}
	arg, err := abi.GetParser(a, "getProposal")
	if err != nil {
		return message.ProposalStatus{}, err
	}
	asMap := map[string]interface{}{}
	err = arg.UnpackIntoMap(asMap, tx.ConstantResult[0])
	if err != nil {
		return message.ProposalStatus{}, err
	}
	var asStruct message.ProposalStatus
	if err := mapstructure.Decode(asMap, &asStruct); err != nil {
		return message.ProposalStatus{}, err
	}

	return asStruct, nil
}

func (c *BridgeContract) GetThreshold() (uint8, error) {
	log.Debug().Msg("Getting threshold")

	tx, err := c.grpc.TriggerContract(
		c.signer.CommonAddress().String(),
		c.contractAddress.String(),
		"_relayerThreshold()",
		"",
		0,
		0,
		"",
		0,
	)
	if err != nil {
		return 0, err
	}
	if len(tx.Result.Message) != 0 {
		log.Error().Bytes("message", tx.Result.Message)
	}
	lastIdx := len(tx.ConstantResult[0]) - 1
	return tx.ConstantResult[0][lastIdx], nil
}

func (c *BridgeContract) ContractAddress() *address.Address {
	return &c.contractAddress
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
	if len(tx.Result.Message) != 0 {
		log.Error().Bytes("message", tx.Result.Message)
	}
	return append([]byte{address.TronBytePrefix}, common.TrimLeftZeroes(tx.ConstantResult[0])...), nil
}

func idAndNonce(srcId uint8, nonce uint64) *big.Int {
	var data []byte
	data = append(data, big.NewInt(int64(nonce)).Bytes()...)
	data = append(data, srcId)
	return big.NewInt(0).SetBytes(data)
}
