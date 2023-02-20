package tvm

import (
	"fmt"

	"github.com/ChainSafe/chainbridge-core/relayer/message"
	"github.com/ChainSafe/chainbridge-core/types"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
	"github.com/rs/zerolog/log"
)

type DummyExecutor struct{}

func (d DummyExecutor) Execute(message *message.Message) error {
	log.Debug().Msgf("get message for execution: %+v", message)
	return nil
}

type DummyMatcher struct{}

func (d DummyMatcher) GetHandlerAddressForResourceID(resourceID types.ResourceID) (address.Address, error) {
	log.Debug().Msgf("get handler address for resource id: %s", string(resourceID[:]))
	return nil, fmt.Errorf("dummy error")
}
