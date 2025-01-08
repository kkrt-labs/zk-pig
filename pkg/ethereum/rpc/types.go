package rpc

import (
	"encoding/json"
	"errors"
	"math/big"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)

type Header struct {
	Number           *hexutil.Big         `json:"number"`
	ParentHash       gethcommon.Hash      `json:"parentHash"`
	Nonce            gethtypes.BlockNonce `json:"nonce"`
	MixHash          gethcommon.Hash      `json:"mixHash"`
	UncleHash        gethcommon.Hash      `json:"sha3Uncles"`
	LogsBloom        gethtypes.Bloom      `json:"logsBloom"`
	Root             gethcommon.Hash      `json:"stateRoot"`
	Miner            gethcommon.Address   `json:"miner"`
	Difficulty       *hexutil.Big         `json:"difficulty"`
	Extra            hexutil.Bytes        `json:"extraData"`
	GasLimit         hexutil.Uint64       `json:"gasLimit"`
	GasUsed          hexutil.Uint64       `json:"gasUsed"`
	Time             hexutil.Uint64       `json:"timestamp"`
	TxRoot           gethcommon.Hash      `json:"transactionsRoot"`
	ReceiptsRoot     gethcommon.Hash      `json:"receiptsRoot"`
	BaseFee          *hexutil.Big         `json:"baseFeePerGas,omitempty"`
	WithdrawalsRoot  *gethcommon.Hash     `json:"withdrawalsRoot,omitempty"`
	BlobGasUsed      *hexutil.Uint64      `json:"blobGasUsed,omitempty"`
	ExcessBlobGas    *hexutil.Uint64      `json:"excessBlobGas,omitempty"`
	ParentBeaconRoot *gethcommon.Hash     `json:"parentBeaconBlockRoot,omitempty"`
	RequestsRoot     *gethcommon.Hash     `json:"requestsRoot,omitempty"`
	Hash             gethcommon.Hash      `json:"hash"`
}

func (h *Header) Header() *gethtypes.Header {
	header := &gethtypes.Header{
		Number:           (*big.Int)(h.Number),
		ParentHash:       h.ParentHash,
		UncleHash:        h.UncleHash,
		Root:             h.Root,
		TxHash:           h.TxRoot,
		ReceiptHash:      h.ReceiptsRoot,
		Time:             uint64(h.Time),
		Extra:            h.Extra,
		Difficulty:       (*big.Int)(h.Difficulty),
		GasLimit:         uint64(h.GasLimit),
		GasUsed:          uint64(h.GasUsed),
		Nonce:            h.Nonce,
		MixDigest:        h.MixHash,
		Coinbase:         h.Miner,
		Bloom:            h.LogsBloom,
		BaseFee:          (*big.Int)(h.BaseFee),
		WithdrawalsHash:  h.WithdrawalsRoot,
		ParentBeaconRoot: h.ParentBeaconRoot,
		RequestsHash:     h.RequestsRoot,
	}

	if h.BlobGasUsed != nil {
		header.BlobGasUsed = (*uint64)(h.BlobGasUsed)
	}

	if h.ExcessBlobGas != nil {
		header.ExcessBlobGas = (*uint64)(h.ExcessBlobGas)
	}

	return header
}

func (h *Header) FromHeader(header *gethtypes.Header) *Header {
	h.Number = (*hexutil.Big)(header.Number)
	h.ParentHash = header.ParentHash
	h.UncleHash = header.UncleHash
	h.Root = header.Root
	h.TxRoot = header.TxHash
	h.ReceiptsRoot = header.ReceiptHash
	h.Time = hexutil.Uint64(header.Time)
	h.Extra = header.Extra
	h.Difficulty = (*hexutil.Big)(header.Difficulty)
	h.GasLimit = hexutil.Uint64(header.GasLimit)
	h.GasUsed = hexutil.Uint64(header.GasUsed)
	h.Nonce = header.Nonce
	h.MixHash = header.MixDigest
	h.Miner = header.Coinbase
	h.LogsBloom = header.Bloom
	h.BaseFee = (*hexutil.Big)(header.BaseFee)
	h.WithdrawalsRoot = header.WithdrawalsHash
	h.ParentBeaconRoot = header.ParentBeaconRoot
	h.RequestsRoot = header.RequestsHash

	if header.BlobGasUsed != nil {
		h.BlobGasUsed = (*hexutil.Uint64)(header.BlobGasUsed)
	}

	if header.ExcessBlobGas != nil {
		h.ExcessBlobGas = (*hexutil.Uint64)(header.ExcessBlobGas)
	}

	h.Hash = header.Hash()

	return h
}

type Block struct {
	Header
	blockExtraInfo
}

