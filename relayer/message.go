// Copyright 2021 ChainSafe Systems
// SPDX-License-Identifier: LGPL-3.0-only

package relayer

import "errors"

type TransferType string

const (
	FungibleTransfer    TransferType = "FungibleTransfer"
	NonFungibleTransfer TransferType = "NonFungibleTransfer"
	GenericTransfer     TransferType = "GenericTransfer"
)

type ProposalStatus uint8

const (
	ProposalStatusInactive ProposalStatus = iota
	ProposalStatusActive
	ProposalStatusPassed // Ready to be executed
	ProposalStatusExecuted
	ProposalStatusCanceled
)

var (
	StatusMap = map[ProposalStatus]string{ProposalStatusInactive: "inactive", ProposalStatusActive: "active", ProposalStatusPassed: "passed", ProposalStatusExecuted: "executed", ProposalStatusCanceled: "canceled"}
)

type Message struct {
	Source       uint8  // Source where message was initiated
	Destination  uint8  // Destination chain of message
	DepositNonce uint64 // Nonce for the deposit
	ResourceId   [32]byte
	Payload      []interface{} // data associated with event sequence
	Type         TransferType
}

// extractAmountTransferred is a private method to extract and transform the transfer amount
// from the Payload field within the Message struct
func (m *Message) extractAmountTransferred() (int, error) {
	// parse payload field from event log message to obtain transfer amount
	// payload slice of interfaces includes..
	// index 0: amount ([]byte)
	// index 1: destination recipient address ([]byte)

	// cast interface to byte slice
	amountByteSlice, ok := m.Payload[0].([]byte)
	if !ok {
		err := errors.New("could not cast interface to byte slice")
		return 0, err
	}

	return int(amountByteSlice[0]), nil
}
