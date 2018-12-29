package chain

import (
	"log"
	"sync"

	"github.com/omzmarlon/blockfs/bclib"
)

var opSalary int

var noopSalary int

var createFileCost int

var createVerifyNum int

var appendVerifyNum int

var opDifficulty int

var noopDifficulty int

type BlockChainManager struct {
	RwLock               sync.RWMutex
	MapLock              sync.RWMutex
	BlockRoot            bclib.Block
	unaddedBlockChans    map[string]chan bclib.Block
	unverifiedBlockChans map[string]chan bool
	retryBlockChann      chan bclib.Block
}

func InitBlockChainManager(minerID string, ogHash string, opBlockSalary int, noopBlockSalary int, createCost int, createNum int, appendNum int, opD int, noopD int, newBlockChannel chan bclib.Block) *BlockChainManager {
	blockRoot := bclib.NewBlock(ogHash, "", "", nil, 0)

	opSalary = opBlockSalary
	noopSalary = noopBlockSalary
	createFileCost = createCost
	createVerifyNum = createNum
	appendVerifyNum = appendNum
	opDifficulty = opD
	noopDifficulty = noopD

	blockChainManager := BlockChainManager{sync.RWMutex{}, sync.RWMutex{}, blockRoot, map[string]chan bclib.Block{}, map[string]chan bool{}, make(chan bclib.Block, 30)}
	go serveNewBlock(minerID, newBlockChannel, &blockChainManager)

	return &blockChainManager
}

func serveNewBlock(minerID string, newBlockChan chan bclib.Block, bcmInstance *BlockChainManager) {
	for {
		var newBlock bclib.Block
		select {
		case newBlock = <-bcmInstance.retryBlockChann:
			log.Println("ChainManager: Got new block from retry chan")
		case newBlock = <-newBlockChan:
			log.Println("ChainManager: Got new block from new input chan")
		}
		log.Printf("ChainManager: Processing new block. newBlockChann now has length: %d. Retry chan now has length: %d\n", len(newBlockChan), len(bcmInstance.retryBlockChann))
		bcmInstance.RwLock.Lock()
		log.Printf("ChainManager: Write lock acquired. adding %s\n", newBlock.Hash)

		if !bcmInstance.UpdateBlockChain(newBlock) {
			bcmInstance.retryBlockChann <- newBlock
		}

		PrintBlockChain(minerID, bcmInstance.BlockRoot)
		bcmInstance.RwLock.Unlock()
		go bcmInstance.notifyBcUpdate(newBlock)
		// isValidBlock := bcmInstance.isValidBlock(newBlock)
		// if isValidBlock {
		// 	bcmInstance.UpdateBlockChain(newBlock)
		// 	bcmInstance.RwLock.Unlock()
		// 	PrintBlockChain(minerID, bcmInstance.GetBlockRoot())
		// 	bcmInstance.notifyBcUpdate(newBlock)
		// } else {
		// 	bcmInstance.RwLock.Unlock()
		// 	log.Printf("ChainManager:WARNING WARNING WARNING %s block is not valid\n", newBlock.Hash)
		// }
	}
}

func (bcm *BlockChainManager) SetBlockRoot(block bclib.Block) { bcm.BlockRoot = block }

func (bcm *BlockChainManager) UpdateBlockChain(newBlock bclib.Block) bool {
	return UpdateBlockChildren(bcm, newBlock)
}

func (bcm *BlockChainManager) MergeChain(localGenisis string, remote bclib.Block) {
	MergeRoots(bcm, localGenisis, remote)
}

func (bcm *BlockChainManager) notifyBcUpdate(newBlock bclib.Block) {
	bcm.MapLock.RLock()
	for _, v := range bcm.unaddedBlockChans {
		v <- newBlock
	}

	for _, v := range bcm.unverifiedBlockChans {
		v <- true
	}
	bcm.MapLock.RUnlock()
}

func (bcm *BlockChainManager) GenUnaddedBlockChan(opId string) <-chan bclib.Block {
	bcm.MapLock.Lock()
	c := make(chan bclib.Block, 5)
	bcm.unaddedBlockChans[opId] = c
	bcm.MapLock.Unlock()
	return c
}

func (bcm *BlockChainManager) GenUnverifiedBlockChan(opId string) <-chan bool {
	bcm.MapLock.Lock()
	c := make(chan bool, 5)
	bcm.unverifiedBlockChans[opId] = c
	bcm.MapLock.Unlock()
	return c
}

func (bcm *BlockChainManager) RemoveUnaddedBlockChan(opId string) {
	bcm.MapLock.Lock()
	delete(bcm.unaddedBlockChans, opId)
	bcm.MapLock.Unlock()
}

func (bcm *BlockChainManager) RemoveUnverifiedChan(opId string) {
	bcm.MapLock.Lock()
	delete(bcm.unverifiedBlockChans, opId)
	bcm.MapLock.Unlock()
}

func (bcm *BlockChainManager) isValidBlock(block bclib.Block) bool {
	return checkPOW(block) && bcm.checkPrevBlock(block) && bcm.checkSufficientCoin(block) && bcm.checkRfsSemantic(block)
}

func checkPOW(block bclib.Block) bool {
	hash := ComputeHashString(block.PrevHash, block.MinerID, block.Nonce, block.Ops)
	if CheckDifficultyReached(hash, block.Ops, opDifficulty, noopDifficulty) {
		return true
	}
	return false
}

func (bcm *BlockChainManager) checkPrevBlock(block bclib.Block) bool {
	return CheckBlockExists(&bcm.BlockRoot, block.PrevHash)
}

func (bcm *BlockChainManager) checkSufficientCoin(block bclib.Block) bool {
	if len(block.Ops) > 0 {
		// bcm.RwLock.RLock()
		// defer bcm.RwLock.RUnlock()
		return CountCoins(&bcm.BlockRoot, opSalary, noopSalary, createFileCost, block.MinerID) >= 0
	}

	return true
}

func (bcm BlockChainManager) checkRfsSemantic(block bclib.Block) bool {
	//create: check that the file has not been created before
	//append: the file has been created before
	//BUT we have to check for all ops in this block

	log.Printf("ChainManager: checking RFS semantics for block: %s\n", block.Hash)
	result := false

	longestChain := FindLongestChain(&bcm.BlockRoot)

	allFiles := GetAllFileNamesOnChain(&bcm.BlockRoot, longestChain)

	for _, op := range block.Ops {
		if op.OpAction == bclib.CREATE {
			if Contains(allFiles, op.Filename) {
				result = false
				log.Printf("ChainManager: block: %s RFS semantics check done, status: %b\n", block.Hash, result)
				return result
			}
		}

		if op.OpAction == bclib.APPEND {
			if !Contains(allFiles, op.Filename) {
				result = false
				log.Printf("ChainManager: block: %s RFS semantics check done, status: %b\n", block.Hash, result)
				return result
			}

		}
	}
	result = true
	log.Printf("ChainManager: block: %s RFS semantics check done, status: %b\n", block.Hash, result)
	return result
}
