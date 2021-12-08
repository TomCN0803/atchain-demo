package gateway

import (
	"crypto/x509"
	"fmt"

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
