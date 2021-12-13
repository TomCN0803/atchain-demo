package contract

import (
	"fmt"

	"github.com/TomCN0803/atchain-demo/pkg/idemix"
	"github.com/TomCN0803/atchain-demo/pkg/transaction"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

// Echo the argument back as the response
func (s *SmartContract) Echo(meta, arg string) (string, error) {
	res, err := s.checkMetadata(meta)
	if err != nil || !res {
		return "", fmt.Errorf("unauthorized transaction, meta check result: %v, error msg: %w", res, err)
	}

	return fmt.Sprintf("meta check result: %v ", res) + arg, nil
}

func (s *SmartContract) checkMetadata(meta string) (bool, error) {
	metaBytes := []byte(meta)

	metadata := new(transaction.Metadata)
	err := metadata.Deserialize(metaBytes)
	if err != nil {
		return false, fmt.Errorf("failed to check transaction metadata: %w", err)
	}

	csp, err := idemix.NewIdemixCSP()
	if err != nil {
		return false, fmt.Errorf("failed to check transaction metadata: %w", err)
	}

	r1, err := csp.VerifyNymSig(
		metadata.NymPK,
		metadata.IssuerPK,
		metadata.NymSig,
		metadata.Digest,
	)
	if err != nil {
		return false, fmt.Errorf("failed to check transaction metadata: %w", err)
	}

	r2, err := csp.VerifySig(
		metadata.OU,
		metadata.Role,
		metadata.IssuerPK,
		metadata.RevocationPK,
		metadata.Sig,
		metadata.Digest,
	)
	if err != nil {
		return false, fmt.Errorf("failed to check transaction metadata: %w", err)
	}

	return r1 && r2, nil
}
