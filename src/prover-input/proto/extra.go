package proto

import (
	gethcommon "github.com/ethereum/go-ethereum/common"
	input "github.com/kkrt-labs/zk-pig/src/prover-input"
)

func ExtraToProto(extra *input.Extra) *Extra {
	if extra == nil {
		return nil
	}

	return &Extra{
		AccessList: AccessListToProto(extra.AccessList),
		StateDiffs: StateDiffsToProto(extra.StateDiffs),
		Committed:  extra.Committed,
	}
}

func ExtraFromProto(extra *Extra) *input.Extra {
	if extra == nil {
		return nil
	}

	return &input.Extra{
		AccessList: AccessListFromProto(extra.AccessList),
		StateDiffs: StateDiffsFromProto(extra.StateDiffs),
		Committed:  extra.Committed,
	}
}

func StateDiffsToProto(stateDiffs []*input.StateDiff) []*StateDiff {
	if stateDiffs == nil {
		return nil
	}

	protoStateDiffs := make([]*StateDiff, len(stateDiffs))
	for i, stateDiff := range stateDiffs {
		protoStateDiffs[i] = StateDiffToProto(stateDiff)
	}
	return protoStateDiffs
}

func StateDiffsFromProto(protoStateDiffs []*StateDiff) []*input.StateDiff {
	if protoStateDiffs == nil {
		return nil
	}

	stateDiffs := make([]*input.StateDiff, len(protoStateDiffs))
	for i, protoStateDiff := range protoStateDiffs {
		stateDiffs[i] = StateDiffFromProto(protoStateDiff)
	}
	return stateDiffs
}

func StateDiffToProto(stateDiff *input.StateDiff) *StateDiff {
	if stateDiff == nil {
		return nil
	}

	return &StateDiff{
		Address:     addrToBytes(&stateDiff.Address),
		PreAccount:  AccountToProto(stateDiff.PreAccount),
		PostAccount: AccountToProto(stateDiff.PostAccount),
		Storage:     StorageDiffsToProto(stateDiff.Storage),
	}
}

func StateDiffFromProto(protoStateDiff *StateDiff) *input.StateDiff {
	if protoStateDiff == nil {
		return nil
	}

	return &input.StateDiff{
		Address:     gethcommon.BytesToAddress(protoStateDiff.Address),
		PreAccount:  AccountFromProto(protoStateDiff.PreAccount),
		PostAccount: AccountFromProto(protoStateDiff.PostAccount),
		Storage:     StorageDiffsFromProto(protoStateDiff.Storage),
	}
}

func StorageDiffsToProto(storageDiffs []*input.StorageDiff) []*StorageDiff {
	if storageDiffs == nil {
		return nil
	}

	protoStorageDiffs := make([]*StorageDiff, len(storageDiffs))
	for i, storageDiff := range storageDiffs {
		protoStorageDiffs[i] = StorageDiffToProto(storageDiff)
	}
	return protoStorageDiffs
}

func StorageDiffsFromProto(protoStorageDiffs []*StorageDiff) []*input.StorageDiff {
	if protoStorageDiffs == nil {
		return nil
	}

	storageDiffs := make([]*input.StorageDiff, len(protoStorageDiffs))
	for i, protoStorageDiff := range protoStorageDiffs {
		storageDiffs[i] = StorageDiffFromProto(protoStorageDiff)
	}
	return storageDiffs
}

func StorageDiffToProto(storageDiff *input.StorageDiff) *StorageDiff {
	if storageDiff == nil {
		return nil
	}

	return &StorageDiff{
		Slot:      storageDiff.Slot.Bytes(),
		PreValue:  storageDiff.PreValue.Bytes(),
		PostValue: storageDiff.PostValue.Bytes(),
	}
}

func StorageDiffFromProto(protoStorageDiff *StorageDiff) *input.StorageDiff {
	if protoStorageDiff == nil {
		return nil
	}

	return &input.StorageDiff{
		Slot:      gethcommon.BytesToHash(protoStorageDiff.Slot),
		PreValue:  gethcommon.BytesToHash(protoStorageDiff.PreValue),
		PostValue: gethcommon.BytesToHash(protoStorageDiff.PostValue),
	}
}

func AccountToProto(account *input.Account) *Account {
	if account == nil {
		return nil
	}

	return &Account{
		Balance:     bigIntToBytes(account.Balance),
		CodeHash:    account.CodeHash.Bytes(),
		Nonce:       account.Nonce,
		StorageHash: account.StorageHash.Bytes(),
	}
}

func AccountFromProto(account *Account) *input.Account {
	if account == nil {
		return nil
	}

	return &input.Account{
		Balance:     bytesToBigInt(account.Balance),
		CodeHash:    gethcommon.BytesToHash(account.CodeHash),
		Nonce:       account.Nonce,
		StorageHash: gethcommon.BytesToHash(account.StorageHash),
	}
}
