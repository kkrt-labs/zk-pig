package proto

import (
	"math/big"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
	ethrpc "github.com/kkrt-labs/go-utils/ethereum/rpc"
	blockinputs "github.com/kkrt-labs/zk-pig/src/blocks/inputs"
)

// ToProto converts Go ProverInputs to protobuf format
func ToProto(pi *blockinputs.ProverInputs) *ProverInputs {
	return &ProverInputs{
		Block:       blockToProto(pi.Block),
		Ancestors:   ancesterHeadersToProto(pi.Ancestors),
		ChainConfig: chainConfigToProto(pi.ChainConfig),
		Codes:       codesToProto(pi.Codes),
		PreState:    bytesFromHexutil(pi.PreState),
		AccessList:  accessListToProto(pi.AccessList),
	}
}

// FromProto converts protobuf format to Go ProverInputs
func FromProto(p *ProverInputs) *blockinputs.ProverInputs {
	return &blockinputs.ProverInputs{
		Block:       blockFromProto(p.GetBlock()),
		Ancestors:   ancesterHeadersFromProto(p.GetAncestors()),
		ChainConfig: chainConfigFromProto(p.GetChainConfig()),
		Codes:       codesFromProto(p.GetCodes()),
		PreState:    bytesToHexutil(p.GetPreState()),
		AccessList:  accessListFromProto(p.GetAccessList()),
	}
}

func blockToProto(b *ethrpc.Block) *Block {
	withdrawals := make([]*Withdrawal, len(b.Withdrawals))
	for i, w := range b.Withdrawals {
		withdrawals[i] = &Withdrawal{
			Index:          w.Index,
			ValidatorIndex: w.Validator,
			Address:        w.Address.Bytes(),
			Amount:         w.Amount,
		}
	}

	// Convert *hexutil.Uint64 to *uint64
	var excessBlobGas, blobGasUsed *uint64

	if b.ExcessBlobGas != nil {
		value := uint64(*b.ExcessBlobGas)
		excessBlobGas = &value
	}
	if b.BlobGasUsed != nil {
		value := uint64(*b.BlobGasUsed)
		blobGasUsed = &value
	}

	return &Block{
		Number:           b.Number.ToInt().Uint64(),
		ParentHash:       b.ParentHash[:],
		Nonce:            b.Nonce.Uint64(),
		MixHash:          b.MixHash[:],
		UncleHash:        b.UncleHash[:],
		LogsBloom:        b.LogsBloom[:],
		Root:             b.Root[:],
		Miner:            b.Miner[:],
		Difficulty:       b.Difficulty.ToInt().Uint64(),
		Extra:            b.Extra,
		GasLimit:         uint64(b.GasLimit),
		GasUsed:          uint64(b.GasUsed),
		Time:             uint64(b.Time),
		TxRoot:           b.TxRoot[:],
		ReceiptsRoot:     b.ReceiptsRoot[:],
		BaseFee:          b.BaseFee.ToInt().Bytes(),
		WithdrawalsRoot:  b.WithdrawalsRoot[:],
		Hash:             b.Hash[:],
		ParentBeaconRoot: bytesOrNil(b.ParentBeaconRoot),
		RequestsRoot:     bytesOrNil(b.RequestsRoot),
		Transactions:     transactionsToProto(b.Transactions),
		Uncles:           unclesToProto(b.Uncles),
		Withdrawals:      withdrawals,
		ReceiptHash:      b.ReceiptsRoot[:],
		ExcessBlobGas:    excessBlobGas,
		BlobGasUsed:      blobGasUsed,
	}
}

