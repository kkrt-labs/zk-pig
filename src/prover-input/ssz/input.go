package ssz

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"
	proverinput "github.com/kkrt-labs/zk-pig/src/prover-input"
	// ssz "github.com/kkrt-labs/zk-pig/src/prover-input/ssz"
)

type ProverInput struct {
	Version     []byte       `json:"version" ssz-max:"32"`        // Prover Input version as byte array
	Blocks      []*Block     `json:"blocks" ssz-max:"1073741824"` // Block to execute
	Witness     *Witness     `json:"witness"`                     // Ancestors of the block that are accessed during the block execution
	ChainConfig *ChainConfig `json:"chainConfig"`                 // Chain configuration, now optional
}

type Block struct {
	Header       *Header       `json:"header"`                            // Use a custom Header struct
	Transactions []byte        `json:"transactions" ssz-max:"1073741824"` // Use a custom Transaction struct
	Uncles       []*Header     `json:"uncles" ssz-max:"1073741824"`       // Ensure Uncles is a slice of pointers
	Withdrawals  []*Withdrawal `json:"withdrawals" ssz-max:"1073741824"`  // Ensure Withdrawals is a slice of pointers
}

// Define a custom Header struct with SSZ-compatible fields
type Header struct {
	ParentHash  []byte  `ssz-size:"32"`
	UncleHash   []byte  `ssz-size:"32"`
	Coinbase    []byte  `ssz-size:"20"`
	Root        []byte  `ssz-size:"32"`
	TxHash      []byte  `ssz-size:"32"`
	ReceiptHash []byte  `ssz-size:"32"`
	Bloom       []byte  `ssz-size:"256"`
	Difficulty  []byte  `ssz-max:"32"` // Use []byte for numeric values
	Number      []byte  `ssz-max:"32"`
	GasLimit    uint64  `ssz-size:"8"`
	GasUsed     uint64  `ssz-size:"8"`
	Time        uint64  `ssz-size:"8"`
	Extra       []byte  `ssz-max:"32"`
	MixDigest   []byte  `ssz-max:"32"`
	Nonce       [8]byte `ssz-size:"8"`
}

type Transaction_AccessListTransaction struct {
	AccessListTransaction *AccessListTransaction
}

type Transaction_DynamicFeeTransaction struct {
	DynamicFeeTransaction *DynamicFeeTransaction
}

type LegacyTransaction struct {
	Nonce    uint64 `ssz-size:"8"`
	GasPrice []byte `ssz-size:"32"`
	Gas      uint64 `ssz-size:"8"`
	To       []byte `ssz-size:"20"`
	Value    []byte `ssz-size:"32"`
	Data     []byte `ssz-size:"1024"`
	V        []byte `ssz-size:"32"`
	R        []byte `ssz-size:"32"`
	S        []byte `ssz-size:"32"`
}

type AccessListTransaction struct {
	ChainId    []byte        `ssz-size:"8"`
	Nonce      uint64        `ssz-size:"8"`
	GasPrice   []byte        `ssz-size:"32"`
	Gas        uint64        `ssz-size:"8"`
	To         []byte        `ssz-size:"20"`
	Value      []byte        `ssz-size:"32"`
	Data       []byte        `ssz-size:"1024"`
	AccessList []AccessTuple `ssz-max:"4096"`
	V          []byte        `ssz-size:"32"`
	R          []byte        `ssz-size:"32"`
	S          []byte        `ssz-size:"32"`
}

type AccessTuple struct {
	Address     []byte   `ssz-max:"20"`
	StorageKeys [][]byte `ssz-size:"1024,32"`
}

type DynamicFeeTransaction struct {
	ChainId    []byte        `ssz-size:"8"`
	Nonce      uint64        `ssz-size:"8"`
	GasTipCap  []byte        `ssz-size:"32"`
	GasFeeCap  []byte        `ssz-size:"32"`
	Gas        uint64        `ssz-size:"8"`
	To         []byte        `ssz-size:"20"`
	Value      []byte        `ssz-size:"32"`
	Data       []byte        `ssz-size:"1024"`
	AccessList []AccessTuple `ssz-max:"4096"`
	V          []byte        `ssz-size:"32"`
	R          []byte        `ssz-size:"32"`
	S          []byte        `ssz-size:"32"`
}

