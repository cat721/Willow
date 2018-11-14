package chain

import (
	"Willow/block"
	"fmt"
	"testing"
)

func TestTempLedgerChain_NewMainChain(t *testing.T) {
	tempLC := NewTempLC(uint32(2),[32]byte{})
	fmt.Println(tempLC)
}

func TestTempLedgerChain_AddHeadOfLedgerBlock(t *testing.T) {

	tempLC := NewTempLC(uint32(2),[32]byte{})
	owner := []byte("以战止战")
	prehlb :=  block.NewHeadOfLB(uint32(2),uint32(1),uint32(100),owner,[32]byte{},[32]byte{})
	var hlb *block.HeadOfLB
	for i:= 0;i < 10;i++{
		if i <9 {
			hash,err := prehlb.Hash()
			if err != nil{
				t.Error(err)
			}


			hlb := block.NewHeadOfLB(uint32(2),uint32(2),uint32(i),owner,hash,[32]byte{})
			prehlb = hlb
			err = tempLC.AddHeadOfLedgerBlock(hlb)
			if err != nil{
				t.Error(err)
			}
		} else {
			hash,err := prehlb.Hash()
			if err != nil{
				t.Error(err)
			}
			hlb = block.NewHeadOfLB(uint32(2),uint32(3),uint32(i),owner,hash,[32]byte{})
		}
	}


	err := tempLC.ExtractLedgerChain(hlb)
	if err != nil{
		t.Error(err)
	}

    fmt.Println(hlb)
	fmt.Println(tempLC.SingleBlocks)
	fmt.Println(tempLC.MapTree)
	fmt.Println(tempLC.LeavesBlocks)
	for k,v := range tempLC.MapTree{
		fmt.Println("the key is ",k)
		fmt.Println(v.Epoch,v.PreHash)
	}
}

func TestTempLedgerChain_AddHeadOfLedgerBlock2(t *testing.T) {

	tempLC := NewTempLC(uint32(2),[32]byte{})
	owner := []byte("以战止战")
	prehlb :=  block.NewHeadOfLB(uint32(2),uint32(1),uint32(100),owner,[32]byte{},[32]byte{})
	var hlb *block.HeadOfLB
	for i:= 0;i < 10;i++{
		if i <9 {
			hash,err := prehlb.Hash()
			if err != nil{
				t.Error(err)
			}


			hlb := block.NewHeadOfLB(uint32(2),uint32(2),uint32(i),owner,hash,[32]byte{})
			prehlb = hlb
			err = tempLC.AddHeadOfLedgerBlock(hlb)
			if err != nil{
				t.Error(err)
			}
		} else {
			hash,err := prehlb.Hash()
			if err != nil{
				t.Error(err)
			}
			hlb = block.NewHeadOfLB(uint32(2),uint32(3),uint32(i),owner,hash,[32]byte{})
		}
	}


	hlb8 := tempLC.MapTree[hlb.PreHash]
	hlb7 := tempLC.MapTree[hlb8.PreHash]
	hlb6 := tempLC.MapTree[hlb7.PreHash]
	hlb5 := tempLC.MapTree[hlb6.PreHash]
	hlb4 := tempLC.MapTree[hlb5.PreHash]
	hlb3 := tempLC.MapTree[hlb4.PreHash]
	hlb2 := tempLC.MapTree[hlb3.PreHash]
	hlb1 := tempLC.MapTree[hlb2.PreHash]

	hlb9 := block.NewHeadOfLB(uint32(2),uint32(2),uint32(5),owner,hlb5.PreHash,[32]byte{})
	hlb10 := block.NewHeadOfLB(uint32(2),uint32(2),uint32(1),owner,hlb1.PreHash,[32]byte{})
	err := tempLC.AddHeadOfLedgerBlock(hlb9)
	if err != nil{
		t.Error(err)
	}
	err = tempLC.AddHeadOfLedgerBlock(hlb10)
	if err != nil{
		t.Error(err)
	}

	fmt.Println(hlb)
	fmt.Println(tempLC.SingleBlocks)
	fmt.Println(tempLC.MapTree)
	fmt.Println(tempLC.LeavesBlocks)
	for k,v := range tempLC.MapTree{
		fmt.Println("the key is ",k)
		fmt.Println(v.Epoch,v.PreHash)
	}
}

func TestTempLedgerChain_AddHeadOfLedgerBlock3(t *testing.T) {

	tempLC := NewTempLC(uint32(2),[32]byte{})
	owner := []byte("以战止战")
	prehlb :=  block.NewHeadOfLB(uint32(2),uint32(1),uint32(100),owner,[32]byte{},[32]byte{})
	var hash [32]byte

	var hlb *block.HeadOfLB
	for i:= 0;i < 10;i++{
		if i <9 {
			hash,_= prehlb.Hash()

			hlb := block.NewHeadOfLB(uint32(2),uint32(2),uint32(i),owner,hash,[32]byte{})
			prehlb = hlb
			err := tempLC.AddHeadOfLedgerBlock(hlb)
			if err != nil{
				t.Error(err)
			}
		} else {
			hash,_= prehlb.Hash()
			hlb = block.NewHeadOfLB(uint32(2),uint32(2),uint32(i),owner,hash,[32]byte{})
		}
	}

	hlb9 := block.NewHeadOfLB(uint32(2),uint32(2),uint32(9),owner,hash,[32]byte{})

	for i:= 10;i < 20;i++{

			hash,_= hlb9.Hash()
			hlb := block.NewHeadOfLB(uint32(2),uint32(2),uint32(i),owner,hash,[32]byte{})
			hlb9 = hlb
			err := tempLC.AddHeadOfLedgerBlock(hlb)
			if err != nil{
				t.Error(err)
			}
	}

	err := tempLC.AddHeadOfLedgerBlock(hlb)
	if err != nil{
		t.Error(err)
	}

	fmt.Println(hlb)
	fmt.Println(tempLC.SingleBlocks)
	fmt.Println(tempLC.MapTree)
	fmt.Println(tempLC.LeavesBlocks)
	for k,v := range tempLC.MapTree{
		fmt.Println("the key is ",k)
		fmt.Println(v.Epoch,v.PreHash)
	}
}