func chainConfigToProto(c *params.ChainConfig) *ChainConfig {
	daoForkSupport := c.DAOForkSupport

	return &ChainConfig{
		ChainId:                 c.ChainID.Uint64(),
		HomesteadBlock:          toBytesIfNotNil(c.HomesteadBlock),
		DaoForkBlock:            toBytesIfNotNil(c.DAOForkBlock),
		DaoForkSupport:          &daoForkSupport,
		Eip150Block:             toBytesIfNotNil(c.EIP150Block),
		Eip155Block:             toBytesIfNotNil(c.EIP155Block),
		Eip158Block:             toBytesIfNotNil(c.EIP158Block),
		ByzantiumBlock:          toBytesIfNotNil(c.ByzantiumBlock),
		ConstantinopleBlock:     toBytesIfNotNil(c.ConstantinopleBlock),
		PetersburgBlock:         toBytesIfNotNil(c.PetersburgBlock),
		IstanbulBlock:           toBytesIfNotNil(c.IstanbulBlock),
		MuirGlacierBlock:        toBytesIfNotNil(c.MuirGlacierBlock),
		BerlinBlock:             toBytesIfNotNil(c.BerlinBlock),
		LondonBlock:             toBytesIfNotNil(c.LondonBlock),
		ArrowGlacierBlock:       toBytesIfNotNil(c.ArrowGlacierBlock),
		GrayGlacierBlock:        toBytesIfNotNil(c.GrayGlacierBlock),
		MergeNetsplitBlock:      toBytesIfNotNil(c.MergeNetsplitBlock),
		ShanghaiTime:            c.ShanghaiTime,
		CancunTime:              c.CancunTime,
		PragueTime:              c.PragueTime,
		VerkleTime:              c.VerkleTime,
		TerminalTotalDifficulty: toBytesIfNotNil(c.TerminalTotalDifficulty),
		DepositContractAddress:  c.DepositContractAddress.Bytes(),
		Ethash: func() []byte {
			if c.Ethash != nil {
				return []byte("Ethash")
			}
			return nil
		}(),
	}
}

func transactionsToProto(txs []*ethrpc.Transaction) []*Transaction {
	if len(txs) == 0 {
		return nil
	}
	result := make([]*Transaction, len(txs))
	for i, tx := range txs {
		result[i] = transactionToProto(tx)
	}
	return result
}

