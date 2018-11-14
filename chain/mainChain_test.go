package chain

import (
	"Willow/block"
	"fmt"
	"testing"
)

func TestNewMainChain(t *testing.T) {
	mc := NewMainChain()
	fmt.Println(mc)

}

func TestMainChain_AddMainBlock(t *testing.T) {
	mc := NewMainChain()
	premb := block.NewMainBlock(uint32(1),uint32(0),[]byte("以战止战"),[32]byte{},uint64(1))

	for i := 1; i < 11; i++ {
		if i < 10{
			hash,_:=premb.Hash()
			mb := block.NewMainBlock(uint32(1),uint32(i),[]byte("以战止战"),hash,uint64(1))
			mc.AddMainBlock(mb)
			premb = mb
		}

		hash,_:=premb.Hash()
		mb := block.NewMainBlock(uint32(1),uint32(i),[]byte("以战止战"),hash,uint64(1))
		premb = mb
	}

	mb := block.NewMainBlock(premb.BlockType,premb.Round,premb.Owner,premb.PreHash,premb.Nonce)

	for i:= 11;i <21;i++{
		hash,_:=premb.Hash()
		mb := block.NewMainBlock(uint32(1),uint32(i),[]byte("以战止战"),hash,uint64(1))
		mc.AddMainBlock(mb)
		premb = mb
	}

	mc.AddMainBlock(mb)

	fmt.Println(mc)
}

func TestMainChain_LastMainBlock(t *testing.T) {

	var mbs [10]*block.MainBlock

	mc := NewMainChain()
	premb := block.NewMainBlock(uint32(1),uint32(0),[]byte("以战止战"),[32]byte{},uint64(1))

	for i := 1; i < 11; i++ {
			hash,_:=premb.Hash()
			mb := block.NewMainBlock(uint32(1),uint32(i),[]byte("以战止战"),hash,uint64(1))
			mc.AddMainBlock(mb)
			mbs[i-1] = mb
			premb = mb
	}

	hash3,_:= mbs[2].Hash()
	mb4 := block.NewMainBlock(uint32(1),uint32(4),[]byte("以战止战"),hash3,uint64(3))
	mc.AddMainBlock(mb4)

	hash4_1,_:= mb4.Hash()
	mb5_1 := block.NewMainBlock(uint32(1),uint32(5),[]byte("以战止战"),hash4_1,uint64(3))
	mc.AddMainBlock(mb5_1)

	hash4,_ := mbs[3].Hash()
	mb5_2 := block.NewMainBlock(uint32(1),uint32(4),[]byte("以战止战"),hash4,uint64(3))
	mc.AddMainBlock(mb5_2)

	lastBlock := mc.LastMainBlock()

	fmt.Println("The last main block is :")
	fmt.Println(lastBlock)

}

