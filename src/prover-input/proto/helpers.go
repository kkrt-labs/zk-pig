package proto

import (
	"math/big"

	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

func bytesToHashPtr(b []byte) *gethcommon.Hash {
	if b == nil {
		return nil
	}
	hash := gethcommon.BytesToHash(b)
	return &hash
}

func bytesToBigInt(b []byte) *big.Int {
	if b == nil {
		return nil
	}
	return new(big.Int).SetBytes(b)
}

func hexBytesToBytes(hexBytes []hexutil.Bytes) [][]byte {
	if hexBytes == nil {
		return nil
	}
	result := make([][]byte, len(hexBytes))
	for i, b := range hexBytes {
		result[i] = []byte(b)
	}
	return result
}

func addrToBytes(addr *gethcommon.Address) []byte {
	if addr == nil {
		return nil
	}
	return addr.Bytes()
}

// Helper function to convert *big.Int to []byte
func bigIntToBytes(b *big.Int) []byte {
	if b != nil {
		return b.Bytes()
	}
	return nil
}

func bytesToHexutil(b [][]byte) []hexutil.Bytes {
	if b == nil {
		return nil
	}
	result := make([]hexutil.Bytes, len(b))
	for i, b := range b {
		result[i] = hexutil.Bytes(b)
	}
	return result
}
