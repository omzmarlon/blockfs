package flooding

import (
	"cs416/P1-v3d0b-q4d0b/bclib"
	"cs416/P1-v3d0b-q4d0b/bclib/chain"
	"cs416/P1-v3d0b-q4d0b/fdlib"
	"math/rand"
	"net"
	"sync"
	"time"
)

type FlooderConfig struct {
	MinerID            string
	IncomingMinersAddr string // where this flooder will listen for connection request
	OutgoingMinersIP   string // where this flooder will connect to other miner
	PeerMinersAddrs    []string
	FDResponseIPPort   string
	LostMsgThresh      uint8
}

type Flooder struct {
	minerID     string // miner ID of this miner
	opsInput    chan bclib.ROp
	blocksInput chan bclib.Block
	// peer miners
	connectedMiners       []*ConnectedMiner
	rwLockConnectedMiners *sync.RWMutex
	minerConnListener     *net.TCPListener
	// failure detection
	fd                 fdlib.FD
	fdResponseIPPort   string
	lostMsgThresh      uint8
	opsQueueInputChann chan<- bclib.ROp
	bcmInputChann      chan<- bclib.Block
	bcm                *chain.BlockChainManager
	OutgoingMinersIP   string
}

type FloodPkt struct {
	ID        int      // mainly used for debugging
	Recved    []string // list of miner IDs
	FloodType FloodType
	Op        bclib.ROp
	Block     bclib.Block
	Chain     bclib.Block
}

func prepareFloodingPkt(floodType FloodType, op bclib.ROp, block bclib.Block, chain bclib.Block) FloodPkt {
	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)
	return FloodPkt{
		ID:        random.Intn(1000000),
		Recved:    []string{},
		FloodType: floodType,
		Op:        op,
		Block:     block,
		Chain:     chain,
	}
}

// FloodType - enum of flooding packet type
type FloodType int

const (
	FLOOD_OP    FloodType = 1
	FLOOD_BLOCK FloodType = 2
	FLOOD_CHAIN FloodType = 3
)

type ConnectedMiner struct {
	minerID            string
	conn               *net.TCPConn
	lock               *sync.Mutex
	fdRespondingIPPort string // IP & port for failure detection
}

func NewConnectedMiner(minerID string, conn *net.TCPConn, fdRespondingIPPort string) ConnectedMiner {
	return ConnectedMiner{
		minerID:            minerID,
		conn:               conn,
		fdRespondingIPPort: fdRespondingIPPort,
		lock:               &sync.Mutex{},
	}
}

// MinerInitMeta - the intial data to exchange to set up miner peer-to-peer connection
type MinerInitMeta struct {
	MinerID          string
	FDResponseIPPort string // IP & port for failure detection
	BlockChain       bclib.Block
}

// NewMinerInitMeta - MinerInitMeta constructor
func NewMinerInitMeta(minerID string, fdResponseIPPort string, blockchain bclib.Block) MinerInitMeta {
	return MinerInitMeta{
		MinerID:          minerID,
		FDResponseIPPort: fdResponseIPPort,
		BlockChain:       blockchain,
	}
}
