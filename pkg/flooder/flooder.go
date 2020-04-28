package flooder

import (
	"context"
	"log"
	"net"
	"sync"

	guuid "github.com/google/uuid"
	"github.com/omzmarlon/blockfs/pkg/api"
	"github.com/omzmarlon/blockfs/pkg/blockchain"
	"github.com/omzmarlon/blockfs/pkg/domain"
	rpc "github.com/omzmarlon/blockfs/pkg/grpc"
	"github.com/omzmarlon/blockfs/pkg/util"
	"google.golang.org/grpc"
)

// Flooder manages flooding blocks and ops to peers and receiving data from peers
type Flooder struct {
	conf            Conf
	blockchain      *blockchain.Blockchain
	peers           map[string]peerStreams
	peersAccessLock sync.RWMutex // peers list accessing lock
}

// Conf - configuration for Flooder
type Conf struct {
	MinerID                   string
	Address                   string
	PeerOpsFloodBufferSize    int
	PeerBlocksFloodBufferSize int
	HeatbeatRetryMax          int
	Peers                     map[string]string // a map of minerID -> IPPort address
}

// peerStreams - holds grpc streams to a peer
type peerStreams struct {
	blockStream api.Peer_FloodBlockClient
	opStream    api.Peer_FloodOpClient
}

// New is the constructor for a new flooder
func New(conf Conf, blockchain *blockchain.Blockchain) *Flooder {
	flooder := &Flooder{
		conf:       conf,
		blockchain: blockchain,
		peers:      initConnToPeers(conf),
	}
	return flooder
}

// StartFlooderDaemons - kicks off the daemon that receives/send ops & blocks from/to peers
// Takes four parameters:
// blockProcessing channel to submit received flooded blocks for internal processing
// opsProcessing channel to submit received flooded ops for internal processing
// blockFloodingBuffer channel that other internal processes use to submit blocks for flooding
// opFloodingBuffer channel that other internal processes use to submit ops for flooding
func (flooder *Flooder) StartFlooderDaemons(blockProcessing chan<- *domain.Block,
	opsProcessing chan<- *domain.Op,
	blockFloodingBuffer <-chan *domain.Block,
	opsFloodingBuffer <-chan *domain.Op) {

	// a channel for buffering blocks flooded from peers
	blocksReqFloodBuffer := make(chan api.FloodBlockRequest, flooder.conf.PeerBlocksFloodBufferSize)
	// a channel for buffering ops flooded from peers
	opsReqFloodBuffer := make(chan api.FloodOpRequest, flooder.conf.PeerOpsFloodBufferSize)
	// kick off daemon for the grpc service to accept flooding requests from peers
	go func() {
		lis, err := net.Listen("tcp", flooder.conf.Address)
		if err != nil {
			log.Fatalf("flooder failed to listen: %v", err)
		}
		s := grpc.NewServer()
		rpc.RegisterMiner(s, blocksReqFloodBuffer, opsReqFloodBuffer)
		s.Serve(lis)
	}()
	// kick off daemon to flood locally generated blocks to peers
	go flooder.localBlockFlooderDaemon(blockFloodingBuffer)
	// kick off daemon to flood client submitted ops to peers
	go flooder.clientOpFlooderDaemon(opsFloodingBuffer)
	// kick off daemon to handle blocks flooded from peers
	go flooder.peerBlockHandlerDaemon(blocksReqFloodBuffer, blockProcessing)
	// kick off daemon to handle ops flooded from peers
	go flooder.peerOpHandlerDaemon(opsReqFloodBuffer, opsProcessing)
}

func initConnToPeers(conf Conf) map[string]peerStreams {
	result := make(map[string]peerStreams)
	for minerID, addr := range conf.Peers {
		conn, err := grpc.Dial(addr, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("did not connect to peer %s for err: %s", addr, err)
		} else {
			// TODO: don't support dynamically adding/removing peers yet
			// so if anything fails to connect, we just forget them for now
			newPeer := api.NewPeerClient(conn)
			if ps, err := newPeerStreams(newPeer); err == nil {
				result[minerID] = *ps
			}
		}

	}

	keys := make([]string, len(result))
	for k := range result {
		keys = append(keys, k)
	}
	log.Printf("[flooder]: initial peer list: %s", keys)
	return result
}

func (flooder *Flooder) localBlockFlooderDaemon(blockFloodingBuffer <-chan *domain.Block) {
	for {
		block := <-blockFloodingBuffer
		log.Printf("[flooder]: flooding block with %s", block.Hash)
		grpcBlock := util.DomainBlockToGrpcBlock(*block)
		floodReq := &api.FloodBlockRequest{
			RequestId:         guuid.New().String(),
			PreviousReceivers: []string{flooder.conf.MinerID},
			Block:             &grpcBlock,
			MinerID:           flooder.conf.MinerID,
			Address:           flooder.conf.Address,
		}
		for _, peer := range flooder.peers {
			err := peer.blockStream.Send(floodReq)
			if err != nil {
				log.Printf("[flooder]: failed to flood block with hash %s to peer due to err %s", block.Hash, err)
			}
		}
	}
}

