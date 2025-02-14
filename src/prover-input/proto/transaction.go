package proto

import (
	"math/big"
	"time"

	gethcommon "github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto/kzg4844"
	"github.com/holiman/uint256"
)

func TransactionToProto(tx *gethtypes.Transaction) *Transaction {
	if tx == nil {
		return nil
	}

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

func TransactionFromProto(t *Transaction) *gethtypes.Transaction {
	switch tx := t.GetTransactionType().(type) {
	case *Transaction_LegacyTransaction:
		return LegacyTransactionFromProto(tx.LegacyTransaction)
	case *Transaction_AccessListTransaction:
		return AccessListTransactionFromProto(tx.AccessListTransaction)
	case *Transaction_BlobTransaction:
		return BlobTransactionFromProto(tx.BlobTransaction)
	case *Transaction_DynamicFeeTransaction:
		return DynamicFeeTransactionFromProto(tx.DynamicFeeTransaction)
	default:
		return nil
	}
}

func LegacyTransactionToProto(tx *gethtypes.Transaction) *LegacyTransaction {
	if tx == nil {
		return nil
	}

	v, r, s := tx.RawSignatureValues()
	return &LegacyTransaction{
		Nonce:    tx.Nonce(),
		GasPrice: bigIntToBytes(tx.GasPrice()),
		Gas:      tx.Gas(),
		To:       addrToBytes(tx.To()),
		Value:    bigIntToBytes(tx.Value()),
		Data:     tx.Data(),
		V:        bigIntToBytes(v),
		R:        bigIntToBytes(r),
		S:        bigIntToBytes(s),
	}
}

func LegacyTransactionFromProto(t *LegacyTransaction) *gethtypes.Transaction {
	var to *gethcommon.Address
	if t.GetTo() != nil {
		to = new(gethcommon.Address)
		to.SetBytes(t.GetTo())
	}

	tx := gethtypes.NewTx(&gethtypes.LegacyTx{
		Nonce:    t.GetNonce(),
		GasPrice: new(big.Int).SetBytes(t.GetGasPrice()),
		Gas:      t.GetGas(),
		To:       to,
		Value:    new(big.Int).SetBytes(t.GetValue()),
		Data:     t.GetData(),
		V:        new(big.Int).SetBytes(t.GetV()),
		R:        new(big.Int).SetBytes(t.GetR()),
		S:        new(big.Int).SetBytes(t.GetS()),
	})
	tx.SetTime(time.Unix(0, 0))
	return tx
}

func AccessListTransactionToProto(tx *gethtypes.Transaction) *AccessListTransaction {
	v, r, s := tx.RawSignatureValues()
	return &AccessListTransaction{
		ChainId:    tx.ChainId().Bytes(),
		Nonce:      tx.Nonce(),
		GasPrice:   tx.GasPrice().Bytes(),
		Gas:        tx.Gas(),
		To:         addrToBytes(tx.To()),
		Value:      tx.Value().Bytes(),
		Data:       tx.Data(),
		AccessList: AccessListToProto(tx.AccessList()),
		V:          v.Bytes(),
		R:          r.Bytes(),
		S:          s.Bytes(),
	}
}

func AccessListTransactionFromProto(t *AccessListTransaction) *gethtypes.Transaction {
	innerTx := &gethtypes.AccessListTx{
		ChainID:    new(big.Int).SetBytes(t.GetChainId()),
		Nonce:      t.GetNonce(),
		GasPrice:   new(big.Int).SetBytes(t.GetGasPrice()),
		Gas:        t.GetGas(),
		Value:      new(big.Int).SetBytes(t.GetValue()),
		Data:       t.GetData(),
		AccessList: AccessListFromProto(t.GetAccessList()),
		V:          new(big.Int).SetBytes(t.GetV()),
		R:          new(big.Int).SetBytes(t.GetR()),
		S:          new(big.Int).SetBytes(t.GetS()),
	}

	if t.GetTo() != nil {
		innerTx.To = (*gethcommon.Address)(t.GetTo())
	}

	tx := gethtypes.NewTx(innerTx)
	tx.SetTime(time.Unix(0, 0))
	return tx
}

func BlobTransactionToProto(tx *gethtypes.Transaction) *BlobTransaction {
	blobHashes := make([][]byte, len(tx.BlobHashes()))
	for i, hash := range tx.BlobHashes() {
		blobHashes[i] = hash.Bytes()
	}
	v, r, s := tx.RawSignatureValues()
	return &BlobTransaction{
		ChainId:    tx.ChainId().Bytes(),
		Nonce:      tx.Nonce(),
		GasTipCap:  tx.GasTipCap().Bytes(),
		GasFeeCap:  tx.GasFeeCap().Bytes(),
		Gas:        tx.Gas(),
		To:         tx.To().Bytes(),
		Value:      tx.Value().Bytes(),
		Data:       tx.Data(),
		AccessList: AccessListToProto(tx.AccessList()),
		BlobFeeCap: tx.BlobGasFeeCap().Bytes(),
		BlobHashes: blobHashes,
		Sidecar:    BlobSidecarToProto(tx.BlobTxSidecar()),
		V:          v.Bytes(),
		R:          r.Bytes(),
		S:          s.Bytes(),
	}
}

func BlobTransactionFromProto(t *BlobTransaction) *gethtypes.Transaction {
	blobHashes := make([]gethcommon.Hash, len(t.GetBlobHashes()))
	for i, hash := range t.GetBlobHashes() {
		blobHashes[i] = gethcommon.BytesToHash(hash)
	}

	innerTx := &gethtypes.BlobTx{
		ChainID:    new(uint256.Int).SetBytes(t.GetChainId()),
		Nonce:      t.GetNonce(),
		GasTipCap:  new(uint256.Int).SetBytes(t.GetGasTipCap()),
		GasFeeCap:  new(uint256.Int).SetBytes(t.GetGasFeeCap()),
		Gas:        t.GetGas(),
		Value:      uint256.NewInt(0).SetBytes(t.GetValue()),
		Data:       t.GetData(),
		AccessList: AccessListFromProto(t.GetAccessList()),
		BlobFeeCap: new(uint256.Int).SetBytes(t.GetBlobFeeCap()),
		BlobHashes: blobHashes,
		Sidecar:    BlobSidecarFromProto(t.GetSidecar()),
		V:          new(uint256.Int).SetBytes(t.GetV()),
		R:          new(uint256.Int).SetBytes(t.GetR()),
		S:          new(uint256.Int).SetBytes(t.GetS()),
	}

	if t.GetTo() != nil {
		innerTx.To = gethcommon.Address(t.GetTo())
	}

	tx := gethtypes.NewTx(innerTx)
	tx.SetTime(time.Unix(0, 0))
	return tx
}

func BlobSidecarToProto(sidecar *gethtypes.BlobTxSidecar) *BlobTxSidecar {
	if sidecar == nil {
		return nil
	}

	sidecarProto := new(BlobTxSidecar)
	for _, blob := range sidecar.Blobs { //nolint:gocritic // TODO: each iteration copies 131072 bytes (consider pointers or indexing)
		sidecarProto.Blobs = append(sidecarProto.Blobs, blob[:])
	}
	for _, commitment := range sidecar.Commitments {
		sidecarProto.Commitments = append(sidecarProto.Commitments, commitment[:])
	}
	for _, proof := range sidecar.Proofs {
		sidecarProto.Proofs = append(sidecarProto.Proofs, proof[:])
	}
	return sidecarProto
}

func BlobSidecarFromProto(t *BlobTxSidecar) *gethtypes.BlobTxSidecar {
	if t == nil {
		return nil
	}
	sidecar := new(gethtypes.BlobTxSidecar)
	for _, blob := range t.GetBlobs() {
		sidecar.Blobs = append(sidecar.Blobs, kzg4844.Blob(blob))
	}
	for _, commitment := range t.GetCommitments() {
		sidecar.Commitments = append(sidecar.Commitments, kzg4844.Commitment(commitment))
	}
	for _, proof := range t.GetProofs() {
		sidecar.Proofs = append(sidecar.Proofs, kzg4844.Proof(proof))
	}
	return sidecar
}

func DynamicFeeTransactionToProto(tx *gethtypes.Transaction) *DynamicFeeTransaction {
	v, r, s := tx.RawSignatureValues()
	dynamicFeeTx := &DynamicFeeTransaction{
		ChainId:    tx.ChainId().Bytes(),
		Nonce:      tx.Nonce(),
		GasTipCap:  tx.GasTipCap().Bytes(),
		GasFeeCap:  tx.GasFeeCap().Bytes(),
		Gas:        tx.Gas(),
		Value:      tx.Value().Bytes(),
		Data:       tx.Data(),
		AccessList: AccessListToProto(tx.AccessList()),
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

func DynamicFeeTransactionFromProto(t *DynamicFeeTransaction) *gethtypes.Transaction {
	innerTx := &gethtypes.DynamicFeeTx{
		ChainID:    new(big.Int).SetBytes(t.GetChainId()),
		Nonce:      t.GetNonce(),
		GasTipCap:  new(big.Int).SetBytes(t.GetGasTipCap()),
		GasFeeCap:  new(big.Int).SetBytes(t.GetGasFeeCap()),
		Gas:        t.GetGas(),
		Value:      new(big.Int).SetBytes(t.GetValue()),
		Data:       t.GetData(),
		AccessList: AccessListFromProto(t.GetAccessList()),
		V:          new(big.Int).SetBytes(t.GetV()),
		R:          new(big.Int).SetBytes(t.GetR()),
		S:          new(big.Int).SetBytes(t.GetS()),
	}

	if t.GetTo() != nil {
		innerTx.To = (*gethcommon.Address)(t.GetTo())
	}

	tx := gethtypes.NewTx(innerTx)
	tx.SetTime(time.Unix(0, 0))
	return tx
}

func AccessListToProto(al gethtypes.AccessList) []*AccessTuple {
	if al == nil {
		return nil
	}

	result := make([]*AccessTuple, len(al))
	for i, tuple := range al {
		result[i] = AccessTupleToProto(&tuple)
	}
	return result
}

func AccessListFromProto(p []*AccessTuple) gethtypes.AccessList {
	if p == nil {
		return nil
	}

	result := make(gethtypes.AccessList, len(p))
	for i, tuple := range p {
		result[i] = *(AccessTupleFromProto(tuple))
	}
	return result
}

func AccessTupleToProto(tuple *gethtypes.AccessTuple) *AccessTuple {
	storageKeys := make([][]byte, len(tuple.StorageKeys))
	for j, key := range tuple.StorageKeys {
		storageKeys[j] = key.Bytes()
	}
	return &AccessTuple{
		Address:     tuple.Address.Bytes(),
		StorageKeys: storageKeys,
	}
}

func AccessTupleFromProto(p *AccessTuple) *gethtypes.AccessTuple {
	storageKeys := make([]gethcommon.Hash, len(p.GetStorageKeys()))
	for j, key := range p.GetStorageKeys() {
		storageKeys[j] = gethcommon.BytesToHash(key)
	}
	return &gethtypes.AccessTuple{
		Address:     gethcommon.BytesToAddress(p.GetAddress()),
		StorageKeys: storageKeys,
	}
}

func TransactionsToProto(transactions []*gethtypes.Transaction) []*Transaction {
	if transactions == nil {
		return nil
	}

	result := make([]*Transaction, len(transactions))
	for i, t := range transactions {
		result[i] = TransactionToProto(t)
	}
	return result
}

func TransactionsFromProto(transactions []*Transaction) []*gethtypes.Transaction {
	if transactions == nil {
		return nil
	}

	result := make([]*gethtypes.Transaction, len(transactions))
	for i, t := range transactions {
		result[i] = TransactionFromProto(t)
	}
	return result
}
