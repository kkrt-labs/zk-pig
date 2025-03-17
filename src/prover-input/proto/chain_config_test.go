package proto

import (
	"math/big"
	"testing"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
	"github.com/kkrt-labs/go-utils/common"
	"github.com/stretchr/testify/assert"
)

func TestChainConfigToProto(t *testing.T) {
	var testCases = []struct {
		desc        string
		chainConfig *params.ChainConfig
	}{
		{
			desc:        "empty chain config",
			chainConfig: &params.ChainConfig{},
		},
		{
			desc: "chain config with all fields set",
			chainConfig: &params.ChainConfig{
				ChainID:                 big.NewInt(1),
				HomesteadBlock:          big.NewInt(2),
				DAOForkBlock:            big.NewInt(3),
				DAOForkSupport:          true,
				EIP150Block:             big.NewInt(4),
				EIP155Block:             big.NewInt(5),
				EIP158Block:             big.NewInt(6),
				ByzantiumBlock:          big.NewInt(7),
				ConstantinopleBlock:     big.NewInt(8),
				PetersburgBlock:         big.NewInt(9),
				IstanbulBlock:           big.NewInt(10),
				MuirGlacierBlock:        big.NewInt(11),
				BerlinBlock:             big.NewInt(12),
				LondonBlock:             big.NewInt(13),
				ArrowGlacierBlock:       big.NewInt(14),
				GrayGlacierBlock:        big.NewInt(15),
				MergeNetsplitBlock:      big.NewInt(16),
				ShanghaiTime:            common.Ptr(uint64(17)),
				CancunTime:              common.Ptr(uint64(18)),
				PragueTime:              common.Ptr(uint64(19)),
				VerkleTime:              common.Ptr(uint64(20)),
				TerminalTotalDifficulty: big.NewInt(21),
				DepositContractAddress:  gethcommon.HexToAddress("0x123"),
				Ethash:                  &params.EthashConfig{},
				Clique: &params.CliqueConfig{
					Period: 1000000000000000000,
					Epoch:  3000000000000000000,
				},
				BlobScheduleConfig: &params.BlobScheduleConfig{
					Cancun: &params.BlobConfig{
						Target:         1000000000000000000,
						Max:            2000000000000000000,
						UpdateFraction: 3000000000000000000,
					},
					Prague: &params.BlobConfig{
						Target:         4000000000000000000,
						Max:            5000000000000000000,
						UpdateFraction: 6000000000000000000,
					},
					Osaka: &params.BlobConfig{
						Target:         7000000000000000000,
						Max:            8000000000000000000,
						UpdateFraction: 9000000000000000000,
					},
					Verkle: &params.BlobConfig{
						Target:         1000000000000000000,
						Max:            1100000000000000000,
						UpdateFraction: 1200000000000000000,
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			protoChainConfig := ChainConfigToProto(tc.chainConfig)
			chainConfigFromProto := ChainConfigFromProto(protoChainConfig)
			assert.Equal(t, tc.chainConfig, chainConfigFromProto)
		})
	}
}
