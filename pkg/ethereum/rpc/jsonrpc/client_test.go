package jsonrpc

import (
	"context"
	"embed"
	"math/big"
	"testing"

	geth "github.com/ethereum/go-ethereum"
	gethbind "github.com/ethereum/go-ethereum/accounts/abi/bind"
	gethcommon "github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	gethclient "github.com/ethereum/go-ethereum/ethclient/gethclient"
	gethrpc "github.com/ethereum/go-ethereum/rpc"
	"github.com/kkrt-labs/kakarot-controller/pkg/ethereum/rpc"
	jsonrpchttp "github.com/kkrt-labs/kakarot-controller/pkg/jsonrpc/http"
	httptestutils "github.com/kkrt-labs/kakarot-controller/pkg/net/http/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

var (
	//go:embed testdata
	testdataFS embed.FS
)

func TestClientImplementsAllGetInterface(t *testing.T) {
	assert.Implements(t, (*rpc.Client)(nil), new(Client))
}

func TestClientImplementsInterface(t *testing.T) {
	client := new(Client)
	assert.Implements(t, (*gethbind.ContractCaller)(nil), client)
	assert.Implements(t, (*gethbind.ContractTransactor)(nil), client)
	assert.Implements(t, (*gethbind.ContractFilterer)(nil), client)
	assert.Implements(t, (*gethbind.ContractBackend)(nil), client)
	assert.Implements(t, (*geth.ChainReader)(nil), client)
	assert.Implements(t, (*geth.ChainStateReader)(nil), client)
	assert.Implements(t, (*geth.ChainSyncReader)(nil), client)
	assert.Implements(t, (*geth.ContractCaller)(nil), client)
	assert.Implements(t, (*geth.GasEstimator)(nil), client)
	assert.Implements(t, (*geth.GasPricer)(nil), client)
	assert.Implements(t, (*geth.LogFilterer)(nil), client)
	assert.Implements(t, (*geth.PendingContractCaller)(nil), client)
	assert.Implements(t, (*geth.PendingStateReader)(nil), client)
	assert.Implements(t, (*geth.TransactionReader)(nil), client)
	assert.Implements(t, (*geth.TransactionSender)(nil), client)
	assert.Implements(t, (*geth.BlockNumberReader)(nil), client)
	assert.Implements(t, (*geth.ChainIDReader)(nil), client)

}

func TestClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockCli := httptestutils.NewMockSender(ctrl)
	c := NewFromClient(jsonrpchttp.NewClientFromClient(mockCli))

	t.Run("BlockNumber", func(t *testing.T) { testBlockNumber(t, c, mockCli) })
	t.Run("HeaderByNumber", func(t *testing.T) { testHeaderByNumber(t, c, mockCli) })
	t.Run("HeaderByNumber_Finalized", func(t *testing.T) { testBlockByNumberFinalized(t, c, mockCli) })
	t.Run("BlockByNumber", func(t *testing.T) { testBlockByNumber(t, c, mockCli) })
	t.Run("BlockByHash", func(t *testing.T) { testBlockByHash(t, c, mockCli) })
	t.Run("CallContract", func(t *testing.T) { testCallContract(t, c, mockCli) })
	t.Run("NonceAt", func(t *testing.T) { testNonceAt(t, c, mockCli) })
	t.Run("PendingNonceAt", func(t *testing.T) { testPendingNonceAt(t, c, mockCli) })
	t.Run("SuggestGasPrice", func(t *testing.T) { testSuggestGasPrice(t, c, mockCli) })
	t.Run("SuggestGasTipCap", func(t *testing.T) { testSuggestGasTipCap(t, c, mockCli) })
	t.Run("EstimateGas", func(t *testing.T) { testEstimateGas(t, c, mockCli) })
	t.Run("SendTransaction", func(t *testing.T) { testSendTransaction(t, c, mockCli) })

	t.Run("GetProof", func(t *testing.T) { testGetProof(t, c, mockCli) })
}

func testBlockNumber(t *testing.T, c *Client, mockCli *httptestutils.MockSender) {
	req := httptestutils.NewGockRequest()
	req.Post("/").
		JSON([]byte(`{"jsonrpc":"","method":"eth_blockNumber","params":null}`)).
		Reply(200).
		JSON([]byte(`{"jsonrpc":"2.0","result":"0x20","id":0}`))

	mockCli.EXPECT().Gock(req)

	blockNumber, err := c.BlockNumber(context.Background())

	require.NoError(t, err)
	assert.Equal(t, uint64(32), blockNumber)
}

