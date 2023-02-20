package listener

import (
	"context"
	"fmt"
	"time"

	"github.com/ChainSafe/chainbridge-core/chains/tvm/calls/events"
	"github.com/ChainSafe/chainbridge-core/relayer/message"
	"github.com/ChainSafe/chainbridge-core/types"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
	"github.com/rs/zerolog/log"
)

type EventListener interface {
	FetchDeposits(ctx context.Context, contractAddress address.Address, startTime, endTime *time.Time) ([]events.Deposit, error)
}
type DepositHandler interface {
	HandleDeposit(sourceID, destID uint8, nonce uint64, resourceID types.ResourceID, calldata, handlerResponse []byte) (*message.Message, error)
}

type DepositEventHandler struct {
	eventListener  EventListener
	depositHandler DepositHandler

	bridgeAddress address.Address
	domainID      uint8
}

func NewDepositEventHandler(eventListener EventListener, depositHandler DepositHandler, bridgeAddress address.Address, domainID uint8) *DepositEventHandler {
	return &DepositEventHandler{
		eventListener:  eventListener,
		depositHandler: depositHandler,
		bridgeAddress:  bridgeAddress,
		domainID:       domainID,
	}
}

func (eh *DepositEventHandler) HandleEvent(startTime, endTime *time.Time, msgChan chan []*message.Message) error {
	deposits, err := eh.eventListener.FetchDeposits(context.Background(), eh.bridgeAddress, startTime, endTime)
	if err != nil {
		return fmt.Errorf("unable to fetch deposit events because of: %+v", err)
	}

	domainDeposits := make(map[uint8][]*message.Message)
	for _, d := range deposits {
		func(d events.Deposit) {
			defer func() {
				if r := recover(); r != nil {
					log.Error().Err(err).Msgf("panic occured while handling deposit %+v", d)
				}
			}()

			m, err := eh.depositHandler.HandleDeposit(eh.domainID, d.DestinationDomainID, d.DepositNonce, d.ResourceID, d.Data, d.HandlerResponse)
			if err != nil {
				log.Error().Err(err).Str("start time", startTime.String()).Str("end time", endTime.String()).Uint8("domainID", eh.domainID).Msgf("%v", err)
				return
			}

			log.Debug().Msgf("Resolved message %+v in block range: %s-%s", m, startTime.String(), endTime.String())
			domainDeposits[m.Destination] = append(domainDeposits[m.Destination], m)
		}(d)
	}

	for _, deposits := range domainDeposits {
		msgChan <- deposits
	}

	return nil
}
