package main

import "fmt"

const (
	CryptoPath     = "../organization/peerOrganizations/demo.com/"
	MSPID          = "DemoIdemixMSP"
	WalletPath     = "user/wallets/john-client"
	ServerName     = "peer0.demo.com"
	ServerEndpoint = "localhost:18850"
	NetWork        = "atchain-channel"
	Contract       = "atchain-demo-cc"
)

var (
	TLSCertPAth = CryptoPath + "peers/peer0.demo.com/tls/ca.crt"
)

func main() {
	user, err := NewUser(MSPID, WalletPath)
	if err != nil {
		panic(err)
	}

	err = user.InitGateway(TLSCertPAth, ServerName, ServerEndpoint)
	if err != nil {
		panic(err)
	}

	defer func() {
		_ = user.CloseGateway()
	}()

	network := user.Gateway.GetNetwork(NetWork)
	contract := network.GetContract(Contract)
	res, err := contract.SubmitTransaction("Echo", "324242342")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(res)
}
