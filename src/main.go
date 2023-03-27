package main

import (
	"context"
	"github.com/akxecosystem/akxchain/network"
	"github.com/k0kubun/go-ansi"
	progressbar "github.com/schollz/progressbar/v3"
	"log"
	"time"
)

func main() {
	bar := progressbar.NewOptions(1000,
		progressbar.OptionSetWriter(ansi.NewAnsiStdout()),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(15),
		progressbar.OptionSetDescription("[cyan][1/3][reset] Generating kyber crystal akx ecosystem node keys..."),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))
	for i := 0; i < 1000; i++ {
		bar.Add(1)
		time.Sleep(2 * time.Millisecond)
	}

	node, _, err := network.NewNode(context.Background(), nil)
	if err != nil {
		panic(err)
	}
	log.Print(node.ID().String())

	/*kp, s := kem.CreateNewKeysWithSharedSecret()

	spew.Dump(s)
	spew.Dump(kp.)*/

}
