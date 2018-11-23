package chain

import (
	"Willow/block"
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"os"
)

//const RedisAdd = "127.0.0.1:6379"
var RedisAdd string

func init() {
	RedisAdd = os.Getenv("RedisAdd")
}

type TempLedgerChain struct {
	Round uint32
	MainBlockHash [32]byte
	MapTree map[[32]byte]*block.HeadOfLB
	SingleBlocks map[[32]byte]*block.HeadOfLB
	LeavesBlocks map[[32]byte]*block.HeadOfLB
}

func NewTempLC(round uint32,mainChainHash [32]byte) *TempLedgerChain{
	tempLC := TempLedgerChain{
		Round:round,
		MainBlockHash:mainChainHash,
		MapTree:make(map[[32]byte]*block.HeadOfLB),
		SingleBlocks:make(map[[32]byte]*block.HeadOfLB),
		LeavesBlocks:make(map[[32]byte]*block.HeadOfLB),
	}

	return &tempLC
}
func NewTLC() *TempLedgerChain{
	lc := TempLedgerChain{}
	return &lc
}

func (TempLChain *TempLedgerChain) ExtractLedgerChain(hlb *block.HeadOfLB) error {

	if TempLChain.Round != hlb.Round-1 {
		return errors.New("[ExtractLedgerChain]:The first ledgerblock is not belong to next round")
	}
	_,ok := TempLChain.MapTree[hlb.PreHash]
	if !ok {
		return errors.New("[ExtractLedgerChain]:wrong local tempLedgerChain")
	}
	//存入数据库中
	c, err := redis.Dial("tcp", RedisAdd)
	if err != nil {
		fmt.Println("Connect to redis error", err)
		return err
	}

	defer c.Close()

	hlb = TempLChain.MapTree[hlb.PreHash]

	for {
		err :=storehlb(hlb,c)
		if err != nil{
			fmt.Println(err)
			return err
		}
		preHash := hlb.PreHash
	//	fmt.Println(preHash)

		hlb,ok = TempLChain.MapTree[preHash]

		if !ok{
			break
		}

	}
	//fmt.Println("Don not get the last ledger block!")
	return nil
}

func (TempLChain *TempLedgerChain) PreHeadOfLedgerBlock (hlb *block.HeadOfLB) (*block.HeadOfLB,error){
	if TempLChain.Round != hlb.Round{
		return nil,errors.New("The LedgerBlock is not in the correct round!")
	}
	preBlock,ok := TempLChain.MapTree[hlb.PreHash]
	if !ok{
		return nil,errors.New("Not have previous ledgerblock")
	}
	return preBlock,nil

}

func (TempLChain *TempLedgerChain) AddHeadOfLedgerBlock(hlb *block.HeadOfLB) error {

	if TempLChain.Round != hlb.Round{
		return errors.New("The LedgerBlock is not in the correct round!")
	}
	//传入块头的hash
	hash,err := hlb.Hash()
	if err != nil{
		return err
	}

	//如果是第一个块，就直接放在树里。
	if hlb.Epoch == uint32(0){
		TempLChain.MapTree[hash] = hlb
		TempLChain.LeavesBlocks[hash] = hlb
		return nil
	}

	//连向叶子节点
	_,ok := TempLChain.LeavesBlocks[hlb.PreHash]
	if ok {
		TempLChain.MapTree[hash] = hlb
		delete(TempLChain.LeavesBlocks,hlb.PreHash)
		TempLChain.LeavesBlocks[hash] = hlb

		err := TempLChain.updateSingleBlocks(hlb)
		if err != nil{
			return err
		}
		return nil
	}

	//连向中间节点
	_,ok = TempLChain.MapTree[hlb.PreHash]
	if ok{
		TempLChain.MapTree[hash] = hlb
		TempLChain.LeavesBlocks[hash] = hlb
		err := TempLChain.updateSingleBlocks(hlb)
		if err != nil{
			return err
		}
		return nil
	}
	//孤块
	TempLChain.SingleBlocks[hash] = hlb

	return nil
}

func (TempLChain *TempLedgerChain) updateSingleBlocks(hlb *block.HeadOfLB) error {
	//传入块头的hash
	hash,err := hlb.Hash()
	if err != nil{
		return err
	}

	for k,v := range TempLChain.SingleBlocks{
		if v.PreHash == hash{
			TempLChain.LeavesBlocks[k] = v
			TempLChain.MapTree[k] = v
			delete(TempLChain.SingleBlocks,k)
			TempLChain.updateSingleBlocks(v)
		}
		delete(TempLChain.LeavesBlocks,hash)
	}
	return nil
}

func (TempLChain *TempLedgerChain) LastLedgerBlock() *block.HeadOfLB {

	llb := block.NewEmptyHLB()
	if TempLChain == nil{
		return llb
	}
	for _,v := range TempLChain.LeavesBlocks{
		if v.Epoch > llb.Epoch{
			llb = v
		}
	}
	return llb
}

func storehlb(hlb *block.HeadOfLB,c redis.Conn) error{

	hash,_ := hlb.Hash()

//	fmt.Println("adding the ",hlb.Epoch,"epoch")

	_,err := c.Do("HMSET",hash,"BlockType",hlb.BlockType,
								 	 	 	  "Round",hlb.Round,
								 	 	 	  "Epoch",hlb.Epoch,
								 	 	 	  "Owner",hlb.Owner,
								 	 	 	  "PreHash",hlb.PreHash,
								 	 	 	  "MainBlockHash",hlb.MainBlockHash)
	if err != nil {
		fmt.Println("redis set failed:", err)
		return err
	}
	return nil
}