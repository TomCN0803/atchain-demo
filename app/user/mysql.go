package main

import (
	//"crypto/rand"
	"encoding/json"
	"fmt"
	paillier "github.com/TomCN0803/paillier-go"
	_ "github.com/go-sql-driver/mysql"
	"math/big"
	"strconv"
	//"github.com/jmoiron/sqlx"
)
type Users struct {
	UserId   string    `db:"user_id"`
	Username string `db:"username"`
	Userpublic string `db:"public"`
	Userprivate string `db:"private"`
}

type Data struct {
	Number   string    `db:"number"`
	Sum      string    `db:"sum"`

}
type Userpublic struct {
	UserId   string    `db:"user_id"`
	Username string `db:"username"`
	Userpublic string `db:"public"`
}

//数据库指针
//var db *sqlx.DB

func Insertsql(id string,name string,publicKey *paillier.PublicKey,privateKey *paillier.PrivateKey){
	jsonpublic, err := json.Marshal(publicKey)
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Println(string(jsonBytes))
	pu := string(jsonpublic)
	fmt.Println(len(pu))
	//var people paillier.PublicKey
	//json.Unmarshal([]byte(sss), &people)
	//fmt.Println(people)

	jsonprivate, err1 := json.Marshal(privateKey)
	if err1 != nil {
		fmt.Println(err1)
	}
	//fmt.Println(string(jsonBytess))
	pr := string(jsonprivate)
	fmt.Println(len(pr))
	sql := "insert into users(user_id,username,public,private)values (?,?,?,?)"
	value := [4]string{id,name, pu, pr}
	//执行SQL语句
	r, err2 := db.Exec(sql, value[0], value[1], value[2],value[3])
	if err2 != nil {
		fmt.Println("exec failed,", err2)
		return
	}
	//查询最后一天用户ID，判断是否插入成功
	ids, err3 := r.LastInsertId()
	if err3 != nil {
		fmt.Println("exec failed,", err3)
		return
	}
	fmt.Println("insert succ", ids)
}

func Getsql(id string) (*paillier.PublicKey,*paillier.PrivateKey) {
	var users []Users
	sql := "select user_id, username,public,private from users where user_id=? "
	err := db.Select(&users, sql, id)
	if err != nil {
		fmt.Println("exec failed, ", err)
		return nil,nil
	}
	fmt.Println("select succ:", users)

	var publickey paillier.PublicKey
	json.Unmarshal([]byte(users[0].Userpublic), &publickey)
	//fmt.Println(publickey)

	var privatekey paillier.PrivateKey
	json.Unmarshal([]byte(users[0].Userprivate), &privatekey)
	//fmt.Println(privatekey)
	return &publickey,&privatekey
}

func Insertdata(res string,label string,privateKey *paillier.PrivateKey){
	big2 ,err:= new(big.Int).SetString(string(res), 10)
	if err!=true{
		fmt.Println(err)
	}
	s1 :=&paillier.PublicValue{big2}
	ss := scheme.Decrypt(privateKey, s1).Val
	a := ss.String()
	sql := "insert into data(number,sum)values (?,?)"
	value := [2]string{label,a}
	//执行SQL语句
	r, err2 := db.Exec(sql, value[0], value[1])
	if err2 != nil {
		fmt.Println("exec failed,", err2)
		return
	}
	//查询最后一天用户ID，判断是否插入成功
	ids, err3 := r.LastInsertId()
	if err3 != nil {
		fmt.Println("exec failed,", err3)
		return
	}
	fmt.Println("insert succ", ids)
}
func Getdata(label string,n string,p *big.Int) *big.Int {
	t, err := strconv.Atoi(n)
	if err!=nil{
		fmt.Println(err)
	}
	shares := make([]string,0,t)
	for i:=1;i<=t;i++ {
		newStr:=fmt.Sprintf("%03d", i)
		lable_1 := label + newStr
		var datas []Data
		sql := "select number,sum from data where number =? "
		err := db.Select(&datas, sql, lable_1)
		if err != nil {
			fmt.Println("exec failed, ", err)
			return nil
		}
		fmt.Println("select succ:", datas[0])
		shares = append(shares,datas[0].Sum)
	}
	s := Reconstruct_fp(shares,t,p)

    number1 := s.String()//转成string
	num1, err := strconv.Atoi(number1)//string转int

	number2 := p.String()//转成string
	num2, err := strconv.Atoi(number2)//string转int

	p1 := num2/2
	if num1 > p1{
		s.Sub(s,p)
	}


	return s


}


func Insertpublic(id string,name string,publicKey *paillier.PublicKey){
	jsonpublic, err := json.Marshal(publicKey)
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Println(string(jsonBytes))
	pu := string(jsonpublic)
	fmt.Println(len(pu))
	//var people paillier.PublicKey
	//json.Unmarshal([]byte(sss), &people)
	//fmt.Println(people)

	sql := "insert into publics(user_id,username,public)values (?,?,?)"
	value := [3]string{id,name, pu}
	//执行SQL语句
	r, err2 := db.Exec(sql, value[0], value[1], value[2])
	if err2 != nil {
		fmt.Println("exec failed,", err2)
		return
	}
	//查询最后一天用户ID，判断是否插入成功
	ids, err3 := r.LastInsertId()
	if err3 != nil {
		fmt.Println("exec failed,", err3)
		return
	}
	fmt.Println("insert succ", ids)
}
func Getpublic(id string) (*paillier.PublicKey) {
	var userp []Userpublic
	sql := "select user_id, username,public from publics where user_id=? "
	err := db.Select(&userp, sql, id)
	if err != nil {
		fmt.Println("exec failed, ", err)
		return nil
	}
	fmt.Println("select succ:", userp)

	var publickey paillier.PublicKey
	json.Unmarshal([]byte(userp[0].Userpublic), &publickey)
	//fmt.Println(publickey)

	return &publickey
}
func Find(id string) (string) {///3333
	var userp []Userpublic
	sql := "select user_id, username,public from publics where user_id=? "
	err := db.Select(&userp, sql, id)
	if err != nil {
		fmt.Println("exec failed, ", err)
		return "nil"
	}
	fmt.Println("select succ:", userp)

	var publickey paillier.PublicKey
	json.Unmarshal([]byte(userp[0].Userpublic), &publickey)
	//fmt.Println(publickey)

	return userp[0].Username
}

func Deletesql()  {
	sql := "delete from publics"

	res, err := db.Exec(sql)
	if err != nil {
		fmt.Println("exce failed,", err)
		return
	}

	row, err := res.RowsAffected()
	if err != nil {
		fmt.Println("row failed, ", err)
	}
	fmt.Println("delete succ: ", row)

	sql1 := "delete from users"

	res1, err := db.Exec(sql1)
	if err != nil {
		fmt.Println("exce failed,", err)
		return
	}

	row1, err := res1.RowsAffected()
	if err != nil {
		fmt.Println("row failed, ", err)
	}
	fmt.Println("delete succ: ", row1)

	sql2 := "delete from data "

	res2, err := db.Exec(sql2)
	if err != nil {
		fmt.Println("exce failed,", err)
		return
	}

	_, err = res2.RowsAffected()
	if err != nil {
		fmt.Println("row failed, ", err)
	}
	//fmt.Println("delete succ: ", row2)

}

