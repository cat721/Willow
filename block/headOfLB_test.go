package block

import (
	"fmt"
	"testing"
)

func TestHeadOfLB_Hash(t *testing.T) {
	owner := []byte("以战止战")
	hlb := NewHeadOfLB(uint32(2),uint32(1),uint32(100),owner,[32]byte{1,212,1},[32]byte{1,212,1})
	hash,err := hlb.Hash()
	if err != nil{
		fmt.Println(err)
	}
	fmt.Println(hash)
}
