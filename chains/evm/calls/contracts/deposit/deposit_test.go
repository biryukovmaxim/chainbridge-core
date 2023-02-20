package deposit

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

const data = "00000000000000000000000000000000000000000000071cc833b3e9e905e0000000000000000000000000000000000000000000000000000000000000000014a79e8a38afda1f6106565611156655dc779499b7"

func TestDepositDecoding(t *testing.T) {
	//data = append(data, math.PaddedBigBytes(big.NewInt(int64(len(destRecipient))), 32)...) // length of recipient
	//data = append(data, destRecipient...)
	//fmt.Println(len(data))

	bts := common.Hex2Bytes(data)
	r := bytes.NewReader(bts)
	amountBts := make([]byte, 32)
	r.Read(amountBts)
	recipientLenghtBts := make([]byte, 32)
	r.Read(recipientLenghtBts)
	addressBts := make([]byte, 20)
	r.Read(addressBts)

	fmt.Println(common.IsHexAddress(common.BytesToAddress(addressBts).Hex()))
}
