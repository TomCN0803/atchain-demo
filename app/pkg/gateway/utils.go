package gateway

import (
	"crypto/x509"
	"fmt"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"io/ioutil"
)

func LoadCert(certPath string) (*x509.Certificate, error) {
	certPem, err := ioutil.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("cannot retrieve certificate: %w", err)
	}

	return identity.CertificateFromPEM(certPem)
}
