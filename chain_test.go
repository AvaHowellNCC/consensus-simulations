package main

import (
	"testing"
	"time"
)

func TestDifficultyAdjustment(t *testing.T) {
	testingInterval, _ := time.ParseDuration("15s")
	var testingChain chain
	for i := 0; i < readjustInterval; i++ {
		testingChain.blocks = append(testingChain.blocks, block{
			Timestamp: time.Now().Add(testingInterval * time.Duration(i)),
		})
	}

	oldDiff := uint64(1000)
	t.Log(readjustDiff(oldDiff, &testingChain))
}
