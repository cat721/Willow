package peer

import (
	"Willow/Message"
	"Willow/block"
	"Willow/chain"
	"bytes"
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"io/ioutil"
	"math/rand"
	"net"
	"time"
)

//连接
//发消息
//收消息

const (Delay  = 10
		LeastTimeOfMining = 5
	   )

type Peer struct {
	currentMB *block.MainBlock
	ip string
	templc *chain.TempLedgerChain
	//preTempls *chain.TempLedgerChain
	mc     *chain.MainChain
	listener net.Listener
	RemoteIp string
}


//处理新接收到到消息
func (p *Peer) SolveMessage(message *Message.Message) error {
	blockType := message.Header.MsgType
	switch blockType {
	case uint32(1):
		err := p.solveMainBlock(message.Payload)
		if err != nil{
			return err
		}
		return nil
	case uint32(2):
		err := p.solveLedgerBlock(message.Payload)
		if err != nil{
			return err
		}
		return nil
	case uint32(3):
		err := p.solveFirstLB(message.Payload)
		if err != nil{
			return err
		}
		return nil
	default:
		return errors.New("Wrong type block!")
	}
}

//当收到当信息是主块、账本块的时候
func (p *Peer) solveMainBlock(b []byte) error {
	mb := block.NewEmptyMB()
	mb.ToBlock(b)

	//把这个块放mc上
	p.mc.AddMainBlock(mb)
	p.currentMB = p.mc.LastMainBlock()
	return nil
}

func (p *Peer) solveLedgerBlock(b []byte) error {
	lb := block.NewEmptyLB()
	lb.ToBlock(b)

	round := lb.HeadOfLB.Round

	//如果接到的消息比较当前的round要靠后，就等一会
	if round > p.templc.Round{
		time.Sleep(Delay*time.Second)
	}
	if round == p.templc.Round{
		p.templc.AddHeadOfLedgerBlock(lb.HeadOfLB)
		return nil
	}
	return errors.New("wrong round ledger block")
}

func (p *Peer) solveFirstLB(b []byte) error {

	//等一会确认主块收到了
	time.Sleep(Delay *time.Second)
	lb := block.NewEmptyLB()
	lb.ToBlock(b)
	mbHash,_:= p.mc.LastMainBlock().Hash()

	if lb.HeadOfLB.MainBlockHash != mbHash{
		return errors.New("Wrong first Ledger block for current main block!")
	}

	if lb.HeadOfLB.Round != 1{
		//如果当前的last main block 不是这个first ledger block的就哭呗

		round := lb.HeadOfLB.Round
		if round != p.templc.Round + 1{
			return errors.New("Wrong round first Ledger block!")
		}
		err := p.templc.ExtractLedgerChain(lb.HeadOfLB)
		if err != nil{
			return err
		}

		p.templc = chain.NewTempLC(round,mbHash)
		err = p.templc.AddHeadOfLedgerBlock(lb.HeadOfLB)
		if err != nil{
			return err
		}
	} else {
		p.templc = chain.NewTempLC(p.mc.LastMainBlock().Round,mbHash)
		p.templc.AddHeadOfLedgerBlock(lb.HeadOfLB)
	}
	return nil
}

//检查消息之前是否收到过
func (p *Peer) ChechMessage(msg *Message.Message) (bool,error) {
	blockType := msg.Header.MsgType
	switch blockType {
	case 1:
		mb := block.NewEmptyMB()
		mb.ToBlock(msg.Payload)
		hash,err := mb.Hash()
		if err != nil {
			return false,err
		}
		bool,err := p.ChechHash(hash)
		if err != nil{
			return false,err
		}
		return bool,nil

	case 2, 3:
		lb := block.NewEmptyLB()
		lb.ToBlock(msg.Payload)
		hlb := lb.HeadOfLB
		hash,err := hlb.Hash()
		if err != nil {
			return false,err
		}
		bool,err := p.ChechHash(hash)
		if err != nil{
			return false,err
		}
		return bool,nil
	default:
		return false,nil
	}

}

func (p *Peer) ChechHash(hash [32]byte) (bool,error){
	c, err := redis.Dial("tcp", chain.RedisAdd)
	if err != nil {
		fmt.Println("Connect to redis error", err)
		return false,err
	}

	//检查在不在mainchain里
	_,okMC:= p.mc.SingleBlocks[hash]
	okMCT,err := c.Do("EXISTS",hash)
	if err!=nil{
		return false,err
	}

	//检查在不在ledgerChain中
	_,okLCT := p.templc.MapTree[hash]
	_,okLC := p.templc.SingleBlocks[hash]

	//检查在不在preLedger中
	//_,okPLCT := p.preTempls.MapTree[hash]
	//_,okPLC := p.preTempls.SingleBlocks[hash]


	if okMC ||okLC || okLCT || okMCT.(bool){
		return true,nil
	}
	return false,nil
}

//假装挖矿
func (p *Peer) MineBlock() error {


	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)

	miningTime := r.Intn(600)+LeastTimeOfMining*60

	for {
		fmt.Println(miningTime)
	}


	return nil
}

func NewPeer(ip string,RemoteIp string) *Peer {
	p := Peer{
		ip:ip,
		RemoteIp:RemoteIp,
		mc:chain.NewMainChain(),
	}
	return &p
}

func (p *Peer) StartListen() error {
	listener,err := net.Listen("tcp",p.ip)
	if err != nil{
		fmt.Println("can not create listener on ",p.ip)
		return err
	}
	p.listener = listener

	for{
		conn,err := listener.Accept()
		if err != nil{
			fmt.Println("请求监听失败!")
			continue
		}
		fmt.Println("listen",conn.RemoteAddr().String())
		go p.handleMessage(conn)
	}
}

func (p *Peer) SendMessage(msg *Message.Message) error{
	buf := new(bytes.Buffer)
	msg.Serialize(buf)

	conn,err := net.Dial("tcp",p.RemoteIp)
	if err != nil {
		fmt.Println(err)
	}

	defer conn.Close()
	_,err = conn.Write(buf.Bytes())
	if err != nil {
		return err
	}
	fmt.Println("Successful send the message!")
	return nil
}

func (p *Peer)handleMessage(conn net.Conn) error {
	defer func() {
		conn.Close()
		fmt.Println("close connection")
	}()
	b,err := ioutil.ReadAll(conn)
	if err != nil {
		fmt.Println("read code error: ", err)
		return err
	}

	buf := new(bytes.Buffer)
	buf.Write(b)
	m := Message.Message{}
	err = m.Deserialize(buf)
	if err  != nil{
		return err
	}

	exist,err := p.ChechMessage(&m)
	if err  != nil{
		return err
	}

	if exist {
		fmt.Println("Have already recieced the message!")
		return errors.New("Have already recieced the message!")
	}

	err = p.SolveMessage(&m)
	if err  != nil{
		return err
	}

	return nil
}