// type BlobTransaction struct {
// 	ChainId       uint64         `ssz-size:"8"`
// 	Nonce         uint64         `ssz-size:"8"`
// 	GasTipCap     []byte         `ssz-size:"32"`
// 	GasFeeCap     []byte         `ssz-size:"32"`
// 	Gas           uint64         `ssz-size:"8"`
// 	To            []byte         `ssz-size:"20"`
// 	Value         []byte         `ssz-size:"32"`
// 	Data          []byte         `ssz-size:"1024"`
// 	AccessList    []AccessTuple  `ssz-max:"4096"`
// 	BlobFeeCap    []byte         `ssz-size:"32"`
// 	BlobHashes    [][]byte       `ssz-size:"1024"`
// 	BlobTxSidecar *BlobTxSidecar `ssz-max:"3221225472"` // TODO: Fix this
// 	V             []byte         `ssz-size:"32"`
// 	R             []byte         `ssz-size:"32"`
// 	S             []byte         `ssz-size:"32"`
// }

type BlobTxSidecar struct {
	Blobs       [][]byte `ssz-max:"1024,1048576" ssz-size:"?,?"`
	Commitments [][]byte `ssz-max:"1024,1048576" ssz-size:"?,?"`
	Proofs      [][]byte `ssz-max:"1024,1048576" ssz-size:"?,?"`
}

type Witness struct {
	State     [][]byte  `json:"state" ssz-max:"1073741824,1073741824" ssz-size:"?,?"`
	Ancestors []*Header `json:"ancestors" ssz-max:"1073741824"`
	Codes     [][]byte  `json:"codes" ssz-max:"1073741824,1073741824" ssz-size:"?,?"`
}

type ChainConfig struct {
	ChainId                 uint64        `ssz-size:"8"`
	HomesteadBlock          []byte        `ssz-max:"32"`
	DaoForkBlock            []byte        `ssz-max:"32"`
	DaoForkSupport          bool          `ssz-size:"1"`
	Eip150Block             []byte        `ssz-max:"32"`
	Eip155Block             []byte        `ssz-max:"32"`
	Eip158Block             []byte        `ssz-max:"32"`
	ByzantiumBlock          []byte        `ssz-max:"32"`
	ConstantinopleBlock     []byte        `ssz-max:"32"`
	PetersburgBlock         []byte        `ssz-max:"32"`
	IstanbulBlock           []byte        `ssz-max:"32"`
	MuirGlacierBlock        []byte        `ssz-max:"32"`
	BerlinBlock             []byte        `ssz-max:"32"`
	LondonBlock             []byte        `ssz-max:"32"`
	ArrowGlacierBlock       []byte        `ssz-max:"32"`
	GrayGlacierBlock        []byte        `ssz-max:"32"`
	MergeNetsplitBlock      []byte        `ssz-max:"32"`
	ShanghaiTime            uint64        `ssz-size:"8"`
	CancunTime              uint64        `ssz-size:"8"`
	PragueTime              uint64        `ssz-size:"8"`
	VerkleTime              uint64        `ssz-size:"8"`
	TerminalTotalDifficulty []byte        `ssz-max:"32"`
	DepositContractAddress  []byte        `ssz-max:"20"`
	Ethash                  []byte        `ssz-max:"32"`
	Clique                  *CliqueConfig `ssz-max:"1024"`
}

type CliqueConfig struct {
	Period uint64 `ssz-size:"8"`
	Epoch  uint64 `ssz-size:"8"`
}

type Withdrawal struct {
	Index          uint64
	ValidatorIndex uint64
	Address        []byte `ssz-max:"20"`
	Amount         uint64
}

// Helper function to convert uint64 to byte slice
func uint64ToBytes(value uint64) []byte {
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(bytes, value)
	return bytes
}

