package main

import "fmt"

const (
	MSPID          = "DemoIdemixMSP"
	UserName       = "User1"
	WalletPath     = "user/wallets/User1-client"
	ServerName     = "peer0.demo.com"
	ServerEndpoint = "localhost:18850"
	NetWork        = "atchain-channel"
	Contract       = "atchain-demo-cc"
)

func main() {
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

	network := user.GwConf.Gateway.GetNetwork(NetWork)
	contract := network.GetContract(Contract)
	res, err := contract.SubmitTransaction("Echo", "324242342")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(res)
}
