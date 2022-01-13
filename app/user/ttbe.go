package main

import (
	"crypto/rand"

	"encoding/json"
	"fmt"
	"hash/fnv"
	"math/big"
	//"math/big"
	"strconv"

	"bufio"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	bn "github.com/renzhe666/bn256"
	"os"
	"strings"
)


// TPK TTBE公钥
type TPK struct {
	H1, U1, V1, W1, Z1 *bn.G1
	H2, U2, V2, W2, Z2 *bn.G2
}

// TSK TTBE私钥
type TSK struct {
	index uint64
	U, V  *big.Int
}

// TVK TTBE验证密钥
type TVK struct {
	index uint64
	U, V  *bn.G2
}

// Cttbe TTBE cipher text.
type Cttbe struct {
	C1, C2, C3, C4, C5 *bn.G1
}

// AudClue The auditing clue.
type AudClue struct {
	index    uint64
	AC1, AC2 *bn.G1
}

type ErrorCttbeInvalid struct{}

func (eci *ErrorCttbeInvalid) Error() string {
	return "invalid TTBE cipher text"
}


// SetUp TTBE初始化
func SetUp(n, t uint64) (TPK, []TSK, []TVK) {
	var tpk TPK
	tsks := make([]TSK, 0, n)
	tvks := make([]TVK, 0, n)

	h, _ := rand.Int(rand.Reader, bn.Order)
	w, _ := rand.Int(rand.Reader, bn.Order)
	z, _ := rand.Int(rand.Reader, bn.Order)

	// u is the shamir secret of u_1 ... u_n
	// v is the shamir secret of v_1 ... v_n
	// tsk is the shamir secret of tsk_1=(u_1, v_1) ... tsk_n=(u_n, v_n)
	u, _ := rand.Int(rand.Reader, bn.Order)
	v, _ := rand.Int(rand.Reader, bn.Order)
	polyU := GenRandPoly(t, u, bn.Order)//1
	polyV := GenRandPoly(t, v, bn.Order)//2
	us := GenShares(polyU, n, bn.Order)//3
	vs := GenShares(polyV, n, bn.Order)//4

	H1, H2 := new(bn.G1).ScalarBaseMult(h), new(bn.G2).ScalarBaseMult(h)
	U1, U2 := new(bn.G1).ScalarMult(H1, u), new(bn.G2).ScalarMult(H2, u)
	vInv := TinvMod(v, bn.Order) // get the inverse of v i.e. vInv
	V1, V2 := new(bn.G1).ScalarMult(U1, vInv), new(bn.G2).ScalarMult(U2, vInv)
	W1, W2 := new(bn.G1).ScalarMult(H1, w), new(bn.G2).ScalarMult(H2, w)
	Z1, Z2 := new(bn.G1).ScalarMult(V1, z), new(bn.G2).ScalarMult(V2, z)

	for i := uint64(0); i < n; i++ {
		usi, vsi := us[i], vs[i]
		tsks = append(tsks, TSK{i + 1, usi.Y, vsi.Y})
		tvkUi := new(bn.G2).ScalarMult(H2, usi.Y)
		tvkVi := new(bn.G2).ScalarMult(V2, vsi.Y)
		tvks = append(tvks, TVK{i + 1, tvkUi, tvkVi})
	}
	tpk = TPK{H1, U1, V1, W1, Z1, H2, U2, V2, W2, Z2}

	return tpk, tsks, tvks
}

// Encrypt generate TTBE cipher text.
func Encrypt(tpk TPK, tag *big.Int, msg *bn.G1) *Cttbe {
	r1, _ := rand.Int(rand.Reader, bn.Order)
	r2, _ := rand.Int(rand.Reader, bn.Order)

	C1 := new(bn.G1).ScalarMult(tpk.H1, r1)
	C2 := new(bn.G1).ScalarMult(tpk.V1, r2)
	C3 := new(bn.G1).Add(
		msg,
		new(bn.G1).ScalarMult(tpk.U1, new(big.Int).Add(r1, r2)),
	)

	Ut := new(bn.G1).ScalarMult(tpk.U1, tag)

	C4 := new(bn.G1).Add(Ut, tpk.W1)
	C4.ScalarMult(C4, r1)
	C5 := new(bn.G1).Add(Ut, tpk.Z1)
	C5.ScalarMult(C5, r2)

	cttbe := &Cttbe{C1, C2, C3, C4, C5}

	return cttbe
}

// VerCttbe verify whether TTBE cipher text (cttbe) is
// generated from tag.
func VerCttbe(tpk TPK, tag *big.Int, cttbe *Cttbe) bool {
	Ut := new(bn.G2).ScalarMult(tpk.U2, tag)
	b1 := bn.Pair(cttbe.C1, new(bn.G2).Add(Ut, tpk.W2)).String() == bn.Pair(cttbe.C4, tpk.H2).String()
	b2 := bn.Pair(cttbe.C2, new(bn.G2).Add(Ut, tpk.Z2)).String() == bn.Pair(cttbe.C5, tpk.V2).String()

	return b1 && b2
}

// ShareDec return an auditing clue.
func ShareDec(tpk TPK, tsk TSK, t *big.Int, cttbe *Cttbe) (*AudClue, error) {
	if !VerCttbe(tpk, t, cttbe) {
		return nil, new(ErrorCttbeInvalid)
	}

	ac1 := new(bn.G1).ScalarMult(cttbe.C1, tsk.U)
	ac2 := new(bn.G1).ScalarMult(cttbe.C2, tsk.V)
	audClue := &AudClue{tsk.index, ac1, ac2}

	return audClue, nil
}