func ProverInputFromSSZ(p *ProverInput) (*proverinput.ProverInput, error) {
	return &proverinput.ProverInput{
		Version:     string(p.Version),
		Blocks:      BlocksFromSSZ(p.Blocks),
		Witness:     WitnessFromSSZ(p.Witness),
		ChainConfig: ChainConfigFromSSZ(p.ChainConfig),
	}, nil
}

func ChainConfigFromSSZ(c *ChainConfig) *params.ChainConfig {
	return &params.ChainConfig{
		ChainID:                 big.NewInt(0).SetBytes(uint64ToBytes(c.ChainId)),
		HomesteadBlock:          big.NewInt(0).SetBytes(c.HomesteadBlock),
		DAOForkBlock:            big.NewInt(0).SetBytes(c.DaoForkBlock),
		DAOForkSupport:          c.DaoForkSupport,
		EIP150Block:             big.NewInt(0).SetBytes(c.Eip150Block),
		EIP155Block:             big.NewInt(0).SetBytes(c.Eip155Block),
		EIP158Block:             big.NewInt(0).SetBytes(c.Eip158Block),
		ByzantiumBlock:          big.NewInt(0).SetBytes(c.ByzantiumBlock),
		ConstantinopleBlock:     big.NewInt(0).SetBytes(c.ConstantinopleBlock),
		PetersburgBlock:         big.NewInt(0).SetBytes(c.PetersburgBlock),
		IstanbulBlock:           big.NewInt(0).SetBytes(c.IstanbulBlock),
		MuirGlacierBlock:        big.NewInt(0).SetBytes(c.MuirGlacierBlock),
		BerlinBlock:             big.NewInt(0).SetBytes(c.BerlinBlock),
		LondonBlock:             big.NewInt(0).SetBytes(c.LondonBlock),
		ArrowGlacierBlock:       big.NewInt(0).SetBytes(c.ArrowGlacierBlock),
		GrayGlacierBlock:        big.NewInt(0).SetBytes(c.GrayGlacierBlock),
		MergeNetsplitBlock:      big.NewInt(0).SetBytes(c.MergeNetsplitBlock),
		ShanghaiTime:            &c.ShanghaiTime,
		CancunTime:              &c.CancunTime,
		PragueTime:              &c.PragueTime,
		VerkleTime:              &c.VerkleTime,
		TerminalTotalDifficulty: big.NewInt(0).SetBytes(c.TerminalTotalDifficulty),
		DepositContractAddress:  common.BytesToAddress(c.DepositContractAddress),
		Ethash:                  &params.EthashConfig{},
		Clique: &params.CliqueConfig{
			Period: c.Clique.Period,
			Epoch:  c.Clique.Epoch,
		},
	}
}

func BlocksFromSSZ(b []*Block) []*proverinput.Block {
	var blocks []*proverinput.Block
	for _, b := range b {
		blocks = append(blocks, BlockFromSSZ(b))
	}
	return blocks
}

func BlockFromSSZ(b *Block) *proverinput.Block {
	return &proverinput.Block{
		Header:       b.Header.HeaderFromSSZ(),
		Transactions: TransactionsFromSSZ(b.Transactions),
		Uncles:       UnclesFromSSZ(b.Uncles),
		Withdrawals:  WithdrawalsFromSSZ(b.Withdrawals),
	}
}

func UnclesFromSSZ(u []*Header) []*gethtypes.Header {
	var uncles []*gethtypes.Header
	for _, u := range u {
		uncles = append(uncles, u.HeaderFromSSZ())
	}
	return uncles
}

func WithdrawalsFromSSZ(w []*Withdrawal) []*gethtypes.Withdrawal {
	var withdrawals []*gethtypes.Withdrawal
	for _, w := range w {
		withdrawals = append(withdrawals, &gethtypes.Withdrawal{
			Index:     w.Index,
			Validator: w.ValidatorIndex,
			Address:   common.BytesToAddress(w.Address),
			Amount:    w.Amount,
		})
	}
	return withdrawals
}

func WitnessFromSSZ(w *Witness) *proverinput.Witness {
	return &proverinput.Witness{
		State:     convertByteSlicesToHexutil(w.State),
		Ancestors: HeadersFromSSZ(w.Ancestors),
		Codes:     convertByteSlicesToHexutil(w.Codes),
	}
}