func LegacyTransactionToProto(tx *ethrpc.Transaction) *LegacyTransaction {
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

func AccessListTransactionToProto(tx *ethrpc.Transaction) *AccessListTransaction {
	v, r, s := tx.RawSignatureValues()
	return &AccessListTransaction{
		ChainId:    tx.ChainId().Uint64(),
		Nonce:      tx.Nonce(),
		GasPrice:   tx.GasPrice().Bytes(),
		Gas:        tx.Gas(),
		To:         tx.To().Bytes(),
		Value:      tx.Value().Bytes(),
		Data:       tx.Data(),
		AccessList: txAccessListToProto(tx.AccessList()),
		V:          v.Bytes(),
		R:          r.Bytes(),
		S:          s.Bytes(),
	}
}

func BlobTransactionToProto(tx *ethrpc.Transaction) *BlobTransaction {
	blobHashes := make([][]byte, len(tx.BlobHashes()))
	for i, hash := range tx.BlobHashes() {
		blobHashes[i] = hash.Bytes()
	}
	v, r, s := tx.RawSignatureValues()
	return &BlobTransaction{
		ChainId:    tx.ChainId().Uint64(),
		Nonce:      tx.Nonce(),
		GasTipCap:  tx.GasTipCap().Bytes(),
		GasFeeCap:  tx.GasFeeCap().Bytes(),
		Gas:        tx.Gas(),
		To:         tx.To().Bytes(),
		Value:      tx.Value().Bytes(),
		Data:       tx.Data(),
		AccessList: txAccessListToProto(tx.AccessList()),
		BlobFeeCap: tx.BlobGasFeeCap().Bytes(),
		BlobHashes: blobHashes,
		V:          v.Bytes(),
		R:          r.Bytes(),
		S:          s.Bytes(),
	}
}

func DynamicFeeTransactionToProto(tx *ethrpc.Transaction) *DynamicFeeTransaction {
	v, r, s := tx.RawSignatureValues()
	dynamicFeeTx := &DynamicFeeTransaction{
		ChainId:    tx.ChainId().Uint64(),
		Nonce:      tx.Nonce(),
		GasTipCap:  tx.GasTipCap().Bytes(),
		GasFeeCap:  tx.GasFeeCap().Bytes(),
		Gas:        tx.Gas(),
		Value:      tx.Value().Bytes(),
		Data:       tx.Data(),
		AccessList: txAccessListToProto(tx.AccessList()),
		V:          v.Bytes(),
		R:          r.Bytes(),
		S:          s.Bytes(),
	}

	// add To field only if it exists
	if to := tx.To(); to != nil {
		toBytes := to.Bytes()
		dynamicFeeTx.To = toBytes
	}

	return dynamicFeeTx

}

func transactionToProto(tx *ethrpc.Transaction) *Transaction {
	switch tx.Type() {
	case gethtypes.LegacyTxType:
		return &Transaction{
			TransactionType: &Transaction_LegacyTransaction{
				LegacyTransaction: LegacyTransactionToProto(tx),
			},
		}
	case gethtypes.AccessListTxType:
		return &Transaction{
			TransactionType: &Transaction_AccessListTransaction{
				AccessListTransaction: AccessListTransactionToProto(tx),
			},
		}
	case gethtypes.BlobTxType:
		return &Transaction{
			TransactionType: &Transaction_BlobTransaction{
				BlobTransaction: BlobTransactionToProto(tx),
			},
		}
	case gethtypes.DynamicFeeTxType:
		return &Transaction{
			TransactionType: &Transaction_DynamicFeeTransaction{
				DynamicFeeTransaction: DynamicFeeTransactionToProto(tx),
			},
		}
	default:
		return nil
	}
}

func unclesToProto(uncles []gethcommon.Hash) [][]byte {
	result := make([][]byte, len(uncles))
	for i, uncle := range uncles {
		result[i] = uncle.Bytes()
	}
	return result
}

func ancesterHeaderToProto(h *gethtypes.Header) *AncestorHeader {
	header := &AncestorHeader{
		ParentHash:  h.ParentHash.Bytes(),
		UncleHash:   h.UncleHash[:],
		Coinbase:    h.Coinbase[:],
		Root:        h.Root[:],
		TxHash:      h.TxHash[:],
		ReceiptHash: h.ReceiptHash[:],
		Bloom:       h.Bloom[:],
		Difficulty:  h.Difficulty.Uint64(),
		Number:      h.Number.Uint64(),
		GasLimit:    h.GasLimit,
		GasUsed:     h.GasUsed,
		Time:        h.Time,
		Extra:       h.Extra,
		MixDigest:   h.MixDigest.Bytes(),
		Nonce:       h.Nonce[:],
	}

	// Add optional fields only if they exist
	if h.BaseFee != nil {
		header.BaseFee = h.BaseFee.Bytes()
	}
	if h.WithdrawalsHash != nil {
		header.WithdrawalsHash = h.WithdrawalsHash[:]
	}
	if h.BlobGasUsed != nil {
		header.BlobGasUsed = h.BlobGasUsed
	}
	if h.ExcessBlobGas != nil {
		header.ExcessBlobGas = h.ExcessBlobGas
	}
	if h.ParentBeaconRoot != nil {
		header.ParentBeaconRoot = h.ParentBeaconRoot[:]
	}
	if h.RequestsHash != nil {
		header.RequestsHash = h.RequestsHash[:]
	}

	return header
}

func ancesterHeadersToProto(headers []*gethtypes.Header) []*AncestorHeader {
	result := make([]*AncestorHeader, len(headers))
	for i, h := range headers {
		result[i] = ancesterHeaderToProto(h)
	}
	return result
}

func txAccessListToProto(al gethtypes.AccessList) []*AccessTuple {
	if len(al) == 0 {
		return nil
	}
	result := make([]*AccessTuple, len(al))
	for i, tuple := range al {
		storageKeys := make([][]byte, len(tuple.StorageKeys))
		for j, key := range tuple.StorageKeys {
			storageKeys[j] = key.Bytes()
		}
		result[i] = &AccessTuple{
			Address:     tuple.Address.Bytes(),
			StorageKeys: storageKeys,
		}
	}
	return result
}

func accessListToProto(al map[gethcommon.Address][]hexutil.Bytes) map[string]*AccessList {
	if len(al) == 0 {
		return nil
	}
	result := make(map[string]*AccessList, len(al))
	for addr, slots := range al {
		storageSlots := make([][]byte, len(slots))
		for i, slot := range slots {
			storageSlots[i] = []byte(slot)
		}
		result[addr.Hex()] = &AccessList{
			StorageSlots: storageSlots,
		}
	}
	return result
}

func codesToProto(codes []hexutil.Bytes) [][]byte {
	result := make([][]byte, len(codes))
	for i, code := range codes {
		result[i] = []byte(code)
	}
	return result
}

// ************** From proto to Go **************
func withdrawalsFromProto(withdrawals []*Withdrawal) []*gethtypes.Withdrawal {
	result := make([]*gethtypes.Withdrawal, len(withdrawals))
	for i, withdrawal := range withdrawals {
		result[i] = withdrawalFromProto(withdrawal)
	}
	return result
}

func withdrawalFromProto(withdrawal *Withdrawal) *gethtypes.Withdrawal {
	return &gethtypes.Withdrawal{
		Index:     withdrawal.GetIndex(),
		Validator: withdrawal.GetValidatorIndex(),
		Address:   gethcommon.BytesToAddress(withdrawal.GetAddress()),
		Amount:    withdrawal.GetAmount(),
	}
}

func blockFromProto(p *Block) *ethrpc.Block {
	block := &ethrpc.Block{
		Header: ethrpc.Header{
			Number:       (*hexutil.Big)(new(big.Int).SetUint64(p.GetNumber())),
			ParentHash:   gethcommon.BytesToHash(p.GetParentHash()),
			Nonce:        gethtypes.EncodeNonce(p.GetNonce()),
			MixHash:      gethcommon.BytesToHash(p.GetMixHash()),
			UncleHash:    gethcommon.BytesToHash(p.GetUncleHash()),
			LogsBloom:    gethtypes.BytesToBloom(p.GetLogsBloom()),
			Root:         gethcommon.BytesToHash(p.GetRoot()),
			Miner:        gethcommon.BytesToAddress(p.GetMiner()),
			Difficulty:   (*hexutil.Big)(new(big.Int).SetUint64(p.GetDifficulty())),
			Extra:        p.GetExtra(),
			GasLimit:     hexutil.Uint64(p.GetGasLimit()),
			GasUsed:      hexutil.Uint64(p.GetGasUsed()),
			Time:         hexutil.Uint64(p.GetTime()),
			TxRoot:       gethcommon.BytesToHash(p.GetTxRoot()),
			ReceiptsRoot: gethcommon.BytesToHash(p.GetReceiptsRoot()),
			Hash:         gethcommon.BytesToHash(p.GetHash()),
		},
	}

	txs := transactionsFromProto(p.GetTransactions())
	blobGasUsed := hexutil.Uint64(p.GetBlobGasUsed())
	excessBlobGas := hexutil.Uint64(p.GetExcessBlobGas())
	block.Transactions = txs
	block.Withdrawals = withdrawalsFromProto(p.GetWithdrawals())
	block.BlobGasUsed = &blobGasUsed
	block.ExcessBlobGas = &excessBlobGas

	// Add optional fields only if they exist
	if baseFee := p.GetBaseFee(); len(baseFee) > 0 {
		block.BaseFee = (*hexutil.Big)(new(big.Int).SetBytes(baseFee))
	}
	if withdrawalsRoot := p.GetWithdrawalsRoot(); len(withdrawalsRoot) > 0 {
		hash := gethcommon.BytesToHash(withdrawalsRoot)
		block.WithdrawalsRoot = &hash
	}
	if parentBeaconRoot := p.GetParentBeaconRoot(); len(parentBeaconRoot) > 0 {
		hash := gethcommon.BytesToHash(parentBeaconRoot)
		block.ParentBeaconRoot = &hash
	}
	if requestsRoot := p.GetRequestsRoot(); len(requestsRoot) > 0 {
		hash := gethcommon.BytesToHash(requestsRoot)
		block.RequestsRoot = &hash
	}

	return block
}

func transactionsFromProto(txs []*Transaction) []*ethrpc.Transaction {
	result := make([]*ethrpc.Transaction, len(txs))
	for i, tx := range txs {
		result[i] = transactionFromProto(tx)
	}
	return result
}

func transactionFromProto(tx *Transaction) *ethrpc.Transaction {
	switch t := tx.TransactionType.(type) {
	case *Transaction_LegacyTransaction:
		gethTx := legacyTransactionFromProto(t.LegacyTransaction)
		return ethrpc.NewTransactionFromGeth(gethTx)
	case *Transaction_AccessListTransaction:
		gethTx := accessListTransactionFromProto(t.AccessListTransaction)
		return ethrpc.NewTransactionFromGeth(gethTx)
	case *Transaction_BlobTransaction:
		gethTx := blobTransactionFromProto(t.BlobTransaction)
		return ethrpc.NewTransactionFromGeth(gethTx)
	case *Transaction_DynamicFeeTransaction:
		gethTx := dynamicFeeTransactionFromProto(t.DynamicFeeTransaction)
		return ethrpc.NewTransactionFromGeth(gethTx)
	default:
		return nil
	}
}

func ancesterHeaderFromProto(h *AncestorHeader) *gethtypes.Header {
	header := &gethtypes.Header{
		ParentHash:  gethcommon.BytesToHash(h.GetParentHash()),
		UncleHash:   gethcommon.BytesToHash(h.GetUncleHash()),
		Coinbase:    gethcommon.BytesToAddress(h.GetCoinbase()),
		Root:        gethcommon.BytesToHash(h.GetRoot()),
		TxHash:      gethcommon.BytesToHash(h.GetTxHash()),
		ReceiptHash: gethcommon.BytesToHash(h.GetReceiptHash()),
		Bloom:       gethtypes.Bloom(h.GetBloom()),
		Difficulty:  new(big.Int).SetUint64(h.GetDifficulty()),
		Number:      new(big.Int).SetUint64(h.GetNumber()),
		GasLimit:    h.GetGasLimit(),
		GasUsed:     h.GetGasUsed(),
		Time:        h.GetTime(),
		Extra:       h.GetExtra(),
		MixDigest:   gethcommon.BytesToHash(h.GetMixDigest()),
		Nonce:       gethtypes.BlockNonce(h.GetNonce()),
	}

	if baseFee := h.GetBaseFee(); len(baseFee) > 0 {
		header.BaseFee = new(big.Int).SetBytes(baseFee)
	}
	if withdrawalsHash := h.GetWithdrawalsHash(); len(withdrawalsHash) > 0 {
		hash := gethcommon.BytesToHash(withdrawalsHash)
		header.WithdrawalsHash = &hash
	}
	if blobGasUsed := h.GetBlobGasUsed(); blobGasUsed != 0 {
		header.BlobGasUsed = &blobGasUsed
	}
	if excessBlobGas := h.GetExcessBlobGas(); excessBlobGas != 0 {
		header.ExcessBlobGas = &excessBlobGas
	}
	if parentBeaconRoot := h.GetParentBeaconRoot(); len(parentBeaconRoot) > 0 {
		hash := gethcommon.BytesToHash(parentBeaconRoot)
		header.ParentBeaconRoot = &hash
	}
	if requestsHash := h.GetRequestsHash(); len(requestsHash) > 0 {
		hash := gethcommon.BytesToHash(requestsHash)
		header.RequestsHash = &hash
	}

	return header
}

func ancesterHeadersFromProto(headers []*AncestorHeader) []*gethtypes.Header {
	result := make([]*gethtypes.Header, len(headers))
	for i, h := range headers {
		result[i] = ancesterHeaderFromProto(h)
	}
	return result
}

func accessListFromProto(al map[string]*AccessList) map[gethcommon.Address][]hexutil.Bytes {
	result := make(map[gethcommon.Address][]hexutil.Bytes)
	for addrHex, slots := range al {
		result[gethcommon.HexToAddress(addrHex)] = bytesToHexutil(slots.StorageSlots)
	}
	return result
}

func codesFromProto(codes [][]byte) []hexutil.Bytes {
	result := make([]hexutil.Bytes, len(codes))
	for i, code := range codes {
		result[i] = hexutil.Bytes(code)
	}
	return result
}

func legacyTransactionFromProto(t *LegacyTransaction) *gethtypes.Transaction {
	return gethtypes.NewTx(&gethtypes.LegacyTx{
		Nonce:    t.GetNonce(),
		GasPrice: new(big.Int).SetBytes(t.GetGasPrice()),
		Gas:      t.GetGas(),
		To:       (*gethcommon.Address)(t.GetTo()),
		Value:    new(big.Int).SetBytes(t.GetValue()),
		Data:     t.GetData(),
		V:        new(big.Int).SetBytes(t.GetV()),
		R:        new(big.Int).SetBytes(t.GetR()),
		S:        new(big.Int).SetBytes(t.GetS()),
	})
}

func accessListTransactionFromProto(t *AccessListTransaction) *gethtypes.Transaction {
	accessList := make(gethtypes.AccessList, len(t.GetAccessList()))
	for i, tuple := range t.GetAccessList() {
		storageKeys := make([]gethcommon.Hash, len(tuple.GetStorageKeys()))
		for j, key := range tuple.GetStorageKeys() {
			storageKeys[j] = gethcommon.BytesToHash(key)
		}
		accessList[i] = gethtypes.AccessTuple{
			Address:     gethcommon.BytesToAddress(tuple.GetAddress()),
			StorageKeys: storageKeys,
		}
	}

	return gethtypes.NewTx(&gethtypes.AccessListTx{
		ChainID:    big.NewInt(int64(t.GetChainId())),
		Nonce:      t.GetNonce(),
		GasPrice:   new(big.Int).SetBytes(t.GetGasPrice()),
		Gas:        t.GetGas(),
		To:         (*gethcommon.Address)(t.GetTo()),
		Value:      new(big.Int).SetBytes(t.GetValue()),
		Data:       t.GetData(),
		AccessList: accessList,
		V:          new(big.Int).SetBytes(t.GetV()),
		R:          new(big.Int).SetBytes(t.GetR()),
		S:          new(big.Int).SetBytes(t.GetS()),
	})
}

func blobTransactionFromProto(t *BlobTransaction) *gethtypes.Transaction {
	accessList := make(gethtypes.AccessList, len(t.GetAccessList()))
	for i, tuple := range t.GetAccessList() {
		storageKeys := make([]gethcommon.Hash, len(tuple.GetStorageKeys()))
		for j, key := range tuple.GetStorageKeys() {
			storageKeys[j] = gethcommon.BytesToHash(key)
		}
		accessList[i] = gethtypes.AccessTuple{
			Address:     gethcommon.BytesToAddress(tuple.GetAddress()),
			StorageKeys: storageKeys,
		}
	}

	blobHashes := make([]gethcommon.Hash, len(t.GetBlobHashes()))
	for i, hash := range t.GetBlobHashes() {
		blobHashes[i] = gethcommon.BytesToHash(hash)
	}
	return gethtypes.NewTx(&gethtypes.BlobTx{
		ChainID:    uint256.NewInt(t.GetChainId()),
		Nonce:      t.GetNonce(),
		GasTipCap:  new(uint256.Int).SetBytes(t.GetGasTipCap()),
		GasFeeCap:  new(uint256.Int).SetBytes(t.GetGasFeeCap()),
		Gas:        t.GetGas(),
		To:         gethcommon.BytesToAddress(t.GetTo()),
		Value:      uint256.NewInt(0).SetBytes(t.GetValue()),
		Data:       t.GetData(),
		AccessList: accessList,
		BlobFeeCap: new(uint256.Int).SetBytes(t.GetBlobFeeCap()),
		BlobHashes: blobHashes,
		V:          new(uint256.Int).SetBytes(t.GetV()),
		R:          new(uint256.Int).SetBytes(t.GetR()),
		S:          new(uint256.Int).SetBytes(t.GetS()),
	})
}

func dynamicFeeTransactionFromProto(t *DynamicFeeTransaction) *gethtypes.Transaction {
	accessList := make(gethtypes.AccessList, len(t.GetAccessList()))
	for i, tuple := range t.GetAccessList() {
		storageKeys := make([]gethcommon.Hash, len(tuple.GetStorageKeys()))
		for j, key := range tuple.GetStorageKeys() {
			storageKeys[j] = gethcommon.BytesToHash(key)
		}
		accessList[i] = gethtypes.AccessTuple{
			Address:     gethcommon.BytesToAddress(tuple.GetAddress()),
			StorageKeys: storageKeys,
		}
	}

	tx := &gethtypes.DynamicFeeTx{
		ChainID:    big.NewInt(int64(t.GetChainId())),
		Nonce:      t.GetNonce(),
		GasTipCap:  new(big.Int).SetBytes(t.GetGasTipCap()),
		GasFeeCap:  new(big.Int).SetBytes(t.GetGasFeeCap()),
		Gas:        t.GetGas(),
		Value:      new(big.Int).SetBytes(t.GetValue()),
		Data:       t.GetData(),
		AccessList: accessList,
		V:          new(big.Int).SetBytes(t.GetV()),
		R:          new(big.Int).SetBytes(t.GetR()),
		S:          new(big.Int).SetBytes(t.GetS()),
	}

	if t.GetTo() != nil {
		tx.To = (*gethcommon.Address)(t.GetTo())
	}

	return gethtypes.NewTx(tx)
}

func chainConfigFromProto(c *ChainConfig) *params.ChainConfig {
	if c == nil {
		return nil
	}

	// Helper function to convert []byte to *big.Int
	toBigIntIfNotEmpty := func(b []byte) *big.Int {
		if len(b) == 0 {
			return nil
		}
		return new(big.Int).SetBytes(b)
	}

	// Helper function to convert uint64 pointer to *uint64
	toUint64Ptr := func(t uint64) *uint64 {
		if t == 0 {
			return nil
		}
		return &t
	}

	chainConfig := &params.ChainConfig{
		ChainID:                 new(big.Int).SetUint64(c.GetChainId()),
		HomesteadBlock:          toBigIntIfNotEmpty(c.GetHomesteadBlock()),
		DAOForkBlock:            toBigIntIfNotEmpty(c.GetDaoForkBlock()),
		DAOForkSupport:          c.GetDaoForkSupport(),
		EIP150Block:             toBigIntIfNotEmpty(c.GetEip150Block()),
		EIP155Block:             toBigIntIfNotEmpty(c.GetEip155Block()),
		EIP158Block:             toBigIntIfNotEmpty(c.GetEip158Block()),
		ByzantiumBlock:          toBigIntIfNotEmpty(c.GetByzantiumBlock()),
		ConstantinopleBlock:     toBigIntIfNotEmpty(c.GetConstantinopleBlock()),
		PetersburgBlock:         toBigIntIfNotEmpty(c.GetPetersburgBlock()),
		IstanbulBlock:           toBigIntIfNotEmpty(c.GetIstanbulBlock()),
		MuirGlacierBlock:        toBigIntIfNotEmpty(c.GetMuirGlacierBlock()),
		BerlinBlock:             toBigIntIfNotEmpty(c.GetBerlinBlock()),
		LondonBlock:             toBigIntIfNotEmpty(c.GetLondonBlock()),
		ArrowGlacierBlock:       toBigIntIfNotEmpty(c.GetArrowGlacierBlock()),
		GrayGlacierBlock:        toBigIntIfNotEmpty(c.GetGrayGlacierBlock()),
		MergeNetsplitBlock:      toBigIntIfNotEmpty(c.GetMergeNetsplitBlock()),
		ShanghaiTime:            toUint64Ptr(c.GetShanghaiTime()),
		CancunTime:              toUint64Ptr(c.GetCancunTime()),
		PragueTime:              toUint64Ptr(c.GetPragueTime()),
		VerkleTime:              toUint64Ptr(c.GetVerkleTime()),
		TerminalTotalDifficulty: toBigIntIfNotEmpty(c.GetTerminalTotalDifficulty()),
		DepositContractAddress:  gethcommon.BytesToAddress(c.GetDepositContractAddress()),
	}

	// Handle consensus engine configs
	if len(c.GetEthash()) > 0 {
		chainConfig.Ethash = &params.EthashConfig{}
	}
	if len(c.GetClique()) > 0 {
		chainConfig.Clique = &params.CliqueConfig{}
	}

	return chainConfig
}

func bytesOrNil(h *gethcommon.Hash) []byte {
	if h == nil {
		return nil
	}
	return h[:]
}

func bytesFromHexutil(hexBytes []hexutil.Bytes) [][]byte {
	result := make([][]byte, len(hexBytes))
	for i, b := range hexBytes {
		result[i] = []byte(b)
	}
	return result
}

func bytesToHexutil(bytes [][]byte) []hexutil.Bytes {
	result := make([]hexutil.Bytes, len(bytes))
	for i, b := range bytes {
		result[i] = hexutil.Bytes(b)
	}
	return result
}

// Helper function to convert *big.Int to []byte
func toBytesIfNotNil(b *big.Int) []byte {
	if b != nil {
		return b.Bytes()
	}
	return nil
}
