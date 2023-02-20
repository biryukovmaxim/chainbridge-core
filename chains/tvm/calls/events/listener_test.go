package events

import (
	"context"
	"testing"
	"time"

	"github.com/fbsobreira/gotron-sdk/pkg/address"
	"github.com/stretchr/testify/require"
)

const (
	BridgeAddress = "TZ2xAEKqHup6hzEQGPhmXQFiXBDBQxSVZG"
	testnet       = "https://api.shasta.trongrid.io/v1"
	testKey       = "bc78f8bb-2e8d-4e18-9165-e4777e6e3fb6"
)

func TestFetchDeposits(t *testing.T) {
	l := NewListener(testnet, testKey)
	addr, err := address.Base58ToAddress(BridgeAddress)
	require.NoError(t, err)
	startTime := time.UnixMilli(1676642178000)
	endTime := time.UnixMilli(1676642178000)
	deposits, err := l.FetchDeposits(context.TODO(), addr, &startTime, &endTime)
	require.NoError(t, err)
	require.Greater(t, len(deposits), 0)
	deposit := deposits[0]

	expectedRes := make([]byte, 32)
	copy(expectedRes, "some text")
	expectedResFixed := (*[32]byte)(expectedRes)
	expectedCreator, err := address.Base58ToAddress("TE2gVgMNYvD6UABUUSYeLk1y6ePhrBer7Q")
	require.NoError(t, err)
	expectedCreator = expectedCreator.Bytes()[1:]

	require.Equal(t, deposit, Deposit{
		DestinationDomainID: 0,
		ResourceID:          *expectedResFixed,
		DepositNonce:        1,
		SenderAddress:       expectedCreator,
		Data:                []byte("some data"),
		HandlerResponse:     []byte("handler"),
	})
}
