package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"math/big"
	"sync"
	"time"
)

var (
	readjustInterval  = 50                       // 50 blocks every...
	readjustTarget, _ = time.ParseDuration("1m") // 1 minute
)

// simple block type, used for demonstration. No transactions.
type block struct {
	PrevBlockHash [sha256.Size]byte
	Timestamp     time.Time
	Work          uint64
	Nonce         []byte
	Difficulty    uint64
}

// bytes() serializes the block data into a byte slice for hashing
func (b *block) bytes() []byte {
	buf := new(bytes.Buffer)

	buf.Write(b.PrevBlockHash[:])
	tstampBytes, err := b.Timestamp.MarshalBinary()
	if err != nil {
		panic(err)
	}
	buf.Write(tstampBytes)
	workbuf := make([]byte, 8)
	binary.LittleEndian.PutUint64(workbuf, b.Work)
	buf.Write(workbuf)
	diffBuf := make([]byte, 8)
	binary.LittleEndian.PutUint64(diffBuf, b.Difficulty)
	buf.Write(diffBuf)

	return buf.Bytes()
}

// simple chain type, tracks a slice of blocks.
type chain struct {
	startingDifficulty uint64
	mu                 sync.Mutex
	blocks             []block
}

// a simplistic difficulty adjustment algorithm
//
// difficulty = difficulty_1_target / current_target
func readjustDiff(oldDiff uint64, c *chain) uint64 {
	blocksSince := c.blocks[len(c.blocks)-readjustInterval:]

	var timeSinceLastAdjust int64
	for i := 0; i < len(blocksSince)-1; i++ {
		timeSinceLastAdjust += blocksSince[i+1].Timestamp.Unix() - blocksSince[i].Timestamp.Unix()
	}
	if timeSinceLastAdjust == 0 {
		return oldDiff / 2
	}
	retargetRatio := float64(timeSinceLastAdjust) / readjustTarget.Seconds()

	return uint64(float64(oldDiff) * retargetRatio)
}

func (c *chain) work() *big.Int {
	c2Work := new(big.Int)
	for _, b := range c.blocks {
		c2Work.Add(c2Work, new(big.Int).SetUint64(b.Work))
	}
	return c2Work
}

// mine grinds a nonce until we find a SHA256 hash that is below the current
// difficultyTarget. `throttle` throttles the rate at which we generate
// candidate hashes, and can effectively be used to control the hashrate.
func (c *chain) mine(throttle time.Duration) ([]byte, uint64) {
	// work is the number of hashes
	work := uint64(0)

	var prevBlockHash [sha256.Size]byte
	difficultyTarget := c.startingDifficulty
	if len(c.blocks) != 0 {
		prevBlockHash = sha256.Sum256(c.blocks[len(c.blocks)-1].bytes())
		difficultyTarget = c.blocks[len(c.blocks)-1].Difficulty
	}

	// readjust difficulty if needed
	if len(c.blocks)%readjustInterval == 0 && len(c.blocks) != 0 {
		difficultyTarget = readjustDiff(difficultyTarget, c)
	}

	for {
		if throttle != time.Duration(0) {
			time.Sleep(throttle)
		}

		nonce := make([]byte, 16)
		rand.Read(nonce)
		blockCandidate := block{
			Timestamp:     time.Now(),
			Nonce:         nonce,
			Work:          work,
			PrevBlockHash: prevBlockHash,
			Difficulty:    difficultyTarget,
		}
		hash := sha256.Sum256(blockCandidate.bytes())

		hashVal := binary.LittleEndian.Uint64(hash[:8])
		if hashVal < difficultyTarget {
			// found block, add it to the chain and return
			c.mu.Lock()
			c.blocks = append(c.blocks, blockCandidate)
			c.mu.Unlock()
			return nonce, work
		}
		work++
	}
}

func (c *chain) getBlocks() []block {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.blocks
}
