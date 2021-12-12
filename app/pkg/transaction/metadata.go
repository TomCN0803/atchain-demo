package transaction

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
