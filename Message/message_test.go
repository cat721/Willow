package Message

import (
	"Willow/block"
	"bytes"
	"crypto/sha256"
	"fmt"
	"testing"
)

func TestNewMessage(t *testing.T) {
	ledgerType := uint32(1)
	round := uint32(1)
	owner := []byte("以战止战")
	prehash := sha256.Sum256(owner)
	nonce := uint64(2333)

	mainBlock := block.NewMainBlock(ledgerType,round,owner,prehash,nonce)
	b,_ := mainBlock.ToJson()

	msg := NewMessage(uint32(1),b)
	fmt.Println(msg)
}

func TestMessage_Serialize(t *testing.T) {
	ledgerType := uint32(1)
	round := uint32(1)
	owner := []byte("以战止战")
	prehash := sha256.Sum256(owner)
	nonce := uint64(2333)

	mainBlock := block.NewMainBlock(ledgerType,round,owner,prehash,nonce)
	b,_ := mainBlock.ToJson()

	msg := NewMessage(uint32(1),b)

	buf := new(bytes.Buffer)
	msg.Deserialize(buf)

	err := msg.Serialize(buf)
	if err != nil{
		t.Error(err)
	}

	m := Message{}
	err = m.Deserialize(buf)
	if err != nil{
		t.Error(err)
	}

	var mb block.MainBlock
	mb.ToBlock(m.Payload)

    fmt.Println("mainBlock BlockType",mainBlock.BlockType)
	fmt.Println("mb BlockType",mb.BlockType)

	fmt.Println("mainBlock Round",mainBlock.Round)
	fmt.Println("mb Round",mb.Round)

	fmt.Println("mainBlock Nonce",mainBlock.Nonce)
	fmt.Println("mb Nonce",mb.Nonce)

	fmt.Println("mainBlock Owner",mainBlock.Owner)
	fmt.Println("mb Owner",mb.Owner)

	fmt.Println("mainBlock PreHash",mainBlock.PreHash)
	fmt.Println("mb PreHash",mb.PreHash)

}
