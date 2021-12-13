package idemix

import (
	"crypto/sha256"
	"io/ioutil"
	"log"
	"path"
	"strconv"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/msp"
)

const (
	SignerConfigPath = "test/wallet/User1-client/user/SignerConfig"
	IssuerPubKeyPath = "test/wallet/User1-client/msp"
)

func TestNymKey(t *testing.T) {
	csp, err := NewIdemixCSP()
	if err != nil {
		log.Fatalln(err)
	}

	signer, err := getSignerConf()
	if err != nil {
		log.Fatalln(err)
	}

	issuerPK, err := ioutil.ReadFile(path.Join(IssuerPubKeyPath, "IssuerPublicKey"))
	if err != nil {
		log.Fatalln(err)
	}

	nymSK, nymPK, err := csp.DeriveNymKeyPair(signer.Sk, issuerPK)
	if err != nil {
		log.Fatalln(err)
	}

	nymSKIns, err := csp.importUserNymSK(nymSK, nymPK)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(nymSKIns)
}

func TestCSP(t *testing.T) {
	csp, err := NewIdemixCSP()
	if err != nil {
		log.Fatalln(err)
	}

	signer, err := getSignerConf()
	if err != nil {
		log.Fatalln(err)
	}

	issuerPK, err := ioutil.ReadFile(path.Join(IssuerPubKeyPath, "IssuerPublicKey"))
	if err != nil {
		log.Fatalln(err)
	}

	revocPK, err := ioutil.ReadFile(path.Join(IssuerPubKeyPath, "RevocationPublicKey"))
	if err != nil {
		log.Fatalln(err)
	}

	nymSK, nymPK, err := csp.DeriveNymKeyPair(signer.Sk, issuerPK)
	if err != nil {
		log.Fatalln(err)
	}

	timestamp := time.Now().UnixNano()
	dig := strconv.Itoa(int(timestamp)) + string(nymPK)
	digHash := sha256.Sum256([]byte(dig))

	sig, err := csp.Sign(
		signer.Sk,
		nymSK,
		nymPK,
		issuerPK,
		signer.Cred,
		signer.CredentialRevocationInformation,
		digHash[:],
	)
	if err != nil {
		log.Fatalln(err)
	}

	nymSig, err := csp.NymSign(signer.Sk, nymSK, nymPK, issuerPK, digHash[:])
	if err != nil {
		log.Fatalln(err)
	}

	r1, err := csp.VerifySig(
		signer.OrganizationalUnitIdentifier,
		int(signer.Role),
		issuerPK,
		revocPK,
		sig,
		digHash[:],
	)
	if err != nil {
		log.Fatalln(err)
	}

	r2, err := csp.VerifyNymSig(nymPK, issuerPK, nymSig, digHash[:])
	if err != nil {
		log.Fatalln(err)
	}

	if !r1 || !r2 {
		log.Fatalf("VerifySig: %v\tVerifyNymSig: %v\n", r1, r2)
	}

	log.Println("succeed")
}

func getSignerConf() (*msp.IdemixMSPSignerConfig, error) {
	signConfBytes, err := ioutil.ReadFile(SignerConfigPath)
	if err != nil {
		return nil, err
	}

	signerConf := new(msp.IdemixMSPSignerConfig)
	err = proto.Unmarshal(signConfBytes, signerConf)
	if err != nil {
		return nil, err
	}

	return signerConf, err
}
