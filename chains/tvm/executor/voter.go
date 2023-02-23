package executor

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/fbsobreira/gotron-sdk/pkg/address"

	"github.com/ChainSafe/chainbridge-core/chains/tvm/executor/proposal"
	"github.com/ChainSafe/chainbridge-core/relayer/message"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog/log"
)

const (
	maxSimulateVoteChecks = 5
	maxShouldVoteChecks   = 40
	shouldVoteCheckPeriod = 15
)

type ChainClient interface {
	CommonAddress() address.Address
}

type MessageHandler interface {
	HandleMessage(m *message.Message) (*proposal.Proposal, error)
}

type BridgeContract interface {
	IsProposalVotedBy(by address.Address, p *proposal.Proposal) (bool, error)
	VoteProposal(proposal *proposal.Proposal) (*common.Hash, error)
	SimulateVoteProposal(proposal *proposal.Proposal) error
	ProposalStatus(p *proposal.Proposal) (message.ProposalStatus, error)
	GetThreshold() (uint8, error)
}

type TVMVoter struct {
	mh                   MessageHandler
	client               ChainClient
	bridgeContract       BridgeContract
	pendingProposalVotes map[common.Hash]uint8
}

// NewVoter creates an instance of TVMVoter that votes for proposal on chain.
//
// It is created without pending proposal subscription and is a fallback
// for nodes that don't support pending transaction subscription and will vote
// on proposals that already satisfy threshold.
func NewVoter(mh MessageHandler, client ChainClient, bridgeContract BridgeContract) *TVMVoter {
	return &TVMVoter{
		mh:                   mh,
		client:               client,
		bridgeContract:       bridgeContract,
		pendingProposalVotes: make(map[common.Hash]uint8),
	}
}

// Execute checks if relayer already voted and is threshold
// satisfied and casts a vote if it isn't.
func (v *TVMVoter) Execute(m *message.Message) error {
	prop, err := v.mh.HandleMessage(m)
	if err != nil {
		return err
	}

	votedByTheRelayer, err := v.bridgeContract.IsProposalVotedBy(v.client.CommonAddress(), prop)
	if err != nil {
		log.Error().Err(err).Msgf("Fetching is proposal %v voted by relayer failed", prop)
		return err
	}
	if votedByTheRelayer {
		return nil
	}

	shouldVote, err := v.shouldVoteForProposal(prop, 0)
	if err != nil {
		log.Error().Err(err).Msgf("Should vote for proposal %v failed", prop)
		return err
	}

	if !shouldVote {
		log.Debug().Msgf("Proposal %+v already satisfies threshold", prop)
		return nil
	}
	err = v.repetitiveSimulateVote(prop, 0)
	if err != nil {
		log.Error().Err(err).Msgf("Simulating proposal %+v vote failed", prop)
		return err
	}

	hash, err := v.bridgeContract.VoteProposal(prop)
	if err != nil {
		log.Error().Err(err).Msgf("voting for proposal %+v failed", prop)
		return fmt.Errorf("voting failed. Err: %w", err)
	}

	log.Debug().Str("hash", hash.String()).Uint64("nonce", prop.DepositNonce).Msgf("Voted")
	return nil
}

// shouldVoteForProposal checks if proposal already has threshold with pending
// proposal votes from other relayers.
// Only works properly in conjuction with NewVoterWithSubscription as without a subscription
// no pending txs would be received and pending vote count would be 0.
func (v *TVMVoter) shouldVoteForProposal(prop *proposal.Proposal, tries int) (bool, error) {
	propID := prop.GetID()
	defer delete(v.pendingProposalVotes, propID)

	// random delay to prevent all relayers checking for pending votes
	// at the same time and all of them sending another tx
	time.Sleep(time.Duration(rand.Intn(shouldVoteCheckPeriod)) * time.Second)

	ps, err := v.bridgeContract.ProposalStatus(prop)
	if err != nil {
		return false, err
	}

	if ps.Status == message.ProposalStatusExecuted || ps.Status == message.ProposalStatusCanceled {
		return false, nil
	}

	threshold, err := v.bridgeContract.GetThreshold()
	if err != nil {
		return false, err
	}

	if ps.YesVotesTotal+v.pendingProposalVotes[propID] >= threshold && tries < maxShouldVoteChecks {
		// Wait until proposal status is finalized to prevent missing votes
		// in case of dropped txs
		tries++
		return v.shouldVoteForProposal(prop, tries)
	}

	return true, nil
}

// repetitiveSimulateVote repeatedly tries(5 times) to simulate vore proposal call until it succeeds
func (v *TVMVoter) repetitiveSimulateVote(prop *proposal.Proposal, tries int) error {
	err := v.bridgeContract.SimulateVoteProposal(prop)
	if err != nil {
		if tries < maxSimulateVoteChecks {
			tries++
			return v.repetitiveSimulateVote(prop, tries)
		}
		return err
	} else {
		return nil
	}
}
