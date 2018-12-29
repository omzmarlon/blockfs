package chain

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/omzmarlon/blockfs/bclib"
)

type bfsObject struct {
	hashes []string
	block  *bclib.Block
}

func newBfsObject(hashes []string, block *bclib.Block) bfsObject {
	return bfsObject{
		hashes: hashes,
		block:  block,
	}
}

const appendCost = 1

//bfs and returns a list of
//only returns the 1 longest chain
func FindLongestChain(blockRoot *bclib.Block) []string {
	// fmt.Println("###########Trying to find the longest chain##############")
	candidates := make(map[int][]string)
	candidates = findLongestChainHelper(*blockRoot, make([]string, 0), candidates)
	max := -1
	for k := range candidates {
		if k > max {
			max = k
		}
	}
	// fmt.Println("###########done finding the longest chain##############")
	// fmt.Println(candidates)
	// fmt.Println(len(candidates[max]))
	return candidates[max][1:]
}

func findLongestChainHelper(blockRoot bclib.Block, statesss []string, candidatesss map[int][]string) map[int][]string {
	candidates := make(map[int][]string)
	for k, v := range candidatesss {
		copyy := make([]string, len(v))
		for i, s := range v {
			copyy[i] = s
		}
		candidates[k] = copyy
	}

	// fmt.Print("lalala:   ")
	// for k, v := range candidates {
	// 	fmt.Print("[")
	// 	fmt.Print(k)
	// 	fmt.Print(", [")
	// 	for _, s := range v {
	// 		if len(s) > 5 {
	// 			fmt.Print(s[:5] + "..., ")
	// 		} else {
	// 			fmt.Print(s + "..., ")
	// 		}
	// 	}
	// 	fmt.Println("]")
	// }
	// fmt.Println("]")

	states := make([]string, 0)
	for _, s := range statesss {
		states = append(states, s)
	}

	states = append(states, blockRoot.Hash)
	if len(blockRoot.Children) == 0 {
		// fmt.Printf("This guy has no children: %s\n", blockRoot.Hash)
		// fmt.Println(len(states))
		// fmt.Println(states)
		max := 0
		for k := range candidates {
			if k > max {
				max = k
			}
		}
		// fmt.Println("Curr Max is ", max)
		if len(states) > max {
			// fmt.Println("added")
			copy := make([]string, 0)
			for _, s := range states {
				copy = append(copy, s)
			}
			candidates[len(states)] = copy
			// candidates[len(states)] = states
		}
		// for k, v := range candidates {
		// 	fmt.Print("[")
		// 	fmt.Print(k)
		// 	fmt.Print(", [")
		// 	for _, s := range v {
		// 		if len(s) > 5 {
		// 			fmt.Print(s[:5] + "..., ")
		// 		} else {
		// 			fmt.Print(s + "..., ")
		// 		}
		// 	}
		// 	fmt.Println("]")
		// }
		// fmt.Println("]")
		return candidates
	}
	for i := range blockRoot.Children {
		child := blockRoot.Children[i]
		candidates = findLongestChainHelper(child, states, candidates)
	}
	return candidates
}

func DepthAtBlock(blockRoot *bclib.Block, blockHash string) int {
	var blockQueue []*bclib.Block

	if blockRoot.Hash == blockHash {
		longestChain := FindLongestChain(blockRoot)
		return len(longestChain)
	}

	for i := range blockRoot.Children {
		blockQueue = enqueueBlock(&blockRoot.Children[i], blockQueue)
	}

	for len(blockQueue) > 0 {

		currBlock, newQueue := dequeueBlock(blockQueue)

		if currBlock.Hash == blockHash {
			longestChain := FindLongestChain(currBlock)
			return len(longestChain)
		}

		currChildren := currBlock.Children
		blockQueue = newQueue

		for i := range currChildren {
			blockQueue = enqueueBlock(&currChildren[i], blockQueue)
		}
	}

	return 0
}

func CheckRfsSemanticsOnChain(blockRoot *bclib.Block, longestChain []string, ops []bclib.ROp) bool {

	allFiles := GetAllFileNamesOnChain(blockRoot, longestChain)
	for _, op := range ops {
		if op.OpAction == bclib.CREATE {
			if Contains(allFiles, op.Filename) {
				return false
			}
		} else {
			if !Contains(allFiles, op.Filename) {
				return false
			}
		}
	}
	return true
}

