package centrifuge

import (
	"github.com/ChainSafe/chainbridge-core/chains/evm/calls/centrifuge"
	"github.com/ChainSafe/chainbridge-core/chains/evm/cli/init"
	"github.com/ChainSafe/chainbridge-core/chains/evm/cli/logger"
	"github.com/ChainSafe/chainbridge-core/chains/evm/evmtransaction"
	"github.com/ChainSafe/chainbridge-core/util"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy centrifuge asset store contract",
	Long:  "This command can be used to deploy Centrifuge asset store contract that represents bridged Centrifuge assets.",
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
		return DeployCentrifugeAssetStoreCmd(cmd, args, centrifuge.NewAssetStoreContract(c, common.Address{}, t))
	},
}

func BindDeployCmdFlags(cmd *cobra.Command) {}

func init() {
	BindDeployCmdFlags(deployCmd)
}

func DeployCentrifugeAssetStoreCmd(cmd *cobra.Command, args []string, contract *centrifuge.AssetStoreContract) error {
	assetStoreAddress, err := contract.DeployContract()
	if err != nil {
		return err
	}

	log.Info().Msgf("Deployed Centrifuge asset store to address: %s", assetStoreAddress.String())
	return nil
}
