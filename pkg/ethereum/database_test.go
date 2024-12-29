package ethereum

import (
	"testing"

	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
)

func TestFillDBWithBytecode(t *testing.T) {
	db := rawdb.NewMemoryDatabase()
	codes := [][]byte{
		[]byte("code1"),
		[]byte("code2"),
	}

	// Fill the database with the bytecodes
	WriteCodes(db, codes...)

	code1 := rawdb.ReadCode(db, crypto.Keccak256Hash(codes[0]))
	assert.Equal(t, codes[0], code1, "Expected code1 to be correct")
	code2 := rawdb.ReadCode(db, crypto.Keccak256Hash(codes[1]))
	assert.Equal(t, codes[1], code2, "Expected code2 to be correct")
}