type blockExtraInfo struct {
	Size         hexutil.Uint64        `json:"size"`
	Transactions []*Transaction        `json:"transactions"`
	Uncles       []gethcommon.Hash     `json:"uncles"`
	Withdrawals  gethtypes.Withdrawals `json:"withdrawals,omitempty"`
}

func (b *Block) UnmarshalJSON(msg []byte) error {
	if err := json.Unmarshal(msg, &b.Header); err != nil {
		return err
	}
	return json.Unmarshal(msg, &b.blockExtraInfo)
}

func (b *Block) Block() *gethtypes.Block {
	body := gethtypes.Body{
		Withdrawals: b.Withdrawals,
	}

	for _, tx := range b.Transactions {
		body.Transactions = append(body.Transactions, tx.Tx())
	}

	return gethtypes.NewBlockWithHeader(b.Header.Header()).WithBody(body)
}

func (b *Block) FromBlock(block *gethtypes.Block, config *params.ChainConfig) *Block {
	b.Header.FromHeader(block.Header())
	b.Size = hexutil.Uint64(block.Size())
	b.Withdrawals = block.Withdrawals()

	signer := gethtypes.MakeSigner(config, block.Number(), block.Time())

	for _, tx := range block.Transactions() {
		from, _ := gethtypes.Sender(signer, tx)
		b.Transactions = append(b.Transactions, &Transaction{
			Transaction: tx,
			txExtraInfo: txExtraInfo{
				BlockNumber: b.Number,
				BlockHash:   &b.Hash,
				From:        &from,
			},
		})
	}

	for _, uncle := range block.Uncles() {
		b.Uncles = append(b.Uncles, uncle.Hash())
	}

	return b
}

type Transaction struct {
	*gethtypes.Transaction
	txExtraInfo
}

type txExtraInfo struct {
	BlockNumber *hexutil.Big        `json:"blockNumber,omitempty"`
	BlockHash   *gethcommon.Hash    `json:"blockHash,omitempty"`
	From        *gethcommon.Address `json:"from,omitempty"`
}

func (tx *Transaction) Tx() *gethtypes.Transaction {
	if tx.From != nil && tx.BlockHash != nil {
		setSender(tx.Transaction, *tx.From, *tx.BlockHash)
	}
	return tx.Transaction
}

func (tx *Transaction) UnmarshalJSON(msg []byte) error {
	if err := json.Unmarshal(msg, &tx.Transaction); err != nil {
		return nil
	}

	return json.Unmarshal(msg, &tx.txExtraInfo)
}

func NewTransactionFromGeth(tx *gethtypes.Transaction) *Transaction {
	return &Transaction{
		Transaction: tx,
	}
}

// func (tx *RPCTransaction) MarshalJSON() ([]byte, error) {
// 	return json.Marshal(tx)
// }

// // newRPCTransactionFromBlockIndex returns a transaction that will serialize to the RPC representation.
// func newRPCTransactionFromBlockIndex(b *gethtypes.Block, index uint64, config *params.ChainConfig) *RPCTransaction {
// 	txs := b.Transactions()
// 	if index >= uint64(len(txs)) {
// 		return nil
// 	}
// 	return newRPCTransaction(txs[index], b.Hash(), b.NumberU64(), b.Time(), index, b.BaseFee(), config)
// }

// // newRPCTransaction returns a transaction that will serialize to the RPC
// // representation, with the given location metadata set (if available).
// func newRPCTransaction(tx *gethtypes.Transaction, blockHash gethcommon.Hash, blockNumber uint64, blockTime uint64, index uint64, baseFee *big.Int, config *params.ChainConfig) *RPCTransaction {
// 	signer := gethtypes.MakeSigner(config, new(big.Int).SetUint64(blockNumber), blockTime)
// 	from, _ := gethtypes.Sender(signer, tx)
// 	v, r, s := tx.RawSignatureValues()
// 	result := &RPCTransaction{
// 		Type:     hexutil.Uint64(tx.Type()),
// 		From:     from,
// 		Gas:      hexutil.Uint64(tx.Gas()),
// 		GasPrice: (*hexutil.Big)(tx.GasPrice()),
// 		Hash:     tx.Hash(),
// 		Input:    hexutil.Bytes(tx.Data()),
// 		Nonce:    hexutil.Uint64(tx.Nonce()),
// 		To:       tx.To(),
// 		Value:    (*hexutil.Big)(tx.Value()),
// 		V:        (*hexutil.Big)(v),
// 		R:        (*hexutil.Big)(r),
// 		S:        (*hexutil.Big)(s),
// 	}
// 	if blockHash != (gethcommon.Hash{}) {
// 		result.BlockHash = &blockHash
// 		result.BlockNumber = (*hexutil.Big)(new(big.Int).SetUint64(blockNumber))
// 		result.TransactionIndex = (*hexutil.Uint64)(&index)
// 	}

