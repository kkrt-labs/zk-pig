package blocks

import (
	"fmt"
	"math/big"

	aws "github.com/kkrt-labs/kakarot-controller/pkg/aws"
	jsonrpcmrgd "github.com/kkrt-labs/kakarot-controller/pkg/jsonrpc/merged"
	store "github.com/kkrt-labs/kakarot-controller/pkg/store"
	compressstore "github.com/kkrt-labs/kakarot-controller/pkg/store/compress"
	filestore "github.com/kkrt-labs/kakarot-controller/pkg/store/file"
	multistore "github.com/kkrt-labs/kakarot-controller/pkg/store/multi"
	s3store "github.com/kkrt-labs/kakarot-controller/pkg/store/s3"
	blockstore "github.com/kkrt-labs/kakarot-controller/src/blocks/store"
	"github.com/kkrt-labs/kakarot-controller/src/config"
)

type ChainConfig struct {
	ID  *big.Int
	RPC *jsonrpcmrgd.Config
}

type StoreConfig struct {
	Format      store.ContentType
	Compression store.ContentEncoding
}

// Config is the configuration for the RPCPreflight.
type Config struct {
	Chain                  ChainConfig
	BaseDir                string `json:"blocks-dir"` // Base directory for storing block data
	HeavyProverInputsStore blockstore.HeavyProverInputsStore
	ProverInputsStore      blockstore.ProverInputsStore
}

func (cfg *Config) SetDefault() *Config {
	if cfg.BaseDir == "" {
		cfg.BaseDir = "data/blocks"
	}

	if cfg.Chain.RPC != nil {
		cfg.Chain.RPC.SetDefault()
	}

	return cfg
}

func FromGlobalConfig(gcfg *config.Config) (*Service, error) {
	contentEncoding, err := store.ParseContentEncoding(gcfg.Store.ContentEncoding)
	if err != nil {
		return nil, fmt.Errorf("failed to parse content encoding: %v", err)
	}
	contentType, err := store.ParseContentType(gcfg.Store.ContentType)
	if err != nil {
		return nil, fmt.Errorf("failed to parse content type: %v", err)
	}

	cfg := &Config{
		Chain:   ChainConfig{},
		BaseDir: gcfg.DataDir,
	}

	if gcfg.Chain.ID != "" {
		cfg.Chain.ID = new(big.Int)
		if _, ok := cfg.Chain.ID.SetString(gcfg.Chain.ID, 10); !ok {
			return nil, fmt.Errorf("invalid chain id %q", gcfg.Chain.ID)
		}
	}

	if gcfg.Chain.RPC.URL != "" {
		cfg.Chain.RPC = &jsonrpcmrgd.Config{
			Addr: gcfg.Chain.RPC.URL,
		}
	}

	var multiConfig multistore.Config

	switch gcfg.Store.Location {
	case "file", "":
		multiConfig.FileConfig = &filestore.Config{
			DataDir: cfg.BaseDir,
		}
	case "s3":
		multiConfig.S3Config = &s3store.Config{
			Bucket:    gcfg.AWS.S3.Bucket,
			KeyPrefix: gcfg.AWS.S3.KeyPrefix + "/",
			ProviderConfig: &aws.ProviderConfig{
				Region: gcfg.AWS.S3.Region,
				Credentials: &aws.CredentialsConfig{
					AccessKey: gcfg.AWS.S3.AccessKey,
					SecretKey: gcfg.AWS.S3.SecretKey,
				},
			},
		}
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", gcfg.Store.Location)
	}

	compressStore, err := compressstore.New(compressstore.Config{
		MultiConfig:     multiConfig,
		ContentEncoding: contentEncoding,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create compress store: %v", err)
	}

	cfg.HeavyProverInputsStore, err = blockstore.NewHeavyProverInputsStore(&blockstore.HeavyProverInputsStoreConfig{
		FileConfig: &filestore.Config{
			DataDir: cfg.BaseDir,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create heavy prover inputs store: %v", err)
	}

	cfg.ProverInputsStore = blockstore.NewFromStore(compressStore, contentType)

	return New(cfg)
}