func testBlockByNumberFinalized(t *testing.T, c *Client, mockCli *httptestutils.MockSender) {
	res, _ := testdataFS.ReadFile("testdata/eth_getBlockByNumber_finalized_true.json")
	require.NotEmpty(t, res, "response should not be empty (check typo in testdata filename)")

	req := httptestutils.NewGockRequest()
	req.Post("/").
		JSON([]byte(`{"jsonrpc":"","method":"eth_getBlockByNumber","params":["finalized",true]}`)).
		Reply(200).
		JSON(res)

	mockCli.EXPECT().Gock(req)

	block, err := c.BlockByNumber(context.Background(), big.NewInt(int64(gethrpc.FinalizedBlockNumber)))

	require.NoError(t, err)
	assert.Equal(
		t,
		&gethtypes.Header{
			ParentHash:  gethcommon.HexToHash("0x6019a4b3e4e3ba7b7b43d28d68492f99226b86e7dff0c607a16ef4d16a617503"),
			UncleHash:   gethcommon.HexToHash("0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347"),
			Coinbase:    gethcommon.HexToAddress("0x52bc44d5378309EE2abF1539BF71dE1b7d7bE3b5"),
			Root:        gethcommon.HexToHash("0x4a4e5f11b8e837adb24fb764ab93f33ed21efa279df4fe59b5bed3c3885e9fae"),
			TxHash:      gethcommon.HexToHash("0x5cb8acbd8a0d2f3c489e47d8267c86a718203da8a5a34f0511918c13cbb14c1b"),
			ReceiptHash: gethcommon.HexToHash("0x081119bc627ccedade0b6321984146672ad1a15b0769b08f7a91ea22474c7bd9"),
			Bloom:       gethtypes.BytesToBloom(gethcommon.FromHex("0x1fbb5f53e8e63cedffe45fd8bf1217fdee15d39bbebf275136afb8ffb99fdd9b92556ffb2ceeb1345a3bf1dd730ebfc6bf4c814119e6faaef2f9fa9b50ffe8fd838eb2bed773592efb0ffc7efd142fe37fe65117f5f4f7bb2f037671a4ff52d443a7044a1be25ec1fb1b13a9aabf6afdd278f4bf4abda64e3293cb9480f97d11c9558ded275cdf8ed5ef7f43398e9fb5fe4e2e0d79257cecebf95bd36e99a8f7bbdab5323febe6baceb1dfdda71cbe21dfbcc6a3feee6702fd85a6bd3ee9f8dc757ca4bacdf3a47ef119c3d95feb5d2f65acffdb9effa17ebb5fdb1b3afe64dfd8fcf3bfa8787f882e660d33cfe7fb9220ef6226efd5dffafcc7daa3b6967faf")),
			Difficulty:  big.NewInt(12795344477503252),
			Number:      big.NewInt(14082406),
			GasLimit:    29999972,
			GasUsed:     29984188,
			Time:        uint64(1643215331),
			Extra:       gethcommon.FromHex("0x6e616e6f706f6f6c2e6f7267"),
			MixDigest:   gethcommon.HexToHash("0x274264e3a69256c43beb4632b6bf8ac2de6534dd6c4fb09dad1a0541eb8ed356"),
			Nonce:       gethtypes.EncodeNonce(3448329947143578346),
			BaseFee:     big.NewInt(121064488104),
		},
		block.Header(),
	)
	assert.Equal(t, 277, block.Transactions().Len())
}

func testBlockByNumber(t *testing.T, c *Client, mockCli *httptestutils.MockSender) {
	res, _ := testdataFS.ReadFile("testdata/eth_getBlockByNumber_0xd6e166_true.json")
	require.NotEmpty(t, res, "response should not be empty (check typo in testdata filename)")

	req := httptestutils.NewGockRequest()
	req.Post("/").
		JSON([]byte(`{"jsonrpc":"","method":"eth_getBlockByNumber","params":["0xd6e166",true]}`)).
		Reply(200).
		JSON(res)

	mockCli.EXPECT().Gock(req)

	block, err := c.BlockByNumber(context.Background(), big.NewInt(14082406))

	require.NoError(t, err)
	assert.Equal(
		t,
		&gethtypes.Header{
			ParentHash:  gethcommon.HexToHash("0x6019a4b3e4e3ba7b7b43d28d68492f99226b86e7dff0c607a16ef4d16a617503"),
			UncleHash:   gethcommon.HexToHash("0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347"),
			Coinbase:    gethcommon.HexToAddress("0x52bc44d5378309EE2abF1539BF71dE1b7d7bE3b5"),
			Root:        gethcommon.HexToHash("0x4a4e5f11b8e837adb24fb764ab93f33ed21efa279df4fe59b5bed3c3885e9fae"),
			TxHash:      gethcommon.HexToHash("0x5cb8acbd8a0d2f3c489e47d8267c86a718203da8a5a34f0511918c13cbb14c1b"),
			ReceiptHash: gethcommon.HexToHash("0x081119bc627ccedade0b6321984146672ad1a15b0769b08f7a91ea22474c7bd9"),
			Bloom:       gethtypes.BytesToBloom(gethcommon.FromHex("0x1fbb5f53e8e63cedffe45fd8bf1217fdee15d39bbebf275136afb8ffb99fdd9b92556ffb2ceeb1345a3bf1dd730ebfc6bf4c814119e6faaef2f9fa9b50ffe8fd838eb2bed773592efb0ffc7efd142fe37fe65117f5f4f7bb2f037671a4ff52d443a7044a1be25ec1fb1b13a9aabf6afdd278f4bf4abda64e3293cb9480f97d11c9558ded275cdf8ed5ef7f43398e9fb5fe4e2e0d79257cecebf95bd36e99a8f7bbdab5323febe6baceb1dfdda71cbe21dfbcc6a3feee6702fd85a6bd3ee9f8dc757ca4bacdf3a47ef119c3d95feb5d2f65acffdb9effa17ebb5fdb1b3afe64dfd8fcf3bfa8787f882e660d33cfe7fb9220ef6226efd5dffafcc7daa3b6967faf")),
			Difficulty:  big.NewInt(12795344477503252),
			Number:      big.NewInt(14082406),
			GasLimit:    29999972,
			GasUsed:     29984188,
			Time:        uint64(1643215331),
			Extra:       gethcommon.FromHex("0x6e616e6f706f6f6c2e6f7267"),
			MixDigest:   gethcommon.HexToHash("0x274264e3a69256c43beb4632b6bf8ac2de6534dd6c4fb09dad1a0541eb8ed356"),
			Nonce:       gethtypes.EncodeNonce(3448329947143578346),
			BaseFee:     big.NewInt(121064488104),
		},
		block.Header(),
	)
	assert.Equal(t, 277, block.Transactions().Len())
}

