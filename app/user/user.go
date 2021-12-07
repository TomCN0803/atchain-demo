package main

import (
	"fmt"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"google.golang.org/grpc"
	"io/ioutil"
	"path"

	"github.com/TomCN0803/atchain-demo/app/pkg/gateway"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-gateway/pkg/idemix"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"github.com/hyperledger/fabric-protos-go/msp"
)

type User struct {
	*msp.IdemixMSPSignerConfig

	Identity  *idemix.Identity
	Signer    identity.Sign
	NymSigner identity.Sign
	IssuerPK  []byte
	Gateway   *client.Gateway
	grpcConn  *grpc.ClientConn
}

func NewUser(mspID, walletPath string) (*User, error) {
	confPath := path.Join(walletPath, "user", "SignerConfig")
	ipkPath := path.Join(walletPath, "msp", "IssuerPublicKey")

	signerConf, err := getIdemixSignerConf(confPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create new idemix identity: %w", err)
	}

	issuerPKBytes, err := ioutil.ReadFile(ipkPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create new idemix identity: %w", err)
	}

	user := new(User)
	user.IssuerPK = issuerPKBytes
	user.IdemixMSPSignerConfig = signerConf
	user.Identity = idemix.NewIdemixIdentity(mspID, signerConf.Cred)
	user.Signer = identity.NewIdemixSign(
		signerConf.Sk,
		issuerPKBytes,
		signerConf.Cred,
		signerConf.CredentialRevocationInformation,
	)
	user.NymSigner = identity.NewIdemixNymKeySign(signerConf.Sk, issuerPKBytes)

	return user, nil
}

func (u *User) InitGateway(tlsCertPath, serverName, serverEndpoint string) error {
	connection, err := gateway.NewConnection(tlsCertPath, serverName, serverEndpoint)
	if err != nil {
		return fmt.Errorf("failed to initialize gateway: %w", err)
	}

	u.grpcConn = connection

	gw, err := gateway.NewGateway(u.Identity, u.Signer, connection)
	if err != nil {
		return fmt.Errorf("failed to initialize gateway: %w", err)
	}

	u.Gateway = gw

	return nil
}

func (u *User) CloseGateway() error {
	err := u.grpcConn.Close()
	if err != nil {
		return fmt.Errorf("failed to close gateway: %w", err)
	}

	err = u.Gateway.Close()
	if err != nil {
		return fmt.Errorf("failed to close gateway: %w", err)
	}

	return nil
}

func getIdemixSignerConf(confPath string) (*msp.IdemixMSPSignerConfig, error) {
	signerConfBytes, err := ioutil.ReadFile(confPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read idemix SignerConfig file from %s: %w", confPath, err)
	}

	signerConf := &msp.IdemixMSPSignerConfig{}
	err = proto.Unmarshal(signerConfBytes, signerConf)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal the SignerConfig bytes: %w", err)
	}

	return signerConf, nil
}
