package chain

import (
	"Willow/block"
	"fmt"
	"github.com/garyburd/redigo/redis"
)

type MainChain struct {
	LongestLeaves map[[32]byte] *block.MainBlock
	SingleBlocks map[[32]byte] *block.MainBlock
}

func NewMainChain() *MainChain {
	mc := MainChain{
		LongestLeaves: make(map[[32]byte] *block.MainBlock),
		SingleBlocks:make(map[[32]byte] *block.MainBlock),
	}

	return &mc
}

func NewMC() *MainChain {
	mc := MainChain{}
	return &mc
}

func (mc *MainChain) AddMainBlock(mb *block.MainBlock) error {
	c, err := redis.Dial("tcp", RedisAdd)
	if err != nil {
		fmt.Println("Connect to redis error", err)
		return err
	}

	hash,err := mb.Hash()
	if err != nil{
		return err
	}
	fmt.Println("the hash is",hash)
	
	defer c.Close()

	//如果是第一个块，就直接放在树里。
	if mb.Round == uint32(1){

		err := storemb(mb,c)
		if err != nil{
			return err
		}
		mc.LongestLeaves[hash] = mb


		fmt.Println("第一个块",mb.PreHash)
		fmt.Println("prehash",mb.PreHash)
		fmt.Println(mc.SingleBlocks)
		fmt.Println(mc.LongestLeaves)
		return nil
	}

	//连向叶子节点
	_,ok := mc.LongestLeaves[mb.PreHash]
	if ok {

		err := storemb(mb,c)
		if err != nil{
			return err
		}

		delete(mc.LongestLeaves,mb.PreHash)
		mc.LongestLeaves[hash] = mb

		err = mc.updateSingleBlocks(mb,c)
		if err != nil{
			return err
		}


		fmt.Println("连向叶子节点",mb.Round)
		fmt.Println("prehash",mb.PreHash)
		fmt.Println(mc.SingleBlocks)
		fmt.Println(mc.LongestLeaves)
		return nil
	}

	//连向中间节点
	 preHash := mb.PreHash
	 exist,_ := existmb(preHash,c)

	if exist {
		storemb(mb,c)
		mc.LongestLeaves[hash] = mb
		err := mc.updateSingleBlocks(mb,c)
		if err != nil{
			return err
		}

		fmt.Println("连向中间节点",mb.Round)
		fmt.Println("prehash",mb.PreHash)
		fmt.Println(mc.SingleBlocks)
		fmt.Println(mc.LongestLeaves)
		return nil
	}
	//孤块
	mc.SingleBlocks[hash] = mb

	return nil
}

func storemb(mb *block.MainBlock,c redis.Conn) error{

	hash,_ := mb.Hash()

	fmt.Println("adding the ",mb.Round,"round")

	_,err := c.Do("HMSET",hash,"BlockType",mb.BlockType,
											 "Round",mb.Round,
											 "Owner",mb.Owner,
											 "PreHash",mb.PreHash,
											 "Nonce",mb.Nonce)
	if err != nil {
		fmt.Println("redis set failed:", err)
		return err
	}
	return nil
}

func delmb(mb *block.MainBlock,c redis.Conn) error{

	hash,_ := mb.Hash()

	fmt.Println("delet the ",mb.Round,"round")

	_,err := c.Do("DEL",hash)
	if err != nil {
		fmt.Println("redis set failed:", err)
		return err
	}
	return nil
}

func preHash(mb *block.MainBlock,c redis.Conn) ([32]byte,error){

	hash,_ := mb.Hash()

	fmt.Println("Find the preHash of ",hash,"in",mb.Round,"round")

	premb,err := c.Do("HGET",hash,"PreHash")
	if err != nil {
		fmt.Println("redis set failed:", err)
		return [32]byte{},err
	}

	return premb.([32]byte),nil
}

func existmb(hash interface{},c redis.Conn) (bool,error) {

	fmt.Println("is the ",hash,"exist?")
	ok,err := c.Do("EXISTS",hash)
	is := ok.(bool)

	if err != nil {
		fmt.Println("redis set failed:", err)
		return false,err
	}

	return is,nil
}

func (mc *MainChain) updateSingleBlocks(mb *block.MainBlock,c redis.Conn) error {
	hash,err := mb.Hash()
	if err != nil{
		return err
	}

	for k,v := range mc.SingleBlocks{
		if v.PreHash == hash{
			mc.LongestLeaves[k] = v
			err :=storemb(v,c)
			if err != nil{
				return err
			}
			mc.updateSingleBlocks(v,c)
		}
		delete(mc.LongestLeaves,hash)
	}
	return nil
}

func (mc *MainChain) LastMainBlock() *block.MainBlock {

	lmb := block.NewMainBlock(uint32(1),uint32(0),[]byte("以战止战"),[32]byte{},uint64(0))
	for _,v := range mc.LongestLeaves{
		if v.Round > lmb.Round{
			lmb = v
		}
	}
	return lmb
}