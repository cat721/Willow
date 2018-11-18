package peer

import (
	"Willow/Message"
	"Willow/block"
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"sync"
	"testing"
	"time"
)

func TestPeer_RecieveMessage(t *testing.T) {
	p := NewPeer("localhost:8000","localhost:8888",[]byte("以战止战"))
	p.StartListen()

	mb := block.NewMainBlock(uint32(1),uint32(1),[]byte("以战止战"),[32]byte{},uint64(2))
	b,err := mb.ToJson()
	if err != nil{
		fmt.Println(err)
	}
	msg := Message.NewMessage(mb.BlockType,b)
	buf := new(bytes.Buffer)
	msg.Serialize(buf)

	conn,err := net.Dial("tcp","localhost:8000")
	if err != nil {
		fmt.Println(err)
	}

	defer conn.Close()
	num,err := conn.Write(buf.Bytes())
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("the length is",num)
}

func TestPeer_SolveMessage(t *testing.T) {
	mb := block.NewMainBlock(uint32(1),uint32(1),[]byte("以战止战"),[32]byte{},uint64(2))
	b,err := mb.ToJson()
	if err != nil{
		fmt.Println(err)
	}

	hash,_ := mb.Hash()

	msg := Message.NewMessage(mb.BlockType,b)

	p := NewPeer("localhost:8000","localhost:8888",[]byte("以战止战"))

	err = p.SolveMessage(msg)
	fmt.Println(err)

	lb := block.NewLedgerBlock(uint32(3),uint32(1),uint32(0),[]byte("以战止战"),[32]byte{},hash)
	b,_ = lb.ToJson()

	msg = Message.NewMessage(lb.HeadOfLB.BlockType,b)
	err = p.SolveMessage(msg)
	fmt.Println(err)

	phash,_ := lb.HeadOfLB.Hash()

	lb = block.NewLedgerBlock(uint32(2),uint32(1),uint32(1),[]byte("以战止战"),phash,hash)
	b,_ = lb.ToJson()

	msg = Message.NewMessage(lb.HeadOfLB.BlockType,b)
	err = p.SolveMessage(msg)
	fmt.Println(err)
}

func TestPeer_mine(t *testing.T) {
	p := NewPeer("localhost:8000","localhost:8888",[]byte("以战止战"))
	listener,err := net.Listen("tcp","localhost:8888")
	if err != nil{
		fmt.Println("can not create listener on 8888\n because of",err)
	}

	defer listener.Close()

	go func() {
		for{

			conn,err := listener.Accept()
			if err != nil{
				fmt.Println("请求监听失败!")
				continue
			}
			defer conn.Close()

			b,err := ioutil.ReadAll(conn)
			if err != nil {
				fmt.Println("read code error: ", err)
			}

			buf := new(bytes.Buffer)
			buf.Write(b)
			m := Message.Message{}
			err = m.Deserialize(buf)
			if err  != nil{
				fmt.Println(err)
			}

			fmt.Println("receive a message!")

		}
	}()


	p.Mine()

}

func Test_createLB(t *testing.T){
	p := NewPeer("localhost:8000","localhost:8888",[]byte("以战止战"))

	listener,err := net.Listen("tcp","localhost:8888")
	if err != nil{
		fmt.Println("can not create listener on 8888\n because of",err)
	}
	p.listener = listener

	go func() {
		for{
			conn,err := listener.Accept()
			if err != nil{
				fmt.Println("请求监听失败!")
				continue
			}
		//fmt.Println("listen",conn.RemoteAddr().String())
			defer conn.Close()
	}}()


	//time.Sleep(time.Second*5)
	mb := block.NewMainBlock(uint32(1),uint32(1),[]byte("以战止战"),[32]byte{},uint64(2))
	prehash,_ :=mb.Hash()
	b,err := mb.ToJson()
	if err != nil{
		fmt.Println(err)
	}
	msg := Message.NewMessage(mb.BlockType,b)
	err = p.SolveMessage(msg)
//	<- p.flag1
	go p.createLB(mb)


	time.Sleep(time.Second*10)
	mb = block.NewMainBlock(uint32(1),uint32(2),[]byte("以战止战"),prehash,uint64(2))
	b,err = mb.ToJson()
	if err != nil{
		fmt.Println(err)
	}
	msg = Message.NewMessage(mb.BlockType,b)
	err = p.SolveMessage(msg)
	go p.createLB(mb)

	time.Sleep(time.Second*10)
	p.flag2<-1

	fmt.Println("finnished!")
}

func TestPeer_1(t *testing.T) {

	var wg sync.WaitGroup
	wg.Add(1)

	p := NewPeer("localhost:8000","localhost:8888",[]byte("以战止战"))
	go p.StartListen()
	go p.Mine()

	wg.Wait()

}

func TestPeer_2(t *testing.T){
	var wg sync.WaitGroup
	wg.Add(1)

	p2 := NewPeer("localhost:8888","localhost:8000",[]byte("至尊宝"))
	go p2.StartListen()
//	go p2.Mine()

	wg.Wait()
}
