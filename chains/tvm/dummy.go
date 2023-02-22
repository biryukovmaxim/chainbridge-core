package tvm

import (
	"github.com/ChainSafe/chainbridge-core/relayer/message"
	"github.com/rs/zerolog/log"
)

type DummyExecutor struct{}

func (d DummyExecutor) Execute(message *message.Message) error {
	log.Debug().Msgf("get message for execution: %+v", message)
	return nil
}
