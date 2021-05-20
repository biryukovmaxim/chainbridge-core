// Copyright 2021 ChainSafe Systems
// SPDX-License-Identifier: LGPL-3.0-only

package listener

import (
	erc20Handler "github.com/ChainSafe/chainbridgev2/bindings/eth/bindings/ERC20Handler"
	erc721Handler "github.com/ChainSafe/chainbridgev2/bindings/eth/bindings/ERC721Handler"
	genericHandler "github.com/ChainSafe/chainbridgev2/bindings/eth/bindings/GenericHandler"
	"github.com/ChainSafe/chainbridgev2/chains/evm"
	"github.com/ChainSafe/chainbridgev2/relayer"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

func HandleErc20DepositedEvent(sourceID, destId uint8, nonce uint64, handlerContractAddress string, backend ChainReader) (relayer.XCMessager, error) {
	contract, err := erc20Handler.NewERC20HandlerCaller(common.HexToAddress(handlerContractAddress), backend)
	if err != nil {
		return nil, err
	}
	record, err := contract.GetDepositRecord(&bind.CallOpts{}, uint64(nonce), uint8(destId))
	if err != nil {
		return nil, err
	}

	return &evm.EVMMessage{
		Source:       sourceID,
		Destination:  destId,
		DepositNonce: nonce,
		ResourceId:   record.ResourceID,
		Payload: []interface{}{
			record.Amount.Bytes(),
			record.DestinationRecipientAddress,
		},
	}, nil
}

func HandleErc721DepositedEvent(sourceID, destId uint8, nonce uint64, handlerContractAddress string, backend ChainReader) (relayer.XCMessager, error) {
	contract, err := erc721Handler.NewERC721HandlerCaller(common.HexToAddress(handlerContractAddress), backend)
	if err != nil {
		return nil, err
	}
	record, err := contract.GetDepositRecord(&bind.CallOpts{}, uint64(nonce), uint8(destId))
	if err != nil {
		return nil, err
	}
	return &evm.EVMMessage{
		Source:       sourceID,
		Destination:  destId,
		DepositNonce: nonce,
		ResourceId:   record.ResourceID,
		Payload: []interface{}{
			record.TokenID.Bytes(),
			record.DestinationRecipientAddress,
			record.MetaData,
		},
	}, nil
}

func HandleGenericDepositedEvent(sourceID, destId uint8, nonce uint64, handlerContractAddress string, backend ChainReader) (relayer.XCMessager, error) {
	contract, err := genericHandler.NewGenericHandlerCaller(common.HexToAddress(handlerContractAddress), backend)
	if err != nil {
		return nil, err
	}
	record, err := contract.GetDepositRecord(&bind.CallOpts{}, uint64(nonce), uint8(destId))
	if err != nil {
		return nil, err
	}
	return &evm.EVMMessage{
		Source:       sourceID,
		Destination:  destId,
		DepositNonce: nonce,
		ResourceId:   record.ResourceID,
		Payload: []interface{}{
			record.MetaData,
		},
	}, nil
}