// VerAudClue verify the TTBE auditing clue.
func VerAudClue(tpk TPK, tvk TVK, tag *big.Int, cttbe *Cttbe, clue *AudClue) bool {
	if !VerCttbe(tpk, tag, cttbe) {
		return false
	}

	b1 := bn.Pair(clue.AC1, tpk.H2).String() == bn.Pair(cttbe.C1, tvk.U).String()
	b2 := bn.Pair(clue.AC2, tpk.V2).String() == bn.Pair(cttbe.C2, tvk.V).String()

	return b1 && b2
}

func Combine(audClues []*AudClue, cttbe *Cttbe) (*bn.G1, error) {
	den := new(bn.G1)
	indxs := make([]*big.Int, 0, len(audClues))
	for _, ac := range audClues {
		indxs = append(indxs, big.NewInt(int64(ac.index)))
	}

	for _, ac := range audClues {
		indexBig := big.NewInt(int64(ac.index))
		lagcoeff := LagCoeff(indexBig, indxs, bn.Order)
		c1 := new(bn.G1).ScalarMult(ac.AC1, lagcoeff)
		c2 := new(bn.G1).ScalarMult(ac.AC2, lagcoeff)
		d := new(bn.G1).Add(c1, c2)
		den.Add(den, d)
	}
	den.Neg(den)

	return new(bn.G1).Add(cttbe.C3, den), nil
}

// invMod find the inverse of a mod p
func TinvMod(a, p *big.Int) *big.Int {
	return new(big.Int).Exp(a, new(big.Int).Sub(p, big.NewInt(2)), p)
}

func Hash(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}
func Temencrypt(tempub string,tag []byte) []byte { //11111

	s := Hash(tempub)
	hashmac := new(big.Int).SetUint64(s)
	S := new(bn.G1).ScalarBaseMult(hashmac)
	Tag := string(tag)
	t := Hash(Tag)
	T := new(big.Int).SetUint64(t)
	c :=  Encrypt(tpk, T, S)

	jsonBytes, err := json.Marshal(c)
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Println("tag")
	fmt.Println(tag)
	//fmt.Println(T.String())
	//fmt.Println("miwen")
	
    fmt.Println(jsonBytes)
	//fmt.Println(string(jsonBytes))
	return jsonBytes
}
func Audit(contract *client.Contract,tags string,cct string,tsks []TSK, tvks []TVK,user_number int,audituser int) (string,string) {
	tag ,err1:= new(big.Int).SetString(tags,10)
	if err1 != true {
		fmt.Println(err1)
	}
	jsonBytess := []byte(cct)
	var c *Cttbe
	json.Unmarshal(jsonBytess, &c)

	r1 := VerCttbe(tpk, tag, c)
	fmt.Printf("密文c：%v.\n密文是否有效：%v.\n", c, r1)

	ac1, _ := ShareDec(tpk, tsks[0], tag, c)
	ac2, _ := ShareDec(tpk, tsks[1], tag, c)
	ac3, _ := ShareDec(tpk, tsks[2], tag, c)
	ac4, _ := ShareDec(tpk, tsks[3], tag, c)
	ac5, _ := ShareDec(tpk, tsks[4], tag, c)
	acs := []*AudClue{ac1,ac2, ac3,ac4, ac5}

	for _, ac := range acs {
		tvk := tvks[ac.index-1]
		r := VerAudClue(tpk, tvk, tag, c, ac)
		fmt.Printf("审计线索%d：%v\t 有效性：%v.\n", ac.index, ac, r)
	}

    jsoncom, err := json.Marshal(acs)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(jsoncom))

	MRecv,err11 := contract.EvaluateTransaction("TTBE_combine", string(jsoncom), cct,strconv.Itoa(audituser))
	if err11 != nil {
		fmt.Println(err11)
		return "-1","-1"
	}
     

	var id string
	for i:=1;i<=user_number;i++{
		userid := strconv.Itoa(i)
		id = "User"+userid
		s := Hash(id)
		idmac := new(big.Int).SetUint64(s)
		S := new(bn.G1).ScalarBaseMult(idmac)
		if S.String() == string(MRecv){
			break
		}
	}
	name :=Find(id)
	return name,id
	//fmt.Printf("该可疑交易发起方name=%s,id=%s\n",name,id)

}

func Tagstring() string {
	reader := bufio.NewReader(os.Stdin)
	msg, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println(err.Error())
	}
	context := strings.Fields(msg)

	ss := make([]byte,0,len(context))
	for i:=0;i<len(context);i++ {

		src,err6 := strconv.Atoi(context[i])
		if err6 != nil {
			fmt.Println(err6)
		}
		sss := uint8(src)
		ss = append(ss,sss)
	}

	fmt.Println(len(string(ss)))
	fmt.Println(string(ss))
	return string(ss)
}

func Ttbestring() string {
	reader := bufio.NewReader(os.Stdin)
	msg, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println(err.Error())
	}
	context := strings.Fields(msg)

	ss := make([]byte,0,len(context))
	for i:=0;i<len(context);i++ {

		src,err6 := strconv.Atoi(context[i])
		if err6 != nil {
			fmt.Println(err6)
		}
		sss := uint8(src)
		ss = append(ss,sss)
	}
	Tag := string(ss)
	t := Hash(Tag)
	T := new(big.Int).SetUint64(t)
	return T.String()
}




