package events

import (
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/ChainSafe/chainbridge-core/types"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
)

//type EventSig string
//
//func (es EventSig) GetTopic() common.Hash {
//	return crypto.Keccak256Hash([]byte(es))
//}

// Deposit struct holds event data with all necessary parameters and a handler response
// https://github.com/ChainSafe/chainbridge-solidity/blob/develop/contracts/Bridge.sol#L47
type Deposit struct {
	// ID of chain deposit will be bridged to
	DestinationDomainID uint8
	// ResourceID used to find address of handler to be used for deposit
	ResourceID types.ResourceID
	// Nonce of deposit
	DepositNonce uint64
	// Address of sender (msg.sender: user)
	SenderAddress address.Address
	// Additional data to be passed to specified handler
	Data []byte
	// ERC20Handler: responds with empty data
	// ERC721Handler: responds with deposited token metadata acquired by calling a tokenURI method in the token contract
	// GenericHandler: responds with the raw bytes returned from the call to the target contract
	HandlerResponse []byte
}

// Deposit struct holds event data with all necessary parameters and a handler response
// https://github.com/ChainSafe/chainbridge-solidity/blob/develop/contracts/Bridge.sol#L47
type depositRaw struct {
	Field1              string `json:"0"`
	DestinationDomainID string `json:"destinationDomainID"`
	Field3              string `json:"1"`
	ResourceID          string `json:"resourceID"`
	Field5              string `json:"2"`
	DepositNonce        string `json:"depositNonce"`
	Field7              string `json:"3"`
	Field8              string `json:"4"`
	Data                string `json:"data"`
	Field10             string `json:"5"`
	HandlerResponse     string `json:"handlerResponse"`
	User                string `json:"user"`
}

func (d depositRaw) convert() (Deposit, error) {
	domainID, err := strconv.ParseUint(d.DestinationDomainID, 10, 64)
	if err != nil {
		return Deposit{}, err
	}
	resBts, err := hex.DecodeString(d.ResourceID)
	if err != nil {
		return Deposit{}, err
	}
	if len(resBts) < 32 {
		return Deposit{}, fmt.Errorf("resource id bytes length less than expected")
	}
	res := (*[32]byte)(resBts)

	depositNonce, err := strconv.ParseUint(d.DepositNonce, 10, 64)
	if err != nil {
		return Deposit{}, err
	}

	dataBts, err := hex.DecodeString(d.Data)
	if err != nil {
		return Deposit{}, err
	}
	handler, err := hex.DecodeString(d.HandlerResponse)
	if err != nil {
		return Deposit{}, err
	}
	return Deposit{
		DestinationDomainID: uint8(domainID),
		ResourceID:          *res,
		DepositNonce:        depositNonce,
		SenderAddress:       address.HexToAddress(d.User),
		Data:                dataBts,
		HandlerResponse:     handler,
	}, nil
}

type Common struct {
	BlockNumber           int    `json:"block_number"`
	BlockTimestamp        int64  `json:"block_timestamp"`
	CallerContractAddress string `json:"caller_contract_address"`
	ContractAddress       string `json:"contract_address"`
	EventIndex            int    `json:"event_index"`
	EventName             string `json:"event_name"`
	ResultType            struct {
		DestinationDomainID string `json:"destinationDomainID"`
		ResourceID          string `json:"resourceID"`
		DepositNonce        string `json:"depositNonce"`
		Data                string `json:"data"`
		HandlerResponse     string `json:"handlerResponse"`
		User                string `json:"user"`
	} `json:"result_type"`
	Event         string `json:"event"`
	TransactionId string `json:"transaction_id"`
}
