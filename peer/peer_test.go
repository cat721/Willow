package peer

import (
	"Willow/Message"
	"Willow/block"
	"bytes"
	"fmt"
	"net"
	"testing"
)

func TestPeer_RecieveMessage(t *testing.T) {
	p := NewPeer("localhost:8000","localhost:8888")
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

	p := NewPeer("localhost:8000","localhost:8888")

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

