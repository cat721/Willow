package peer

import (
	"Willow/Message"
	"Willow/block"
	"Willow/chain"
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"strconv"
	"time"
)

//连接
//发消息
//收消息
/*
var Delay int
var	LeastTimeOfMining int
var LongestTimeOfMining int
var	numOfPeer int
var IntervalOfLB int

func init() {
	Delay ,_ = strconv.Atoi(os.Getenv("Delay"))
	LeastTimeOfMining,_  = strconv.Atoi(os.Getenv("LeastTimeOfMining"))
	LongestTimeOfMining,_ = strconv.Atoi(os.Getenv("LongestTimeOfMining"))
	numOfPeer,_= strconv.Atoi(os.Getenv("numOfPeer"))
	IntervalOfLB,_ = strconv.Atoi(os.Getenv("IntervalOfLB"))
}
*/

const (Delay  = 1
	LeastTimeOfMining = 100
	LongestTimeOfMining = 300
	numOfPeer = 2
	IntervalOfLB = 3
	W = 2
)

type Peer struct {
	currentMB *block.MainBlock
	minedMB *block.MainBlock
	flag1 chan int
	flag2 map[int]chan int
	ip string
	templc map[uint32]*chain.TempLedgerChain
	//preTempls *chain.TempLedgerChain
	mc     *chain.MainChain
	listener net.Listener
	RemoteIp string
	owner []byte
	bias int64
	isMining bool
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
	_,ok := p.templc[round]
	if!ok {
		log.Println("\n\n[solveLedgerBlock]:Fail add ledger block:in round",lb.HeadOfLB.Round,"epoch",lb.HeadOfLB.Epoch,"\n\n")
		return nil
	}
	err := p.templc[round].AddHeadOfLedgerBlock(lb.HeadOfLB)
	if err !=nil {
		log.Println("\n\n[solveLedgerBlock]:Fail add ledger block:in round",lb.HeadOfLB.Round,"epoch",lb.HeadOfLB.Epoch,"::",err,"\n\n")
		//log.Fatal(err)
	}

	log.Println("[solveLedgerBlock]:successfully add ledger block in round",lb.HeadOfLB.Round,"epoch",lb.HeadOfLB.Epoch)
	return nil

}

func (p *Peer) solveFirstLB(b []byte) error {
	fmt.Println("=============Start solve first ledger block!===========")
	//等一会确认主块收到了
	//time.Sleep(Delay*time.Second)
	lb := block.NewEmptyLB()
	lb.ToBlock(b)
	mbHash,_:= p.currentMB.Hash()
	round := lb.HeadOfLB.Round

	if lb.HeadOfLB.MainBlockHash != mbHash{
		log.Println("\n\n[error]-[solveFirstLB]:Wrong first Ledger block for current main block!\n\n")
		log.Println("\n\n[error]-[solveFirstLB]:current tlc round is",round)
		log.Println("[error]-[solveFirstLB]:recieved lb round is",lb.HeadOfLB.Round,"\n\n")
		return errors.New("\n\n[error]-[solveFirstLB]:Wrong first Ledger block for current main block!\n\n")
	}

	if lb.HeadOfLB.Round > W{
		//如果当前的last main block 不是这个first ledger block的就哭呗


		err := p.templc[round-W].ExtractLedgerChain(lb.HeadOfLB)
		if err != nil{
			fmt.Println("\n\n[error]-[solveFirstLB]:Fail extract ledger chain of round",lb.HeadOfLB.Round,err,"\n\n")
			return err
		}

		delete(p.templc,round-W)

		p.templc[round] = chain.NewTempLC(round,mbHash)
		err = p.templc[round].AddHeadOfLedgerBlock(lb.HeadOfLB)
		if err != nil{
			fmt.Println("\n\n[error]-[solveFirstLB]:Fail add first ledger chain of round",lb.HeadOfLB.Round,err,"\n\n")
			return err
		}

		log.Println("[solveFirstLB]:successfully add the first ledger block in round",lb.HeadOfLB.Round)
	} else {
		p.templc[round] = chain.NewTempLC(round,mbHash)
		err := p.templc[round].AddHeadOfLedgerBlock(lb.HeadOfLB)
		if err != nil{
			fmt.Println("\n\n[error]-[solveFirstLB]:Fail add first ledger chain of round",lb.HeadOfLB.Round,err,"\n\n")
			return err
		}
		log.Println("[solveFirstLB]:successfully add the first ledger block in round",lb.HeadOfLB.Round)
	}
	return nil
}

