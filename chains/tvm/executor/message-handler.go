package executor

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"

	"github.com/ChainSafe/chainbridge-core/chains/tvm/executor/proposal"
	"github.com/ChainSafe/chainbridge-core/chains/tvm/utils"
	"github.com/ChainSafe/chainbridge-core/relayer/message"
	"github.com/ChainSafe/chainbridge-core/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
	"github.com/rs/zerolog/log"
)

type HandlerMatcher interface {
	GetHandlerAddressForResourceID(resourceID types.ResourceID) (address.Address, error)
	ContractAddress() *address.Address
}

type MessageHandlerFunc func(m *message.Message, handlerAddr, bridgeAddress address.Address) (*proposal.Proposal, error)

// NewTVMMessageHandler creates an instance of TVMMessageHandler that contains
// message handler functions for converting deposit message into a chain specific
// proposal
func NewTVMMessageHandler(handlerMatcher HandlerMatcher) *TVMMessageHandler {
	return &TVMMessageHandler{
		handlerMatcher: handlerMatcher,
	}
}

type TVMMessageHandler struct {
	handlerMatcher HandlerMatcher
	handlers       map[[21]byte]MessageHandlerFunc
}

func (mh *TVMMessageHandler) HandleMessage(m *message.Message) (*proposal.Proposal, error) {
	// Matching resource ID with handler.
	addr, err := mh.handlerMatcher.GetHandlerAddressForResourceID(m.ResourceId)
	if err != nil {
		return nil, err
	}
	// Based on handler that registered on BridgeContract
	handleMessage, err := mh.MatchAddressWithHandlerFunc(addr)
	if err != nil {
		return nil, err
	}
	log.Info().Str("type", string(m.Type)).Uint8("src", m.Source).Uint8("dst", m.Destination).Uint64("nonce", m.DepositNonce).Str("resourceID", fmt.Sprintf("%x", m.ResourceId)).Msg("Handling new message")
	prop, err := handleMessage(m, addr, *mh.handlerMatcher.ContractAddress())
	if err != nil {
		return nil, err
	}
	return prop, nil
}

func (mh *TVMMessageHandler) MatchAddressWithHandlerFunc(addr address.Address) (MessageHandlerFunc, error) {
	h, ok := mh.handlers[utils.ToFixed(addr)]
	if !ok {
		return nil, fmt.Errorf("no corresponding message handler for this address %s exists", addr.String())
	}
	return h, nil
}

// RegisterMessageHandler registers a message handler by associating a handler function to a specified address
func (mh *TVMMessageHandler) RegisterMessageHandler(addrBase58 string, handler MessageHandlerFunc) {
	if addrBase58 == "" {
		return
	}
	if mh.handlers == nil {
		mh.handlers = make(map[[21]byte]MessageHandlerFunc)
	}

	log.Info().Msgf("Registered message handler for address %s", addrBase58)
	addr, _ := address.Base58ToAddress(addrBase58)
	mh.handlers[utils.ToFixed(addr)] = handler
}

func ERC20MessageHandler(m *message.Message, handlerAddr, bridgeAddress address.Address) (*proposal.Proposal, error) {
	if len(m.Payload) != 2 {
		return nil, errors.New("malformed payload. Len  of payload should be 2")
	}
	amount, ok := m.Payload[0].([]byte)
	if !ok {
		return nil, errors.New("wrong payload amount format")
	}
	recipient, ok := m.Payload[1].([]byte)
	if !ok {
		return nil, errors.New("wrong payload recipient format")
	}
	var data []byte
	data = append(data, common.LeftPadBytes(amount, 32)...) // amount (uint256)
	recipientLen := big.NewInt(int64(len(recipient))).Bytes()
	data = append(data, common.LeftPadBytes(recipientLen, 32)...) // length of recipient (uint256)
	data = append(data, recipient...)                             // recipient ([]byte)
	return proposal.NewProposal(m.Source, m.Destination, m.DepositNonce, m.ResourceId, data, handlerAddr, bridgeAddress, m.Metadata), nil
}

func ERC721MessageHandler(msg *message.Message, handlerAddr, bridgeAddress address.Address) (*proposal.Proposal, error) {
	if len(msg.Payload) != 3 {
		return nil, errors.New("malformed payload. Len  of payload should be 3")
	}
	tokenID, ok := msg.Payload[0].([]byte)
	if !ok {
		return nil, errors.New("wrong payload tokenID format")
	}
	recipient, ok := msg.Payload[1].([]byte)
	if !ok {
		return nil, errors.New("wrong payload recipient format")
	}
	metadata, ok := msg.Payload[2].([]byte)
	if !ok {
		return nil, errors.New("wrong payload metadata format")
	}
	data := bytes.Buffer{}
	data.Write(common.LeftPadBytes(tokenID, 32))
	recipientLen := big.NewInt(int64(len(recipient))).Bytes()
	data.Write(common.LeftPadBytes(recipientLen, 32))
	data.Write(recipient)
	metadataLen := big.NewInt(int64(len(metadata))).Bytes()
	data.Write(common.LeftPadBytes(metadataLen, 32))
	data.Write(metadata)
	return proposal.NewProposal(msg.Source, msg.Destination, msg.DepositNonce, msg.ResourceId, data.Bytes(), handlerAddr, bridgeAddress, msg.Metadata), nil
}

func GenericMessageHandler(msg *message.Message, handlerAddr, bridgeAddress address.Address) (*proposal.Proposal, error) {
	if len(msg.Payload) != 1 {
		return nil, errors.New("malformed payload. Len  of payload should be 1")
	}
	metadata, ok := msg.Payload[0].([]byte)
	if !ok {
		return nil, errors.New("wrong payload metadata format")
	}
	data := bytes.Buffer{}
	metadataLen := big.NewInt(int64(len(metadata))).Bytes()
	data.Write(common.LeftPadBytes(metadataLen, 32)) // length of metadata (uint256)
	data.Write(metadata)
	return proposal.NewProposal(msg.Source, msg.Destination, msg.DepositNonce, msg.ResourceId, data.Bytes(), handlerAddr, bridgeAddress, msg.Metadata), nil
}