func testBlockByHash(t *testing.T, c *Client, mockCli *httptestutils.MockSender) {
	res, _ := testdataFS.ReadFile("testdata/eth_getBlockByHash_0x0fb6d5609c9edab75bf587ea7449e6e6940d6e3df1992a1bd96ca8b74ffd16fc_true.json")
	require.NotEmpty(t, res, "response should not be empty (check typo in testdata filename)")

	req := httptestutils.NewGockRequest()
	req.Post("/").
		JSON([]byte(`{"jsonrpc":"","method":"eth_getBlockByHash","params":["0x0fb6d5609c9edab75bf587ea7449e6e6940d6e3df1992a1bd96ca8b74ffd16fc",true]}`)).
		Reply(200).
		JSON(res)

	mockCli.EXPECT().Gock(req)

	block, err := c.BlockByHash(context.Background(), gethcommon.HexToHash("0x0fb6d5609c9edab75bf587ea7449e6e6940d6e3df1992a1bd96ca8b74ffd16fc"))

	require.NoError(t, err)
	assert.Equal(
		t,
		&gethtypes.Header{
			ParentHash:  gethcommon.HexToHash("0x6019a4b3e4e3ba7b7b43d28d68492f99226b86e7dff0c607a16ef4d16a617503"),
			UncleHash:   gethcommon.HexToHash("0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347"),
			Coinbase:    gethcommon.HexToAddress("0x52bc44d5378309EE2abF1539BF71dE1b7d7bE3b5"),
			Root:        gethcommon.HexToHash("0x4a4e5f11b8e837adb24fb764ab93f33ed21efa279df4fe59b5bed3c3885e9fae"),
			TxHash:      gethcommon.HexToHash("0x5cb8acbd8a0d2f3c489e47d8267c86a718203da8a5a34f0511918c13cbb14c1b"),
			ReceiptHash: gethcommon.HexToHash("0x081119bc627ccedade0b6321984146672ad1a15b0769b08f7a91ea22474c7bd9"),
			Bloom:       gethtypes.BytesToBloom(gethcommon.FromHex("0x1fbb5f53e8e63cedffe45fd8bf1217fdee15d39bbebf275136afb8ffb99fdd9b92556ffb2ceeb1345a3bf1dd730ebfc6bf4c814119e6faaef2f9fa9b50ffe8fd838eb2bed773592efb0ffc7efd142fe37fe65117f5f4f7bb2f037671a4ff52d443a7044a1be25ec1fb1b13a9aabf6afdd278f4bf4abda64e3293cb9480f97d11c9558ded275cdf8ed5ef7f43398e9fb5fe4e2e0d79257cecebf95bd36e99a8f7bbdab5323febe6baceb1dfdda71cbe21dfbcc6a3feee6702fd85a6bd3ee9f8dc757ca4bacdf3a47ef119c3d95feb5d2f65acffdb9effa17ebb5fdb1b3afe64dfd8fcf3bfa8787f882e660d33cfe7fb9220ef6226efd5dffafcc7daa3b6967faf")),
			Difficulty:  big.NewInt(12795344477503252),
			Number:      big.NewInt(14082406),
			GasLimit:    29999972,
			GasUsed:     29984188,
			Time:        uint64(1643215331),
			Extra:       gethcommon.FromHex("0x6e616e6f706f6f6c2e6f7267"),
			MixDigest:   gethcommon.HexToHash("0x274264e3a69256c43beb4632b6bf8ac2de6534dd6c4fb09dad1a0541eb8ed356"),
			Nonce:       gethtypes.EncodeNonce(3448329947143578346),
			BaseFee:     big.NewInt(121064488104),
		},
		block.Header(),
	)
	assert.Equal(t, 277, block.Transactions().Len())
}

