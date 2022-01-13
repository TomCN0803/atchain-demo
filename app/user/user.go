package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	paillier "github.com/TomCN0803/paillier-go"
	"io/ioutil"
	"os/exec"
	"path"
	"strconv"
	"time"

	"github.com/TomCN0803/atchain-demo/app/pkg/gateway"
	"github.com/TomCN0803/atchain-demo/pkg/idemix"
	"github.com/TomCN0803/atchain-demo/pkg/transaction"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-gateway/pkg/client"
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

	GwConf  *userGatewayConf
	Gateway *client.Gateway

	CSP *idemix.CSPWrapper

	IssuerPK     []byte
	RevocationPK []byte
}

type userGatewayConf struct {
	grpcConn *grpc.ClientConn
	gwSigner identity.Sign
	identity identity.Identity
}


func Start(id string)  {
	dir := "./userKeyGen.sh client "+id+" 12344"
	command := exec.Command("sh","-c",dir)
	outinfo := bytes.Buffer{}
	command.Stdout = &outinfo
	command.Dir = "user"
	err := command.Start()
	if err != nil{
		fmt.Println(err.Error())
	}
	if err = command.Wait();err!=nil{
		fmt.Println(err.Error())
	}else{
		//fmt.Println(command.ProcessState.Pid())
		//fmt.Println(outinfo.String())
	}

}

func New(id string) (*User,*client.Contract) {//555

	UserNames := id
	WalletPaths :="user/wallets/"+id+"-client"
	//fmt.Println(WalletPaths)

	user, err2 := NewUser(MSPID, UserNames, WalletPaths)
	if err2 != nil {
		panic(err2)
	}

	err := user.InitGateway(ServerName, ServerEndpoint)
	if err != nil {
		panic(err)
	}

	network := user.Gateway.GetNetwork(NetWork)
	contract := network.GetContract(Contract)
	return user,contract
}

func NewUser(mspID, name, walletPath string) (*User, error) {
	confPath := path.Join(walletPath, "user", "SignerConfig")
	ipkPath := path.Join(walletPath, "msp", "IssuerPublicKey")
	revPKPath := path.Join(walletPath, "msp", "RevocationPublicKey")

	signerConf, err := getIdemixSignerConf(confPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create new user: %w", err)
	}

	issuerPKBytes, err := ioutil.ReadFile(ipkPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create new user: %w", err)
	}

	revPKBytes, err := ioutil.ReadFile(revPKPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create new user: %w", err)
	}

	user := &User{
		IdemixMSPSignerConfig: signerConf,
		MSPID:                 mspID,
		Name:                  name,
		WalletPath:            walletPath,
		IssuerPK:              issuerPKBytes,
		RevocationPK:          revPKBytes,
	}

	user.CSP, err = idemix.NewIdemixCSP()
	if err != nil {
		return nil, fmt.Errorf("failed to create new user: %w", err)
	}

	return user, nil
}

func (u *User) InitGateway(serverName, serverEndpoint string) error {
	tlsCertPath := path.Join(u.WalletPath, "conn", "tls", "ca.crt")
	signcertName := u.Name + "@" + DomainName + "-cert.pem"
	signcertPath := path.Join(u.WalletPath, "conn", "msp", "signcerts", signcertName)
	keyPath := path.Join(u.WalletPath, "conn", "msp", "keystore", "key.pem")

	connection, err := gateway.NewConnection(tlsCertPath, serverName, serverEndpoint)
	if err != nil {
		return fmt.Errorf("failed to initialize gateway: %w", err)
	}

	gwConf := new(userGatewayConf)

	gwConf.grpcConn = connection

	id, err := gateway.NewIdentity(u.MSPID, signcertPath)
	if err != nil {
		return fmt.Errorf("failed to initialize gateway: %w", err)
	}

	gwConf.identity = id

	signer, err := gateway.NewSigner(keyPath)
	if err != nil {
		return fmt.Errorf("failed to initialize gateway: %w", err)
	}

	gwConf.gwSigner = signer
	u.GwConf = gwConf

	gw, err := gateway.NewGateway(id, signer, connection)
	if err != nil {
		return fmt.Errorf("failed to initialize gateway: %w", err)
	}

	u.Gateway = gw

	return nil
}

func (u *User) CloseGateway() error {
	err := u.Gateway.Close()
	if err != nil {
		return fmt.Errorf("failed to close gateway: %w", err)
	}

	err = u.GwConf.grpcConn.Close()
	if err != nil {
		return fmt.Errorf("failed to close gateway: %w", err)
	}

	return nil
}

