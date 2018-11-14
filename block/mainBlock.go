package block

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
)


func init() {
	for i,_ := range nilPayload{
		nilPayload[i] = 1
	}
}

type MainBlock struct {
	BlockType uint32 `json:"blockType"`
	Round uint32 `round`
	Owner []byte `owner`
	PreHash [SizeOfHash]byte `preHash`
	Nonce uint64 `nonce`
}

func (mb *MainBlock)ToJson() ([]byte,error) {
	jsonBytes, err := json.Marshal(mb)
	if err != nil {
		fmt.Println(err)
		return nil,err
	}
	return jsonBytes,err
}

func (mb *MainBlock)Hash() ([32]byte,error) {
	bytes,err := mb.ToJson()
	if err != nil {
		return [32]byte{},err
	}

	hash := sha256.Sum256(bytes)
	return  hash,nil
}

func (mb *MainBlock)ToBlock(b []byte) error {
	err := json.Unmarshal(b,mb)
	if err != nil{
		return err
	}
	return nil
}

func NewMainBlock(blocktype uint32,round uint32,owner []byte,preHash [SizeOfHash]byte,nonce uint64,) *MainBlock {
	 mb := MainBlock{
	 	BlockType:blocktype,
	 	Round:round,
	 	Owner:owner,
	 	PreHash:preHash,
	 	Nonce:nonce,
	 }
	 return &mb
}

func NewEmptyMB() *MainBlock {
	mb := MainBlock{}
	return &mb
}


