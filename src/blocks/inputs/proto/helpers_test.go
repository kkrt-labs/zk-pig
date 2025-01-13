package proto

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	blockinputs "github.com/kkrt-labs/kakarot-controller/src/blocks/inputs"
)

func TestFromGoToProtoProverInputs(t *testing.T) {
	// Load test data
	testData, err := os.ReadFile("../testdata/Ethereum_Mainnet_21465322.json")
	require.NoError(t, err)

	var data struct {
		ProverInputs *blockinputs.ProverInputs `json:"proverInputs"`
	}
	err = json.Unmarshal(testData, &data)
	require.NoError(t, err)

	// Convert HeavyProverInputs to ProverInputs
	goInputs := &blockinputs.ProverInputs{
		Block:       data.ProverInputs.Block,
		Ancestors:   data.ProverInputs.Ancestors,
		ChainConfig: data.ProverInputs.ChainConfig,
		Codes:       data.ProverInputs.Codes,
		PreState:    data.ProverInputs.PreState,
		AccessList:  data.ProverInputs.AccessList,
	}

	// Convert to proto
	protoInputs := ToProto(goInputs)
	require.NotNil(t, protoInputs)

	// Convert back to Go
	backToGo := FromProto(protoInputs)
	require.NotNil(t, backToGo)

	normalizedGoInputs := blockinputs.NormalizeProverInputs(goInputs)
	normalisedBackToGo := blockinputs.NormalizeProverInputs(backToGo)

	assert.Equal(t, normalizedGoInputs.Codes, normalisedBackToGo.Codes)
	assert.Equal(t, normalizedGoInputs.PreState, normalisedBackToGo.PreState)
	assert.Equal(t, normalizedGoInputs.AccessList, normalisedBackToGo.AccessList)
	assert.Equal(t, normalizedGoInputs.ChainConfig, normalisedBackToGo.ChainConfig)
	assert.Equal(t, normalizedGoInputs.Block.BaseFee, normalisedBackToGo.Block.BaseFee)
	assert.Equal(t, normalizedGoInputs.Block.BlobGasUsed, normalisedBackToGo.Block.BlobGasUsed)
	assert.Equal(t, normalizedGoInputs.Block.ExcessBlobGas, normalisedBackToGo.Block.ExcessBlobGas)
	assert.Equal(t, normalizedGoInputs.Block.Withdrawals, normalisedBackToGo.Block.Withdrawals)
	assert.Equal(t, normalizedGoInputs.Block.Nonce, normalisedBackToGo.Block.Nonce)
	assert.Equal(t, len(normalizedGoInputs.Block.Transactions), len(normalisedBackToGo.Block.Transactions))

	for i := range normalizedGoInputs.Block.Transactions {
		assert.Equal(t, normalizedGoInputs.Block.Transactions[i].Hash(), normalisedBackToGo.Block.Transactions[i].Hash())
		assert.Equal(t, normalizedGoInputs.Block.Transactions[i].Gas(), normalisedBackToGo.Block.Transactions[i].Gas())
		assert.Equal(t, normalizedGoInputs.Block.Transactions[i].Type(), normalisedBackToGo.Block.Transactions[i].Type())
		assert.Equal(t, normalizedGoInputs.Block.Transactions[i].To(), normalisedBackToGo.Block.Transactions[i].To())
		assert.Equal(t, normalizedGoInputs.Block.Transactions[i].From, normalisedBackToGo.Block.Transactions[i].From)
		v, r, s := normalizedGoInputs.Block.Transactions[i].RawSignatureValues()
		v1, r1, s1 := normalisedBackToGo.Block.Transactions[i].RawSignatureValues()
		assert.Equal(t, v, v1)
		assert.Equal(t, r, r1)
		assert.Equal(t, s, s1)
		assert.Equal(t, normalizedGoInputs.Block.Transactions[i].ChainId(), normalisedBackToGo.Block.Transactions[i].ChainId())
		assert.Equal(t, normalizedGoInputs.Block.Transactions[i].GasPrice(), normalisedBackToGo.Block.Transactions[i].GasPrice())
		assert.Equal(t, normalizedGoInputs.Block.Transactions[i].GasTipCap(), normalisedBackToGo.Block.Transactions[i].GasTipCap())
		assert.Equal(t, normalizedGoInputs.Block.Transactions[i].GasFeeCap(), normalisedBackToGo.Block.Transactions[i].GasFeeCap())
		assert.Equal(t, normalizedGoInputs.Block.Transactions[i].AccessList(), normalisedBackToGo.Block.Transactions[i].AccessList())
		assert.Equal(t, normalizedGoInputs.Block.Transactions[i].Data(), normalisedBackToGo.Block.Transactions[i].Data())
		assert.Equal(t, normalizedGoInputs.Block.Transactions[i].Value(), normalisedBackToGo.Block.Transactions[i].Value())
	}
}