func testHeaderByNumber(t *testing.T, c *Client, mockCli *httptestutils.MockSender) {
	req := httptestutils.NewGockRequest()
	req.Post("/").
		JSON([]byte(`{"jsonrpc":"","method":"eth_getBlockByNumber","params":["0xd6e166",false]}`)).
		Reply(200).
		JSON([]byte(`{"jsonrpc":"2.0","id":1,"result":{"baseFeePerGas":"0x1c30017ca8","difficulty":"0x2d754c4a5c3f14","extraData":"0x6e616e6f706f6f6c2e6f7267","gasLimit":"0x1c9c364","gasUsed":"0x1c985bc","hash":"0x0fb6d5609c9edab75bf587ea7449e6e6940d6e3df1992a1bd96ca8b74ffd16fc","logsBloom":"0x1fbb5f53e8e63cedffe45fd8bf1217fdee15d39bbebf275136afb8ffb99fdd9b92556ffb2ceeb1345a3bf1dd730ebfc6bf4c814119e6faaef2f9fa9b50ffe8fd838eb2bed773592efb0ffc7efd142fe37fe65117f5f4f7bb2f037671a4ff52d443a7044a1be25ec1fb1b13a9aabf6afdd278f4bf4abda64e3293cb9480f97d11c9558ded275cdf8ed5ef7f43398e9fb5fe4e2e0d79257cecebf95bd36e99a8f7bbdab5323febe6baceb1dfdda71cbe21dfbcc6a3feee6702fd85a6bd3ee9f8dc757ca4bacdf3a47ef119c3d95feb5d2f65acffdb9effa17ebb5fdb1b3afe64dfd8fcf3bfa8787f882e660d33cfe7fb9220ef6226efd5dffafcc7daa3b6967faf","miner":"0x52bc44d5378309ee2abf1539bf71de1b7d7be3b5","mixHash":"0x274264e3a69256c43beb4632b6bf8ac2de6534dd6c4fb09dad1a0541eb8ed356","nonce":"0x2fdaedd11fd5a2ea","number":"0xd6e166","parentHash":"0x6019a4b3e4e3ba7b7b43d28d68492f99226b86e7dff0c607a16ef4d16a617503","receiptsRoot":"0x081119bc627ccedade0b6321984146672ad1a15b0769b08f7a91ea22474c7bd9","sha3Uncles":"0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347","size":"0x43a1d","stateRoot":"0x4a4e5f11b8e837adb24fb764ab93f33ed21efa279df4fe59b5bed3c3885e9fae","timestamp":"0x61f179e3","totalDifficulty":"0x873cd0f1a366947ae8d","transactions":[],"transactionsRoot":"0x5cb8acbd8a0d2f3c489e47d8267c86a718203da8a5a34f0511918c13cbb14c1b","uncles":[]},"id":0}`))

	mockCli.EXPECT().Gock(req)

	header, err := c.HeaderByNumber(context.Background(), big.NewInt(14082406))

	require.NoError(t, err)
	assert.Equal(
		t,
		&gethtypes.Header{
			ParentHash:  gethcommon.HexToHash("0x6019a4b3e4e3ba7b7b43d28d68492f99226b86e7dff0c607a16ef4d16a617503"),
			UncleHash:   gethcommon.HexToHash("0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347"),
			Coinbase:    gethcommon.HexToAddress("0x52bc44d5378309EE2abF1539BF71dE1b7d7bE3b5"),
			Root:        gethcommon.HexToHash("0x4a4e5f11b8e837adb24fb764ab93f33ed21efa279df4fe59b5bed3c3885e9fae"),
			TxHash:      gethcommon.HexToHash("0x5cb8acbd8a0d2f3c489e47d8267c86a718203da8a5a34f0511918c13cbb14c1b"),
			ReceiptHash: gethcommon.HexToHash("0x081119bc627ccedade0b6321984146672ad1a15b0769b08f7a91ea22474c7bd9"),
			Bloom:       gethtypes.BytesToBloom(gethcommon.FromHex("0x1fbb5f53e8e63cedffe45fd8bf1217fdee15d39bbebf275136afb8ffb99fdd9b92556ffb2ceeb1345a3bf1dd730ebfc6bf4c814119e6faaef2f9fa9b50ffe8fd838eb2bed773592efb0ffc7efd142fe37fe65117f5f4f7bb2f037671a4ff52d443a7044a1be25ec1fb1b13a9aabf6afdd278f4bf4abda64e3293cb9480f97d11c9558ded275cdf8ed5ef7f43398e9fb5fe4e2e0d79257cecebf95bd36e99a8f7bbdab5323febe6baceb1dfdda71cbe21dfbcc6a3feee6702fd85a6bd3ee9f8dc757ca4bacdf3a47ef119c3d95feb5d2f65acffdb9effa17ebb5fdb1b3afe64dfd8fcf3bfa8787f882e660d33cfe7fb9220ef6226efd5dffafcc7daa3b6967faf")),
			Difficulty:  big.NewInt(12795344477503252),
			Number:      big.NewInt(14082406),
			GasLimit:    29999972,
			GasUsed:     29984188,
			Time:        uint64(1643215331),
			Extra:       gethcommon.FromHex("0x6e616e6f706f6f6c2e6f7267"),
			MixDigest:   gethcommon.HexToHash("0x274264e3a69256c43beb4632b6bf8ac2de6534dd6c4fb09dad1a0541eb8ed356"),
			Nonce:       gethtypes.EncodeNonce(3448329947143578346),
			BaseFee:     big.NewInt(121064488104),
		},
		header,
	)
}