func FindLongestChains(blockRoot *bclib.Block) [][]string {

	var blockQueue []*bfsObject
	var currLongestChain []string
	var longestChains [][]string
	//deal with the root

	for i := range blockRoot.Children {

		elToQueue := newBfsObject([]string{}, &blockRoot.Children[i])
		blockQueue = enqueueBFSObj(&elToQueue, blockQueue)

	}

	for len(blockQueue) > 0 {

		currBfsObj, newQueue := dequeueBFSObj(blockQueue)
		currChildren := currBfsObj.block.Children
		blockQueue = newQueue

		if len(currChildren) < 1 {

			currHash := append(currBfsObj.hashes, currBfsObj.block.Hash)

			if len(currHash) > len(currLongestChain) {
				longestChains = [][]string{currHash}
				currLongestChain = currHash
			} else if len(currHash) == len(currLongestChain) {
				longestChains = append(longestChains, currHash)
			}
		}

		blocks := currBfsObj.block.Children
		currBlockHash := currBfsObj.block.Hash
		currBlockPath := currBfsObj.hashes

		for i := range blocks {

			elToQueue := newBfsObject(append(currBlockPath, currBlockHash), &blocks[i])
			blockQueue = enqueueBFSObj(&elToQueue, blockQueue)

		}
	}

	return longestChains
}

func PrintBlockChain(minerID string, blockRoot bclib.Block) {
	fmt.Println("***** start of blockchain *****")

	var blockQueue []*bclib.Block
	stringToPrint := ""
	var currLevel int
	var nextLevel int
	prevHash := ""
	//deal with the root

	fmt.Println(blockRoot.Hash)

	blockQueue = queueChildrenBlocks(blockRoot.Children, blockQueue)

	currLevel = len(blockRoot.Children)

	for len(blockQueue) > 0 {
		if currLevel == 0 {
			fmt.Println(stringToPrint)
			stringToPrint = ""
			currLevel = nextLevel
			nextLevel = 0
			prevHash = ""
		}

		currLevel--

		currBlock, newQueue := dequeueBlock(blockQueue)

		if prevHash == "" {
			prevHash = currBlock.PrevHash
		}

		if prevHash != currBlock.PrevHash {
			stringToPrint += "|  "
			prevHash = currBlock.PrevHash
		}

		theHash := ""
		if len(currBlock.Hash) > 5 {
			theHash = currBlock.Hash[:5] + "..."
		} else {
			theHash = currBlock.Hash
		}

		stringToPrint += currBlock.MinerID + ":" + theHash

		if len(currBlock.PrevHash) > 5 {
			stringToPrint += "(Prev: " + currBlock.PrevHash[:5] + "...)"
		} else {
			stringToPrint += "(Prev: " + currBlock.PrevHash + "...)"
		}

		stringToPrint += ":" + "["
		for _, op := range currBlock.Ops {
			stringToPrint += op.OpID + ":" + strconv.Itoa(int(op.OpAction))
		}
		stringToPrint += "] + "

		currChildren := currBlock.Children
		blockQueue = newQueue

		if len(currChildren) > 0 {
			blockQueue = queueChildrenBlocks(currChildren, blockQueue)
			nextLevel += len(currChildren)
		}
	}

	fmt.Println(stringToPrint)

	fmt.Println("***** end of blockchain *****")
}

func queueChildren(bfsEl bfsObject, queue []*bfsObject) []*bfsObject {

	blocks := bfsEl.block.Children
	currBlockHash := bfsEl.block.Hash
	currBlockPath := bfsEl.hashes

	for i := range blocks {

		elToQueue := newBfsObject(append(currBlockPath, currBlockHash), &blocks[i])
		queue = enqueueBFSObj(&elToQueue, queue)

	}

	return queue
}

func enqueueBFSObj(obj *bfsObject, queue []*bfsObject) []*bfsObject {
	return append(queue, obj)
}

func dequeueBFSObj(queue []*bfsObject) (*bfsObject, []*bfsObject) {
	el := queue[0]
	return el, queue[1:]
}

func enqueueBlock(block *bclib.Block, queue []*bclib.Block) []*bclib.Block {
	return append(queue, block)
}

func dequeueBlock(queue []*bclib.Block) (*bclib.Block, []*bclib.Block) {
	el := queue[0]
	return el, queue[1:]
}

func queueChildrenBlocks(blocks []bclib.Block, queue []*bclib.Block) []*bclib.Block {
	for i := range blocks {
		queue = append(queue, &blocks[i])
	}

	return queue
}

