package proto

import (
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
)

func ChainConfigToProto(c *params.ChainConfig) *ChainConfig {
	if c == nil {
		return nil
	}

	return &ChainConfig{
		ChainId:                 bigIntToBytes(c.ChainID),
		HomesteadBlock:          bigIntToBytes(c.HomesteadBlock),
		DaoForkBlock:            bigIntToBytes(c.DAOForkBlock),
		DaoForkSupport:          &(c.DAOForkSupport),
		Eip150Block:             bigIntToBytes(c.EIP150Block),
		Eip155Block:             bigIntToBytes(c.EIP155Block),
		Eip158Block:             bigIntToBytes(c.EIP158Block),
		ByzantiumBlock:          bigIntToBytes(c.ByzantiumBlock),
		ConstantinopleBlock:     bigIntToBytes(c.ConstantinopleBlock),
		PetersburgBlock:         bigIntToBytes(c.PetersburgBlock),
		IstanbulBlock:           bigIntToBytes(c.IstanbulBlock),
		MuirGlacierBlock:        bigIntToBytes(c.MuirGlacierBlock),
		BerlinBlock:             bigIntToBytes(c.BerlinBlock),
		LondonBlock:             bigIntToBytes(c.LondonBlock),
		ArrowGlacierBlock:       bigIntToBytes(c.ArrowGlacierBlock),
		GrayGlacierBlock:        bigIntToBytes(c.GrayGlacierBlock),
		MergeNetsplitBlock:      bigIntToBytes(c.MergeNetsplitBlock),
		ShanghaiTime:            c.ShanghaiTime,
		CancunTime:              c.CancunTime,
		PragueTime:              c.PragueTime,
		VerkleTime:              c.VerkleTime,
		TerminalTotalDifficulty: bigIntToBytes(c.TerminalTotalDifficulty),
		DepositContractAddress:  c.DepositContractAddress.Bytes(),
		Ethash: func() []byte {
			if c.Ethash != nil {
				return []byte("Ethash")
			}
			return nil
		}(),
		Clique:       CliqueConfigToProto(c.Clique),
		BlobSchedule: BlobScheduleConfigToProto(c.BlobScheduleConfig),
	}
}

func ChainConfigFromProto(c *ChainConfig) *params.ChainConfig {
	if c == nil {
		return nil
	}

	chainConfig := &params.ChainConfig{
		ChainID:                 bytesToBigInt(c.GetChainId()),
		HomesteadBlock:          bytesToBigInt(c.GetHomesteadBlock()),
		DAOForkBlock:            bytesToBigInt(c.GetDaoForkBlock()),
		DAOForkSupport:          c.GetDaoForkSupport(),
		EIP150Block:             bytesToBigInt(c.GetEip150Block()),
		EIP155Block:             bytesToBigInt(c.GetEip155Block()),
		EIP158Block:             bytesToBigInt(c.GetEip158Block()),
		ByzantiumBlock:          bytesToBigInt(c.GetByzantiumBlock()),
		ConstantinopleBlock:     bytesToBigInt(c.GetConstantinopleBlock()),
		PetersburgBlock:         bytesToBigInt(c.GetPetersburgBlock()),
		IstanbulBlock:           bytesToBigInt(c.GetIstanbulBlock()),
		MuirGlacierBlock:        bytesToBigInt(c.GetMuirGlacierBlock()),
		BerlinBlock:             bytesToBigInt(c.GetBerlinBlock()),
		LondonBlock:             bytesToBigInt(c.GetLondonBlock()),
		ArrowGlacierBlock:       bytesToBigInt(c.GetArrowGlacierBlock()),
		GrayGlacierBlock:        bytesToBigInt(c.GetGrayGlacierBlock()),
		MergeNetsplitBlock:      bytesToBigInt(c.GetMergeNetsplitBlock()),
		ShanghaiTime:            c.ShanghaiTime,
		CancunTime:              c.CancunTime,
		PragueTime:              c.PragueTime,
		VerkleTime:              c.VerkleTime,
		TerminalTotalDifficulty: bytesToBigInt(c.GetTerminalTotalDifficulty()),
		DepositContractAddress:  gethcommon.BytesToAddress(c.GetDepositContractAddress()),
		Clique:                  CliqueConfigFromProto(c.GetClique()),
		BlobScheduleConfig:      BlobScheduleConfigFromProto(c.GetBlobSchedule()),
	}

	// Handle consensus engine configs
	if len(c.GetEthash()) > 0 {
		chainConfig.Ethash = &params.EthashConfig{}
	}

	return chainConfig
}

func CliqueConfigToProto(c *params.CliqueConfig) *CliqueConfig {
	if c == nil {
		return nil
	}

	return &CliqueConfig{
		Period: c.Period,
		Epoch:  c.Epoch,
	}
}

func CliqueConfigFromProto(c *CliqueConfig) *params.CliqueConfig {
	if c == nil {
		return nil
	}

	return &params.CliqueConfig{
		Period: c.Period,
		Epoch:  c.Epoch,
	}
}

func BlobScheduleConfigToProto(c *params.BlobScheduleConfig) *BlobScheduleConfig {
	if c == nil {
		return nil
	}

	return &BlobScheduleConfig{
		Cancun: BlobConfigToProto(c.Cancun),
		Prague: BlobConfigToProto(c.Prague),
		Osaka:  BlobConfigToProto(c.Osaka),
		Verkle: BlobConfigToProto(c.Verkle),
	}
}

func BlobScheduleConfigFromProto(c *BlobScheduleConfig) *params.BlobScheduleConfig {
	if c == nil {
		return nil
	}

	return &params.BlobScheduleConfig{
		Cancun: BlobConfigFromProto(c.GetCancun()),
		Prague: BlobConfigFromProto(c.GetPrague()),
		Osaka:  BlobConfigFromProto(c.GetOsaka()),
		Verkle: BlobConfigFromProto(c.GetVerkle()),
	}
}

func BlobConfigToProto(c *params.BlobConfig) *BlobConfig {
	if c == nil {
		return nil
	}

	return &BlobConfig{
		Target:         uint64(c.Target),
		Max:            uint64(c.Max),
		UpdateFraction: c.UpdateFraction,
	}
}

func BlobConfigFromProto(c *BlobConfig) *params.BlobConfig {
	if c == nil {
		return nil
	}

	return &params.BlobConfig{
		Target:         int(c.Target),
		Max:            int(c.Max),
		UpdateFraction: c.UpdateFraction,
	}
}