// Helper function to convert [][]byte to []hexutil.Bytes
func convertByteSlicesToHexutil(byteSlices [][]byte) []hexutil.Bytes {
	hexBytes := make([]hexutil.Bytes, len(byteSlices))
	for i, b := range byteSlices {
		hexBytes[i] = hexutil.Bytes(b) // Convert []byte to hexutil.Bytes
	}
	return hexBytes
}

func HeadersFromSSZ(hs []*Header) []*gethtypes.Header {
	var headers []*gethtypes.Header
	for _, h := range hs {
		headers = append(headers, h.HeaderFromSSZ())
	}
	return headers
}

// Convert custom Header back to gethtypes.Header
func (h *Header) HeaderFromSSZ() *gethtypes.Header {
	return &gethtypes.Header{
		ParentHash:  common.BytesToHash(h.ParentHash),
		UncleHash:   common.BytesToHash(h.UncleHash),
		Coinbase:    common.BytesToAddress(h.Coinbase),
		Root:        common.BytesToHash(h.Root),
		TxHash:      common.BytesToHash(h.TxHash),
		ReceiptHash: common.BytesToHash(h.ReceiptHash),
		Bloom:       gethtypes.BytesToBloom(h.Bloom),
		Difficulty:  new(big.Int).SetBytes(h.Difficulty),
		Number:      new(big.Int).SetBytes(h.Number),
		GasLimit:    h.GasLimit,
		GasUsed:     h.GasUsed,
		Time:        h.Time,
		Extra:       h.Extra,
		MixDigest:   common.BytesToHash(h.MixDigest),
		Nonce:       gethtypes.BlockNonce(h.Nonce),
	}
}

func TransactionsFromSSZ(t []byte) []*gethtypes.Transaction {
	var txs []*gethtypes.Transaction
	txs = append(txs, TransactionFromSSZ(t))
	return txs
}

func TransactionFromSSZ(t []byte) *gethtypes.Transaction {
	var tx gethtypes.Transaction
	err := rlp.DecodeBytes(t, &tx)
	if err != nil {
		// Handle the error appropriately, e.g., log it or return nil
		return nil
	}
	return &tx
}

func DynamicFeeTransactionFromSSZ(tx *Transaction_DynamicFeeTransaction) *gethtypes.Transaction {
	var to *common.Address
	if len(tx.DynamicFeeTransaction.To) > 0 {
		address := common.BytesToAddress(tx.DynamicFeeTransaction.To)
		to = &address
	}
	return gethtypes.NewTx(&gethtypes.DynamicFeeTx{
		ChainID:    big.NewInt(0).SetBytes(tx.DynamicFeeTransaction.ChainId),
		Nonce:      tx.DynamicFeeTransaction.Nonce,
		GasTipCap:  new(big.Int).SetBytes(tx.DynamicFeeTransaction.GasTipCap),
		GasFeeCap:  new(big.Int).SetBytes(tx.DynamicFeeTransaction.GasFeeCap),
		Gas:        tx.DynamicFeeTransaction.Gas,
		To:         to,
		Value:      new(big.Int).SetBytes(tx.DynamicFeeTransaction.Value),
		Data:       tx.DynamicFeeTransaction.Data,
		AccessList: gethtypes.AccessList{}, // Convert AccessList if needed
		V:          new(big.Int).SetBytes(tx.DynamicFeeTransaction.V),
		R:          new(big.Int).SetBytes(tx.DynamicFeeTransaction.R),
		S:          new(big.Int).SetBytes(tx.DynamicFeeTransaction.S),
	})
}

// func BlobTransactionFromSSZ(tx *BlobTransaction) *gethtypes.Transaction {
// 	var to *common.Address
// 	if len(tx.To) > 0 {
// 		address := common.BytesToAddress(tx.To)
// 		to = &address
// 	}
// 	// Handle BlobTransaction conversion
// 	return nil // Implement conversion logic
// }

// To SSZ functions

