package chain

import (
	"fmt"
	"math/big"
	"time"

	"github.com/creasty/defaults"

	"github.com/mitchellh/mapstructure"
)

type TVMConfig struct {
	GeneralChainConfig GeneralChainConfig
	Bridge             string
	Erc20Handler       string
	//Erc721Handler      string
	//GenericHandler     string
	TVMEventsConfig TVMEventsConfig
	//MaxGasPrice        *big.Int
	//GasMultiplier      *big.Float
	GasLimit           *big.Int
	StartBlock         *big.Int
	BlockConfirmations *big.Int
	BlockInterval      *big.Int
	BlockRetryInterval time.Duration
}

type TVMEventsConfig struct {
	EventsBaseUrl string `json:"eventsBaseUrl"`
	ApiKey        string `json:"eventsApiKey"`
}

type RawTVMConfig struct {
	GeneralChainConfig `mapstructure:",squash"`
	TVMEventsConfig    `mapstructure:",squash"`
	Bridge             string  `mapstructure:"bridge"`
	Erc20Handler       string  `mapstructure:"erc20Handler"`
	Erc721Handler      string  `mapstructure:"erc721Handler"`
	GenericHandler     string  `mapstructure:"genericHandler"`
	MaxGasPrice        int64   `mapstructure:"maxGasPrice" default:"20000000000"`
	GasMultiplier      float64 `mapstructure:"gasMultiplier" default:"1"`
	GasLimit           int64   `mapstructure:"gasLimit" default:"4000000000"`
	StartBlock         int64   `mapstructure:"startBlock"`
	BlockConfirmations int64   `mapstructure:"blockConfirmations" default:"10"`
	BlockInterval      int64   `mapstructure:"blockInterval" default:"5"`
	BlockRetryInterval uint64  `mapstructure:"blockRetryInterval" default:"5"`
}

func (c *RawTVMConfig) Validate() error {
	if err := c.GeneralChainConfig.Validate(); err != nil {
		return err
	}
	if c.Bridge == "" {
		return fmt.Errorf("required field chain.Bridge empty for chain %v", *c.Id)
	}
	if c.BlockConfirmations != 0 && c.BlockConfirmations < 1 {
		return fmt.Errorf("blockConfirmations has to be >=1")
	}
	return nil
}

// NewTVMConfig decodes and validates an instance of an EVMConfig from
// raw chain config
func NewTVMConfig(chainConfig map[string]interface{}) (*TVMConfig, error) {
	var c RawTVMConfig
	err := mapstructure.Decode(chainConfig, &c)
	if err != nil {
		return nil, err
	}

	err = defaults.Set(&c)
	if err != nil {
		return nil, err
	}

	err = c.Validate()
	if err != nil {
		return nil, err
	}

	c.GeneralChainConfig.ParseFlags()
	config := &TVMConfig{
		GeneralChainConfig: c.GeneralChainConfig,
		Erc20Handler:       c.Erc20Handler,
		//Erc721Handler:      c.Erc721Handler,
		//GenericHandler:     c.GenericHandler,
		Bridge:             c.Bridge,
		TVMEventsConfig:    c.TVMEventsConfig,
		BlockRetryInterval: time.Duration(c.BlockRetryInterval) * time.Second,
		GasLimit:           big.NewInt(c.GasLimit),
		//MaxGasPrice:        big.NewInt(c.MaxGasPrice),
		//GasMultiplier:      big.NewFloat(c.GasMultiplier),
		StartBlock:         big.NewInt(c.StartBlock),
		BlockConfirmations: big.NewInt(c.BlockConfirmations),
		BlockInterval:      big.NewInt(c.BlockInterval),
	}

	return config, nil
}
