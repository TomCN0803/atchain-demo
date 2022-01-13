package main

import (
	"bufio"
	"crypto/rand"
	//"encoding/json"
	"fmt"
	paillier "github.com/TomCN0803/paillier-go"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	

	//"math/big"
	"os"
	"strconv"
	"strings"
	//bn "github.com/renzhe666/bn256"
)

var db *sqlx.DB

//初始化数据库连接，init()方法系统会在动在main方法之前执行。
func init() {
	database, err := sqlx.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/mytest")
	if err != nil {
		fmt.Println("open mysql failed,", err)
	}
	db = database
}


var scheme paillier.PaillierScheme
var tpk TPK

const (
	MSPID          = "DemoMSP"
	UserName       = "User1"
	WalletPath     = "user/wallets/User1-client"
	ServerName     = "peer0.demo.com"
	ServerEndpoint = "localhost:18850"
	NetWork        = "atchain-channel"
	Contract       = "atchain-demo-cc"
)

//const (
//	MSPID1          = "DemoMSP"
//	UserName1       = "User2"
//	WalletPath1     = "user/wallets/User2-client"
//)
//
//const (
//	MSPID2          = "DemoMSP"
//	UserName2       = "User3"
//	WalletPath2     = "user/wallets/User3-client"
//)


func main() {
	Deletesql()
	user_number := 0 //333
	audituser := 5 
	tpk1, tsks, tvks := SetUp(uint64(audituser), uint64(audituser))
	tpk = tpk1
	p, err := rand.Prime(rand.Reader, 128)
	if err != nil {
		panic(err)
	}

	scheme = paillier.GetInstance(rand.Reader, 128)

	fmt.Println("欢迎使用本系统,选项如下：")
	fmt.Println("create -name：创建一个新账号")
	fmt.Println("launch -id -label -identifier -number -datafomat -user1...:根据协商的编号、身份号和交易总数发起交易")
	fmt.Println("operate -id -label -identifier -number -operation:根据协商的编号、身份号和交易总数，对数据进行安全多方计算")
	fmt.Println("gain -id -label -number：获得计算结果")
	fmt.Println("query -id -label：根据交易标号查询交易")

	fmt.Println("audit -id ：审定指定交易中的用户")//222

	fmt.Println("end:结束")
	//user, err2 := NewUser(MSPID, UserName, WalletPath)
	//if err2 != nil {
	//	panic(err2)
	//}
	//
	//err = user.InitGateway(ServerName, ServerEndpoint)
	//if err != nil {
	//	panic(err)
	//}
	//
	//defer func() {
	//	_ = user.CloseGateway()
	//}()
	//network := user.Gateway.GetNetwork(NetWork)
	//contract := network.GetContract(Contract)

	for{
		reader := bufio.NewReader(os.Stdin)
		msg, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err.Error())
		}
		context := strings.Fields(msg)
		if len(context)>0 {
			if context[0] == "end" {
				break
			} else if context[0] == "create" {//333
				user_number++
				usernb := strconv.Itoa(user_number)
				id := "User"+usernb
				fmt.Printf("id:%s\n",id)

				Start(id)


				scheme1 := paillier.GetInstance(rand.Reader, 128)
				privateKey := scheme1.GenKeypair()
				publicKey := privateKey.PublicKey
				Insertsql(id,context[1],publicKey,privateKey)
				Insertpublic(id,context[1],publicKey)
			} else if context[0] == "launch" {

				context[5] = Datastring(context[5])

				user,contract :=New(context[1])
				defer func() {
					_ = user.CloseGateway()
				}()

				num, err := strconv.Atoi(context[4])
				if err!=nil{
					fmt.Println(err)
				}
				ident, err1 := strconv.Atoi(context[3])
				if err1!=nil{
					fmt.Println(err1)
				}
				publics := make([]*paillier.PublicKey,0,num)
				publickey :=Getpublic(context[1])

				j :=0
				for i:=0;i<num;i++{
					if i != (ident-1){
						publickeys := Getpublic(context[6+j])
						publics = append(publics,publickeys)
						j = j+1
					}else {
						publics = append(publics,publickey)
					}
				}
				key := make([]string,0,num)
				identstr:=fmt.Sprintf("%03d", ident)
				for i:=1;i<=num;i++{
					newStr:=fmt.Sprintf("%03d", i)
					k :=context[2]+identstr+newStr
					key = append(key,k)
				}
				data, err3 := strconv.ParseInt(context[5], 10, 64) 
				if err3!=nil{
					fmt.Println(err3)
				}
				Share_f(user,contract,context[1],data,num,p,publics,key)
			} else if context[0] == "operate"{
				user,contract :=New(context[1])
				defer func() {
					_ = user.CloseGateway()
				}()

				ident, err1 := strconv.Atoi(context[3])
				if err1!=nil{
					fmt.Println(err1)
				}
				identstr:=fmt.Sprintf("%03d", ident)
				publicKey,privateKey :=Getsql(context[1])
				number := user.ComputContract(contract,"Compute",publicKey,privateKey,context[5],context[2],identstr,context[4],p.String())
				fmt.Println(number)

			} else if context[0] == "gain" {
				sum := Getdata(context[2],context[3],p)
				fmt.Println(sum)
			} else if context[0] == "query" {
				user,contract :=New(context[1])
				defer func() {
					_ = user.CloseGateway()
				}()

				result :=user.GetTransaction(contract, "Get",context[2])
				if err != nil {
					fmt.Println(err)
					fmt.Println("-1")
				}
				fmt.Println(result)

  			}else if context[0] == "audit"{//2222333
				user,contract :=New(context[1])
				defer func() {
					_ = user.CloseGateway()
				}()
                fmt.Println("请输入需要审计的交易的标签：")
				tag := Ttbestring()
				fmt.Println("请输入需要审计的交易的密文")
				label := Tagstring()

				name ,id := Audit(contract,tag,label,tsks,tvks,user_number,audituser)
				fmt.Printf("该可疑交易发起方name=%s,id=%s\n",name,id)

				

			}else {
				break
			}
		}else {
			break
		}

	}
}