//function for finding a block with a certain hash(prevHash) in the block chain
//AND THEN adding the newBlock to that block
func UpdateBlockChildren(bcmInstance *BlockChainManager, newBlock bclib.Block) bool {
	var blockQueue []*bclib.Block
	prevHash := newBlock.PrevHash
	blockRoot := &bcmInstance.BlockRoot

	if blockRoot.Hash == prevHash {
		exits := false
		for _, b := range blockRoot.Children {
			if b.Hash == newBlock.Hash {
				exits = true
			}
		}

		if !exits {
			blockRoot.Children = append(blockRoot.Children, bclib.NewBlock(newBlock.Hash, newBlock.PrevHash, newBlock.MinerID, newBlock.Ops, newBlock.Nonce))
			log.Println("ChainManager: Attached ", newBlock.Hash, " to parent", blockRoot.Hash)
			return true
		} else {
			log.Println("ChainManager: ", newBlock.Hash, " already exits for "+blockRoot.Hash+", skipping...")
			return true
		}
	}

	for i := range blockRoot.Children {
		blockQueue = enqueueBlock(&blockRoot.Children[i], blockQueue)
	}

	for len(blockQueue) > 0 {

		currBlock, newQueue := dequeueBlock(blockQueue)

		if currBlock.Hash == prevHash {
			exits := false
			for _, b := range currBlock.Children {
				if b.Hash == newBlock.Hash {
					exits = true
				}
			}

			if !exits {
				currBlock.Children = append(currBlock.Children, bclib.NewBlock(newBlock.Hash, newBlock.PrevHash, newBlock.MinerID, newBlock.Ops, newBlock.Nonce))
				log.Println("ChainManager: Attached ", newBlock.Hash, " to parent", currBlock.Hash)
				return true
			} else {
				log.Println("ChainManager: ", newBlock.Hash, " already exits for "+currBlock.Hash+", skipping...")
				return true
			}
		}

		currChildren := currBlock.Children
		blockQueue = newQueue

		blockQueue = queueChildrenBlocks(currChildren, blockQueue)
	}
	log.Println("ChainManager: Did not find parent for: ", newBlock.Hash)
	return false
}

func MergeRoots(bcmInstance *BlockChainManager, localGenisis string, remote bclib.Block) {
	if localGenisis != remote.Hash {
		log.Printf("WARNING WARNING WARNING different genisis block local: %s, remote: %s", localGenisis, remote.Hash)
		return
	}
	var blockQueue []*bclib.Block
	currBlock := &remote
	for i := range currBlock.Children {
		blockQueue = enqueueBlock(&currBlock.Children[i], blockQueue)
	}

	for len(blockQueue) > 0 {
		currBlock, blockQueue = dequeueBlock(blockQueue)
		for i := range currBlock.Children {
			blockQueue = enqueueBlock(&currBlock.Children[i], blockQueue)
		}
		UpdateBlockChildren(bcmInstance, bclib.NewBlock(currBlock.Hash, currBlock.PrevHash, currBlock.MinerID, currBlock.Ops, currBlock.Nonce))

	}
}

func CheckBlockExists(blockRoot *bclib.Block, blockHash string) bool {

	var blockQueue []*bclib.Block

	if blockRoot.Hash == blockHash {
		return true
	}

	for i := range blockRoot.Children {
		blockQueue = enqueueBlock(&blockRoot.Children[i], blockQueue)

	}

	for len(blockQueue) > 0 {

		currBlock, newQueue := dequeueBlock(blockQueue)

		if currBlock.Hash == blockHash {
			return true
		}

		currChildren := currBlock.Children
		blockQueue = newQueue

		for i := range currChildren {
			blockQueue = enqueueBlock(&currChildren[i], blockQueue)

		}
	}

	return false
}

func CountCoins(blockRoot *bclib.Block, opCoinSalary int, noopCoinSalary int, createOpCost int, minerID string) int {

	totalCoins := 0

	currBlock := blockRoot

	longestChain := FindLongestChain(blockRoot)

	for _, blockHash := range longestChain {
		childBlock, err := GetBlockFromChildren(currBlock.Children, blockHash)
		if err != nil {
			break
		}
		currOps := childBlock.Ops
		if childBlock.MinerID == minerID {
			if len(currOps) == 0 || currOps == nil {
				totalCoins += noopCoinSalary
			} else {
				totalCoins = totalCoins + countOpBlockCoins(currOps, minerID, createOpCost) + opCoinSalary
			}
		} else {
			//the current miner did not make this block but we still need to see if theres an op that this miner requested
			//in some other minders block
			totalCoins = totalCoins + countOpBlockCoins(currOps, minerID, createOpCost)
		}

		currBlock = childBlock
	}

	return totalCoins

}

