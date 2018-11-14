package block

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

const (
	SizeOfPayload = 1 << 17 //1M byte
	SizeOfHash = 32
)

var nilPayload [SizeOfPayload]byte

func init() {
	for i,_ := range nilPayload{
		nilPayload[i] = 1
	}
}

type LedgerBlock struct {
	HeadOfLB *HeadOfLB
	Payload [SizeOfPayload]byte `payload`
}

func (lb *LedgerBlock) ToJson() ([]byte,error) {
	jsonBytes, err := json.Marshal(lb)
	if err != nil {
		fmt.Println(err)
		return nil,err
	}
	return jsonBytes,err
}

func (lb LedgerBlock) Hash() ([32]byte,error) {
	bytes,err := lb.ToJson()
	if err != nil {
		return [32]byte{},err
	}

	hash := sha256.Sum256(bytes)
	return  hash,nil
}

func (lb *LedgerBlock) ToBlock(b []byte) error {
	err := json.Unmarshal(b,lb)
	if err != nil{
		return err
	}
	return nil
	}


func NewLedgerBlock(blocktype uint32,round uint32,epoch uint32,owner []byte,preHash [SizeOfHash]byte,mainBlockHash [SizeOfHash]byte) *LedgerBlock {

	headOfLB := NewHeadOfLB(blocktype,round,epoch,owner,preHash,mainBlockHash)
	mb := LedgerBlock{
		HeadOfLB:headOfLB,
		Payload:nilPayload,
	}
	return &mb
}

func NewEmptyLB() *LedgerBlock {
	lb := LedgerBlock{}
	return &lb
}

