package main

import (
	"fmt"
	"io/ioutil"
	"path"

	schemes "github.com/IBM/idemix/bccsp/schemes"
	"github.com/TomCN0803/atchain-demo/app/pkg/gateway"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/idemix"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"github.com/hyperledger/fabric-protos-go/msp"
	"google.golang.org/grpc"
)

const DomainName = "demo.com"

type User struct {
	*msp.IdemixMSPSignerConfig

	MSPID      string
	Name       string
	WalletPath string
	GwConf     *UserGatewayConf

	CSP      *idemix.CSPWrapper
	IssuerPK schemes.Key
	IdxSK    schemes.Key
	IdxNymSK schemes.Key
}

type UserGatewayConf struct {
	Gateway  *client.Gateway
	grpcConn *grpc.ClientConn
	GwSigner identity.Sign
	Identity identity.Identity
}

func NewUser(mspID, name, walletPath string) (*User, error) {
	confPath := path.Join(walletPath, "user", "SignerConfig")
	ipkPath := path.Join(walletPath, "msp", "IssuerPublicKey")

	signerConf, err := getIdemixSignerConf(confPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create new user: %w", err)
	}

	issuerPKBytes, err := ioutil.ReadFile(ipkPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create new user: %w", err)
	}

	user := new(User)
	user.IdemixMSPSignerConfig = signerConf
	user.MSPID = mspID
	user.Name = name
	user.WalletPath = walletPath

	user.CSP, err = idemix.NewIdemixCSP()
	if err != nil {
		return nil, fmt.Errorf("failed to create new user: %w", err)
	}

	user.IssuerPK, err = user.CSP.GetIssuerPK(issuerPKBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to create new user: %w", err)
	}

	user.IdxSK, err = user.CSP.GetUserSK(signerConf.Sk)
	if err != nil {
		return nil, fmt.Errorf("failed to create new user: %w", err)
	}

	return user, nil
}

func (u *User) InitGateway(serverName, serverEndpoint string) error {
	tlsCertPath := path.Join(u.WalletPath, "conn", "tls", "ca.crt")
	signcertName := u.Name + "@" + DomainName + "-cert.pem"
	signcertPath := path.Join(u.WalletPath, "conn", "msp", "signcerts", signcertName)
	keyPath := path.Join(u.WalletPath, "conn", "msp", "keystore", "priv_sk")

	connection, err := gateway.NewConnection(tlsCertPath, serverName, serverEndpoint)
	if err != nil {
		return fmt.Errorf("failed to initialize gateway: %w", err)
	}

	u.GwConf.grpcConn = connection

	id, err := gateway.NewIdentity(u.MSPID, signcertPath)
	if err != nil {
		return fmt.Errorf("failed to initialize gateway: %w", err)
	}

	u.GwConf.Identity = id

	signer, err := gateway.NewSigner(keyPath)
	if err != nil {
		return fmt.Errorf("failed to initialize gateway: %w", err)
	}

	u.GwConf.GwSigner = signer

	gw, err := gateway.NewGateway(id, signer, connection)
	if err != nil {
		return fmt.Errorf("failed to initialize gateway: %w", err)
	}

	u.GwConf.Gateway = gw

	return nil
}

func (u *User) CloseGateway() error {
	err := u.GwConf.grpcConn.Close()
	if err != nil {
		return fmt.Errorf("failed to close gateway: %w", err)
	}

	err = u.GwConf.Gateway.Close()
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
