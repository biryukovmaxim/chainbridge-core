// Copyright 2021 ChainSafe Systems
// SPDX-License-Identifier: LGPL-3.0-only

package app

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ChainSafe/chainbridge-core/chains/evm"
	"github.com/ChainSafe/chainbridge-core/chains/evm/calls/contracts/bridge"
	"github.com/ChainSafe/chainbridge-core/chains/evm/calls/events"
	"github.com/ChainSafe/chainbridge-core/chains/evm/calls/evmclient"
	"github.com/ChainSafe/chainbridge-core/chains/evm/calls/evmtransaction"
	"github.com/ChainSafe/chainbridge-core/chains/evm/calls/transactor/signAndSend"
	"github.com/ChainSafe/chainbridge-core/chains/evm/executor"
	"github.com/ChainSafe/chainbridge-core/chains/evm/listener"
	"github.com/ChainSafe/chainbridge-core/chains/tvm"
	tvmEvents "github.com/ChainSafe/chainbridge-core/chains/tvm/calls/events"
	"github.com/ChainSafe/chainbridge-core/chains/tvm/calls/tvmclient"
	tvmListener "github.com/ChainSafe/chainbridge-core/chains/tvm/listener"
	"github.com/ChainSafe/chainbridge-core/config"
	"github.com/ChainSafe/chainbridge-core/config/chain"
	secp256k12 "github.com/ChainSafe/chainbridge-core/crypto/secp256k1"
	"github.com/ChainSafe/chainbridge-core/e2e/dummy"
	"github.com/ChainSafe/chainbridge-core/flags"
	"github.com/ChainSafe/chainbridge-core/lvldb"
	"github.com/ChainSafe/chainbridge-core/opentelemetry"
	"github.com/ChainSafe/chainbridge-core/relayer"
	"github.com/ChainSafe/chainbridge-core/store"
	"github.com/ethereum/go-ethereum/common"
	secp256k1 "github.com/ethereum/go-ethereum/crypto"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func Run() error {
	configuration, err := config.GetConfig(viper.GetString(flags.ConfigFlagName))
	if err != nil {
		panic(err)
	}

	db, err := lvldb.NewLvlDB(viper.GetString(flags.BlockstoreFlagName))
	if err != nil {
		panic(err)
	}
	blockstore := store.NewBlockStore(db)

	chains := []relayer.RelayedChain{}
	for _, chainConfig := range configuration.ChainConfigs {
		switch chainConfig["type"] {
		case "evm":
			{
				config, err := chain.NewEVMConfig(chainConfig)
				if err != nil {
					panic(err)
				}

				privateKey, err := secp256k1.HexToECDSA(config.GeneralChainConfig.Key)
				if err != nil {
					panic(err)
				}

				kp := secp256k12.NewKeypair(*privateKey)
				fmt.Println(kp.Address())
				client, err := evmclient.NewEVMClient(config.GeneralChainConfig.Endpoint, kp)
				if err != nil {
					panic(err)
				}

				dummyGasPricer := dummy.NewStaticGasPriceDeterminant(client, nil)
				t := signAndSend.NewSignAndSendTransactor(evmtransaction.NewTransaction, dummyGasPricer, client)
				bridgeContract := bridge.NewBridgeContract(client, common.HexToAddress(config.Bridge), t)

				depositHandler := listener.NewETHDepositHandler(bridgeContract)
				depositHandler.RegisterDepositHandler(config.Erc20Handler, listener.Erc20DepositHandler)
				depositHandler.RegisterDepositHandler(config.Erc721Handler, listener.Erc721DepositHandler)
				depositHandler.RegisterDepositHandler(config.GenericHandler, listener.GenericDepositHandler)
				eventListener := events.NewListener(client)
				eventHandlers := make([]listener.EventHandler, 0)
				eventHandlers = append(eventHandlers, listener.NewDepositEventHandler(eventListener, depositHandler, common.HexToAddress(config.Bridge), *config.GeneralChainConfig.Id))
				evmListener := listener.NewEVMListener(client, eventHandlers, blockstore, *config.GeneralChainConfig.Id, config.BlockRetryInterval, config.BlockConfirmations, config.BlockInterval)

				mh := executor.NewEVMMessageHandler(bridgeContract)
				mh.RegisterMessageHandler(config.Erc20Handler, executor.ERC20MessageHandler)
				mh.RegisterMessageHandler(config.Erc721Handler, executor.ERC721MessageHandler)
				mh.RegisterMessageHandler(config.GenericHandler, executor.GenericMessageHandler)

				var evmVoter *executor.EVMVoter
				evmVoter, err = executor.NewVoterWithSubscription(mh, client, bridgeContract)
				if err != nil {
					log.Error().Msgf("failed creating voter with subscription: %s. Falling back to default voter.", err.Error())
					evmVoter = executor.NewVoter(mh, client, bridgeContract)
				}

				chain := evm.NewEVMChain(evmListener, evmVoter, blockstore, *config.GeneralChainConfig.Id, config.StartBlock, config.GeneralChainConfig.LatestBlock, config.GeneralChainConfig.FreshStart)

				chains = append(chains, chain)
			}
		case "tvm":
			const (
				BridgeAddress = "TPH9cWgafMHhGmzL3ccaWX5gF7e8kbicZr"
				testnet       = "https://api.shasta.trongrid.io/v1"
				testKey       = "bc78f8bb-2e8d-4e18-9165-e4777e6e3fb6"
			)
			eventListener := tvmEvents.NewFetcher(testnet, testKey)
			depositHandler := tvmListener.NewTronDepositHandler(tvm.DummyMatcher{})
			//depositHandler.RegisterDepositHandler(config.Erc20Handler, listener.Erc20DepositHandler)
			//depositHandler.RegisterDepositHandler(config.Erc721Handler, listener.Erc721DepositHandler)
			//depositHandler.RegisterDepositHandler(config.GenericHandler, listener.GenericDepositHandler)
			eventHandlers := make([]tvmListener.EventHandler, 0)
			brAddr, _ := address.Base58ToAddress(BridgeAddress)
			eventHandlers = append(eventHandlers, tvmListener.NewDepositEventHandler(eventListener, depositHandler, brAddr, 2))
			client := tvmclient.NewTronClient("")
			if err := client.Start(); err != nil {
				log.Panic().Err(err)
			}
			defer client.Stop()
			evmListener := tvmListener.NewTVMListener(client, eventHandlers, blockstore, 2, time.Duration(0), big.NewInt(1), big.NewInt(10))

			newChain := tvm.NewTVMChain(evmListener, tvm.DummyExecutor{}, blockstore, 2, big.NewInt(31494927), false, false)
			chains = append(chains, newChain)
		default:
			log.Warn().Msgf("type '%s' not recognized", chainConfig["type"])
			continue
		}
	}

	r := relayer.NewRelayer(
		chains,
		&opentelemetry.ConsoleTelemetry{},
	)

	errChn := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go r.Start(ctx, errChn)

	sysErr := make(chan os.Signal, 1)
	signal.Notify(sysErr,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGHUP,
		syscall.SIGQUIT)

	select {
	case err := <-errChn:
		log.Error().Err(err).Msg("failed to listen and serve")
		return err
	case sig := <-sysErr:
		log.Info().Msgf("terminating got ` [%v] signal", sig)
		return nil
	}
}
