package main

import (
	"fmt"
	"os"
	"time"
)

// demonstrate a timewarping vulnerability: timestamp manipulation allows for
// manipulation of the mining difficulty. It is very important to validate
// timestamps.
func timewarp() {
	// create a chain, mine some blocks on it
	startingDifficulty := uint64(100000000000000)

	c := &chain{
		startingDifficulty: startingDifficulty,
	}

	for i := 0; i < readjustInterval-1; i++ {
		fmt.Println("mined block", len(c.blocks))
		c.mine(0)
		if i%5 == 0 {
			c.blocks[len(c.blocks)-1].Timestamp = c.blocks[len(c.blocks)-1].Timestamp.Add(time.Hour * 10)
			fmt.Println("inserted evil timestamp")
		}
	}

	fmt.Println("inserted a few crafted timestamps, we should be able to mine a ton of blocks now")
	for i := 5; i > 0; i-- {
		time.Sleep(time.Second)
		fmt.Printf("%v...\n", i)
	}
	time.Sleep(time.Second)

	for i := 0; i < 500; i++ {
		fmt.Println("mined block", len(c.blocks))
		c.mine(0)
	}
}

// demonstrate the property that the longest chain is not necessarily the chain
// with the most work
func longestChain() {
	startingDifficulty := uint64(1000000000000000)

	c1 := &chain{
		startingDifficulty: startingDifficulty,
	}
	c2 := &chain{
		startingDifficulty: startingDifficulty,
	}

	// honest miner, doing the maximum amount of work possible, consider this
	// analogous to the 'main mining network'
	go func() {
		for {
			c1.mine(0)
		}
	}()

	// evil miner, mining slowly isolated from the main chain but increasing
	// hashrate after each difficulty adjustment period.
	go func() {
		throttle := time.Microsecond * 5
		for {
			c2.mine(throttle)
			if len(c2.blocks)%readjustInterval == 0 {
				fmt.Println("increasing attacker hashrate")
				throttle /= 4
			}
		}
	}()

	for {
		time.Sleep(time.Second)
		attackerBlocks := c2.getBlocks()
		mainBlocks := c1.getBlocks()
		fmt.Println("main at block: ", len(mainBlocks))
		fmt.Println("attacker at block: ", len(attackerBlocks))
		if len(attackerBlocks) > len(mainBlocks) {
			fmt.Print("ATTACK SUCCESS")

			fmt.Printf(`
----
attacker length: %v,
attacker work: %v,

main length: %v,
main work: %v,
---`, len(attackerBlocks), c2.work().String(), len(mainBlocks), c1.work().String())

			return
		}
	}
}
func main() {
	if len(os.Args) != 2 {
		fmt.Println("usage: pow-attacks [timewarp/chain]")
		os.Exit(-1)
	}
	if os.Args[1] == "timewarp" {
		timewarp()
	}
	if os.Args[1] == "chain" {
		longestChain()
	}
}
