package tvm

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ChainSafe/chainbridge-core/relayer/message"
	"github.com/ChainSafe/chainbridge-core/store"
	"github.com/rs/zerolog/log"
)

type EventListener interface {
	ListenToEvents(ctx context.Context, startBlock *big.Int, msgChan chan []*message.Message, errChan chan<- error)
}

type ProposalExecutor interface {
	Execute(message *message.Message) error
}

// TVMChain is struct that aggregates all data required for
type TVMChain struct {
	listener   EventListener
	writer     ProposalExecutor
	blockstore *store.BlockStore

	domainID    uint8
	startBlock  *big.Int
	freshStart  bool
	latestBlock bool
}

func NewTVMChain(listener EventListener, writer ProposalExecutor, blockstore *store.BlockStore, domainID uint8, startBlock *big.Int, latestBlock bool, freshStart bool) *TVMChain {
	return &TVMChain{
		listener:    listener,
		writer:      writer,
		blockstore:  blockstore,
		domainID:    domainID,
		startBlock:  startBlock,
		latestBlock: latestBlock,
		freshStart:  freshStart,
	}
}

// PollEvents is the goroutine that polls blocks and searches Deposit events in them.
// Events are then sent to eventsChan.
func (c *TVMChain) PollEvents(ctx context.Context, sysErr chan<- error, msgChan chan []*message.Message) {
	log.Info().Msg("Polling Blocks...")

	startBlock, err := c.blockstore.GetStartBlock(
		c.domainID,
		c.startBlock,
		c.latestBlock,
		c.freshStart,
	)
	if err != nil {
		sysErr <- fmt.Errorf("error %w on getting last stored block", err)
		return
	}

	go c.listener.ListenToEvents(ctx, startBlock, msgChan, sysErr)
}

func (c *TVMChain) Write(msg []*message.Message) {
	for _, msg := range msg {
		go func(msg *message.Message) {
			err := c.writer.Execute(msg)
			if err != nil {
				log.Err(err).Msgf("Failed writing message %v", msg)
			}
		}(msg)
	}
}

func (c *TVMChain) DomainID() uint8 {
	return c.domainID
}
