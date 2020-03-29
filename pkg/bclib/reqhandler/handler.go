package reqhandler

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
	"sync"

	"github.com/omzmarlon/blockfs/pkg/bclib"
	"github.com/omzmarlon/blockfs/pkg/bclib/chain"
	"github.com/omzmarlon/blockfs/pkg/bclib/flooding"
	"github.com/omzmarlon/blockfs/pkg/bclib/op"
)

type Record [512]byte

type ResponseMessage struct {
	NumRecs   uint16
	Record    Record
	Fnames    []string
	RecordNum uint16
	Error     string
}

type RequestMessage struct {
	OpName    string
	Fname     string
	RecordNum uint16
	// we changed this to Record not *Record because we can't send pointers over the wire
	Record Record
}

type ReqHandler Record

var minerID string

var createVerifyNum int

var appendVerifyNum int

var verifyFailThreshold = 15

var opIdCounter = 0

var genOpLock *sync.Mutex

var bcmInstance *chain.BlockChainManager

var createFileCost int

var flooderInstance *flooding.Flooder

var noopSalary int

var opSalary int

var opFlooder chan<- bclib.ROp

//initializes the req handler
//takes a ip and port together as a string
//creates a tcp connection to listen for tcp connections from the rfs clients
//returns error if can not create a tcp connection with the given addr
func InitReqHandler(addr string, minerId string, createVerify int, appendVerify int, opMoney int, noopMoney int, createCost int, flooder *flooding.Flooder, bcm *chain.BlockChainManager) error {

	reqHandler := new(ReqHandler)
	rpc.Register(reqHandler)
	rpc.HandleHTTP()

	l, e := net.Listen("tcp", addr)
	if e != nil {
		return e
	}

	bcmInstance = bcm
	createFileCost = createCost
	minerID = minerId
	createVerifyNum = createVerify
	appendVerifyNum = appendVerify
	flooderInstance = flooder
	noopSalary = noopMoney
	opSalary = opMoney
	opFlooder = flooderInstance.GetFloodingOpsInputChann()

	genOpLock = &sync.Mutex{}

	go http.Serve(l, nil)
	return nil
}

func (r *ReqHandler) HandleLocalRequest(reqMsg *RequestMessage, resMsg *ResponseMessage) error {
	if flooderInstance.GenNumberOfPeers() < 1 {
		resMsg.Error = "DISCONNECTED"
		return nil
	}

	files := handleList()
	switch reqMsg.OpName {

	case "LIST":
		resMsg.Fnames = files
		return nil

	case "TOTALREC":
		if chain.Contains(files, reqMsg.Fname) {
			totalRecs := handleTotalRec(reqMsg.Fname)
			resMsg.NumRecs = totalRecs
		} else {
			resMsg.Error = "FILEDNE"
		}
		return nil

	case "READREC":
		file, err := handleReadRec(reqMsg.Fname, int(reqMsg.RecordNum))
		if err != nil {
			resMsg.Error = "FILEDNE"
		} else {
			resMsg.Record = file

		}
		return nil
	}

	return nil
}

func (r *ReqHandler) HandleRemoteRequest(reqMsg *RequestMessage, resMsg *ResponseMessage) error {
	if flooderInstance.GenNumberOfPeers() < 1 {
		resMsg.Error = "DISCONNECTED"
		return nil
	}
	files := handleList()

	switch reqMsg.OpName {

	case "CREATE":

		if chain.Contains(files, reqMsg.Fname) {
			resMsg.Error = "FILEEXISTS"
		} else {
			_, err := handleRemote(reqMsg.Fname, reqMsg.Record, bclib.CREATE)
			if err != nil {
				resMsg.Error = "FILEEXISTS"
			}
		}

		return nil
	case "APPENDREC":
		if !chain.Contains(files, reqMsg.Fname) {
			resMsg.Error = "FILEDNE"

		} else {
			indexNum, _ := handleRemote(reqMsg.Fname, reqMsg.Record, bclib.APPEND)
			resMsg.RecordNum = indexNum
		}
	}

	return nil
}

func handleRemote(fname string, record Record, ins bclib.ROpAction) (uint16, error) {
	var cost int
	if ins == bclib.CREATE {
		cost = createFileCost
	} else {
		cost = 1
	}
	opQueueChan := op.GetOpsInputChann()
	opId := genOpId()
	unaddedBlockChan := bcmInstance.GenUnaddedBlockChan(opId)

	for {
		if checkSufficientCoin(cost) {
			break
		}
		<-unaddedBlockChan
	}

	for {
		newRop := bclib.NewROp(opId, minerID, ins, fname, record[:])

		opQueueChan <- newRop
		opFlooder <- newRop

		ourBlock, err := checkAndReturnAdded(opId, unaddedBlockChan, ins, fname)

		if err != nil {
			return 0, errors.New("")
		}

		//stage 2: we know that our block has been added to the block chain, next we see if we
		// are verfied

		err = checkVerified(opId, ourBlock, ins)
		if err == nil {
			if ins == bclib.APPEND {
				return handleTotalRec(fname) - 1, nil
			} else {
				return 0, nil
			}
		}
		//retrying with new opID
		fmt.Println("RETRYING, RETRYING, RETRYING, RETRYING, RETRYING, RETRYING, RETRYING")

		opId = genOpId()
	}
}

