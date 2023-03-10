package tvmclient

import (
	"github.com/ChainSafe/chainbridge-core/chains/evm/calls/evmclient"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
)

type Adapter struct {
	evmclient.Signer
}

func NewAdapter(signer evmclient.Signer) *Adapter {
	return &Adapter{Signer: signer}
}

func (a Adapter) CommonAddress() address.Address {
	return append([]byte{address.TronBytePrefix}, a.Signer.CommonAddress().Bytes()[:]...)
}
