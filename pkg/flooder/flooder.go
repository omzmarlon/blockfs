package flooder

import (
	"context"
	"log"
	"net"
	"strconv"
	"time"

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
	conf       Conf
	blockchain *blockchain.Blockchain
	peers      map[string]api.PeerClient
}

// Conf - configuration for Flooder
type Conf struct {
	MinerID                   string
	Port                      int // port to serve grpc service
	PeerOpsFloodBufferSize    int
	PeerBlocksFloodBufferSize int
	HeatbeatRetryMax          int
	Peers                     map[string]string // a map of minerID -> IPPort address
}

// New is the constructor for a new flooder
func New(conf Conf, blockchain *blockchain.Blockchain) *Flooder {
	flooder := &Flooder{
		conf:       conf,
		blockchain: blockchain,
		peers:      initConnToPeers(conf.Peers),
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
		lis, err := net.Listen("tcp", ":"+strconv.Itoa(flooder.conf.Port))
		if err != nil {
			log.Fatalf("flooder failed to listen: %v", err)
		}
		s := grpc.NewServer()
		rpc.RegisterMiner(s, blocksReqFloodBuffer, opsReqFloodBuffer)
		s.Serve(lis)

	}()
	// kick off daemon to flood locally generated blocks to peers
	go func() {
		for {
			block := <-blockFloodingBuffer
			log.Println("Got a block to flood to peers")
			grpcBlock := util.DomainBlockToGrpcBlock(*block)
			floodReq := &api.FloodBlockRequest{
				RequestId:         guuid.New().String(),
				PreviousReceivers: []string{flooder.conf.MinerID},
				Block:             &grpcBlock,
			}
			for _, peer := range flooder.peers {
				sendFloodBlockRequestHelper(peer, floodReq)
			}
		}
	}()
	// kick off daemon to flood client submitted ops to peers
	go func() {
		for {
			op := <-opsFloodingBuffer
			log.Println("Got an op to flood to peers")
			grpcOp := util.DomainOpToGrpcOp(*op)
			floodReq := &api.FloodOpRequest{
				RequestId:         guuid.New().String(),
				PreviousReceivers: []string{flooder.conf.MinerID},
				Op:                &grpcOp,
			}
			for _, peer := range flooder.peers {
				sendFloodOpRequestHelper(peer, floodReq)
			}
		}
	}()
	// kick off daemon to handle blocks flooded from peers
	go func() {
		for {
			peerFloodReq := <-blocksReqFloodBuffer
			log.Println("Received a block from peers")
			domainBlock := util.GrpcBlockToDomainBlock(*peerFloodReq.GetBlock())
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
					}
					sendFloodBlockRequestHelper(peer, newReq)
				}
			}
		}
	}()
	// kick off daemon to handle ops flooded from peers
	go func() {
		for {
			peerFloodReq := <-opsReqFloodBuffer
			log.Println("Received an op from peers")
			domainOp := util.GrpcOpToDomainOp(*peerFloodReq.GetOp())
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
					}
					sendFloodOpRequestHelper(peer, newReq)
				}
			}
		}
	}()
}

func sendFloodBlockRequestHelper(client api.PeerClient, request *api.FloodBlockRequest) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := client.FloodBlock(ctx, request)
	if err != nil {
		log.Fatal("flooding to one of the clients failed")
	}
}

func sendFloodOpRequestHelper(client api.PeerClient, request *api.FloodOpRequest) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := client.FloodOp(ctx, request)
	if err != nil {
		log.Fatal("flooding to one of the clients failed")
	}
}

func initConnToPeers(peerAddresses map[string]string) map[string]api.PeerClient {
	result := make(map[string]api.PeerClient)
	for minerID, addr := range peerAddresses {
		conn, err := grpc.Dial(addr, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("did not connect to peer %s for err: %s", addr, err)
		} else {
			// TODO: don't support dynamically adding/removing peers yet
			// so if anything fails to connect, we just forget them for now
			result[minerID] = api.NewPeerClient(conn)
		}

	}
	return result
}

// TODO need to do heart beat and then manage active peers
// TODO the peer connection is still one way at this point