func testCallContract(t *testing.T, c *Client, mockCli *httptestutils.MockSender) {
	req := httptestutils.NewGockRequest()
	req.Post("/").
		JSON([]byte(`{"jsonrpc":"","method":"eth_call","params":[{"data":"0x0123456789","from":"0x52bc44d5378309ee2abf1539bf71de1b7d7be3b5","to":null},"0xd6e166"]}`)).
		Reply(200).
		JSON([]byte(`{"jsonrpc":"2.0","result":"0xabcdef","id":0}`))

	mockCli.EXPECT().Gock(req)

	res, err := c.CallContract(
		context.Background(),
		geth.CallMsg{
			From: gethcommon.HexToAddress("0x52bc44d5378309EE2abF1539BF71dE1b7d7bE3b5"),
			Data: gethcommon.FromHex("0x0123456789"),
		},
		big.NewInt(14082406),
	)

	require.NoError(t, err)
	assert.Equal(t, gethcommon.FromHex("0xabcdef"), res)
}

func testNonceAt(t *testing.T, c *Client, mockCli *httptestutils.MockSender) {
	req := httptestutils.NewGockRequest()
	req.Post("/").
		JSON([]byte(`{"jsonrpc":"","method":"eth_getTransactionCount","params":["0x52bc44d5378309ee2abf1539bf71de1b7d7be3b5","0xd6e6f3"]}`)).
		Reply(200).
		JSON([]byte(`{"jsonrpc":"2.0","id":1,"result":"0x11189c7"}`))

	mockCli.EXPECT().Gock(req)

	nonce, err := c.NonceAt(
		context.Background(),
		gethcommon.HexToAddress("0x52bc44d5378309ee2abf1539bf71de1b7d7be3b5"),
		big.NewInt(14083827),
	)

	require.NoError(t, err)
	assert.Equal(t, uint64(17926599), nonce)
}

func testPendingNonceAt(t *testing.T, c *Client, mockCli *httptestutils.MockSender) {
	req := httptestutils.NewGockRequest()
	req.Post("/").
		JSON([]byte(`{"jsonrpc":"","method":"eth_getTransactionCount","params":["0x52bc44d5378309ee2abf1539bf71de1b7d7be3b5","pending"]}`)).
		Reply(200).
		JSON([]byte(`{"jsonrpc":"2.0","id":1,"result":"0x11189c7"}`))

	mockCli.EXPECT().Gock(req)

	nonce, err := c.PendingNonceAt(
		context.Background(),
		gethcommon.HexToAddress("0x52bc44d5378309ee2abf1539bf71de1b7d7be3b5"),
	)

	require.NoError(t, err)
	assert.Equal(t, uint64(17926599), nonce)
}

func testSuggestGasPrice(t *testing.T, c *Client, mockCli *httptestutils.MockSender) {
	req := httptestutils.NewGockRequest()
	req.Post("/").
		JSON([]byte(`{"jsonrpc":"","method":"eth_gasPrice","params":null}`)).
		Reply(200).
		JSON([]byte(`{"jsonrpc":"2.0","id":1,"result":"0x2fbbd1aa9b"}`))

	mockCli.EXPECT().Gock(req)

	p, err := c.SuggestGasPrice(context.Background())

	require.NoError(t, err)
	assert.Equal(t, big.NewInt(205014543003), p)
}

func testSuggestGasTipCap(t *testing.T, c *Client, mockCli *httptestutils.MockSender) {
	req := httptestutils.NewGockRequest()
	req.Post("/").
		JSON([]byte(`{"jsonrpc":"","method":"eth_maxPriorityFeePerGas","params":null}`)).
		Reply(200).
		JSON([]byte(`{"jsonrpc":"2.0","id":1,"result":"0x2fbbd1aa9b"}`))

	mockCli.EXPECT().Gock(req)

	p, err := c.SuggestGasTipCap(context.Background())

	require.NoError(t, err)
	assert.Equal(t, big.NewInt(205014543003), p)
}

func testEstimateGas(t *testing.T, c *Client, mockCli *httptestutils.MockSender) {
	req := httptestutils.NewGockRequest()
	req.Post("/").
		JSON([]byte(`{"jsonrpc":"","method":"eth_estimateGas","params":[{"data":"0x0123456789","from":"0x52bc44d5378309ee2abf1539bf71de1b7d7be3b5","to":null}]}`)).
		Reply(200).
		JSON([]byte(`{"jsonrpc":"2.0","result":"0xabcdef","id":0}`))

	mockCli.EXPECT().Gock(req)

	gas, err := c.EstimateGas(
		context.Background(),
		geth.CallMsg{
			From: gethcommon.HexToAddress("0x52bc44d5378309EE2abF1539BF71dE1b7d7bE3b5"),
			Data: gethcommon.FromHex("0x0123456789"),
		},
	)

	require.NoError(t, err)
	assert.Equal(t, uint64(11259375), gas)
}

