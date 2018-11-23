package main

import (
	"Willow/peer"
	"os"
	"strconv"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(1)

	localAdd := os.Getenv("LocalHost")
	remoteAdd := os.Getenv("RemoteHost")
	bias := os.Getenv("Bias")

	strBias,_ := strconv.ParseInt(bias, 10, 64)

	p := peer.NewPeer(localAdd,remoteAdd,[]byte("以战止战"),strBias)
	go p.StartListen()
	go p.Mine()
	wg.Wait()
}
