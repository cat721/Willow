package block

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

type HeadOfLB struct {
	BlockType uint32 `json:"blockType"`
	Round uint32 `round`
	Epoch uint32 `epoch`
	Owner []byte `owner`
	PreHash [SizeOfHash]byte `preHash`
	MainBlockHash [SizeOfHash]byte `mainBlockHash`

}

func (hlb HeadOfLB) Hash() ([32]byte,error) {
	bytes,err := hlb.ToJson()
	if err != nil {
		return [32]byte{},err
	}

	hash := sha256.Sum256(bytes)
	return  hash,nil
}

func (hlb *HeadOfLB) ToJson () ([]byte,error) {
	jsonBytes, err := json.Marshal(hlb)
	if err != nil {
		fmt.Println(err)
		return nil,err
	}
	return jsonBytes,err
}

func (hlb *HeadOfLB) ToBlock(b []byte) error {
	err := json.Unmarshal(b,hlb)
	if err != nil{
		return err
	}
	return nil
}

func NewHeadOfLB(blocktype uint32,round uint32,epoch uint32,owner []byte,preHash [SizeOfHash]byte,mainBlockHash [SizeOfHash]byte) *HeadOfLB{
	flb := HeadOfLB{
		BlockType:blocktype,
		Round:round,
		Epoch:epoch,
		Owner:owner,
		PreHash:preHash,
		MainBlockHash:mainBlockHash,
	}
	return &flb
}

func NewEmptyHLB() *HeadOfLB{
	hlb := HeadOfLB{}
	return &hlb
}
