package block

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

type FirstLedgerBlock struct{
	BlockType uint32 `json:"blockType"`
	Round uint32 `round`
	Epoch uint32 `epoch`
	Owner []byte `owner`
	PreHash [SizeOfHash]byte `preHash`
	MainBlockHash [SizeOfHash]byte `mainBlockHash`
}

func (flb FirstLedgerBlock) Hash() ([32]byte,error) {
	bytes,err := flb.ToJson()
	if err != nil {
		return [32]byte{},err
	}

	hash := sha256.Sum256(bytes)
	return  hash,nil
}

func (flb *FirstLedgerBlock) ToJson () ([]byte,error) {
	jsonBytes, err := json.Marshal(flb)
	if err != nil {
		fmt.Println(err)
		return nil,err
	}
	return jsonBytes,err
}

func (flb *FirstLedgerBlock) ToBlock(b []byte) error {
	err := json.Unmarshal(b,flb)
	if err != nil{
		return err
	}
	return nil
}

func NewFirstLB(blocktype uint32,round uint32,epoch uint32,owner []byte,preHash [SizeOfHash]byte,mainBlockHash [SizeOfHash]byte) FirstLedgerBlock{
	flb := FirstLedgerBlock{
		BlockType:blocktype,
		Round:round,
		Epoch:epoch,
		Owner:owner,
		PreHash:preHash,
		MainBlockHash:mainBlockHash,
	}
	return flb
}