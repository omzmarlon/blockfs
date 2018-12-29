package generation

import (
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/omzmarlon/blockfs/bclib"
	"github.com/omzmarlon/blockfs/bclib/chain"
	"github.com/omzmarlon/blockfs/bclib/flooding"
)

// ********************* PUBLIC *********************

// InitBlockGenerator - Initializes block generator routine
// mId - current miner's ID
// opD - the difficulty for op block
// noopD - the difficulty for no-op block
// blockOutput - channel to put a ready block
func InitBlockGenerator(mID string, genesisBlockHash string, opD uint8, noopD uint8, confirmsPerFileCreate uint8, confirmsPerFileAppend uint8,
	blockOutput chan<- bclib.Block, Bcm *chain.BlockChainManager, opsOutput <-chan []bclib.ROp, Flooder *flooding.Flooder) {
	minerID = mID
	opDifficulty = opD
	noopDifficulty = noopD
	ConfirmsPerFileCreate = confirmsPerFileCreate
	ConfirmsPerFileAppend = confirmsPerFileAppend
	GenesisBlockHash = genesisBlockHash

	bcm = Bcm
	flooder = Flooder

	go blockGeneratorDaemon(opsOutput, blockOutput)
}

// ********************* PRIVATE *********************
var (
	minerID               string
	opDifficulty          uint8
	noopDifficulty        uint8
	bcm                   *chain.BlockChainManager
	flooder               *flooding.Flooder
	ConfirmsPerFileCreate uint8
	ConfirmsPerFileAppend uint8
	GenesisBlockHash      string
)

func blockGeneratorDaemon(opsInput <-chan []bclib.ROp, blockOutput chan<- bclib.Block) {
	for {
		var ops []bclib.ROp
		select {
		case ops = <-opsInput:
		default:
			// no ops available, use empty ops for no-op block
			ops = []bclib.ROp{}
		}

		var prevHash string
		bcm.RwLock.RLock()
		root := bcm.BlockRoot
		candidates := chain.FindLongestChains(&root)
		bcm.RwLock.RUnlock()
		if len(candidates) == 0 {
			prevHash = GenesisBlockHash
		} else if len(candidates) == 1 || len(ops) == 0 {
			prevHash = candidates[0][len(candidates[0])-1]
		} else {
			for _, candidate := range candidates {
				if chain.CheckRfsSemanticsOnChain(&root, candidate, ops) {
					prevHash = candidate[len(candidate)-1]
					break
				}
			}
			prevHash = candidates[0][len(candidates[0])-1]
			log.Printf("BLOCK GEN: WARNING WARNING WARNING No proper prevHash found\n")
		}

		opsString := "["
		for _, op := range ops {
			opsString += op.OpID + ":"
			opsString += strconv.Itoa(int(op.OpAction))
			opsString += ", "
		}
		opsString += "]"
		log.Printf("BLOCK GEN: start mining new block using prevHash: %s, ops: %s\n", prevHash, opsString)

		var nonce uint32
		var hash string
		for {
			// TODO: logic for intercepting no-op block generation and start working on op block?
			nonce = randomNonce()
			hash = chain.ComputeHashString(prevHash, minerID, nonce, ops)
			if chain.CheckDifficultyReached(hash, ops, int(opDifficulty), int(noopDifficulty)) {
				break
			}
		}
		newBlock := bclib.NewBlock(hash, prevHash, minerID, ops, nonce)
		log.Printf("BLOCK GEN: Successfully created a new block with hash: %s\n", hash)

		blockOutput <- newBlock
		log.Printf("BLOCK GEN: Successfully forwarded the new block with hash: %s to bcm. BCM input chan has length: %d\n", hash, len(blockOutput))
		flooder.GetFloodingBlocksInputChann() <- newBlock
		log.Printf("BLOCK GEN: Successfully forwarded the new block with hash: %s to flooder\n", hash)
		time.Sleep(500 * time.Millisecond)
	}
}

func randomNonce() uint32 {
	source := rand.NewSource(time.Now().UnixNano())
	ran := rand.New(source)
	return ran.Uint32()
}
