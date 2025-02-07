package src

import (
	"fmt"
	"math/big"
	"path/filepath"

	aws "github.com/kkrt-labs/go-utils/aws"
	jsonrpcmrgd "github.com/kkrt-labs/go-utils/jsonrpc/merged"
	store "github.com/kkrt-labs/go-utils/store"
	filestore "github.com/kkrt-labs/go-utils/store/file"
	multistore "github.com/kkrt-labs/go-utils/store/multi"
	s3store "github.com/kkrt-labs/go-utils/store/s3"
	"github.com/kkrt-labs/zk-pig/src/config"
	inputstore "github.com/kkrt-labs/zk-pig/src/store"
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
	Chain              ChainConfig
	DataDir            string
	PreflightDataStore inputstore.PreflightDataStoreConfig
	ProverInputStore   inputstore.ProverInputStoreConfig
}

func (cfg *Config) SetDefault() *Config {
	if cfg.DataDir == "" {
		cfg.DataDir = "data"
	}

	if cfg.Chain.RPC != nil {
		cfg.Chain.RPC.SetDefault()
	}

	return cfg
}

func FromGlobalConfig(gcfg *config.Config) (*Service, error) {
	// Initialize configuration with default values
	cfg := &Config{
		Chain:   ChainConfig{},
		DataDir: gcfg.DataDir,
	}

	// Set Chain ID if provided
	var err error
	if gcfg.Chain.ID != "" {
		if cfg.Chain.ID, err = parseChainID(gcfg.Chain.ID); err != nil {
			return nil, err
		}
	}

	// --- Set RPC configuration if URL is provided ---
	if gcfg.Chain.RPC.URL != "" {
		cfg.Chain.RPC = &jsonrpcmrgd.Config{Addr: gcfg.Chain.RPC.URL}
	}

	// --- Set Preflight Data Store configuration ---
	if gcfg.PreflightDataStore.File.Dir != "" {
		cfg.PreflightDataStore = inputstore.PreflightDataStoreConfig{
			FileConfig: &filestore.Config{DataDir: filepath.Join(gcfg.DataDir, ChainID(gcfg), gcfg.PreflightDataStore.File.Dir)},
		}
	}

	// --- Set Prover Input Store configuration
	contentEncoding, err := store.ParseContentEncoding(gcfg.ProverInputStore.ContentEncoding)
	if err != nil {
		return nil, fmt.Errorf("failed to parse content encoding: %v", err)
	}
	contentType, err := store.ParseContentType(gcfg.ProverInputStore.ContentType)
	if err != nil {
		return nil, fmt.Errorf("failed to parse content type: %v", err)
	}

	var proverInputStoreCfg multistore.Config

	// If File Dir config is set file store
	if gcfg.ProverInputStore.File.Dir != "" {
		proverInputStoreCfg.FileConfig = &filestore.Config{DataDir: filepath.Join(gcfg.DataDir, ChainID(gcfg), gcfg.ProverInputStore.File.Dir)}
	}

	// Configure S3 store
	if gcfg.ProverInputStore.S3.Bucket != "" {
		proverInputStoreCfg.S3Config = &s3store.Config{
			Bucket:    gcfg.ProverInputStore.S3.Bucket,
			KeyPrefix: gcfg.ProverInputStore.S3.BucketKeyPrefix,
			ProviderConfig: &aws.ProviderConfig{
				Region: gcfg.ProverInputStore.S3.AWSProvider.Region,
				Credentials: &aws.CredentialsConfig{
					AccessKey: gcfg.ProverInputStore.S3.AWSProvider.Credentials.AccessKey,
					SecretKey: gcfg.ProverInputStore.S3.AWSProvider.Credentials.SecretKey,
				},
			},
		}
	}

	// Set prover inputs store configuration
	cfg.ProverInputStore = inputstore.ProverInputStoreConfig{
		StoreConfig:     proverInputStoreCfg,
		ContentEncoding: contentEncoding,
		ContentType:     contentType,
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

func ChainID(gcfg *config.Config) string {
	if gcfg.Chain.ID == "" {
		return "default"
	}
	return gcfg.Chain.ID
}