func ToSSZ(p *proverinput.ProverInput) ([]byte, error) {
	var proverInput ProverInput

	proverInput.Version = []byte(p.Version)
	proverInput.Blocks = BlocksToSSZ(p.Blocks)
	proverInput.Witness = WitnessToSSZ(p.Witness)
	proverInput.ChainConfig = ChainConfigToSSZ(p.ChainConfig)

	return proverInput.MarshalSSZ()
}

func BlocksToSSZ(b []*proverinput.Block) []*Block {
	var blocks []*Block
	for _, b := range b {
		blocks = append(blocks, BlockToSSZ(b))
	}
	return blocks
}

func BlockToSSZ(b *proverinput.Block) *Block {
	var block Block

	block.Header = HeaderToSSZ(b.Header)
	block.Transactions = TransactionsToSSZ(b.Transactions)
	block.Uncles = HeadersToSSZ(b.Uncles) // TODO: Fix this; issue: it generates blank array for uncles instead of the actual uncles
	block.Withdrawals = WithdrawalsToSSZ(b.Withdrawals)

	return &block
}

func ChainConfigToSSZ(c *params.ChainConfig) *ChainConfig {
	chainConfig := &ChainConfig{
		ChainId:                 c.ChainID.Uint64(),
		HomesteadBlock:          c.HomesteadBlock.Bytes(),
		DaoForkBlock:            c.DAOForkBlock.Bytes(),
		DaoForkSupport:          c.DAOForkSupport,
		Eip150Block:             c.EIP150Block.Bytes(),
		Eip155Block:             c.EIP155Block.Bytes(),
		Eip158Block:             c.EIP158Block.Bytes(),
		ByzantiumBlock:          c.ByzantiumBlock.Bytes(),
		ConstantinopleBlock:     c.ConstantinopleBlock.Bytes(),
		PetersburgBlock:         c.PetersburgBlock.Bytes(),
		IstanbulBlock:           c.IstanbulBlock.Bytes(),
		MuirGlacierBlock:        c.MuirGlacierBlock.Bytes(),
		BerlinBlock:             c.BerlinBlock.Bytes(),
		LondonBlock:             c.LondonBlock.Bytes(),
		ArrowGlacierBlock:       c.ArrowGlacierBlock.Bytes(),
		GrayGlacierBlock:        c.GrayGlacierBlock.Bytes(),
		ShanghaiTime:            uint64(*c.ShanghaiTime),
		CancunTime:              uint64(*c.CancunTime),
		TerminalTotalDifficulty: c.TerminalTotalDifficulty.Bytes(),
		DepositContractAddress:  c.DepositContractAddress.Bytes(),
	}

	if c.VerkleTime != nil {
		chainConfig.VerkleTime = uint64(*c.VerkleTime)
	}

	if c.PragueTime != nil {
		chainConfig.PragueTime = uint64(*c.PragueTime)
	}

	if c.MergeNetsplitBlock != nil {
		chainConfig.MergeNetsplitBlock = c.MergeNetsplitBlock.Bytes()
	}

	if c.Ethash != nil {
		ethashBytes, err := json.Marshal(c.Ethash)
		if err == nil {
			chainConfig.Ethash = ethashBytes
		}
	}

	if c.Clique != nil {
		chainConfig.Clique = &CliqueConfig{
			Period: c.Clique.Period,
			Epoch:  c.Clique.Epoch,
		}
	}

	return chainConfig
}

func WitnessToSSZ(w *proverinput.Witness) *Witness {
	return &Witness{
		State:     convertHexutilBytesSlice(w.State),
		Ancestors: HeadersToSSZ(w.Ancestors),
		Codes:     convertHexutilBytesSlice(w.Codes),
	}
}

// Helper function to convert []hexutil.Bytes to [][]byte
func convertHexutilBytesSlice(hexBytes []hexutil.Bytes) [][]byte {
	byteSlices := make([][]byte, len(hexBytes))
	for i, hb := range hexBytes {
		byteSlices[i] = hb[:] // Convert hexutil.Bytes to []byte
	}
	return byteSlices
}

