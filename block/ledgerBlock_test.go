package block

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

func TestNewLedgerBlock(t *testing.T) {
	ledgerType := uint32(2)
	round := uint32(1)
	epoch := uint32(3)
	owner := []byte("以战止战")
	prehash := sha256.Sum256(owner)
	mcHash := [32]byte{}

	ledgerBlock := NewLedgerBlock(ledgerType,round,epoch,owner,prehash,mcHash)
	fmt.Printf("The ledgerType is %d\n",ledgerBlock.HeadOfLB.BlockType)
	fmt.Printf("The round is %d\n",ledgerBlock.HeadOfLB.Round)
	fmt.Printf("The epoch is %d\n",ledgerBlock.HeadOfLB.Epoch)
	fmt.Printf("The owner is %s\n",string(ledgerBlock.HeadOfLB.Owner))
	fmt.Println("The prehash is ",ledgerBlock.HeadOfLB.PreHash)
	fmt.Println("The payload is ",ledgerBlock.Payload)
	fmt.Println("Successfully new a ledgerblock!!")
}

func TestLedgerBlock_LedgerBlockToByte(t *testing.T) {
	ledgerType := uint32(2)
	round := uint32(1)
	epoch := uint32(3)
	owner := []byte("以战止战")
	prehash := sha256.Sum256(owner)
	mcHash := [32]byte{}

	ledgerBlock := NewLedgerBlock(ledgerType,round,epoch,owner,prehash,mcHash)

	bytes,err := ledgerBlock.ToJson()
	if err != nil{
		fmt.Println(err)
	}
	fmt.Println("The byte of ledgerblock is ",bytes)

	hash,err := ledgerBlock.Hash()
	if err != nil{
		fmt.Println(err)
	}
	fmt.Println("The hash of ledgerblock is ",hash)

	var restorelb LedgerBlock
	err = restorelb.ToBlock(bytes)
	if err != nil{
		fmt.Println(err)
	}
	fmt.Printf("The ledgerType of restorelb is %d\n",restorelb.HeadOfLB.BlockType)
	fmt.Printf("The round of restorelb is %d\n",restorelb.HeadOfLB.Round)
	fmt.Printf("The epoch of restorelb is %d\n",restorelb.HeadOfLB.Epoch)
	fmt.Printf("The owner of restorelb is %s\n",string(restorelb.HeadOfLB.Owner))
	fmt.Println("The nonce of restorelb is ",restorelb.HeadOfLB.PreHash)
	fmt.Println("The payload of restorelb is ",restorelb.Payload)
	fmt.Println("Successfully  restore the ledgerblock!!")



}

func TestLedgerBlock_Hash(t *testing.T) {
	ledgerType := uint32(2)
	round := uint32(1)
	epoch := uint32(3)
	owner := []byte("以战止战")
	prehash := sha256.Sum256(owner)
	mcHash := [32]byte{}

	ledgerBlock := NewLedgerBlock(ledgerType,round,epoch,owner,prehash,mcHash)

	hash,err := ledgerBlock.Hash()
	if err != nil{
		fmt.Println(err)
	}
	fmt.Println(hash)

}
