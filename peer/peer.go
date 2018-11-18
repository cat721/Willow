package peer

import (
	"Willow/Message"
	"Willow/block"
	"Willow/chain"
	"bytes"
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/johnnadratowski/golang-neo4j-bolt-driver/log"
	"io/ioutil"
	"math/rand"
	"net"
	"time"
)

//连接
//发消息
//收消息

const (Delay  = 2
	   LeastTimeOfMining = 0
	   LongestTimeOfMining = 1
	   numOfPeer =1
	   )

type Peer struct {
	currentMB *block.MainBlock
	flag1 chan int
	flag2 chan int
	ip string
	templc *chain.TempLedgerChain
	//preTempls *chain.TempLedgerChain
	mc     *chain.MainChain
	listener net.Listener
	RemoteIp string
	owner []byte
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
		return errors.New("Error:Wrong type block!")
	}
}

//当收到当信息是主块、账本块的时候
func (p *Peer) solveMainBlock(b []byte) error {
	mb := block.NewEmptyMB()
	mb.ToBlock(b)

	//把这个块放mc上
	p.mc.AddMainBlock(mb)
	lastMainBlock := p.mc.LastMainBlock()
	//如果接收了一个新的最后的block，更新current Block
	if 	p.currentMB.Round < lastMainBlock.Round{
		p.currentMB = lastMainBlock
		if string(mb.Owner) != string(p.owner) {
			p.flag1 <- 1
		}
	}
	return nil
}

func (p *Peer) solveLedgerBlock(b []byte) error {
	lb := block.NewEmptyLB()
	lb.ToBlock(b)

	round := lb.HeadOfLB.Round
	fmt.Println("Now the templedger is round",round)
	//如果接到的消息比较当前的round要靠后，就等一会
	//if round > p.templc.Round{
	//	time.Sleep(Delay*time.Second)
	//}
	if round == p.templc.Round{
		if string(p.currentMB.Owner) == string(lb.HeadOfLB.Owner){
			p.templc.AddHeadOfLedgerBlock(lb.HeadOfLB)
			return nil
		}
	}
	//return errors.New("wrong round ledger block")
	return nil
}

func (p *Peer) solveFirstLB(b []byte) error {

	//等一会确认主块收到了
	time.Sleep(Delay*time.Second)
	lb := block.NewEmptyLB()
	lb.ToBlock(b)
	mbHash,_:= p.currentMB.Hash()

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
		p.flag2 <- 1
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
	defer c.Close()
	//检查在不在mainchain里
	_,okMC:= p.mc.SingleBlocks[hash]
	okMCT,err := c.Do("EXISTS",hash)
	if err!=nil{
		return false,err
	}

	okmct := false
	if okMCT.(int64) != 0{
		okmct = true
	}

	//检查在不在ledgerChain中
	_,okLCT := p.templc.MapTree[hash]
	_,okLC := p.templc.SingleBlocks[hash]

	//检查在不在preLedger中
	//_,okPLCT := p.preTempls.MapTree[hash]
	//_,okPLC := p.preTempls.SingleBlocks[hash]


	if okMC ||okLC || okLCT || okmct {
		return true,nil
	}
	return false,nil
}

