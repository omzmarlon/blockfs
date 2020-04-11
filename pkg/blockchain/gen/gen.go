package gen

import (
	"log"
	"time"

	"github.com/omzmarlon/blockfs/pkg/domain"
	"github.com/omzmarlon/blockfs/pkg/util"
)

// GenerateBlock generates a new block satisfying provided difficulty
func GenerateBlock(prevHash string, minerID string, ops []domain.Op,
	opDifficulty int, noopDifficulty int) *domain.Block {
	log.Printf("[gen]: Generating the a new block...")
	start := time.Now().Unix()
	children := make([]*domain.Block, 0)
	block := &domain.Block{
		Hash:     "",
		PrevHash: prevHash,
		MinerID:  minerID,
		Ops:      ops,
		Nonce:    0,
		Children: &children,
	}

	var difficulty int
	if len(ops) == 0 {
		difficulty = noopDifficulty
	} else {
		difficulty = opDifficulty
	}

	for !util.IsDifficultyReached(util.ComputeBlockHash(*block), difficulty) {
		block.Nonce = util.RandomNonce()
	}
	block.Hash = util.ComputeBlockHash(*block)

	log.Printf("New block generated! Hash: %s. Took %d seconds", block.Hash, time.Now().Unix()-start)
	return block
}
