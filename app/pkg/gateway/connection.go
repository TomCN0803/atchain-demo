package gateway

import (
	"crypto/x509"
	"fmt"
	"io/ioutil"

	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func NewConnection(tlsCertPath, serverName, serverEndpoint string) (*grpc.ClientConn, error) {
	errln := "failed to initiate a new connection: "

	tlsCert, err := LoadCert(tlsCertPath)
	if err != nil {
		return nil, fmt.Errorf(errln + err.Error())
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(tlsCert)
	transCreds := credentials.NewClientTLSFromCert(certPool, serverName)

	connection, err := grpc.Dial(serverEndpoint, grpc.WithTransportCredentials(transCreds))
	if err != nil {
		return nil, fmt.Errorf(errln + err.Error())
	}

	return connection, nil
}

func LoadCert(certPath string) (*x509.Certificate, error) {
	certPem, err := ioutil.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("cannot retrieve certificate: %w", err)
	}

	return identity.CertificateFromPEM(certPem)
}
