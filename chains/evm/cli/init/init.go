package init

import (
	"fmt"
	"github.com/ChainSafe/chainbridge-core/chains/evm/calls/client"
	"github.com/ChainSafe/chainbridge-core/chains/evm/calls/transactor"
	"github.com/ChainSafe/chainbridge-core/chains/evm/evmclient"
	"github.com/ChainSafe/chainbridge-core/chains/evm/evmgaspricer"
	"github.com/ChainSafe/chainbridge-core/crypto/secp256k1"
	"github.com/rs/zerolog/log"
	"math/big"
)

func InitializeClient(
	url string,
	senderKeyPair *secp256k1.Keypair,
) (*evmclient.EVMClient, error) {
	ethClient, err := evmclient.NewEVMClientFromParams(
		url, senderKeyPair.PrivateKey())
	if err != nil {
		log.Error().Err(fmt.Errorf("eth client intialization error: %v", err))
		return nil, err
	}
	return ethClient, nil
}

func InitializeTransactor(
	gasPrice *big.Int,
	txFabric client.TxFabric,
	client *evmclient.EVMClient,
) (transactor.Transactor, error) {
	gasPricer := evmgaspricer.NewLondonGasPriceClient(
		client,
		&evmgaspricer.GasPricerOpts{UpperLimitFeePerGas: gasPrice},
	)

	trans := transactor.NewSignAndSendTransactor(txFabric, gasPricer, client)
	return trans, nil
}