//检查消息之前是否收到过
func (p *Peer) ChechMessage(hash [32]byte) (bool,error) {
	c, err := redis.Dial("tcp", chain.RedisAdd)
	if err != nil {
		log.Fatal("Connect to redis error", err)
		return false,err
	}
	defer c.Close()
	ok,err := c.Do("SISMEMBER","msg",hash)
	exist := ok.(int64)

	if exist == 1 {
		return true,nil
	}
	return false,nil
}

func (p *Peer) StartListen() error {
	listener,err := net.Listen("tcp",p.ip)
	if err != nil{
		log.Println("can not create listener on ",p.ip)
		log.Println(err)
		return err
	}
	p.listener = listener
	defer listener.Close()

	for{
		conn,err := listener.Accept()
		if err != nil{
			log.Println("请求监听失败!")
			continue
		}
		//log.Println("listen",conn.RemoteAddr().String())
		conn.SetReadDeadline(time.Now().Add(time.Duration(10) * time.Second))
		go p.handleMessage(conn)
	}
}

func (p *Peer) SendMessage(msg *Message.Message) error{
	buf := new(bytes.Buffer)
	msg.Serialize(buf)

	conn,err := net.Dial("tcp",p.RemoteIp)
	if err != nil {
		log.Fatal("\n\n[error]-[SendMessage]:can not create the connenct:",err,"\n\n")
		return err
	}

	defer func() {
		err := conn.Close()
		if err != nil{
			log.Fatal("\n\n[error]-[SendMessage]:close connnection false:",err,"\n\n")
		}
	}()

	hash := sha256.Sum256(buf.Bytes())
	err = p.addHashOfMessage(hash)
	if err != nil {
		log.Fatal("\n\n[error]-[SendMessage]:Add the message to redis wrong:",err,"\n\n")
		return err
	}

	_,err = conn.Write(buf.Bytes())
	if err != nil {
		log.Fatal("\n\n[error]-[SendMessage]:write the message wrong:",err,"\n\n")
		return err
	}

//	fmt.Println("Successful send the message!")
	return nil
}

func (p *Peer) addHashOfMessage(hash [32]byte) error {
	c, err := redis.Dial("tcp", chain.RedisAdd)
	if err != nil {
		log.Println("Connect to redis error", err)
		return err
	}
	defer c.Close()

	_,err = c.Do("SADD","msg",hash)

	if err!=nil{
		return err
	}

	return nil
}

func (p *Peer) handleMessage(conn net.Conn) error {
	defer func() {
		err:=conn.Close()
	//	log.Println("close connection")
		if err != nil{
			log.Fatal("\n\n[error]-[handleMessage]write the message wrong:",err,"\n\n")
		}
	}()
	b,err := ioutil.ReadAll(conn)

	if err != nil {
		log.Println("\n\n[error]-[handleMessage]:read code error: ", err,"\n\n")
		return err
	}

	buf := new(bytes.Buffer)
	buf.Write(b)
	m := Message.Message{}
	err = m.Deserialize(buf)
	if err  != nil{
		return err
	}

	hash := sha256.Sum256(b)
	exist,err := p.ChechMessage(hash)
	if err  != nil{
		return err
	}

	if exist {
		log.Println("[info]-[handleMessage]:Have already recieced the message!")
		return nil
	}

	log.Println("Recieve",readMessage(&m))

	err = p.SolveMessage(&m)
	if err  != nil{
		log.Println("[error]-[handleMessage]:failed solve message::",err)
	}
	log.Println("[info]-[handleMessage]:Recieve a new message!")

	err = p.SendMessage(&m)
	if err != nil{
		log.Println(err)
		return nil
	}

	return nil
}

//假装挖矿
func (p *Peer) Mine() error {
	for {
		round := int(p.currentMB.Round)+1
		mb:= p.MineBlock()
		flag := make(chan int,1)
		p.flag2[round] = flag


		if mb == nil{
			if p.isMining == true{
				_,ok := p.flag2[round-W]
				if ok{
					p.flag2[round-W] <- 1
				}
			}
			continue
		}

		b,err:= mb.ToJson()
		log.Println("Get new block round",mb.Round)
		//将新产生的main block加入到本地视图

		if p.isMining == true{
			_,ok := p.flag2[round-W]
			if ok{
				p.flag2[round-W] <- 1
			}

		}

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
		time.Sleep(time.Second*time.Duration(IntervalOfLB))
		//产生ledger block
		go p.createLB(mb)
	}
	return nil
}

