package proto

import (
	input "github.com/kkrt-labs/zk-pig/src/prover-input"
)

// ToProto converts Go input.ProverInput to protobuf format
func ToProto(pi *input.ProverInput) *ProverInput {
	if pi == nil {
		return nil
	}

	return &ProverInput{
		Version:     pi.Version,
		Blocks:      BlocksToProto(pi.Blocks),
		Witness:     WitnessToProto(pi.Witness),
		ChainConfig: ChainConfigToProto(pi.ChainConfig),
	}
}

func FromProto(pi *ProverInput) *input.ProverInput {
	if pi == nil {
		return nil
	}

	return &input.ProverInput{
		Version:     pi.Version,
		Blocks:      BlocksFromProto(pi.Blocks),
		Witness:     WitnessFromProto(pi.Witness),
		ChainConfig: ChainConfigFromProto(pi.ChainConfig),
	}
}

func WitnessToProto(w *input.Witness) *Witness {
	if w == nil {
		return nil
	}

	return &Witness{
		Ancestors: HeadersToProto(w.Ancestors),
		State:     hexBytesToBytes(w.State),
		Codes:     hexBytesToBytes(w.Codes),
	}
}

func WitnessFromProto(w *Witness) *input.Witness {
	if w == nil {
		return nil
	}

	return &input.Witness{
		Ancestors: HeadersFromProto(w.Ancestors),
		State:     bytesToHexutil(w.State),
		Codes:     bytesToHexutil(w.Codes),
	}
}
