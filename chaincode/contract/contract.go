package contract

import (
	"crypto/rand"
	"encoding/json"
	"strconv"

	"errors"
	//"debug/elf"
	"fmt"
	//"github.com/IBM/idemix/bccsp/keystore"
	//"golang.org/x/text/number"
	"math/big"

	"github.com/TomCN0803/atchain-demo/pkg/idemix"
	"github.com/TomCN0803/atchain-demo/pkg/transaction"
	paillier "github.com/TomCN0803/paillier-go"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	//bn "github.com/renzhe666/bn256"
)

type SmartContract struct {
	contractapi.Contract
}

// Echo the argument back as the response
func (s *SmartContract) Echo(meta, arg string)  error {
	res, err := s.checkMetadata(meta)
	if err != nil || !res {
		return  fmt.Errorf("unauthorized transaction, meta check result: %v, error msg: %w", res, err)
	}
    
	return nil
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

func (s *SmartContract) Get(ctx contractapi.TransactionContextInterface, key string) (string,error) {
	
    fmt.Println(key)
	existing, err := ctx.GetStub().GetState(key)
	if err != nil {
		return "", errors.New("Unable to interact with world state")
	}
	if existing == nil {
		return "", fmt.Errorf("Cannot read world state pair with key %s. Does not exist", key)
	}
	fmt.Println(string(existing))
	return string(existing), nil

}

func (s *SmartContract) Insert(ctx contractapi.TransactionContextInterface, meta, key string, value string) error {
	res, err := s.checkMetadata(meta)

	fmt.Println(key)
	fmt.Println(value)

	if err != nil || !res {
		return  fmt.Errorf("unauthorized transaction, meta check result: %v, error msg: %w", res, err)
	}

	err = ctx.GetStub().PutState(key, []byte(value))
	if err != nil {
		return errors.New("Unable to interact with world state")
	}
	return nil
}

func (s *SmartContract)  Compute(ctx contractapi.TransactionContextInterface,publickey string,operation string,label string,keys string,num string,sp string) (string,error){
	scheme := paillier.GetInstance(rand.Reader, 64)
	public := []byte(publickey)
	var publicKey *paillier.PublicKey
	json.Unmarshal(public, &publicKey)

	n, err := strconv.Atoi(num)
	if err!=nil{
		fmt.Println(err)
	}
	key := make([]string,0,n)
	for i:=1;i<=n;i++{
		newStr:=fmt.Sprintf("%03d", i)
		k :=label+newStr+keys
		key = append(key,k)
		fmt.Println(k)
	}

	data := make([]*paillier.PublicValue,0,n)
	for i:=0;i<len(key);i++{
		existing, err := ctx.GetStub().GetState(key[i])
		if err != nil {
			return "", errors.New("Unable to interact with world state")
		}
		if existing == nil {
			return "", fmt.Errorf("Cannot read world state pair with key %s. Does not exist", key[i])
		}
		big1 ,err1:= new(big.Int).SetString(string(existing), 10)
		if err1!=true{
			fmt.Println(err1)
		}
		s :=&paillier.PublicValue{big1}
		data = append(data, s)
	}

	if operation == "add"{
		sum := data[0]
		for i:=1;i<n;i++ {
			sum = scheme.Add(sum, data[i], publicKey)
		}
		Sum:= (sum.Val).String()
		return Sum,nil
	}else if operation == "sub"{
		sum := data[0]
		paa ,err12:= new(big.Int).SetString(sp, 10)
		if err12!=true{
			fmt.Println(err12)
		}
		pa := &paillier.PrivateValue{Val: paa}
		ps := scheme.Encrypt(publicKey, pa)
		for i:=1;i<n;i++ {
			sum = scheme.Sub(sum, data[i], publicKey)
			sum = scheme.Add(sum, ps,      publicKey)
		}
		Sum:= (sum.Val).String()
		return Sum,nil
	} else {
		return "nil",nil
	}

}


func (s *SmartContract)  Mul(ctx contractapi.TransactionContextInterface,publickey string,label string,key string,num string) (string,error){
	scheme := paillier.GetInstance(rand.Reader, 256)
	public := []byte(publickey)
	var publicKey *paillier.PublicKey
	json.Unmarshal(public, &publicKey)
	keys := label + key
	existing, err := ctx.GetStub().GetState(keys)
	if err != nil {
		return "", errors.New("Unable to interact with world state")
	}
	if existing == nil {
		return "", fmt.Errorf("Cannot read world state pair with key %s. Does not exist", key)
	}
	big1 ,err1:= new(big.Int).SetString(string(existing), 10)
	if err1!=true{
		fmt.Println(err1)
	}
	ss :=&paillier.PublicValue{big1}
	number ,err2 := new(big.Int).SetString(num, 10)
	if err2!=true{
		fmt.Println(err2)
	}
	sum := scheme.Mul(ss,number,publicKey)
	Sum:= (sum.Val).String()
	return Sum,nil

}

func (s *SmartContract) TTBE_combine(clue string, cc string,n string) (string,error)   {

	jsonBytess := []byte(cc)
	var c_1 *Cttbe
	json.Unmarshal(jsonBytess, &c_1)
	jsoncoms := []byte(clue)
	var acs_1 []*AudClue
	json.Unmarshal(jsoncoms, &acs_1)

	num, err := strconv.Atoi(n)
	if err != nil {
		fmt.Println(err)
	}
	for i:=uint64(1);i<=uint64(num);i++ {
		acs_1[i-1].index = i

	}

	MRecv, _ := Combine(acs_1, c_1)
	return MRecv.String(),nil
}