func testSendTransaction(t *testing.T, c *Client, mockCli *httptestutils.MockSender) {
	req := httptestutils.NewGockRequest()
	req.Post("/").
		JSON([]byte(`{"jsonrpc":"","method":"eth_sendRawTransaction","params":["0xf86d8202b38477359400825208944592d8f8d7b001e72cb26a73e4fa1806a51ac79d880de0b6b3a7640000802ba0699ff162205967ccbabae13e07cdd4284258d46ec1051a70a51be51ec2bc69f3a04e6944d508244ea54a62ebf9a72683eeadacb73ad7c373ee542f1998147b220e"]}`)).
		Reply(200).
		JSON([]byte(`{"jsonrpc":"2.0","result":"0x679bdd54941acaebcf592035101606b56087048ebb7ea12a02df4a6be426f8dd","id":0}`))

	mockCli.EXPECT().Gock(req)

	tx := &gethtypes.Transaction{}
	_ = tx.UnmarshalBinary(gethcommon.FromHex("0xf86d8202b38477359400825208944592d8f8d7b001e72cb26a73e4fa1806a51ac79d880de0b6b3a7640000802ba0699ff162205967ccbabae13e07cdd4284258d46ec1051a70a51be51ec2bc69f3a04e6944d508244ea54a62ebf9a72683eeadacb73ad7c373ee542f1998147b220e"))

	err := c.SendTransaction(
		context.Background(),
		tx,
	)

	require.NoError(t, err)
}

