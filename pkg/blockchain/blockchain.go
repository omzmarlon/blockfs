package blockchain

import (
	"log"
	"sync"

	"github.com/omzmarlon/blockfs/pkg/domain"
	"github.com/omzmarlon/blockfs/pkg/util"
)

// TODO:
// also need coin count
// this class performs the block legitimacy check
// ideally do not run any daemon in this class, it is just a data structure to be changed by other class's daemons

// Blockchain - the blockchain hold on this miner
type Blockchain struct {
	Conf        Conf
	genesis     *domain.Block
	ChainRWLock sync.RWMutex // Mutexes usually work without pointers

	// TODO
	// read/write lock when accessing data?

}

// Conf - configurations for the blockchain
type Conf struct {
	GenesisHash    string
	MinerID        string
	OpDiffculty    uint8
	NoopDifficulty uint8
}

// AppendBlockResult - result code for trying to append new block to blockchain
type AppendBlockResult int

const (
	APPEND_RESULT_SUCCESS          AppendBlockResult = 0
	APPEND_RESULT_NOT_FOUND        AppendBlockResult = 1
	APPEND_RESULT_INVALID_BLOCK    AppendBlockResult = 2
	APPEND_RESULT_INVALID_SEMANTIC AppendBlockResult = 3
)

// New - initialize the blockchain
func New(conf Conf) *Blockchain {
	ops := make([]domain.Op, 0)
	children := make([]domain.Block, 0)
	genesis := &domain.Block{
		Hash:     conf.GenesisHash,
		PrevHash: "",
		MinerID:  conf.MinerID,
		Ops:      &ops,
		Nonce:    util.RandomNonce(),
		Children: &children,
	}
	ret := &Blockchain{
		Conf:    conf,
		genesis: genesis,
	}
	log.Printf("Genesis block generated: %v", ret)
	return ret
}

// AppendBlock appends a new block to the blockchain
func (blockchain *Blockchain) AppendBlock(block *domain.Block) AppendBlockResult {
	// TODO

	// verify
	// find the place to append
	// blockfs semantic check
	// return codes for: success, could not find prev block, invalid block
	return APPEND_RESULT_SUCCESS
}

// GetBlockHash returns the hash of the last block on the longest chain
func (blockchain *Blockchain) GetBlockHash() string {
	// TODO
	if len(*blockchain.genesis.Children) == 0 {
		return blockchain.genesis.Hash
	}
	return ""
}

// TODO:
// method for finding the last block on the longest chain
// method for locating a block
