package blocks

import (
	"fmt"
	"math/big"

	aws "github.com/kkrt-labs/go-utils/aws"
	jsonrpcmrgd "github.com/kkrt-labs/go-utils/jsonrpc/merged"
	store "github.com/kkrt-labs/go-utils/store"
	filestore "github.com/kkrt-labs/go-utils/store/file"
	multistore "github.com/kkrt-labs/go-utils/store/multi"
	s3store "github.com/kkrt-labs/go-utils/store/s3"
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
	Chain                        ChainConfig
	BaseDir                      string `json:"blocks-dir"` // Base directory for storing block data
	HeavyProverInputsStoreConfig blockstore.HeavyProverInputsStoreConfig
	ProverInputsStoreConfig      blockstore.ProverInputsStoreConfig
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
	// Parse content encoding and type with error handling
	contentEncoding, err := store.ParseContentEncoding(gcfg.ProverInputsStore.ContentEncoding)
	if err != nil {
		return nil, fmt.Errorf("failed to parse content encoding: %v", err)
	}
	contentType, err := store.ParseContentType(gcfg.ProverInputsStore.ContentType)
	if err != nil {
		return nil, fmt.Errorf("failed to parse content type: %v", err)
	}

	// Initialize configuration with default values
	cfg := &Config{
		Chain:   ChainConfig{},
		BaseDir: gcfg.DataDir.Root,
	}

	// Set Chain ID if provided
	if gcfg.Chain.ID != "" {
		if cfg.Chain.ID, err = parseChainID(gcfg.Chain.ID); err != nil {
			return nil, err
		}
	}

	// Set RPC configuration if URL is provided
	if gcfg.Chain.RPC.URL != "" {
		cfg.Chain.RPC = &jsonrpcmrgd.Config{Addr: gcfg.Chain.RPC.URL}
	}

	// Configure multi-store settings
	multiStoreConfig := configureMultiStore(gcfg, cfg.BaseDir)

	// Set heavy prover inputs store configuration
	cfg.HeavyProverInputsStoreConfig = blockstore.HeavyProverInputsStoreConfig{
		FileConfig: &filestore.Config{DataDir: gcfg.DataDir.Root + "/" + ChainID(gcfg) + "/" + gcfg.DataDir.Preflight},
	}

	// Set prover inputs store configuration
	cfg.ProverInputsStoreConfig = blockstore.ProverInputsStoreConfig{
		MultiStoreConfig: multiStoreConfig,
		ContentEncoding:  contentEncoding,
		ContentType:      contentType,
	}

	return New(cfg)
}

// Helper function to parse chain ID
func parseChainID(chainID string) (*big.Int, error) {
	id := new(big.Int)
	if _, ok := id.SetString(chainID, 10); !ok {
		return nil, fmt.Errorf("invalid chain id %q", chainID)
	}
	return id, nil
}

// Helper function to configure multi-store settings
func configureMultiStore(gcfg *config.Config, baseDir string) multistore.Config {
	var multiStoreConfig multistore.Config
	// Configure file store
	if gcfg.DataDir.Inputs != "" {
		multiStoreConfig.FileConfig = &filestore.Config{DataDir: baseDir + "/" + ChainID(gcfg) + "/" + gcfg.DataDir.Inputs}
	}

	// Configure S3 store
	if gcfg.ProverInputsStore.S3.AWSProvider.Bucket != "" {
		multiStoreConfig.S3Config = &s3store.Config{
			Bucket:    gcfg.ProverInputsStore.S3.AWSProvider.Bucket,
			KeyPrefix: gcfg.ProverInputsStore.S3.AWSProvider.KeyPrefix,
			ProviderConfig: &aws.ProviderConfig{
				Region: gcfg.ProverInputsStore.S3.AWSProvider.Region,
				Credentials: &aws.CredentialsConfig{
					AccessKey: gcfg.ProverInputsStore.S3.AWSProvider.Credentials.AccessKey,
					SecretKey: gcfg.ProverInputsStore.S3.AWSProvider.Credentials.SecretKey,
				},
			},
		}
	}

	return multiStoreConfig
}

func ChainID(gcfg *config.Config) string {
	if gcfg.Chain.ID == "" {
		return "default"
	}
	return gcfg.Chain.ID
}
