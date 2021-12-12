package transaction

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

type Metadata struct {
	Cttbe  []byte
	Sig    []byte
	NymSig []byte

	Digest       []byte
	OU           string
	Role         int
	NymPK        []byte
	IssuerPK     []byte
	RevocationPK []byte
}

func (m *Metadata) Serialize() ([]byte, error) {
	metaBuffer := new(bytes.Buffer)
	enc := gob.NewEncoder(metaBuffer)
	err := enc.Encode(m)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize metadata: %w", err)
	}

	return metaBuffer.Bytes(), nil
}

func (m *Metadata) Deserialize(metaBytes []byte) error {
	metaBuffer := bytes.NewBuffer(metaBytes)
	dec := gob.NewDecoder(metaBuffer)
	err := dec.Decode(m)
	if err != nil {
		return fmt.Errorf("failed to deserialize metadata: %w", err)
	}

	return nil
}