func checkSufficientCoin(cost int) bool {
	bcmInstance.RwLock.RLock()
	defer bcmInstance.RwLock.RUnlock()
	return chain.CountCoins(&bcmInstance.BlockRoot, opSalary, noopSalary, createFileCost, minerID) >= cost
}

func checkAndReturnAdded(opId string, unaddedBlockChan <-chan bclib.Block, action bclib.ROpAction, fname string) (bclib.Block, error) {
	// stage 1, we are going to see if our op is on the block chain or not yet

	var count = 0

	for {
		newBlock := <-unaddedBlockChan
		count++
		fmt.Println("CHECK AND RETURN ADDED, CHECK AND RETURN ADDED, CHECK AND RETURN ADDED")

		if count > 10 && action == bclib.CREATE {
			bcmInstance.RwLock.RLock()
			longestChain := chain.FindLongestChain(&bcmInstance.BlockRoot)
			allFiles := chain.GetAllFileNamesOnChain(&bcmInstance.BlockRoot, longestChain)
			bcmInstance.RwLock.RUnlock()
			if chain.Contains(allFiles, fname) {
				return bclib.Block{}, errors.New("")
			}
		}

		if blockContainsOps(opId, newBlock) {
			bcmInstance.RemoveUnaddedBlockChan(opId)
			return newBlock, nil
		}
	}
}

func checkVerified(opId string, block bclib.Block, opType bclib.ROpAction) error {

	addedBlocksCount := 0

	unverifiedBlockChan := bcmInstance.GenUnverifiedBlockChan(opId)

	var verifyNum int

	if opType == bclib.CREATE {
		verifyNum = createVerifyNum
	} else {
		verifyNum = appendVerifyNum
	}

	//stage 2 we are going to see if our block that has been added becomes the "longest chain"

	for {
		<-unverifiedBlockChan
		addedBlocksCount++
		fmt.Println("CHECK VERIFIED, CHECK VERIFIED, CHECK VERIFIED, CHECK VERIFIED, CHECK VERIFIED")
		bcmInstance.RwLock.RLock()
		longestChain := chain.DepthAtBlock(&bcmInstance.BlockRoot, block.Hash)
		bcmInstance.RwLock.RUnlock()
		if longestChain >= verifyNum {
			bcmInstance.RemoveUnverifiedChan(opId)
			return nil
		}

		if addedBlocksCount > verifyFailThreshold {
			return errors.New("we are not on the lognest chain, retry")
		}

	}
}

func blockContainsOps(blockId string, newBlock bclib.Block) bool {

	for _, o := range newBlock.Ops {
		if o.OpID == blockId {
			return true
		}
	}
	return false
}

func genOpId() string {
	genOpLock.Lock()
	opIdCounter++
	opId := minerID + strconv.Itoa(opIdCounter)
	genOpLock.Unlock()
	return opId
}

//handles the LIST op from the rfs client
func handleList() []string {
	bcmInstance.RwLock.RLock()
	defer bcmInstance.RwLock.RUnlock()
	longestChain := chain.FindLongestChain(&bcmInstance.BlockRoot)
	allFiles := chain.GetAllFileNamesOnChain(&bcmInstance.BlockRoot, longestChain)

	return allFiles
}

//find the total number of recs in the block chain
func handleTotalRec(fname string) uint16 {

	bcmInstance.RwLock.RLock()
	defer bcmInstance.RwLock.RUnlock()
	num := chain.CountRecNumWithFname(&bcmInstance.BlockRoot, fname, createVerifyNum, appendVerifyNum)

	return uint16(num)
}

func handleReadRec(fname string, recNum int) (Record, error) {

	chanName := fname + strconv.Itoa(recNum) + genOpId()
	retryChan := bcmInstance.GenUnaddedBlockChan(chanName)

	for {
		bcmInstance.RwLock.RLock()
		if chain.CountRecNumWithFname(&bcmInstance.BlockRoot, fname, createVerifyNum, appendVerifyNum)-1 >= recNum {
			longestChain := chain.FindLongestChain(&bcmInstance.BlockRoot)
			allFiles := chain.GetAllFileNamesOnChain(&bcmInstance.BlockRoot, longestChain)

			if !chain.Contains(allFiles, fname) {
				bcmInstance.RwLock.RUnlock()
				return Record{}, errors.New("FILEDNE")
			}

			recordBytes, err := chain.ReadRec(&bcmInstance.BlockRoot, fname, recNum, appendVerifyNum)

			if err == nil {
				var readRecord [512]byte
				copy(readRecord[:], recordBytes)
				bcmInstance.RemoveUnaddedBlockChan(chanName)
				bcmInstance.RwLock.RUnlock()
				return readRecord, nil
			}
		}
		bcmInstance.RwLock.RUnlock()
		<-retryChan
	}

}