func HeaderToSSZ(h *gethtypes.Header) *Header {
	return &Header{
		ParentHash:  h.ParentHash.Bytes(),
		UncleHash:   h.UncleHash[:],
		Coinbase:    h.Coinbase[:],
		Root:        h.Root[:],
		TxHash:      h.TxHash[:],
		ReceiptHash: h.ReceiptHash[:],
		Bloom:       h.Bloom[:],
		Difficulty:  h.Difficulty.Bytes(),
		Number:      h.Number.Bytes(),
		GasLimit:    h.GasLimit,
		GasUsed:     h.GasUsed,
		Time:        h.Time,
	}
}

// TODO: Fix this
// issue: it generates blank array for transactions instead of the actual transactions
func TransactionsToSSZ(txs []*gethtypes.Transaction) []byte {
	buf := new(bytes.Buffer)
	for _, tx := range txs {
		tx.EncodeRLP(buf)
	}
	return buf.Bytes()
}

func LegacyTransactionToSSZ(tx *gethtypes.Transaction) *LegacyTransaction {
	v, r, s := tx.RawSignatureValues()

	return &LegacyTransaction{
		Nonce:    tx.Nonce(),
		GasPrice: tx.GasPrice().Bytes(),
		Gas:      tx.Gas(),
		To:       tx.To().Bytes(),
		Value:    tx.Value().Bytes(),
		Data:     tx.Data(),
		V:        v.Bytes(),
		R:        r.Bytes(),
		S:        s.Bytes(),
	}
}

func AccessListTransactionToSSZ(tx *gethtypes.Transaction) *AccessListTransaction {
	v, r, s := tx.RawSignatureValues()

	return &AccessListTransaction{
		ChainId:    tx.ChainId().Bytes(),
		Nonce:      tx.Nonce(),
		GasPrice:   tx.GasPrice().Bytes(),
		Gas:        tx.Gas(),
		To:         tx.To().Bytes(),
		Value:      tx.Value().Bytes(),
		Data:       tx.Data(),
		AccessList: AccessListToSSZ(tx.AccessList()),
		V:          v.Bytes(),
		R:          r.Bytes(),
		S:          s.Bytes(),
	}
}

func DynamicFeeTransactionToSSZ(tx *gethtypes.Transaction) *DynamicFeeTransaction {
	v, r, s := tx.RawSignatureValues()

	return &DynamicFeeTransaction{
		ChainId:    tx.ChainId().Bytes(),
		Nonce:      tx.Nonce(),
		GasTipCap:  tx.GasTipCap().Bytes(),
		GasFeeCap:  tx.GasFeeCap().Bytes(),
		Gas:        tx.Gas(),
		To:         tx.To().Bytes(),
		Value:      tx.Value().Bytes(),
		Data:       tx.Data(),
		AccessList: AccessListToSSZ(tx.AccessList()),
		V:          v.Bytes(),
		R:          r.Bytes(),
		S:          s.Bytes(),
	}
}

func AccessListToSSZ(al gethtypes.AccessList) []AccessTuple {
	var accessList []AccessTuple
	for _, at := range al {
		accessList = append(accessList, *AccessTupleToSSZ(&at))
	}
	return accessList
}

func AccessTupleToSSZ(at *gethtypes.AccessTuple) *AccessTuple {
	storageKeys := make([][]byte, len(at.StorageKeys))
	for i, key := range at.StorageKeys {
		storageKeys[i] = key.Bytes() // Convert common.Hash to []byte
	}

	return &AccessTuple{
		Address:     at.Address.Bytes(),
		StorageKeys: storageKeys,
	}
}

func HeadersToSSZ(hs []*gethtypes.Header) []*Header {
	var headers []*Header
	for _, h := range hs {
		headers = append(headers, HeaderToSSZ(h))
	}
	return headers
}

func WithdrawalsToSSZ(ws []*gethtypes.Withdrawal) []*Withdrawal {
	var withdrawals []*Withdrawal
	for _, w := range ws {
		withdrawals = append(withdrawals, &Withdrawal{
			Index:          w.Index,
			ValidatorIndex: w.Validator,
			Address:        w.Address[:],
			Amount:         w.Amount,
		})
	}
	return withdrawals
}
