package deposit

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math/big"
	"testing"

	callsUtil "github.com/ChainSafe/chainbridge-core/chains/evm/calls"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
	"github.com/stretchr/testify/require"
)

const data = "00000000000000000000000000000000000000000000071cc833b3e9e905e0000000000000000000000000000000000000000000000000000000000000000014a79e8a38afda1f6106565611156655dc779499b7"

func TestDepositDecoding(t *testing.T) {
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

func TestEncodingTrcDeposit(t *testing.T) {
	base58 := "TE2gVgMNYvD6UABUUSYeLk1y6ePhrBer7Q"
	amount := "0.0002"
	addr, err := address.Base58ToAddress(base58)
	require.NoError(t, err)
	realAmount, err := callsUtil.UserAmountToWei(amount, big.NewInt(18))
	require.NoError(t, err)

	data := ConstructErc20DepositData(addr.Bytes()[1:], realAmount)
	hexData := hexutil.Encode(data)
	fmt.Println(hexData)

	h256h := sha256.New()
	h256h.Write(data)
	hash := h256h.Sum(nil)
	hexHash := hexutil.Encode(hash)
	fmt.Println(hexHash)
}

func TestEncodingErcDeposit(t *testing.T) {
	hexed := "0xB1184d7c47eccAE188726A2C3C9AA8E2151B227A"
	amount := "0.0012"
	addr := common.HexToAddress(hexed)
	realAmount, err := callsUtil.UserAmountToWei(amount, big.NewInt(18))
	//fmt.Println(realAmount.String())
	realAmount = big.NewInt(199999999994000)
	require.NoError(t, err)

	data := ConstructErc20DepositData(addr.Bytes(), realAmount)
	hexData := hexutil.Encode(data)
	fmt.Println(hexData)

	h256h := sha256.New()
	h256h.Write(data)
	hash := h256h.Sum(nil)
	hexHash := hexutil.Encode(hash)
	fmt.Println(hexHash)
}
