package erc721

import (
	"fmt"
	"github.com/ChainSafe/chainbridge-core/chains/evm/calls/bridge"
	"github.com/ChainSafe/chainbridge-core/chains/evm/calls/transactor"
	"github.com/ChainSafe/chainbridge-core/chains/evm/cli/init"
	"github.com/ChainSafe/chainbridge-core/chains/evm/evmtransaction"
	"github.com/ChainSafe/chainbridge-core/util"
	"math/big"
	"strconv"

	"github.com/ChainSafe/chainbridge-core/chains/evm/cli/flags"
	"github.com/ChainSafe/chainbridge-core/chains/evm/cli/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var depositCmd = &cobra.Command{
	Use:   "deposit",
	Short: "Deposit ERC721 token",
	Long:  "Deposit ERC721 token",
	PreRun: func(cmd *cobra.Command, args []string) {
		logger.LoggerMetadata(cmd.Name(), cmd.Flags())
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return util.CallPersistentPreRun(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := init.InitializeClient(url, senderKeyPair)
		if err != nil {
			return err
		}
		t, err := init.InitializeTransactor(gasPrice, evmtransaction.NewTransaction, c)
		if err != nil {
			return err
		}
		return DepositCmd(cmd, args, bridge.NewBridgeContract(c, bridgeAddr, t))
	},
	Args: func(cmd *cobra.Command, args []string) error {
		err := ValidateDepositFlags(cmd, args)
		if err != nil {
			return err
		}

		err = ProcessDepositFlags(cmd, args)
		return err
	},
}

func BindDepositCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&Recipient, "recipient", "", "address of recipient")
	cmd.Flags().StringVar(&Bridge, "bridge", "", "address of bridge contract")
	cmd.Flags().StringVar(&DestionationID, "destId", "", "destination domain ID")
	cmd.Flags().StringVar(&ResourceID, "resourceId", "", "resource ID for transfer")
	cmd.Flags().StringVar(&TokenId, "tokenId", "", "ERC721 token ID")
	cmd.Flags().StringVar(&Metadata, "metadata", "", "ERC721 metadata")
	flags.MarkFlagsAsRequired(cmd, "recipient", "bridge", "destId", "resourceId", "tokenId")
}

func init() {
	BindDepositCmdFlags(depositCmd)
}

func ValidateDepositFlags(cmd *cobra.Command, args []string) error {
	if !common.IsHexAddress(Recipient) {
		return fmt.Errorf("invalid recipient address")
	}
	if !common.IsHexAddress(Bridge) {
		return fmt.Errorf("invalid bridge address")
	}
	return nil
}

func ProcessDepositFlags(cmd *cobra.Command, args []string) error {
	var err error

	recipientAddr = common.HexToAddress(Recipient)
	bridgeAddr = common.HexToAddress(Bridge)

	destinationID, err = strconv.Atoi(DestionationID)
	if err != nil {
		log.Error().Err(fmt.Errorf("destination ID conversion error: %v", err))
		return err
	}

	var ok bool
	tokenId, ok = big.NewInt(0).SetString(TokenId, 10)
	if !ok {
		return fmt.Errorf("invalid token id value")
	}

	resourceId, err = flags.ProcessResourceID(ResourceID)
	return err
}

func DepositCmd(cmd *cobra.Command, args []string, bridgeContract *bridge.BridgeContract) error {
	txHash, err := bridgeContract.Erc721Deposit(
		tokenId, Metadata, recipientAddr, resourceId, uint8(destinationID), transactor.TransactOptions{})
	if err != nil {
		return err
	}

	log.Info().Msgf(
		`erc721 deposit hash: %s
		%s token were transferred to %s from %s`,
		txHash.Hex(),
		tokenId.String(),
		recipientAddr.Hex(),
		senderKeyPair.CommonAddress().String(),
	)
	return nil
}
