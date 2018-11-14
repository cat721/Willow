package block

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

func TestNewMainBlock(t *testing.T) {
	ledgerType := uint32(1)
	round := uint32(1)
	owner := []byte("以战止战")
	prehash := sha256.Sum256(owner)
	nonce := uint64(2333)

	mainBlock := NewMainBlock(ledgerType,round,owner,prehash,nonce)
	fmt.Printf("The ledgerType is %d\n",mainBlock.BlockType)
	fmt.Printf("The round is %d\n",mainBlock.Round)
	fmt.Printf("The owner is %s\n",string(mainBlock.Owner))
	fmt.Println("The prehash is ",mainBlock.PreHash)
	fmt.Println("The nonce is ",mainBlock.Nonce)
	fmt.Println("Successfully new a mainBlock!!")
}

func TestLedgerBlock_LedgerBlockToJson(t *testing.T) {
	ledgerType := uint32(1)
	round := uint32(1)
	owner := []byte("以战止战")
	prehash := sha256.Sum256(owner)
	nonce := uint64(2333)

	mainBlock := NewMainBlock(ledgerType,round,owner,prehash,nonce)

	bytes,err := mainBlock.ToJson()
	if err != nil{
		fmt.Println(err)
	}
	fmt.Println("The byte of mainBlock is ",bytes)

	hash ,err := mainBlock.Hash()
	if err != nil{
		fmt.Println(err)
	}
	fmt.Println("The hash of mainBlock is ",hash)

	var restorelb MainBlock
	err = restorelb.ToBlock(bytes)
	if err != nil{
		fmt.Println(err)
	}
	fmt.Printf("The ledgerType of restorelb is %d\n",restorelb.BlockType)
	fmt.Printf("The round of restorelb is %d\n",restorelb.Round)
	fmt.Printf("The owner of restorelb is %s\n",string(restorelb.Owner))
	fmt.Printf("The prehash of restorelb is %s\n",restorelb.PreHash)
	fmt.Println("The nonce of restorelb is ",restorelb.Nonce)
	fmt.Println("Successfully  restore the ledgerblock!!")



}
