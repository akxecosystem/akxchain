package main

import (
	"akxchain/internal/libs/kem"
	"github.com/davecgh/go-spew/spew"
	"github.com/k0kubun/go-ansi"
	progressbar "github.com/schollz/progressbar/v3"
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

	kp, err := kem.GenerateNewKeyPair()
	if err != nil {
		panic(err)
	}
	spew.Dump(kp)

}
