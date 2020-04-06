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
	blocksQueue     chan *domain.Block // queue for blocks generated and sent from other miner peer
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
		blocksQueue:     make(chan *domain.Block, conf.BlockQueueSize),
		retryBlockQueue: make(chan *domain.Block, conf.RetryBlockQueueMaxSize),
	}

	go queues.blockProcessorDaemon()
	go queues.opProcessorDaemon()

	return queues
}

func (queues *Queues) blockProcessorDaemon() {
	for {
		// TODO
		if len(queues.blocksQueue) != 0 {
			// append to blockchain, react based on result code
			// maybe I can try to append asyncly so that they all get processed at once
			// the async function can check if it were successful, depends of the result, send the block to retry queue
			log.Printf("Got block: %+v", <-queues.blocksQueue)
		}
		if len(queues.retryBlockQueue) != 0 {

		}
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
			queues.blocksQueue <- gen.GenerateBlock(queues.blockchain.GetBlockHash(),
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
			queues.blocksQueue <- gen.GenerateBlock(queues.blockchain.GetBlockHash(),
				queues.blockchain.Conf.MinerID, &ops, queues.blockchain.Conf.OpDiffculty,
				queues.blockchain.Conf.NoopDifficulty)
			restartTimer = true
		}
	}
}
