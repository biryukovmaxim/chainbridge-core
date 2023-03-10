// Copyright 2021 ChainSafe Systems
// SPDX-License-Identifier: LGPL-3.0-only

package listener

import (
	"context"
	"math/big"
	"time"

	"github.com/ChainSafe/chainbridge-core/relayer/message"
	"github.com/ChainSafe/chainbridge-core/store"

	"github.com/rs/zerolog/log"
)

type EventHandler interface {
	HandleEvent(startTime, endTime *time.Time, msgChan chan []*message.Message) error
}

type ChainClient interface {
	LatestBlock() (number *big.Int, timestamp time.Time, err error)
	BlockTsByNum(num int64) (time.Time, error)
}

type TVMListener struct {
	client        ChainClient
	eventHandlers []EventHandler

	domainID           uint8
	blockstore         *store.BlockStore
	blockRetryInterval time.Duration
	blockConfirmations *big.Int
	blockInterval      *big.Int
}

// NewTVMListener creates an TVMListener that listens to deposit events on chain
// and calls event handler when one occurs
func NewTVMListener(
	client ChainClient,
	eventHandlers []EventHandler,
	blockstore *store.BlockStore,
	domainID uint8,
	blockRetryInterval time.Duration,
	blockConfirmations *big.Int,
	blockInterval *big.Int) *TVMListener {
	return &TVMListener{
		client:             client,
		eventHandlers:      eventHandlers,
		blockstore:         blockstore,
		domainID:           domainID,
		blockRetryInterval: blockRetryInterval,
		blockConfirmations: blockConfirmations,
		blockInterval:      blockInterval,
	}
}

// ListenToEvents goes block by block of a network and executes event handlers that are
// configured for the listener.
func (l *TVMListener) ListenToEvents(ctx context.Context, startBlock *big.Int, msgChan chan []*message.Message, errChn chan<- error) {
	endBlock := big.NewInt(0)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			head, _, err := l.client.LatestBlock()
			if err != nil {
				log.Error().Err(err).Msg("Unable to get latest block")
				time.Sleep(l.blockRetryInterval)
				continue
			}
			if startBlock == nil {
				startBlock = head
			}
			endBlock.Add(startBlock, l.blockInterval)
			if endBlock.Cmp(head) == 1 {
				endBlock = head
			}
			// Sleep if the difference is less than needed block confirmations; (latest - current) < BlockDelay
			if new(big.Int).Sub(head, endBlock).Cmp(l.blockConfirmations) == -1 {
				time.Sleep(l.blockRetryInterval)
				continue
			}
			startTime, err := l.client.BlockTsByNum(startBlock.Int64())
			if err != nil {
				log.Error().Err(err).Msgf("Unable to fetch ts by block number %s", startBlock.String())
				time.Sleep(l.blockRetryInterval)
				continue
			}

			endTime, err := l.client.BlockTsByNum(new(big.Int).Sub(endBlock, big.NewInt(1)).Int64())
			if err != nil {
				log.Error().Err(err).Msgf("Unable to fetch ts by block before %s", endBlock.String())
				time.Sleep(l.blockRetryInterval)
				continue
			}
			for _, handler := range l.eventHandlers {
				err := handler.HandleEvent(&startTime, &endTime, msgChan)
				if err != nil {
					log.Error().Err(err).Str("DomainID", string(l.domainID)).Msgf("Unable to handle events")
					continue
				}
			}

			//Write to block store. Not a critical operation, no need to retry
			err = l.blockstore.StoreBlock(endBlock, l.domainID)
			if err != nil {
				log.Error().Str("block", endBlock.String()).Err(err).Msg("Failed to write latest block to blockstore")
			}

			startBlock.Add(startBlock, l.blockInterval)
		}
	}
}
