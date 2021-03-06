package coordinator

import (
	"log"
	"time"

	"github.com/omzmarlon/blockfs/pkg/blockchain"
	"github.com/omzmarlon/blockfs/pkg/blockchain/gen"
	"github.com/omzmarlon/blockfs/pkg/domain"
	"github.com/omzmarlon/blockfs/pkg/util"
)

// Coordinator coordinates blocks and ops internal processing
type Coordinator struct {
	conf            Conf
	blockchain      *blockchain.Blockchain
	retryBlockQueue chan *domain.Block // in case if newer blocks get received before older blocks
}

// Conf - configurations for various queues
type Conf struct {
	MaxOpsQueueSize        int // exceeding this size, the ops in the queue should be packed into a block
	RetryBlockQueueMaxSize int
	WaitOpsTimeoutMilli    int // timeout for waiting for enough ops to generate a new block
}

// New is the constructor for a new coordinator
func New(conf Conf, blockchain *blockchain.Blockchain) *Coordinator {
	coordinator := &Coordinator{
		conf:            conf,
		blockchain:      blockchain,
		retryBlockQueue: make(chan *domain.Block, conf.RetryBlockQueueMaxSize),
	}
	return coordinator
}

// StartCoordinatorDaemons kicks off background daemons that coordinates blocks and ops processing
// Takes three parameters:
// blocksBuffer channel that internal services use to submit blocks for processing
// opsBuffer channel that internal services use to submit ops for processing
// floodingBuffer channel to submit generated blocks to flooding to peers
func (coordinator *Coordinator) StartCoordinatorDaemons(blocksBuffer chan *domain.Block, opsBuffer <-chan *domain.Op, floodingBuffer chan<- *domain.Block) {
	log.Println("Initializing coordinator...")
	defer log.Println("Coordinator initialized.")

	go coordinator.blockProcessorDaemon(blocksBuffer)
	go coordinator.opProcessorDaemon(blocksBuffer, opsBuffer, floodingBuffer)
}

func (coordinator *Coordinator) blockProcessorDaemon(blocksBuffer <-chan *domain.Block) {
	// TODO: do I need to prioritize regular block processing? less priority or
	// exp back off for retry?
	for {
		if len(blocksBuffer) != 0 {
			block := <-blocksBuffer
			log.Printf("[blockProcessorDaemon]: processing block: %s", block.String())
			go coordinator.processBlockHelper(block)
		}
		if len(coordinator.retryBlockQueue) != 0 {
			block := <-coordinator.retryBlockQueue
			//log.Printf("[blockProcessorDaemon]: retrying block: %s", block.String())
			go coordinator.processBlockHelper(block)
		}
	}
}

func (coordinator *Coordinator) processBlockHelper(block *domain.Block) {
	result := coordinator.blockchain.AppendBlock(block)
	switch result {
	case blockchain.APPEND_RESULT_SUCCESS:
		log.Println("Block appended successfully")
	case blockchain.APPEND_RESULT_NOT_FOUND:
		coordinator.retryBlockQueue <- block
	case blockchain.APPEND_RESULT_INVALID_BLOCK:
		log.Println("Append failed due to invalid block")
		// TODO: may need to get back to client?
	case blockchain.APPEND_RESULT_INVALID_SEMANTIC:
		log.Println("Append failed due to invalid semantics")
		// TODO: may need to get back to client?
		// TODO: or just a default block to catch all non-success, non-notfound codes
	}
}

func (coordinator *Coordinator) opProcessorDaemon(blockBuffer chan<- *domain.Block, opsBuffer <-chan *domain.Op, floodingBuffer chan<- *domain.Block) {
	duration := time.Duration(coordinator.conf.WaitOpsTimeoutMilli) * time.Millisecond
	restartTimer := false
	timer := time.NewTimer(duration)
	for {
		if restartTimer {
			timer = time.NewTimer(duration)
			restartTimer = false
		}
		select {
		case <-timer.C:
			log.Println("[coordinator]: Time out for Ops reached, starting block generation...")
			numOps := util.MinInt(len(opsBuffer), coordinator.conf.MaxOpsQueueSize)
			ops := make([]domain.Op, 0)
			opIDSet := make(map[string]bool)
			for i := 0; i < numOps; i++ {
				op := *(<-opsBuffer)
				if _, exists := opIDSet[op.OpID]; !exists {
					opIDSet[op.OpID] = true
					ops = append(ops, op)
				}
			}

			block := gen.GenerateBlock(coordinator.blockchain.GetBlockHash(),
				coordinator.blockchain.Conf.MinerID, ops, coordinator.blockchain.Conf.OpDiffculty,
				coordinator.blockchain.Conf.NoopDifficulty)
			blockBuffer <- block
			floodingBuffer <- block
			restartTimer = true
		default:
		}
		if len(opsBuffer) >= coordinator.conf.MaxOpsQueueSize {
			log.Println("[coordinator]: Max queue size for Ops reached, starting block generation...")
			ops := make([]domain.Op, 0)
			opIDSet := make(map[string]bool)
			for i := 0; i < coordinator.conf.MaxOpsQueueSize; i++ {
				op := *(<-opsBuffer)
				if _, exists := opIDSet[op.OpID]; !exists {
					opIDSet[op.OpID] = true
					ops = append(ops, op)
				}
			}

			block := gen.GenerateBlock(coordinator.blockchain.GetBlockHash(),
				coordinator.blockchain.Conf.MinerID, ops, coordinator.blockchain.Conf.OpDiffculty,
				coordinator.blockchain.Conf.NoopDifficulty)
			blockBuffer <- block
			floodingBuffer <- block
			restartTimer = true
		}
	}
}
