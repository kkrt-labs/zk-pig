package store

import (
	"context"

	"github.com/kkrt-labs/go-utils/app/svc"
	"github.com/kkrt-labs/go-utils/log"
	"github.com/kkrt-labs/go-utils/tag"
	input "github.com/kkrt-labs/zk-pig/src/prover-input"
	"github.com/kkrt-labs/zk-pig/src/steps"

	"go.uber.org/zap"
)

type taggedProverInputStore struct {
	s      ProverInputStore
	tagged *svc.Tagged
}

func ProverInputStoreWithTags(s ProverInputStore) ProverInputStore {
	return &taggedProverInputStore{
		s:      s,
		tagged: svc.NewTagged(),
	}
}

func (s *taggedProverInputStore) WithTags(tags ...*tag.Tag) {
	s.tagged.WithTags(tags...)
}

func (s *taggedProverInputStore) StoreProverInput(ctx context.Context, inputs *input.ProverInput) error {
	return s.s.StoreProverInput(s.context(ctx, inputs.ChainConfig.ChainID.Uint64(), inputs.Blocks[0].Header.Number.Uint64()), inputs)
}

func (s *taggedProverInputStore) LoadProverInput(ctx context.Context, chainID, blockNumber uint64) (*input.ProverInput, error) {
	return s.s.LoadProverInput(s.context(ctx, chainID, blockNumber), chainID, blockNumber)
}

func (s *taggedProverInputStore) context(ctx context.Context, chainID, blockNumber uint64) context.Context {
	return s.tagged.Context(ctx, tag.Key("chain.id").Int64(int64(chainID)), tag.Key("block.number").Int64(int64(blockNumber)))
}

type loggedProverInputStore struct {
	s ProverInputStore
}

func ProverInputStoreWithLog(s ProverInputStore) ProverInputStore {
	return &loggedProverInputStore{
		s: s,
	}
}

func (s *loggedProverInputStore) StoreProverInput(ctx context.Context, inputs *input.ProverInput) error {
	log.LoggerFromContext(ctx).Debug("Storing prover input")
	err := s.s.StoreProverInput(ctx, inputs)
	if err != nil {
		log.LoggerFromContext(ctx).Error("Failed to store prover input", zap.Error(err))
	}
	log.LoggerFromContext(ctx).Debug("Prover input successfully stored")
	return err
}

func (s *loggedProverInputStore) LoadProverInput(ctx context.Context, chainID, blockNumber uint64) (*input.ProverInput, error) {
	log.LoggerFromContext(ctx).Debug("Loading prover input")
	inputs, err := s.s.LoadProverInput(ctx, chainID, blockNumber)
	if err != nil {
		log.LoggerFromContext(ctx).Error("Failed to load prover input", zap.Error(err))
	}
	log.LoggerFromContext(ctx).Debug("Prover input successfully loaded")
	return inputs, err
}

type taggedPreflightDataStore struct {
	s      PreflightDataStore
	tagged *svc.Tagged
}

func PreflightDataStoreWithTags(s PreflightDataStore) PreflightDataStore {
	return &taggedPreflightDataStore{
		s:      s,
		tagged: svc.NewTagged(),
	}
}

func (s *taggedPreflightDataStore) WithTags(tags ...*tag.Tag) {
	s.tagged.WithTags(tags...)
}

func (s *taggedPreflightDataStore) StorePreflightData(ctx context.Context, data *steps.PreflightData) error {
	return s.s.StorePreflightData(s.context(ctx, data.ChainConfig.ChainID.Uint64(), data.Block.Header.Nonce.Uint64()), data)
}

func (s *taggedPreflightDataStore) LoadPreflightData(ctx context.Context, chainID, blockNumber uint64) (*steps.PreflightData, error) {
	return s.s.LoadPreflightData(s.context(ctx, chainID, blockNumber), chainID, blockNumber)
}

func (s *taggedPreflightDataStore) context(ctx context.Context, chainID, blockNumber uint64) context.Context {
	return s.tagged.Context(ctx, tag.Key("chain.id").Int64(int64(chainID)), tag.Key("block.number").Int64(int64(blockNumber)))
}

type loggedPreflightDataStore struct {
	s PreflightDataStore
}

func PreflightDataStoreWithLog(s PreflightDataStore) PreflightDataStore {
	return &loggedPreflightDataStore{
		s: s,
	}
}

func (s *loggedPreflightDataStore) StorePreflightData(ctx context.Context, data *steps.PreflightData) error {
	log.LoggerFromContext(ctx).Debug("Storing preflight data")
	err := s.s.StorePreflightData(ctx, data)
	if err != nil {
		log.LoggerFromContext(ctx).Error("Failed to store preflight data", zap.Error(err))
	}
	log.LoggerFromContext(ctx).Debug("Preflight data successfully stored")
	return err
}

func (s *loggedPreflightDataStore) LoadPreflightData(ctx context.Context, chainID, blockNumber uint64) (*steps.PreflightData, error) {
	log.LoggerFromContext(ctx).Debug("Loading preflight data")
	data, err := s.s.LoadPreflightData(ctx, chainID, blockNumber)
	if err != nil {
		log.LoggerFromContext(ctx).Error("Failed to load preflight data", zap.Error(err))
	}
	log.LoggerFromContext(ctx).Debug("Preflight data successfully loaded")
	return data, err
}
