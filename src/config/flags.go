package config

import (
	"github.com/kkrt-labs/kakarot-controller/pkg/common"
	"github.com/kkrt-labs/kakarot-controller/pkg/spf13"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	configFileFlag = &spf13.StringArrayFlag{
		ViperKey:    "config",
		Name:        "config",
		Shorthand:   "c",
		Env:         "CONFIG",
		Description: "Configuration file (yaml format)",
	}
)

func AddConfigFileFlag(v *viper.Viper, f *pflag.FlagSet) {
	configFileFlag.Add(v, f)
}

var (
	chainIDFlag = &spf13.StringFlag{
		ViperKey:    "chain.id",
		Name:        "chain-id",
		Env:         "CHAIN_ID",
		Description: "Chain ID (decimal)",
	}
	chainRPCURLFlag = &spf13.StringFlag{
		ViperKey:    "chain.rpc.url",
		Name:        "chain-rpc-url",
		Env:         "CHAIN_RPC_URL",
		Description: "Chain JSON-RPC URL",
	}
	dataDirFlag = &spf13.StringFlag{
		ViperKey:     "data-dir",
		Name:         "data-dir",
		Env:          "DATA_DIR",
		Description:  "Path to data directory",
		DefaultValue: common.Ptr("data/inputs"),
	}
)

func AddChainFlags(v *viper.Viper, f *pflag.FlagSet) {
	chainIDFlag.Add(v, f)
	chainRPCURLFlag.Add(v, f)
}

func AddProverInputsFlags(v *viper.Viper, f *pflag.FlagSet) {
	AddChainFlags(v, f)
	dataDirFlag.Add(v, f)
}
