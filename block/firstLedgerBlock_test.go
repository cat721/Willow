package block

import (
	"fmt"
	"testing"
)

func TestFirstLedgerBlock_Hash(t *testing.T) {
	owner := []byte("以战止战")
	flb := NewFirstLB(uint32(2),uint32(1),uint32(100),owner,[32]byte{1,212,1},[32]byte{1,212,1})

	fmt.Println(flb)

	hash,err := flb.Hash()
	if err != nil{
		fmt.Println(err)
	}
	fmt.Println(hash)
}