func (flooder *Flooder) clientOpFlooderDaemon(opsFloodingBuffer <-chan *domain.Op) {
	for {
		op := <-opsFloodingBuffer
		log.Printf("[flooder]: flooding op on file %s with action %d", op.Filename, op.OpAction)
		grpcOp := util.DomainOpToGrpcOp(*op)
		floodReq := &api.FloodOpRequest{
			RequestId:         guuid.New().String(),
			PreviousReceivers: []string{flooder.conf.MinerID},
			Op:                &grpcOp,
			MinerID:           flooder.conf.MinerID,
			Address:           flooder.conf.Address,
		}
		for _, peer := range flooder.peers {
			err := peer.opStream.Send(floodReq)
			if err != nil {
				log.Printf("[flooder]: failed to flood op to peer due to err %s", err)
			}
		}
	}
}

func (flooder *Flooder) peerBlockHandlerDaemon(blocksReqFloodBuffer chan api.FloodBlockRequest,
	blockProcessing chan<- *domain.Block) {
	for {
		peerFloodReq := <-blocksReqFloodBuffer
		domainBlock := util.GrpcBlockToDomainBlock(*peerFloodReq.GetBlock())
		log.Printf("[flooder]: Received a block from peers: %s", domainBlock.String())
		flooder.addNewPeerIfNotExists(peerFloodReq.MinerID, peerFloodReq.Address)
		blockProcessing <- &domainBlock
		for peerID, peer := range flooder.peers {
			inPrevList := false
			for _, prevID := range peerFloodReq.GetPreviousReceivers() {
				if peerID == prevID {
					inPrevList = true
				}
			}
			if !inPrevList {
				newReceivers := append(peerFloodReq.GetPreviousReceivers(), flooder.conf.MinerID)
				newReq := &api.FloodBlockRequest{
					RequestId:         peerFloodReq.GetRequestId(),
					PreviousReceivers: newReceivers,
					Block:             peerFloodReq.GetBlock(),
					MinerID:           flooder.conf.MinerID,
					Address:           flooder.conf.Address,
				}
				err := peer.blockStream.Send(newReq)
				if err != nil {
					log.Printf("[flooder]: failed to flood block with hash %s to peer due to err %s",
						peerFloodReq.GetBlock().Hash, err)
				}
			}
		}
	}
}

func (flooder *Flooder) peerOpHandlerDaemon(opsReqFloodBuffer chan api.FloodOpRequest, opsProcessing chan<- *domain.Op) {
	for {
		peerFloodReq := <-opsReqFloodBuffer
		log.Println("Received an op from peers")
		domainOp := util.GrpcOpToDomainOp(*peerFloodReq.GetOp())
		flooder.addNewPeerIfNotExists(peerFloodReq.MinerID, peerFloodReq.Address)
		opsProcessing <- &domainOp
		for peerID, peer := range flooder.peers {
			inPrevList := false
			for _, prevID := range peerFloodReq.GetPreviousReceivers() {
				if peerID == prevID {
					inPrevList = true
				}
			}
			if !inPrevList {
				newReceivers := append(peerFloodReq.GetPreviousReceivers(), flooder.conf.MinerID)
				newReq := &api.FloodOpRequest{
					RequestId:         peerFloodReq.GetRequestId(),
					PreviousReceivers: newReceivers,
					Op:                peerFloodReq.GetOp(),
					MinerID:           flooder.conf.MinerID,
					Address:           flooder.conf.Address,
				}
				err := peer.opStream.Send(newReq)
				if err != nil {
					log.Printf("[flooder]: failed to flood op to peer due to err %s", err)
				}
			}
		}
	}
}

func (flooder *Flooder) peerExists(minerID string) bool {
	flooder.peersAccessLock.RLock()
	defer flooder.peersAccessLock.RUnlock()
	_, ok := flooder.peers[minerID]
	return ok
}

func (flooder *Flooder) addNewPeer(minerID string, streams peerStreams) {
	flooder.peersAccessLock.Lock()
	defer flooder.peersAccessLock.Unlock()
	flooder.peers[minerID] = streams
}

func (flooder *Flooder) addNewPeerIfNotExists(minerID string, address string) {
	if !flooder.peerExists(minerID) {
		log.Printf("[flooder]: adding new peer %s", minerID)
		conn, err := grpc.Dial(address, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("did not connect to peer %s for err: %s", address, err)
		} else {
			newPeer := api.NewPeerClient(conn)
			if ps, err := newPeerStreams(newPeer); err == nil {
				flooder.addNewPeer(minerID, *ps)
				// need to flood the local blockchain to the new peer
				// flood each child of the genesis block as everyone has the same genesis
				clone := flooder.blockchain.CloneChain()
				for i := 0; i < len(*clone.Children); i++ {
					child := util.DomainBlockToGrpcBlock(*(*clone.Children)[i])
					req := api.FloodBlockRequest{
						RequestId:         guuid.New().String(),
						PreviousReceivers: []string{flooder.conf.MinerID},
						Block:             &child,
						MinerID:           flooder.conf.MinerID,
						Address:           flooder.conf.Address,
					}
					ps.blockStream.Send(&req)
				}
			}
		}
	}
}

func newPeerStreams(peerClient api.PeerClient) (*peerStreams, error) {
	blockStream, berr := peerClient.FloodBlock(context.Background())
	if berr != nil {
		log.Printf("[flooder]: could not set up block flooding stream due to err %s", berr)
		return nil, berr
	}
	opStream, oerr := peerClient.FloodOp(context.Background())
	if oerr != nil {
		log.Printf("[flooder]: could not set up op flooding stream due to err %s", oerr)
		return nil, oerr
	}
	return &peerStreams{
		blockStream: blockStream,
		opStream:    opStream,
	}, nil
}

// TODO need to do heart beat and then manage active peers
// a dedicated daemon to do this
// periodic heartbeat
// heart beat timeout
// max retry
