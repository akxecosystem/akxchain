package chain

import (
	"fmt"
	"time"
)

type Akx interface {
}

type akx struct {
}

func InitBlockchain() {
	// Create instances of the main blockchain, sidechain, and validator
	bc := &MainBlockchain{}
	sc := &SidechainImpl{}
	v := &ValidatorImpl{}

	// Add transactions to the sidechain
	tx1 := &MyTransaction{"hash1", "sender1", "receiver1", 100}
	tx2 := &MyTransaction{"hash2", "sender2", "receiver2", 200}
	sc.AddTransaction(tx1)
	sc.AddTransaction(tx2)

	// Create a block and add it to the main blockchain
	block := &MyBlock{
		"hash",
		"parent_hash",
		time.Now(),
		12345,
		6789,
		[]Transaction{tx1, tx2},
	}
	bc.AddBlock(block)

	// Validate the block using the validator
	isValidBlock := v.ValidateBlock(block)
	fmt.Println("Is block valid?", isValidBlock)

	// Validate the transactions using the validator
	isValidTx1 := v.ValidateTransaction(tx1)
	fmt.Println("Is tx1 valid?", isValidTx1)
	isValidTx2 := v.ValidateTransaction(tx2)
	fmt.Println("Is tx2 valid?", isValidTx2)
}
