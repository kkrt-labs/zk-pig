package proto

import (
	gethcommon "github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	input "github.com/kkrt-labs/zk-pig/src/prover-input"
)

func BlockToProto(b *input.Block) *Block {
	if b == nil {
		return nil
	}

	return &Block{
		Header:       HeaderToProto(b.Header),
		Transactions: TransactionsToProto(b.Transactions),
		Uncles:       HeadersToProto(b.Uncles), // we assume a post-merge
		Withdrawals:  WithdrawalsToProto(b.Withdrawals),
	}
}

func BlockFromProto(b *Block) *input.Block {
	if b == nil {
		return nil
	}

	return &input.Block{
		Header:       HeaderFromProto(b.Header),
		Transactions: TransactionsFromProto(b.Transactions),
		Uncles:       HeadersFromProto(b.Uncles), // we assume a post-merge
		Withdrawals:  WithdrawalsFromProto(b.Withdrawals),
	}
}

func BlocksToProto(blocks []*input.Block) []*Block {
	if blocks == nil {
		return nil
	}

	result := make([]*Block, len(blocks))
	for i, b := range blocks {
		result[i] = BlockToProto(b)
	}
	return result
}

func BlocksFromProto(blocks []*Block) []*input.Block {
	if blocks == nil {
		return nil
	}

	result := make([]*input.Block, len(blocks))
	for i, b := range blocks {
		result[i] = BlockFromProto(b)
	}
	return result
}

func HeadersToProto(headers []*gethtypes.Header) []*Header {
	if headers == nil {
		return nil
	}

	result := make([]*Header, len(headers))
	for i, h := range headers {
		result[i] = HeaderToProto(h)
	}
	return result
}

func HeadersFromProto(headers []*Header) []*gethtypes.Header {
	if headers == nil {
		return nil
	}

	result := make([]*gethtypes.Header, len(headers))
	for i, h := range headers {
		result[i] = HeaderFromProto(h)
	}
	return result
}

func HeaderToProto(h *gethtypes.Header) *Header {
	if h == nil {
		return nil
	}

	header := &Header{
		ParentHash:       h.ParentHash.Bytes(),
		Sha3Uncles:       h.UncleHash[:],
		Miner:            h.Coinbase[:],
		Root:             h.Root[:],
		TransactionsRoot: h.TxHash[:],
		ReceiptsRoot:     h.ReceiptHash[:],
		LogsBloom:        h.Bloom[:],
		Difficulty:       bigIntToBytes(h.Difficulty),
		Number:           bigIntToBytes(h.Number),
		GasLimit:         h.GasLimit,
		GasUsed:          h.GasUsed,
		Timestamp:        h.Time,
		ExtraData:        h.Extra,
		MixHash:          h.MixDigest.Bytes(),
		Nonce:            h.Nonce.Uint64(),
	}

	// Add optional fields only if they exist
	if h.BaseFee != nil {
		header.BaseFeePerGas = h.BaseFee.Bytes()
	}
	if h.WithdrawalsHash != nil {
		header.WithdrawalsRoot = h.WithdrawalsHash[:]
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
		header.RequestsRoot = h.RequestsHash[:]
	}

	return header
}

func HeaderFromProto(h *Header) *gethtypes.Header {
	if h == nil {
		return nil
	}

	header := &gethtypes.Header{
		ParentHash:       gethcommon.BytesToHash(h.GetParentHash()),
		UncleHash:        gethcommon.BytesToHash(h.GetSha3Uncles()),
		Coinbase:         gethcommon.BytesToAddress(h.GetMiner()),
		Root:             gethcommon.BytesToHash(h.GetRoot()),
		TxHash:           gethcommon.BytesToHash(h.GetTransactionsRoot()),
		ReceiptHash:      gethcommon.BytesToHash(h.GetReceiptsRoot()),
		Bloom:            gethtypes.Bloom(h.GetLogsBloom()),
		Difficulty:       bytesToBigInt(h.Difficulty),
		Number:           bytesToBigInt(h.Number),
		GasLimit:         h.GetGasLimit(),
		GasUsed:          h.GetGasUsed(),
		Time:             h.GetTimestamp(),
		Extra:            h.GetExtraData(),
		MixDigest:        gethcommon.BytesToHash(h.GetMixHash()),
		Nonce:            gethtypes.EncodeNonce(h.GetNonce()),
		BaseFee:          bytesToBigInt(h.GetBaseFeePerGas()),
		WithdrawalsHash:  bytesToHashPtr(h.GetWithdrawalsRoot()),
		BlobGasUsed:      h.BlobGasUsed,
		ExcessBlobGas:    h.ExcessBlobGas,
		RequestsHash:     bytesToHashPtr(h.RequestsRoot),
		ParentBeaconRoot: bytesToHashPtr(h.ParentBeaconRoot),
	}

	return header
}

func WithdrawalsToProto(withdrawals []*gethtypes.Withdrawal) []*Withdrawal {
	if withdrawals == nil {
		return nil
	}

	result := make([]*Withdrawal, len(withdrawals))
	for i, w := range withdrawals {
		result[i] = WithdrawalToProto(w)
	}
	return result
}

func WithdrawalsFromProto(withdrawals []*Withdrawal) []*gethtypes.Withdrawal {
	if withdrawals == nil {
		return nil
	}

	result := make([]*gethtypes.Withdrawal, len(withdrawals))
	for i, w := range withdrawals {
		result[i] = WithdrawalFromProto(w)
	}
	return result
}

func WithdrawalToProto(w *gethtypes.Withdrawal) *Withdrawal {
	if w == nil {
		return nil
	}

	return &Withdrawal{
		Index:          w.Index,
		ValidatorIndex: w.Validator,
		Address:        w.Address.Bytes(),
		Amount:         w.Amount,
	}
}

func WithdrawalFromProto(w *Withdrawal) *gethtypes.Withdrawal {
	if w == nil {
		return nil
	}

	return &gethtypes.Withdrawal{
		Index:     w.Index,
		Validator: w.ValidatorIndex,
		Address:   gethcommon.BytesToAddress(w.Address),
		Amount:    w.Amount,
	}
}