// 	switch tx.Type() {
// 	case gethtypes.LegacyTxType:
// 		// if a legacy transaction has an EIP-155 chain id, include it explicitly
// 		if id := tx.ChainId(); id.Sign() != 0 {
// 			result.ChainID = (*hexutil.Big)(id)
// 		}

// 	case gethtypes.AccessListTxType:
// 		al := tx.AccessList()
// 		yparity := hexutil.Uint64(v.Sign())
// 		result.Accesses = &al
// 		result.ChainID = (*hexutil.Big)(tx.ChainId())
// 		result.YParity = &yparity

// 	case gethtypes.DynamicFeeTxType:
// 		al := tx.AccessList()
// 		yparity := hexutil.Uint64(v.Sign())
// 		result.Accesses = &al
// 		result.ChainID = (*hexutil.Big)(tx.ChainId())
// 		result.YParity = &yparity
// 		result.GasFeeCap = (*hexutil.Big)(tx.GasFeeCap())
// 		result.GasTipCap = (*hexutil.Big)(tx.GasTipCap())
// 		// if the transaction has been mined, compute the effective gas price
// 		if baseFee != nil && blockHash != (gethcommon.Hash{}) {
// 			// price = min(gasTipCap + baseFee, gasFeeCap)
// 			result.GasPrice = (*hexutil.Big)(effectiveGasPrice(tx, baseFee))
// 		} else {
// 			result.GasPrice = (*hexutil.Big)(tx.GasFeeCap())
// 		}

// 	case gethtypes.BlobTxType:
// 		al := tx.AccessList()
// 		yparity := hexutil.Uint64(v.Sign())
// 		result.Accesses = &al
// 		result.ChainID = (*hexutil.Big)(tx.ChainId())
// 		result.YParity = &yparity
// 		result.GasFeeCap = (*hexutil.Big)(tx.GasFeeCap())
// 		result.GasTipCap = (*hexutil.Big)(tx.GasTipCap())
// 		// if the transaction has been mined, compute the effective gas price
// 		if baseFee != nil && blockHash != (gethcommon.Hash{}) {
// 			result.GasPrice = (*hexutil.Big)(effectiveGasPrice(tx, baseFee))
// 		} else {
// 			result.GasPrice = (*hexutil.Big)(tx.GasFeeCap())
// 		}
// 		result.MaxFeePerBlobGas = (*hexutil.Big)(tx.BlobGasFeeCap())
// 		result.BlobVersionedHashes = tx.BlobHashes()
// 	}
// 	return result
// }

// // effectiveGasPrice computes the transaction gas fee, based on the given basefee value.
// //
// //	price = min(gasTipCap + baseFee, gasFeeCap)
// func effectiveGasPrice(tx *gethtypes.Transaction, baseFee *big.Int) *big.Int {
// 	fee := tx.GasTipCap()
// 	fee = fee.Add(fee, baseFee)
// 	if tx.GasFeeCapIntCmp(fee) < 0 {
// 		return tx.GasFeeCap()
// 	}
// 	return fee
// }

// senderFromServer is a types.Signer that remembers the sender address returned by the RPC
// server. It is stored in the transaction's sender address cache to avoid an additional
// request in TransactionSender.
type senderFromServer struct {
	addr      gethcommon.Address
	blockhash gethcommon.Hash
}

var errNotCached = errors.New("sender not cached")

func setSender(tx *gethtypes.Transaction, addr gethcommon.Address, block gethcommon.Hash) {
	// Use types.Sender for side-effect to store our signer into the cache.
	_, _ = gethtypes.Sender(&senderFromServer{addr, block}, tx)
}

func (s *senderFromServer) Equal(other gethtypes.Signer) bool {
	os, ok := other.(*senderFromServer)
	return ok && os.blockhash == s.blockhash
}

func (s *senderFromServer) Sender(_ *gethtypes.Transaction) (gethcommon.Address, error) {
	if s.addr == (gethcommon.Address{}) {
		return gethcommon.Address{}, errNotCached
	}
	return s.addr, nil
}

func (s *senderFromServer) ChainID() *big.Int {
	panic("can't sign with senderFromServer")
}
func (s *senderFromServer) Hash(_ *gethtypes.Transaction) gethcommon.Hash {
	panic("can't sign with senderFromServer")
}
func (s *senderFromServer) SignatureValues(_ *gethtypes.Transaction, _ []byte) (_, _, _ *big.Int, _ error) {
	panic("can't sign with senderFromServer")
}
