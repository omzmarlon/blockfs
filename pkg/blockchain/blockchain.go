package blockchain

import (
	"fmt"
	"log"
	"sync"

	"github.com/Workiva/go-datastructures/queue"
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
	chainRWLock sync.RWMutex // Mutexes usually work without pointers

	// TODO
	// read/write lock when accessing data?

}

// Conf - configurations for the blockchain
type Conf struct {
	GenesisHash    string
	MinerID        string
	OpDiffculty    int
	NoopDifficulty int
}

// AppendBlockResult - result code for trying to append new block to blockchain
type AppendBlockResult int

const (
	APPEND_RESULT_UNEXPECTED_ERR   AppendBlockResult = 0
	APPEND_RESULT_SUCCESS          AppendBlockResult = 1
	APPEND_RESULT_NOT_FOUND        AppendBlockResult = 2
	APPEND_RESULT_INVALID_BLOCK    AppendBlockResult = 3
	APPEND_RESULT_INVALID_SEMANTIC AppendBlockResult = 4
	APPEND_RESULT_DUPLICATE        AppendBlockResult = 5
)

// New - initialize the blockchain
func New(conf Conf) *Blockchain {
	ops := make([]domain.Op, 0)
	children := make([]*domain.Block, 0)
	genesis := &domain.Block{
		Hash:     conf.GenesisHash,
		PrevHash: "",
		MinerID:  conf.MinerID,
		Ops:      ops,
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
	blockchain.chainRWLock.Lock()
	defer blockchain.chainRWLock.Unlock()
	if !blockchain.verifyBlock(block) {
		return APPEND_RESULT_INVALID_BLOCK
	}
	q := queue.New(10)
	q.Put(blockchain.genesis)
	for q.Len() != 0 {
		res, err := q.Get(1)
		if err != nil {
			log.Fatalf("Append block failed unexpected with err: %s", err)
			return APPEND_RESULT_UNEXPECTED_ERR
		}
		curr := res[0].(*domain.Block)
		if curr.Hash == block.PrevHash {
			for _, child := range *curr.Children {
				if child.Hash == block.Hash {
					return APPEND_RESULT_DUPLICATE
				}
			}
			// TODO blockfs semantic check before appending
			*curr.Children = append(*curr.Children, block)
			blockchain.PrintBlockchain()
			return APPEND_RESULT_SUCCESS
		}
		for _, child := range *curr.Children {
			q.Put(child)
		}

	}
	return APPEND_RESULT_NOT_FOUND
}

func (blockchain *Blockchain) verifyBlock(block *domain.Block) bool {
	hash := util.ComputeBlockHash(*block)
	if hash != block.Hash {
		return false
	}
	difficulty := blockchain.Conf.OpDiffculty
	if len(block.Ops) == 0 {
		difficulty = blockchain.Conf.NoopDifficulty
	}
	if !util.IsDifficultyReached(hash, difficulty) {
		return false
	}
	if len(*block.Children) == 0 {
		return true
	}
	for _, child := range *block.Children {
		if !blockchain.verifyBlock(child) {
			return false
		}

	}
	return true
}

// GetBlockHash returns the hash of the last block on the longest chain
func (blockchain *Blockchain) GetBlockHash() string {
	blockchain.chainRWLock.RLock()
	defer blockchain.chainRWLock.RUnlock()
	if len(*blockchain.genesis.Children) == 0 {
		return blockchain.genesis.Hash
	}
	_, hash := blockchain.getBlockHashHelper(blockchain.genesis, 1)
	return hash
}

func (blockchain *Blockchain) getBlockHashHelper(block *domain.Block, currDepth int) (int, string) {
	if len(*block.Children) == 0 {
		return currDepth, block.Hash
	}
	max := currDepth
	maxhash := ""
	for _, child := range *block.Children {
		depth, hash := blockchain.getBlockHashHelper(child, currDepth+1)
		if depth > max {
			max = depth
			maxhash = hash
		}
	}
	return max, maxhash
}

// TODO: fix it
func (blockchain *Blockchain) PrintBlockchain() {
	fmt.Println("Printing blockchain...")
	defer fmt.Println("Printing blockchain done...")
	q := queue.New(10)
	q.Put(blockchain.genesis)
	q.Put(&domain.Block{Hash: "placeholder"})
	for q.Len() != 0 {
		res, err := q.Get(1)
		if err != nil {
			log.Fatalf("Print blockchain failed unexpected with err: %s", err)
			return
		}
		curr := res[0].(*domain.Block)

		if curr.Hash == "placeholder" && q.Len() == 0 {
			fmt.Print("\n")
			return
		} else if curr.Hash == "placeholder" {
			q.Put(&domain.Block{Hash: "placeholder"})
			fmt.Print("\n")
		} else {
			fmt.Print(curr.String())
			for _, child := range *curr.Children {
				q.Put(child)
			}
		}

	}
}

// TODO:
// printing the entire chain
