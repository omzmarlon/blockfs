package op

import (
	"log"
	"time"

	"github.com/omzmarlon/blockfs/bclib"
)

// ********************* PUBLIC *********************

// InitOpsQueue - initialize ops queue
// OpsQueue takes ops input from req handler and flooding protocol and add ops to queue
// When # of ops exceeds maximum or a timeout has been reached, ops are sent to block generator for block generation
func InitOpsQueue(opTimeout uint8, qml uint32) (chan<- bclib.ROp, <-chan []bclib.ROp) {
	// // initialize global variables
	opsInput = make(chan bclib.ROp, 1)
	opsQueue = make([]bclib.ROp, 0)

	genOpBlockTimeout = opTimeout
	opsOutput = make(chan []bclib.ROp)
	queueMaxLen = qml

	go opMonitor()

	return opsInput, opsOutput
}

// GetOpsInputChann - getter for the input channel of ops
func GetOpsInputChann() chan<- bclib.ROp {
	return opsInput
}

// ********************* PRIVATE *********************
var (
	opsInput          chan bclib.ROp
	opsQueue          []bclib.ROp
	timer             time.Time
	genOpBlockTimeout uint8
	opsOutput         chan []bclib.ROp
	queueMaxLen       uint32
)

func opMonitor() {
	for {
		select {
		case newOp := <-opsInput:
			log.Println("OP-QUEUE: Got new op: ", newOp.OpID)
			if !checkOpExists(opsQueue, newOp) {
				opsQueue = append(opsQueue, newOp)
			}
			if len(opsQueue) > int(queueMaxLen) {
				go sendOpsForBlockGen(opsQueue, opsOutput) // non-blocking send to create new block
				opsQueue = make([]bclib.ROp, 0)
				setTimer()
			}
		default:
			if len(opsQueue) > 0 && time.Since(timer) > (time.Duration(int(genOpBlockTimeout))*time.Millisecond) {
				log.Println("OP-QUEUE: timer timed out")
				go sendOpsForBlockGen(opsQueue, opsOutput) // non-blocking send to create new block
				opsQueue = make([]bclib.ROp, 0)
				setTimer()
			}
		}
	}
}

func sendOpsForBlockGen(ops []bclib.ROp, opsOutput chan<- []bclib.ROp) {
	opsOutput <- ops
}

func setTimer() {
	timer = time.Now()
	log.Println("OP-QUEUE: timer reset to: ", timer)
}

func checkOpExists(ops []bclib.ROp, newOp bclib.ROp) bool {
	for _, op := range ops {
		if op.OpAction == bclib.CREATE && op.Filename == newOp.Filename {
			return true
		}
	}
	return false
}
