package queue

import (
	"log"
	"time"

	"github.com/omzmarlon/blockfs/pkg/blockchain"
	"github.com/omzmarlon/blockfs/pkg/blockchain/gen"
	"github.com/omzmarlon/blockfs/pkg/domain"
	"github.com/omzmarlon/blockfs/pkg/util"
)

// Queues - contains all the queues that a miner needs to maintain
type Queues struct {
	conf       Conf
	blockchain *blockchain.Blockchain
	// use channel type for below queues to ensure non-blocking operations
	OpsQueue        chan domain.Op     // queue for Ops sent from client and other miner peer
	BlocksQueue     chan *domain.Block // queue for blocks generated and sent from other miner peer
	retryBlockQueue chan *domain.Block // in case if newer blocks get received before older blocks
}

// Conf - configurations for various queues
type Conf struct {
	MaxOpsQueueSize        int // exceeding this size, the ops in the queue should be packed into a block
	RetryBlockQueueMaxSize int
	WaitOpsTimeoutMilli    int // timeout for waiting for enough ops to generate a new block
	BlockQueueSize         int
}

// StartQueueDaemons kicks off all queue processing background daemons
func StartQueueDaemons(conf Conf, blockchain *blockchain.Blockchain) *Queues {
	log.Println("Initializing miner queues...")
	defer log.Println("Miner queues initialized.")

	queues := &Queues{
		conf:            conf,
		blockchain:      blockchain,
		OpsQueue:        make(chan domain.Op, conf.MaxOpsQueueSize),
		BlocksQueue:     make(chan *domain.Block, conf.BlockQueueSize),
		retryBlockQueue: make(chan *domain.Block, conf.RetryBlockQueueMaxSize),
	}

	go queues.blockProcessorDaemon()
	go queues.opProcessorDaemon()

	return queues
}

func (queues *Queues) blockProcessorDaemon() {
	for {
		if len(queues.BlocksQueue) != 0 {
			block := <-queues.BlocksQueue
			log.Printf("Got block: %+v", block)
			go queues.processBlockHelper(block)
		}
		if len(queues.retryBlockQueue) != 0 {
			block := <-queues.retryBlockQueue
			go queues.processBlockHelper(block)
		}
	}
}

func (queues *Queues) processBlockHelper(block *domain.Block) {
	result := queues.blockchain.AppendBlock(block)
	switch result {
	case blockchain.APPEND_RESULT_SUCCESS:
		log.Println("Block appended successfully")
	case blockchain.APPEND_RESULT_NOT_FOUND:
		queues.retryBlockQueue <- block
	case blockchain.APPEND_RESULT_INVALID_BLOCK:
		log.Println("Append failed due to invalid block")
		// TODO: may need to get back to client?
	case blockchain.APPEND_RESULT_INVALID_SEMANTIC:
		log.Println("Append failed due to invalid semantics")
		// TODO: may need to get back to client?
	}
}

func (queues *Queues) opProcessorDaemon() {
	duration := time.Duration(queues.conf.WaitOpsTimeoutMilli) * time.Millisecond
	restartTimer := false
	timer := time.NewTimer(duration)
	for {
		if restartTimer {
			timer = time.NewTimer(duration)
			restartTimer = false
		}
		select {
		case <-timer.C:
			numOps := util.MinInt(len(queues.OpsQueue), queues.conf.MaxOpsQueueSize)
			ops := make([]domain.Op, numOps)
			for i := 0; i < numOps; i++ {
				ops[i] = <-queues.OpsQueue
			}
			log.Println("Time out for Ops reached, starting block generation...")
			queues.BlocksQueue <- gen.GenerateBlock(queues.blockchain.GetBlockHash(),
				queues.blockchain.Conf.MinerID, &ops, queues.blockchain.Conf.OpDiffculty,
				queues.blockchain.Conf.NoopDifficulty)
			restartTimer = true
		default:
		}
		if len(queues.OpsQueue) >= queues.conf.MaxOpsQueueSize {
			ops := make([]domain.Op, queues.conf.MaxOpsQueueSize)
			for i := 0; i < queues.conf.MaxOpsQueueSize; i++ {
				ops[i] = <-queues.OpsQueue
			}
			log.Println("Max queue size for Ops reached, starting block generation...")
			queues.BlocksQueue <- gen.GenerateBlock(queues.blockchain.GetBlockHash(),
				queues.blockchain.Conf.MinerID, &ops, queues.blockchain.Conf.OpDiffculty,
				queues.blockchain.Conf.NoopDifficulty)
			restartTimer = true
		}
	}
}
