package erc20

import (
	"fmt"
	"github.com/ChainSafe/chainbridge-core/chains/evm/calls/bridge"
	"github.com/ChainSafe/chainbridge-core/chains/evm/calls/client"
	"github.com/ChainSafe/chainbridge-core/chains/evm/calls/transactor"
	"github.com/ChainSafe/chainbridge-core/chains/evm/cli/init"
	"github.com/ChainSafe/chainbridge-core/chains/evm/evmtransaction"
	"github.com/ChainSafe/chainbridge-core/util"
	"math/big"

	"github.com/ChainSafe/chainbridge-core/chains/evm/cli/flags"
	"github.com/ChainSafe/chainbridge-core/chains/evm/cli/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var depositCmd = &cobra.Command{
	Use:   "deposit",
	Short: "Initiate a transfer of ERC20 tokens",
	Long:  "Initiate a transfer of ERC20 tokens",
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
		return DepositCmd(cmd, args, bridge.NewBridgeContract(c, erc20Addr, t))
	},
	Args: func(cmd *cobra.Command, args []string) error {
		err := ValidateDepositFlags(cmd, args)
		if err != nil {
			return err
		}

		err = ProcessDepositFlags(cmd, args)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	BindDepositCmdFlags(depositCmd)
}

func BindDepositCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&Recipient, "recipient", "", "address of recipient")
	cmd.Flags().StringVar(&Bridge, "bridge", "", "address of bridge contract")
	cmd.Flags().StringVar(&Amount, "amount", "", "amount to deposit")
	cmd.Flags().Uint64Var(&DomainID, "domainId", 0, "destination domain ID")
	cmd.Flags().StringVar(&ResourceID, "resourceId", "", "resource ID for transfer")
	cmd.Flags().Uint64Var(&Decimals, "decimals", 0, "ERC20 token decimals")
	flags.MarkFlagsAsRequired(cmd, "recipient", "bridge", "amount", "domainId", "resourceId", "decimals")
}

func ValidateDepositFlags(cmd *cobra.Command, args []string) error {
	if !common.IsHexAddress(Recipient) {
		return fmt.Errorf("invalid recipient address %s", Recipient)
	}
	if !common.IsHexAddress(Bridge) {
		return fmt.Errorf("invalid bridge address %s", Bridge)
	}
	return nil
}

func ProcessDepositFlags(cmd *cobra.Command, args []string) error {
	var err error

	recipientAddress = common.HexToAddress(Recipient)
	decimals := big.NewInt(int64(Decimals))
	bridgeAddr = common.HexToAddress(Bridge)
	realAmount, err = client.UserAmountToWei(Amount, decimals)
	if err != nil {
		return err
	}
	resourceIdBytesArr, err = flags.ProcessResourceID(ResourceID)
	return err
}

func DepositCmd(cmd *cobra.Command, args []string, contract *bridge.BridgeContract) error {
	data := bridge.ConstructErc20DepositData(recipientAddress.Bytes(), realAmount)
	hash, err := contract.Deposit(resourceIdBytesArr, uint8(DomainID), data, transactor.TransactOptions{})
	if err != nil {
		log.Error().Err(fmt.Errorf("erc20 deposit error: %v", err))
		return err
	}

	log.Info().Msgf("%s tokens were transferred to %s from %s with hash %s", Amount, recipientAddress.Hex(), senderKeyPair.CommonAddress().String(), hash.Hex())
	return nil
}
