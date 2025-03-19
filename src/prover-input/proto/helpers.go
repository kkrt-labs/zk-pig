package proto

import (
	"math/big"

	gethcommon "github.com/ethereum/go-ethereum/common"
)

func bytesToHashPtr(b []byte) *gethcommon.Hash {
	if b == nil {
		return nil
	}
	hash := gethcommon.BytesToHash(b)
	return &hash
}

func addrToBytes(addr *gethcommon.Address) []byte {
	if addr == nil {
		return nil
	}
	return addr.Bytes()
}

func bytesToBigInt(b []byte) *big.Int {
	if b == nil {
		return nil
	}
	return new(big.Int).SetBytes(b)
}

// Helper function to convert *big.Int to []byte
func bigIntToBytes(b *big.Int) []byte {
	if b != nil {
		return b.Bytes()
	}
	return nil
}