func (u *User) EvaluateTransaction(contract *client.Contract, name string, args ...string) ([]byte, error) {//1111
	
	res, err := contract.EvaluateTransaction(name, args...)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to evaluate transaction %s.%s: %w",
			contract.ChaincodeName(),
			name,
			err,
		)
	}

	return res, nil
}

func (u *User) SubmitTransaction(contract *client.Contract, name string, ID string,args ...string) error {///111

	err := u.prepareTransMeta(name,ID,&args)
	if err != nil {
		return  fmt.Errorf(
			"failed to submit transaction %s.%s: %w",
			contract.ChaincodeName(),
			name,
			err,
		)
	}
	

	r, err1 := contract.SubmitTransaction(name, args...)
	if err1 != nil {
		return fmt.Errorf(
			"failed to submit transaction %s.%s: %w",
			contract.ChaincodeName(),
			name,
			err1,
		)
	}
	if r != nil{
		fmt.Println(r)
	}

	return  nil
}

func (u *User) prepareTransMeta(name string, ID string,argsPtr *[]string) error {//1111
	nymSK, nymPK, err := u.CSP.DeriveNymKeyPair(u.Sk, u.IssuerPK)
	if err != nil {
		return fmt.Errorf("failed to generate transaction metadata: %w", err)
	}

	timestamp := time.Now().UnixNano()
	txDigest := strconv.Itoa(int(timestamp)) + name + string(nymPK)
	txDigestHash := sha256.Sum256([]byte(txDigest))

	sig, err := u.CSP.Sign(u.Sk, nymSK, nymPK, u.IssuerPK, u.Cred, u.CredentialRevocationInformation, txDigestHash[:])
	if err != nil {
		return fmt.Errorf("failed to generate transaction metadata: %w", err)
	}

	nymSig, err := u.CSP.NymSign(u.Sk, nymSK, nymPK, u.IssuerPK, txDigestHash[:])
	if err != nil {
		return fmt.Errorf("failed to generate transaction metadata: %w", err)
	}
	cctbe := Temencrypt(ID,nymPK)

	meta := &transaction.Metadata{
		Cttbe:        cctbe,
		Sig:          sig,
		NymSig:       nymSig,
		Digest:       txDigestHash[:],
		OU:           u.OrganizationalUnitIdentifier,
		Role:         int(u.Role),
		NymPK:        nymPK,
		IssuerPK:     u.IssuerPK,
		RevocationPK: u.RevocationPK,
	}

	metaBytes, err := meta.Serialize()
	if err != nil {
		return fmt.Errorf("failed to generate transaction metadata: %w", err)
	}

	*argsPtr = append(*argsPtr, "")
	copy((*argsPtr)[1:], *argsPtr)
	(*argsPtr)[0] = string(metaBytes)

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

func (u *User) GetTransaction(contract *client.Contract, name string, args ...string) string {

	

	res, err := contract.EvaluateTransaction(name, args[0])
	if err != nil {
		fmt.Println(err)
		return  "-1"
	}
    //fmt.Println(string(res))
	return string(res)
}
func (u *User) ComputContract(contract *client.Contract, name string,publicKey *paillier.PublicKey,privateKey *paillier.PrivateKey,operation string,lable string,key string,n string,p string) string {

	jsonBytes, err2 := json.Marshal(publicKey)
	if err2 != nil {
		fmt.Println(err2)
	}
	fmt.Println(string(jsonBytes))

	if operation == "sub"{
		res, err1 := contract.EvaluateTransaction(name, string(jsonBytes), operation, lable, key, n, p)
		if err1 != nil {
			fmt.Println(err1)
			return "-1"
		}
		n1, err := strconv.Atoi(key)
		if err!=nil{
			fmt.Println(err)
		}
		newStr:=fmt.Sprintf("%03d", n1)
		lable_1 := lable + newStr
		Insertdata(string(res),lable_1,privateKey)
		return string(res)
	}else {
		res, err1 := contract.EvaluateTransaction(name, string(jsonBytes), operation, lable, key, n, p)
		if err1 != nil {
			fmt.Println(err1)
			return "-1"
		}
		n1, err := strconv.Atoi(key)
		if err!=nil{
			fmt.Println(err)
		}
		newStr:=fmt.Sprintf("%03d", n1)
		lable_1 := lable + newStr
		Insertdata(string(res),lable_1,privateKey)
		return string(res)

	}
}