func testGetProof(t *testing.T, c *Client, mockCli *httptestutils.MockSender) {
	res, _ := testdataFS.ReadFile("testdata/eth_getProof.json")
	require.NotEmpty(t, res, "response should not be empty (check typo in testdata filename)")

	req := httptestutils.NewGockRequest()
	req.Post("/").
		JSON([]byte(`{"jsonrpc":"","method":"eth_getProof","params":["0x4aa30adb6d6616ae2c9927050fd740ff4061c0ec",["key1","key2"],"latest"]}`)).
		Reply(200).
		JSON(res)

	mockCli.EXPECT().Gock(req)

	proof, err := c.GetProof(context.Background(), gethcommon.HexToAddress("0x4aa30adb6d6616ae2c9927050fd740ff4061c0ec"), []string{"key1", "key2"}, nil)

	require.NoError(t, err)
	assert.Equal(
		t,
		&gethclient.AccountResult{
			Address: gethcommon.HexToAddress("0x87870bca3f3fd6335c3f4ce8392d69350b4fa4e2"),
			AccountProof: []string{
				"0xf90211a03c5bbf4ac6247bbb6e1e7a6fba238b841ae4735d0d7d9ac870e52af7d76c38cca0fc05e8eb3c266f569273e2d082965a6f3de00df833b723e3bb5037316673c1f8a038f199e69e44b0f343974f07f9a98960f9322c3b43ee341cd00f26055b32bc3da07839b4ef3fe57aa4970b450e4c4db7f2a5a0e4803f84f0270c7030a6b158d92ca0c57e8a05cdc807e634f92b29479d14d0149c9669b2c3302b404cb794f066e49aa0618040af05177beb56db3ff6f1b3e139eff42541489f064c9b6ca8403c90bc2ca05357f48ca00b1e8f74c377b5190b9a19f1c0f0b1943aa13cd44c03eb1c512c1fa07282ac80c9878d2b4ef4acb9f505058444b121326bfd4d6cfc6c9441581b16fea01314ef3cbefdcc707d6cec2372087e6a373e1542fe2740f580a947df78a32db7a0bca2b81164f0c80106af54f49dd8cbf45277ec80d262b3fdbcd9d6478abc8344a0c358eb569aa3729af26d42e4b7391b94d81d24716c23eabce54e188beb0e2c4da046d7c7b7cddcdc0808891e7276d2d2055327f1c9930746e8f3e649efcf95f9a4a098eb734443384f7a20109e387c7f87fcb2825dd17ba74260c4708c243c94eb81a0154e694fe76b89c86f7dba7d6ae74bd63fa2004ea62732fb214f9308b257b5fda09b52db28538c52b7e89e91f67d0784eed03d34129f834d02b35c9564ffcfd4cea0deddb47a9caa153c73ef83e8cc3ab2c27e8409acf6add4d9fc84507d64096a0a80",
				"0xf90211a0447d713534bf06a6b7e7c81750ea31fb7890f8e5effc1df44fe5b5afc6098b13a024c57345b94811710fef0171fef7a222cd4d9de333e73efc317b587ee5d39d6ea056fe3f5383399417bdb3226aa9cd3656e454eaedd0161221c5113147e12dd7e9a0be7952327b3422cfcdbacc3c2d566d4403b6398726aa14882349188c1e58f7d5a037da4d4eeda3e2f5115a91f51d1727c5b3784b4f72a4384d14ce2663520c8a5aa0388f72fc36cb6ab907d04416aa495f641c9fab23440e0de8ae3ac9abd07bdc67a0c59dc0e4a420a4ba51480e3a239bb1626bf13ede19052b99eba6e83124171d13a07339d6243bc734beb6f7eb951f31b332912dc89b78ce6b6797f298e9f8bb0567a0b783b261ff2421b0d2b340ee172c6fb866f871f822ce6cfdd3eb79f9f3b682aaa087cd34bc59731b658c2fde7f42b920f6034d89f0aeea6488bad7c69c30a5711ca084866ce013762ec3815ae39f5e67e0a2a344107e81e4217868d1d06c5358715da015ff43c3e3f54b95df4dbdcd4a5dd7b800198e2c0c47d81c1f016ab946137c85a0722529e2dbdbb1f6983d32b517808a77c96fce45d3b4899f41db746e32e0f551a06c31931cf1fac8508a4d37afade11ff7156cb2122f3e3bad71e6aff492d384b0a04c6cd4e6c11725f85893e0fe2eb9558ad4214ce48eca2a08c2ed78109694a479a0eb85aff2f9d4c2c88cb941c5de62c63a8fc1222c5f684265287e3e3f16bfb10680",
				"0xf90211a0061aeab470a17b0a0cbc58abbd7408482f4a6e5afb256620a3ae5e02eac53176a07a37881a0a281c49da386b24844c21c7cc0f9271a800abfa78a288d8e79bf024a035313e47bfbef64ff062ba71905b274bf794a782d816834986553c5f22535612a0ce83c8dde0365ebbea24f2970d85368da56f9fd9cc5b969d6385e54d2579991ca00655e3712cb7f522bd7ae04d07f5fb903535acd6576c6c0a3bf7f301608cebf2a07fbbac7161ed827c57723c66b6f1b1ed010038318dea7ee1bda2daae7047a848a03595469999c00699c702985e96e2d492c2d7a6498ae72666bea1752d46094478a0791b4f33487d150d7c2381e68faa9d6748e748452736988d454ffbcf0a001586a0efab71ddbb1e7aaed8c16bfeaaf70c1f0d697c78b29ef5f7d5cc30227033cfc5a0706bcd2ec88756010333e04c23609981f326c43e08d496b06c3759382849d595a08a7cc9e6c2edca2d2b92737b1721e2dc00b6466abfd729197dc749346e95c450a02d265a0c7b1801140e38e79c71b3e1b1925eea4e095fa98e1322e2a9e5ca35bda0858299a62155eec466753a8a40d65c55b453f535f4b9292981df9b061df1ba54a07b660bbac5566bccbadd02f79ad5754516e31d9c1b7d9e9f6bfc60d3bd076382a00201669b8cba40973eea164e080da2170e26c1bc7cdde008f69462545050fbf4a0b0ed174219b70c8863ec9d52658b606a177fe9bedb33979826ee12627e2aca5880",
				"0xf90211a01a8eb6d31cce1b3b56a10087cbbd6beed4aa5ac96140faa6d112955505527f50a0ed1f89e6521add0a73167ad53814438f88c71253ac1db17151423f6f048e0fd0a0ce43e6ef559f04b963459398478a94f0ef69f26ced03f97e4a3a07269ac8edcda01b3b28683e8abbe6c4e0cc45ad480da8e26c04c225c65aa5dcb8e82db2e28145a0dc7ad79596c1baac95d94152c116fb46a1dd6e605e0218a818dfdc113ceca5f8a07e713fdc841172bc7744165643368d6c4356b3e8926d7c466828ec4bbbeb1c59a06654c2797437cd480081c482a8076082d5cb5f0bc4e887bd31b5c30d43ed22faa05555c9d73da13f6699fb1d9ae9d0e70516297ad4385b8fb2342c6d681bc66d9aa0da330d7428e11bb5d119e92bd439169e5a078e698ce7b5bfc7c5b2fdd4275f51a0262e634ed15a0964802fc7d0f7222c5e0b735a6df5584ab86d96ca868ecfc45ca048efcb15eb3aa86e6046d9b0c0d266ca353975bcf798f27db440561dd7bd089da0b55652bec0063b77615d397b1b01221a0d18886e819376c3f2d56aa70c017ca9a09de5b708266542928898af483a7b26b9e0aaa962b61b4d5c4cb306f8d6418759a0488881c26f1a83ef85151ed42b808c5a376418e8742f23402e52c91ebaecc96da0904359b9e28b6163b39a9081e2804edf6c1ed8002c79710ff9e06921e70d55d8a05bbd13e24e30f6c65bfe1879fe7d2acdaade77b26f253d8eba0addb6936891fb80",
				"0xf90211a041ac2d07bbd039bc06a1bd6028c9f527e1fb7147b1117509f1480b53740aa7eca05654f2dfc94079451298ffc03aa76f79196756105046db5300b6cb1a40882e9ea0a7dce13db7559a1068fd28e93f6ba82cb6119cdf32581807abe03dbcd0b172d0a0a7d37c52dcadc6086023c7dd027cd9d4c2a88611005d530ca60922bd58c0fbbfa0b39e045b0c1be94a203851aae865ab6b7f136b27d22ac00b4dba576a03ebc477a018210c247992da617f75242cce5e3b03b4021292bb12da89af20cc564b6c957aa0eaa41e07ef5237e51ee496d4973f1c560214f29c24baf34c8566def80c44d495a0adf1977e8c85e5373a05f732514f979e4be1dbe9927b5bb0d9beb334ca8a5fe0a05228a8d7cf648e15e10ec847cb1b0cc465633928e4170bb3dbba6c0553ea77c5a0b4b3768c2e1e343b40912307d0e1fc2b803bfb5c7a4fc675c4edf3e81eb0015ba09feb413a74b355c11df72900a32e45c74ef677d2c1fc490cd171c485a8b1735aa00a63a26d08cd00377417ccdcf5668bdfd134bdde7cf885a78eb506674b40ff4ea058cb612a03954200256c999c96cfb63b2f836c70549a62fd7991ca87ac27098da03fee3c0e6c99e4081f8d996294697c9c03a44ad18c3063eaaccfe9f52caeb06aa0a62484523ee85e26bab4b8a9c2acc30e71c97f44b01a8b18b8290b6bd54163dda0aa85118c8386f340e6101f4f8cb666ff5544fcbf3c057362288ed25a44865ab480",
				"0xf90211a0d0bc96cadc0093312f3dcfa25013c75a26dd0e2f3058192f0fa0a430daee3e3ba01004a21c0ca5a847e8197c96b8d01de2f36047d64d88197a17ae2c714bf02b36a0cd8ad866c168a58035c986cc312eb188e1bd96a3fe60cdf75ba9f95da63d20c9a07ac84c71ad7aedfc86666991cf28029d9552c2512e53179d3958b8dcdab5cd77a02d0cad87483387171c94bdfd77b04124e97aa68fda4376618c02050fb7ca2c02a0e8f90372eef7d0beb24014bce260044b79d08c9f699b77c56bbd68088fc750fca08283b06985b1059546cd8c21fcf439d6643f729cb2ed81739b2a4c79096aaadda0392693f53e29fc6d47a670bf699ee8960ec6661e6070f550e9d32134b2cd32a3a0e9cdce1400fbf964fff0277e633153428d7b96a2f9137ff129f1f6ebb203a6d0a06da6f9fad8f5e4ea9f2cf13c59f9d5fb742e04ddb6de34e11259f925278b5826a0db4edfeb32663b9851eb6f8a7167766ff7f22cfb35a8b2a21d862aedf95637fba05c0e29da6ab5e5ed3d2d505f9da91ecabe791bcf2a014413ee3de6955059d614a05179718e8aeb2462984a9e541fb300e44aebce5530e031c75c2af716744ff070a0c1d7dd289dd0f1ed45d4d4e442880787f9d7c706d7542256b65100ed18e3e6aca0a6c95838cdd7d9dce172d55c0369bd3a8c8ed325e9bc85fdfc76520031e58455a069ac0cd1883fd67d3ca9ca4da3da54beca3ea5b29d8e5a538185cac535dbb93f80",
				"0xf9011180a0e55a052178f9ae1ededbf0b714944e1381b50a70efd7b1b15bc7b8987a5448098080a0734c2845f5a95b69c8cb4a20e8c430710ff6b7a4238c47d9239e035d8c9b91018080a0a0f5a391ab87dba51cc7a5aab9b32c562e1d37af589ec4ff0014536320557b0ba0fcdbd2b4f1d5f09cbe6b55398c8c3258dbcf5eb4e70174b74c3194b72d707c30a07f76d6d3a66ce5647e7f9a40fd3b01b75cc174e038344694097d0bf9cce6245680a0e6a2ee68b6aae5503b146cc64b7713c0dba5aa1977c19f50d3285fdfe7f0c193a07e4717765025d87943e462589d42eff3cc273c2b20febb588db00da5abcbbd7ba0520e1d30823fa9a9ebb162d293fe4080918d805fc221f4eab0b60ae895d324ba808080",
				"0xf871a0ab01a36fb678cac2e770162d442ae64784cdb2b170487c34e22906778552699b8080808080a02d9b2c476dab5102c3582bdcf1211992317c25b4a684a70ca86339c0afee21418080808080a09dd55262b974bb5813e268ff88622a2fc04452ecb0eea67c61118f29cdeeca8780808080",
				"0xf8669d2087e5428cd7092bc5ff81bb600cc7e78b2ee99dea10f0df90ab8920e6b846f8440180a05e0c53c87ffe7e7c4e724486ecd40fe4e6f13630219de7427b0fdcfaed042d7fa096107dc4006b4c7fecd1827cfb275ffeef31e6194cd50466f85f8eb24ccf2679",
			},
			Balance:     big.NewInt(1000000000000000000),
			CodeHash:    gethcommon.HexToHash("0x96107dc4006b4c7fecd1827cfb275ffeef31e6194cd50466f85f8eb24ccf2679"),
			Nonce:       uint64(1),
			StorageHash: gethcommon.HexToHash("0x5e0c53c87ffe7e7c4e724486ecd40fe4e6f13630219de7427b0fdcfaed042d7f"),
			StorageProof: []gethclient.StorageResult{
				{
					Key:   "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
					Value: big.NewInt(10),
					Proof: []string{
						"0x2fbbd1aa9c",
						"0x2fbbd1aa9d",
					},
				},
			},
		},
		proof,
	)
}
