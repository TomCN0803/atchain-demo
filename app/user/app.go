package main

import (
	"crypto/rand"
	"fmt"
	"github.com/olegfomenko/paillier"
	
)


var scheme paillier.PaillierScheme

const (
	MSPID          = "DemoMSP"
	UserName       = "User1"
	WalletPath     = "user/wallets/User1-client"
	ServerName     = "peer0.demo.com"
	ServerEndpoint = "localhost:18850"
	NetWork        = "atchain-channel"
	Contract       = "atchain-demo-cc"
)

const (
	MSPID1          = "DemoMSP"
	UserName1       = "User2"
	WalletPath1     = "user/wallets/User2-client"
)

const (
	MSPID2          = "DemoMSP"
	UserName2       = "User3"
	WalletPath2     = "user/wallets/User3-client"
)


func main() {
	p, err := rand.Prime(rand.Reader, 256)
	if err != nil {
		panic(err)
	}

	scheme = paillier.GetInstance(rand.Reader, 256)
	privateKey := scheme.GenKeypair()
	publicKey := privateKey.PublicKey
	privateKey_1 := scheme.GenKeypair()
	publicKey_1 := privateKey_1.PublicKey
	privateKey_2 := scheme.GenKeypair()
	publicKey_2 := privateKey_2.PublicKey

	public := make([]*paillier.PublicKey,0,2)
	public =append(public,publicKey)
	public =append(public,publicKey_1)
	public =append(public,publicKey_2)

	private := make([]*paillier.PrivateKey,0,2)
	private =append(private,privateKey)
	private =append(private,privateKey_1)
	private =append(private,privateKey_2)


	user, err := NewUser(MSPID, UserName, WalletPath)
	if err != nil {
		panic(err)
	}

	err = user.InitGateway(ServerName, ServerEndpoint)
	if err != nil {
		panic(err)
	}

	defer func() {
		_ = user.CloseGateway()
	}()

	network := user.Gateway.GetNetwork(NetWork)
	contract := network.GetContract(Contract)
    key := make([]string,0,3)
	key = append(key,"11")
	key = append(key,"12")
	key = append(key,"13")
	Share_f(user,contract,450,3,p,public,key)


	user_1, err := NewUser(MSPID1, UserName1, WalletPath1)
	if err != nil {
		panic(err)
	}

	err = user_1.InitGateway(ServerName, ServerEndpoint)
	if err != nil {
		panic(err)
	}

	defer func() {
		_ = user_1.CloseGateway()
	}()

	network_1 := user_1.Gateway.GetNetwork(NetWork)
	contract_1 := network_1.GetContract(Contract)
	key_1 := make([]string,0,3)
	key_1 = append(key_1,"21")
	key_1 = append(key_1,"22")
	key_1 = append(key_1,"23")
	Share_f(user_1,contract_1,120,3,p,public,key_1)

   
    user_2, err := NewUser(MSPID2, UserName2, WalletPath2)
	if err != nil {
		panic(err)
	}

	err = user_2.InitGateway(ServerName, ServerEndpoint)
	if err != nil {
		panic(err)
	}

	defer func() {
		_ = user_2.CloseGateway()
	}()

	network_2 := user_2.Gateway.GetNetwork(NetWork)
	contract_2 := network_2.GetContract(Contract)
    key_2 := make([]string,0,3)
	key_2 = append(key_2,"31")
	key_2 = append(key_2,"32")
    key_2 = append(key_2,"33")
	Share_f(user_2,contract_2,30,3,p,public,key_2) 



	s_11 :=user.GetTransaction(contract, "Get", "11")
	s_12 :=user.GetTransaction(contract, "Get", "12")
	s_13 :=user.GetTransaction(contract, "Get", "13")


	s_21 :=user_1.GetTransaction(contract_1, "Get", "21")
	s_22 :=user_1.GetTransaction(contract_1, "Get", "22")
	s_23 :=user_1.GetTransaction(contract_1, "Get", "23")

	s_31 :=user_2.GetTransaction(contract_2, "Get", "31")
	s_32 :=user_2.GetTransaction(contract_2, "Get", "32")
	s_33 :=user_2.GetTransaction(contract_2, "Get", "33")


	number := Add(p,publicKey,privateKey,s_11,s_21,s_31)
	number_1 := Add(p,publicKey,privateKey_1,s_12,s_22,s_32)
	number_2 := Add(p,publicKey,privateKey_2,s_13,s_23,s_33)

	shares_3 := make([]string,0,3)
	shares_3 =append(shares_3,number)
	shares_3 =append(shares_3,number_1)
	shares_3 =append(shares_3,number_2)



	ss := Reconstruct_f(shares_3,3,p,private)
	fmt.Println(ss)
}
