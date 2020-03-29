package flooding

import (
	"log"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/omzmarlon/blockfs/pkg/bclib"
	"github.com/omzmarlon/blockfs/pkg/bclib/chain"
	"github.com/omzmarlon/blockfs/pkg/fdlib"
)

func (flooder *Flooder) GetFloodingOpsInputChann() chan<- bclib.ROp {
	return flooder.opsInput
}

func (flooder *Flooder) GetFloodingBlocksInputChann() chan<- bclib.Block {
	return flooder.blocksInput
}

func (flooder *Flooder) GenNumberOfPeers() int {
	flooder.rwLockConnectedMiners.RLock()
	defer flooder.rwLockConnectedMiners.RUnlock()
	return len(flooder.connectedMiners)
}

// InitFloodingProtocol - initialize flooding protocol
// sets up listening ip:port for incoming miners
func InitFloodingProtocol(config FlooderConfig, OpsQueueInputChann chan<- bclib.ROp, BCMInputChann chan<- bclib.Block, bcm *chain.BlockChainManager) (*Flooder, error) {
	// set up listening connection
	tcpAddr, tcpAddrErr := net.ResolveTCPAddr("tcp", config.IncomingMinersAddr)
	if tcpAddrErr != nil {
		log.Printf("FLOODER: failed to resolve TCP addr to init flooding protocol: %s\n", tcpAddrErr)
		return nil, tcpAddrErr
	}
	listener, lerr := net.ListenTCP("tcp", tcpAddr)
	if lerr != nil {
		log.Printf("FLOODER: failed to listen tcp to init flooding protocol: %s\n", lerr)
		return nil, lerr
	}

	// initialize failure detector
	source := rand.NewSource(time.Now().UnixNano())
	rand := rand.New(source)
	epochNonce := rand.Uint64()

	failureDetector, notifyCh, fderr := fdlib.Initialize(epochNonce, 50) // TODO: parameterize channel capacity?
	if fderr != nil {
		log.Printf("FLOODER: failed to init FD to init flooding protocol: %s\n", fderr)
		return nil, fderr
	}

	flooder := &Flooder{
		minerID:               config.MinerID,
		opsInput:              make(chan bclib.ROp, 30),
		blocksInput:           make(chan bclib.Block, 30),
		connectedMiners:       make([]*ConnectedMiner, 0),
		rwLockConnectedMiners: &sync.RWMutex{},
		minerConnListener:     listener,
		fd:                    failureDetector,
		fdResponseIPPort:      config.FDResponseIPPort,
		lostMsgThresh:         config.LostMsgThresh,
		opsQueueInputChann:    OpsQueueInputChann,
		bcmInputChann:         BCMInputChann,
		bcm:                   bcm,
		OutgoingMinersIP:      config.OutgoingMinersIP,
	}

	// ***** Kick off flooder *****
	// start flooder failure detector
	fdreserr := flooder.fd.StartResponding(config.FDResponseIPPort)
	if fdreserr != nil {
		log.Printf("FLOODER: failed start fd responding to init flooding protocol: %s\n", fdreserr)
		return nil, fdreserr
	}

	// connect to provided peers
	flooder.connectToPeerMiner(config.PeerMinersAddrs, config.OutgoingMinersIP, config.FDResponseIPPort)

	// start daemons
	go flooder.acceptMinerDaemon(config.FDResponseIPPort)
	go flooder.floodingInputDaemon()
	go flooder.receivePacketDaemon()
	go flooder.minerFailureDetectionDaemon(notifyCh)

	flooder.bcm.RwLock.RLock()
	flooder.floodPacket(prepareFloodingPkt(FLOOD_CHAIN, bclib.ROp{}, bclib.Block{}, flooder.bcm.BlockRoot))
	flooder.bcm.RwLock.RUnlock()
	log.Printf("FLOODER: New miner %s flooded local blockchain to peers.\n", flooder.minerID)

	return flooder, nil
}