func countOpBlockCoins(rops []bclib.ROp, minerID string, createCost int) int {

	totalDefficit := 0

	for _, op := range rops {
		if op.MinerID == minerID {
			if op.OpAction == bclib.CREATE {
				totalDefficit -= createCost
			} else {
				totalDefficit -= appendCost
			}
		}
	}

	return totalDefficit
}

//given a block and a string of the hash of one of the children, return that child
func GetBlockFromChildren(block []bclib.Block, blockID string) (*bclib.Block, error) {

	for i := range block {
		if block[i].Hash == blockID {
			return &block[i], nil
		}
	}
	// TODO this should never happen unless my code is bad
	// TODO: fail silently
	// panic("get block from children  " + blockID)
	log.Println("ChainManager: WARNING WARNING WARNING: get block from children panic " + blockID)
	return &bclib.Block{}, errors.New("e")
}

func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

//from the block chain, get all the files a list of file names
func GetAllFileNamesOnChain(rootBlock *bclib.Block, longestChain []string) []string {

	fileNames := map[string]bool{}
	currBlock := rootBlock

	for _, blockHash := range longestChain {
		// fmt.Printf("Trying to get block from children. Curr: %s, target in children: %s\n", currBlock.Hash, blockHash)
		childBlock, err := GetBlockFromChildren(currBlock.Children, blockHash)
		if err != nil {
			break
		}
		currBlock = childBlock
		childOps := childBlock.Ops

		fNames := listNames(childOps)

		for _, fname := range fNames {
			fileNames[fname] = true
		}
	}

	return getKeysFromMap(fileNames)
}

func CountRecNumWithFname(blockRoot *bclib.Block, fname string, createVerifyNum int, appendVerifyNum int) int {
	totalRecs := 0

	currBlock := blockRoot

	longestChain := FindLongestChain(blockRoot)

	for i, blockHash := range longestChain {

		blocksAfterCurr := len(longestChain) - i - 1

		childBlock, err := GetBlockFromChildren(currBlock.Children, blockHash)
		if err != nil {
			break
		}
		currOps := childBlock.Ops

		totalRecs += countRecInBlockOps(currOps, fname, createVerifyNum, appendVerifyNum, blocksAfterCurr)

		currBlock = childBlock
	}

	return totalRecs
}

func ReadRec(blockRoot *bclib.Block, fname string, recNum int, appendVerifyNum int) ([]byte, error) {
	var recordBytes []byte

	currBlock := blockRoot

	longestChain := FindLongestChain(blockRoot)

	counter := 0

	for i, blockHash := range longestChain {
		blocksAfterCurr := len(longestChain) - i - 1

		childBlock, err := GetBlockFromChildren(currBlock.Children, blockHash)
		if err != nil {
			break
		}
		currBlock = childBlock
		childOps := childBlock.Ops

		bytes, newCounter, read := getRecordBytesOfBlock(childOps, fname, recNum, counter, blocksAfterCurr, appendVerifyNum)
		counter = newCounter
		if newCounter == recNum && read {
			recordBytes = bytes
			break
		}
	}

	if counter != recNum {
		return []byte{}, errors.New("")
	}

	return recordBytes, nil
}

func JoinBlockChains(blockRoot1 bclib.Block, blockRoot2 bclib.Block) bclib.Block {
	block2Children := blockRoot2.Children

	for _, child := range block2Children {
		blockRoot1.Children = append(blockRoot1.Children, child)
	}

	return blockRoot1
}

func countRecInBlockOps(ops []bclib.ROp, fname string, createVerifyNum int, appendVerifyNum int, numChildren int) int {

	recs := 0

	for _, op := range ops {
		if op.Filename == fname {
			if op.OpAction == bclib.APPEND {
				if numChildren >= appendVerifyNum {
					recs++
				}
			}
		}
	}

	return recs
}

//given a list of ROPs, we take all their file names
func listNames(ops []bclib.ROp) []string {
	var fileNames []string
	for _, rop := range ops {
		currName := rop.Filename

		if len(currName) > 0 {
			fileNames = append(fileNames, currName)
		}

	}
	return fileNames
}

//returns all the keys from a map
func getKeysFromMap(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	return keys
}

func getRecordBytesOfBlock(ops []bclib.ROp, fname string, recNum int, counter int, blockAfterCurr int, appendVerifyNum int) ([]byte, int, bool) {
	currCounter := counter
	for _, op := range ops {
		if op.Filename == fname && op.OpAction == bclib.APPEND {
			if blockAfterCurr >= appendVerifyNum {
				if currCounter == recNum {
					return op.Record, currCounter, true
				}
				currCounter++
			} else {
				return []byte{}, currCounter, false
			}
		}
	}

	return []byte{}, currCounter, false
}
