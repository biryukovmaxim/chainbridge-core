package listener_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/ChainSafe/chainbridge-core/chains/tvm/calls/events"
	"github.com/ChainSafe/chainbridge-core/chains/tvm/listener"
	mock_listener "github.com/ChainSafe/chainbridge-core/chains/tvm/listener/mock"
	"github.com/ChainSafe/chainbridge-core/relayer/message"
	"github.com/ChainSafe/chainbridge-core/types"
	common "github.com/fbsobreira/gotron-sdk/pkg/address"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
)

type DepositHandlerTestSuite struct {
	suite.Suite
	depositEventHandler *listener.DepositEventHandler
	mockDepositHandler  *mock_listener.MockDepositHandler
	mockEventListener   *mock_listener.MockEventListener
	domainID            uint8
}

func TestRunDepositHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(DepositHandlerTestSuite))
}

func (s *DepositHandlerTestSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	s.domainID = 1
	s.mockEventListener = mock_listener.NewMockEventListener(ctrl)
	s.mockDepositHandler = mock_listener.NewMockDepositHandler(ctrl)
	s.depositEventHandler = listener.NewDepositEventHandler(s.mockEventListener, s.mockDepositHandler, common.Address{}, s.domainID)
}

func (s *DepositHandlerTestSuite) Test_FetchDepositFails() {
	s.mockEventListener.EXPECT().FetchDeposits(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return([]events.Deposit{}, fmt.Errorf("error"))

	msgChan := make(chan []*message.Message, 1)
	err := s.depositEventHandler.HandleEvent(&time.Time{}, &time.Time{}, msgChan)

	s.NotNil(err)
	s.Equal(len(msgChan), 0)
}

func (s *DepositHandlerTestSuite) Test_HandleDepositFails_ExecutionContinue() {
	d1 := events.Deposit{
		DepositNonce:        1,
		DestinationDomainID: 2,
		ResourceID:          types.ResourceID{},
		HandlerResponse:     []byte{},
		Data:                []byte{},
	}
	d2 := events.Deposit{
		DepositNonce:        2,
		DestinationDomainID: 2,
		ResourceID:          types.ResourceID{},
		HandlerResponse:     []byte{},
		Data:                []byte{},
	}
	deposits := []events.Deposit{d1, d2}
	s.mockEventListener.EXPECT().FetchDeposits(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(deposits, nil)
	s.mockDepositHandler.EXPECT().HandleDeposit(
		s.domainID,
		d1.DestinationDomainID,
		d1.DepositNonce,
		d1.ResourceID,
		d1.Data,
		d1.HandlerResponse,
	).Return(&message.Message{}, fmt.Errorf("error"))
	s.mockDepositHandler.EXPECT().HandleDeposit(
		s.domainID,
		d2.DestinationDomainID,
		d2.DepositNonce,
		d2.ResourceID,
		d2.Data,
		d2.HandlerResponse,
	).Return(
		&message.Message{DepositNonce: 2},
		nil,
	)

	msgChan := make(chan []*message.Message, 2)
	err := s.depositEventHandler.HandleEvent(&time.Time{}, &time.Time{}, msgChan)
	msgs := <-msgChan

	s.Nil(err)
	s.Equal(msgs, []*message.Message{{DepositNonce: 2}})
}

func (s *DepositHandlerTestSuite) Test_HandleDepositPanis_ExecutionContinues() {
	d1 := events.Deposit{
		DepositNonce:        1,
		DestinationDomainID: 2,
		ResourceID:          types.ResourceID{},
		HandlerResponse:     []byte{},
		Data:                []byte{},
	}
	d2 := events.Deposit{
		DepositNonce:        2,
		DestinationDomainID: 2,
		ResourceID:          types.ResourceID{},
		HandlerResponse:     []byte{},
		Data:                []byte{},
	}
	deposits := []events.Deposit{d1, d2}
	s.mockEventListener.EXPECT().FetchDeposits(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(deposits, nil)
	s.mockDepositHandler.EXPECT().HandleDeposit(
		s.domainID,
		d1.DestinationDomainID,
		d1.DepositNonce,
		d1.ResourceID,
		d1.Data,
		d1.HandlerResponse,
	).Do(func(sourceID, destID, nonce, resourceID, calldata, handlerResponse interface{}) {
		panic("error")
	})
	s.mockDepositHandler.EXPECT().HandleDeposit(
		s.domainID,
		d2.DestinationDomainID,
		d2.DepositNonce,
		d2.ResourceID,
		d2.Data,
		d2.HandlerResponse,
	).Return(
		&message.Message{DepositNonce: 2},
		nil,
	)

	msgChan := make(chan []*message.Message, 2)
	err := s.depositEventHandler.HandleEvent(&time.Time{}, &time.Time{}, msgChan)
	msgs := <-msgChan

	s.Nil(err)
	s.Equal(msgs, []*message.Message{{DepositNonce: 2}})
}

func (s *DepositHandlerTestSuite) Test_SuccessfulHandleDeposit() {
	d1 := events.Deposit{
		DepositNonce:        1,
		DestinationDomainID: 2,
		ResourceID:          types.ResourceID{},
		HandlerResponse:     []byte{},
		Data:                []byte{},
	}
	d2 := events.Deposit{
		DepositNonce:        2,
		DestinationDomainID: 2,
		ResourceID:          types.ResourceID{},
		HandlerResponse:     []byte{},
		Data:                []byte{},
	}
	deposits := []events.Deposit{d1, d2}
	s.mockEventListener.EXPECT().FetchDeposits(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(deposits, nil)
	s.mockDepositHandler.EXPECT().HandleDeposit(
		s.domainID,
		d1.DestinationDomainID,
		d1.DepositNonce,
		d1.ResourceID,
		d1.Data,
		d1.HandlerResponse,
	).Return(
		&message.Message{DepositNonce: 1},
		nil,
	)
	s.mockDepositHandler.EXPECT().HandleDeposit(
		s.domainID,
		d2.DestinationDomainID,
		d2.DepositNonce,
		d2.ResourceID,
		d2.Data,
		d2.HandlerResponse,
	).Return(
		&message.Message{DepositNonce: 2},
		nil,
	)

	msgChan := make(chan []*message.Message, 2)
	err := s.depositEventHandler.HandleEvent(&time.Time{}, &time.Time{}, msgChan)
	msgs := <-msgChan

	s.Nil(err)
	s.Equal(msgs, []*message.Message{{DepositNonce: 1}, {DepositNonce: 2}})
}
