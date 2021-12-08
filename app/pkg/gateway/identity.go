package gateway

import (
	"fmt"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
)

func NewIdentity(mspID, certPath string) (identity.Identity, error) {
	cert, err := LoadCert(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create new identity: %w", err)
	}

	id, err := identity.NewX509Identity(mspID, cert)
	if err != nil {
		return nil, fmt.Errorf("failed to create new identity: %w", err)
	}

	return id, nil
}