func NewPeer(ip string,RemoteIp string,owner []byte) *Peer {
	mb := block.MainBlock{
					BlockType:uint32(1),
					Round:uint32(0),
					Owner:owner,
					PreHash:[32]byte{},
					Nonce:uint64(0),
	}



	p := Peer{
		ip:ip,
		RemoteIp:RemoteIp,
		mc:chain.NewMainChain(),
		templc:chain.NewTLC(),
		owner:owner,
		flag1:make(chan int,1),
		flag2:make(chan int,1),
		currentMB:&mb,
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
	defer listener.Close()

	for{
		conn,err := listener.Accept()
		if err != nil{
			fmt.Println("请求监听失败!")
			continue
		}
		fmt.Println("listen",conn.RemoteAddr().String())
		conn.SetReadDeadline(time.Now().Add(time.Duration(10) * time.Second))
		go p.handleMessage(conn)
	}
}

func (p *Peer) SendMessage(msg *Message.Message) error{
	buf := new(bytes.Buffer)
	msg.Serialize(buf)

	conn,err := net.Dial("tcp",p.RemoteIp)
	if err != nil {
		fmt.Println("can not create the connenct:",err)
		return err
	}

	defer func() {
		err := conn.Close()
		if err != nil{
			fmt.Println("close connnection false:",err)
		}
	}()
	_,err = conn.Write(buf.Bytes())
	if err != nil {
		fmt.Println("write the message wrong:",err)
		return err
	}
//	fmt.Println("Successful send the message!")
	return nil
}

func (p *Peer)handleMessage(conn net.Conn) error {
	defer func() {
		err:=conn.Close()
		fmt.Println("close connection")
		if err != nil{
			fmt.Println("write the message wrong:",err)
		}
	}()
	b,err := ioutil.ReadAll(conn)

	fmt.Println("recieve a message!")

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
	fmt.Println("recieve a message!")
	err = p.SendMessage(&m)
	if err != nil{
		fmt.Println(err)
		return nil
	}

	return nil
}

//假装挖矿
func (p *Peer) Mine() error {
	for  {
		mb,err:= p.MineBlock()
		if err != nil{
			fmt.Println(err)
			continue
		}

		if mb == nil{
			continue
		}

		b,err:= mb.ToJson()
		fmt.Println("get new block round",mb.Round)
		//将新产生的main block加入到本地视图
		err = p.solveMainBlock(b)
		if err != nil{
			log.Fatal(err)
			continue
		}
		//将产生的新的main block发送出去
		msg := Message.NewMessage(uint32(1),b)
		err = p.SendMessage(msg)
		if err != nil{
			log.Fatal(err)
			continue
		}

		//产生ledger block
		go p.createLB(mb)
}
	return nil
}

func (p *Peer) MineBlock() (*block.MainBlock,error) {
	c := make(chan *block.MainBlock,1)

	go p.mineBlock(c)

	select {
	case <-p.flag1:
		fmt.Println("recieve a newer main block")
		return nil,errors.New("recieve a newer main block")
	case mb := <-c:
		fmt.Println("mined a newer main block")
		return mb,nil
	}
}

func (p *Peer) mineBlock(c chan *block.MainBlock) error{
	defer close(c)

	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)
	miningTime := r.Intn(60*(LongestTimeOfMining-LeastTimeOfMining)*numOfPeer)+LeastTimeOfMining*60
	fmt.Println("mining time is",miningTime)
 	time.Sleep(time.Duration(miningTime)*time.Second)

//	time.Sleep(time.Second*9)
	hash,err := p.currentMB.Hash()
	if err != nil {
		return nil
	}
	mb := block.NewMainBlock(uint32(1),p.currentMB.Round+1,p.owner,hash,uint64(1))
	c <- mb
	return nil

}

func (p *Peer) createLB(mb *block.MainBlock) error {
	preHash := [32]byte{}
	hash,_:= mb.Hash()

	i := 0
	for {
		select {
		case <- p.flag2:
			return nil
		default:
			if i == 0{
				preHash,_ = p.templc.LastLedgerBlock().Hash()
				lb := block.NewLedgerBlock(uint32(1),mb.Round,uint32(i),p.owner,preHash,hash)
				preHash,_ = lb.HeadOfLB.Hash()

				b,err := lb.ToJson()
				if err != nil{
					log.Fatal(err)
					continue
				}
				msg := Message.NewMessage(uint32(3),b)

				err = p.SolveMessage(msg)
				if err != nil{
					log.Fatal(err)
					continue
				}
				err = p.SendMessage(msg)
				if err != nil{
					log.Fatal(err)
					continue
				}
				time.Sleep(time.Second*1)
				fmt.Println("create round",lb.HeadOfLB.Round,"epoch",lb.HeadOfLB.Epoch)
				i++
			}else {
				lb := block.NewLedgerBlock(uint32(1),mb.Round,uint32(i),p.owner,preHash,hash)
				preHash,_ = lb.HeadOfLB.Hash()
				b,err := lb.ToJson()
				if err != nil{
					log.Fatal(err)
					continue
				}
				msg := Message.NewMessage(uint32(2),b)

				err = p.SolveMessage(msg)
				if err != nil{
					log.Fatal(err)
					continue
				}
				err = p.SendMessage(msg)
				if err != nil{
					log.Fatal(err)
					continue
				}
				time.Sleep(time.Second*1)
				fmt.Println("create round",lb.HeadOfLB.Round,"epoch",lb.HeadOfLB.Epoch)
				i++
			}
		}
	}
	return nil
}