func (p *Peer) MineBlock() *block.MainBlock {
	c := make(chan *block.MainBlock,1)

	go p.mineBlock(c)

	select {
	case <-p.flag1:
		log.Println("\n",string(p.owner),"[info]-[MineBlock]:recieve a newer main block\n")
		return nil
	case mb := <-c:
		log.Println("\n",string(p.owner),"[info]-[MineBlock]:mined a newer main block\n")
		return mb
	}
}

func (p *Peer) mineBlock(c chan *block.MainBlock) error{
	defer close(c)

	s := rand.NewSource(time.Now().Unix()+p.bias)
	r := rand.New(s)
	miningTime := r.Intn((LongestTimeOfMining-LeastTimeOfMining)*numOfPeer)+LeastTimeOfMining
	log.Println("[info]-[mineBlock]:mining time is",miningTime,"and mining round",p.currentMB.Round+1)
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

	fmt.Println("\n\n===========",string(p.currentMB.Owner),"start creating ledger block for",p.currentMB.Round,"=========\n\n")
	defer func() {
		p.isMining = false
		fmt.Println("\n outof create ledger block!\n")
		}()


	preHash := [32]byte{}
	hash,_:= mb.Hash()
	round := mb.Round

	i := 0
	CreateLB:
		for {
			select {
			case <- p.flag2[int(round)]:
				delete(p.flag2,int(round))
				break CreateLB
			default:
				if i == 0{
					var lb *block.LedgerBlock
					_,ok := p.templc[round-W]
					if ok {
						preHash,_ = p.templc[round-W].LastLedgerBlock().Hash()
						lb = block.NewLedgerBlock(uint32(3),mb.Round,uint32(i),p.owner,preHash,hash)
					} else {
						lb = block.NewLedgerBlock(uint32(3),mb.Round,uint32(i),p.owner,preHash,hash)
					}

					preHash,_ = lb.HeadOfLB.Hash()
					b,err := lb.ToJson()
					if err != nil{
						log.Fatal(err)
						return err
					}
					msg := Message.NewMessage(uint32(3),b)

					err = p.SolveMessage(msg)
					if err != nil{
						log.Fatal(err)
						return err
					}
					err = p.SendMessage(msg)
					if err != nil{
						log.Fatal(err)
						return err
					}

					p.isMining = true
					time.Sleep(time.Second*time.Duration(IntervalOfLB))
					//log.Println(string(p.owner),"create round",lb.HeadOfLB.Round,"epoch",lb.HeadOfLB.Epoch)
					i++
				}else {
					lb := block.NewLedgerBlock(uint32(2),round,uint32(i),p.owner,preHash,hash)
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
					time.Sleep(time.Second* time.Duration(IntervalOfLB))
					//log.Println(string(p.owner),"create round",lb.HeadOfLB.Round,"epoch",lb.HeadOfLB.Epoch)
					i++
				}
			}
		}

	fmt.Println("\n=========Stop create ledger block of round",round,"=============\n")
	return nil
}

func NewPeer(ip string,RemoteIp string,owner []byte,bias int64) *Peer {
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
		minedMB:&mb,
		mc:chain.NewMainChain(),
		templc:make(map[uint32]*chain.TempLedgerChain),
		owner:owner,
		flag1:make(chan int,1),
		flag2:make(map[int] chan int),
		currentMB:&mb,
		bias:bias,
		isMining:false,
	}

	return &p
}

func readMessage(msg *Message.Message) string{
	switch msg.MsgType {
	case 1:
		mb := block.NewEmptyMB()
		mb.ToBlock(msg.Payload)
		return string(mb.Owner)+"'s main block in round:"+ strconv.Itoa(int(mb.Round))
	case 2,3:
		lb := block.NewEmptyLB()
		lb.ToBlock(msg.Payload)
		return string(lb.HeadOfLB.Owner)+"'s ledger block "+"of type " + strconv.Itoa(int(msg.Header.MsgType))+" in round: "+ strconv.Itoa(int(lb.HeadOfLB.Round))+" epoch "+strconv.Itoa(int(lb.HeadOfLB.Epoch))
	}
	return "can not read the message"
}


