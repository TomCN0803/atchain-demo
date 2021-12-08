package gateway

import (
	"fmt"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"io/ioutil"
)

func NewSigner(keyPath string) (identity.Sign, error) {
	skPEM, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize a new signer: %w", err)
	}

	sk, err := identity.CertificateFromPEM(skPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize a new signer: %w", err)
	}

	signer, err := identity.NewPrivateKeySign(sk)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize a new signer: %w", err)
	}

	return signer, nil
}
