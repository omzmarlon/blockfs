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
		Nonce:    0,
		Children: &children,
		Metadata: domain.Metadata{
			LongestChainLength: 1,
			Parent:             nil,
		},
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
	blockchain.preprocessBlockMetadata(block)
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
			// preserving some metadata to assist faster lookup
			block.Metadata.Parent = curr
			blockchain.updateParentsMetaHelper(block.Metadata.Parent, uint64(block.Metadata.LongestChainLength))
			blockchain.PrintBlockchain()
			return APPEND_RESULT_SUCCESS
		}
		for _, child := range *curr.Children {
			q.Put(child)
		}

	}
	return APPEND_RESULT_NOT_FOUND
}

func (blockchain *Blockchain) updateParentsMetaHelper(start *domain.Block, subChainLen uint64) {
	if subChainLen+1 > start.Metadata.LongestChainLength {
		start.Metadata.LongestChainLength = subChainLen + 1
	}
	if start.Metadata.Parent != nil {
		blockchain.updateParentsMetaHelper(start.Metadata.Parent, start.Metadata.LongestChainLength)
	}
}

func (blockchain *Blockchain) preprocessBlockMetadata(block *domain.Block) {
	if len(*block.Children) == 0 {
		block.Metadata.LongestChainLength = 1
		return
	}
	length := blockchain.chainLengthHelper(block)
	block.Metadata.LongestChainLength = length
}

func (blockchain *Blockchain) chainLengthHelper(block *domain.Block) uint64 {
	max := uint64(0)
	for _, child := range *block.Children {
		depth := blockchain.chainLengthHelper(child)
		child.Metadata.Parent = block
		child.Metadata.LongestChainLength = depth
		if depth > max {
			max = depth
		}
	}
	return max + 1
}

// CloneChain - make a complete copy of the current local blockchain
func (blockchain *Blockchain) CloneChain() domain.Block {
	blockchain.chainRWLock.RLock()
	defer blockchain.chainRWLock.RUnlock()
	result := domain.Block{
		Hash:     blockchain.Conf.GenesisHash,
		PrevHash: "",
		MinerID:  blockchain.Conf.MinerID,
		Ops:      make([]domain.Op, 0),
		Nonce:    0,
	}
	children := make([]*domain.Block, len(*blockchain.genesis.Children))
	if len(*blockchain.genesis.Children) == 0 {
		result.Children = &children
		return result
	}
	for i := 0; i < len(*blockchain.genesis.Children); i++ {
		child := blockchain.cloneChainHelper(*(*blockchain.genesis.Children)[i])
		children[i] = &child
	}
	result.Children = &children
	return result
}

func (blockchain *Blockchain) cloneChainHelper(root domain.Block) domain.Block {
	ops := make([]domain.Op, len(root.Ops))
	for i := 0; i < len(root.Ops); i++ {
		ops[i] = domain.NewOp(root.Ops[i].OpID, root.Ops[i].MinerID,
			root.Ops[i].OpAction, root.Ops[i].Filename, root.Ops[i].Record)
	}
	result := domain.Block{
		Hash:     root.Hash,
		PrevHash: root.PrevHash,
		MinerID:  root.MinerID,
		Ops:      ops,
		Nonce:    root.Nonce,
		// Children *[]*Block
	}
	children := make([]*domain.Block, len(*root.Children))
	if len(*root.Children) == 0 {
		result.Children = &children
		return result
	}
	for i := 0; i < len(*root.Children); i++ {
		child := blockchain.cloneChainHelper(*(*root.Children)[i])
		children[i] = &child
	}
	result.Children = &children
	return result
}

// GetData returns all the ops on the longest chain
func (blockchain *Blockchain) GetData() [][]domain.Op {
	blockchain.chainRWLock.RLock()
	defer blockchain.chainRWLock.RUnlock()
	result := make([][]domain.Op, 0)
	curr := blockchain.genesis
	for {
		result = append(result, curr.Ops)
		if len(*curr.Children) == 0 {
			break
		}
		curr = (*curr.Children)[0]
		maxLen := curr.Metadata.LongestChainLength
		for _, child := range *curr.Children {
			if child.Metadata.LongestChainLength > maxLen {
				curr = child
			}
		}
	}
	return result
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
