package grpc

import (
	"context"
	"io"
	"log"

	"github.com/omzmarlon/blockfs/pkg/api"

	"google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TODO:
// CHANGE PROGRAMMING MODEL
// use birectional grpc streaming so that I don;t need to handle pairing connection myself
// this peer class handles all the sending/receiving flooding requests/responses
// flooder use channels to communicate with peer class to send/receiving flooding (flooder does not send requests by itself anymore!)
// the authentication can be done on init and then piggypack some token to all the following requests

// Peer is grpc service for receiving flooded blocks, ops, and chains
type Peer struct {
	blocksBuffer chan<- api.FloodBlockRequest // channel to off-load blocks to when received peer flood
	opsBuffer    chan<- api.FloodOpRequest    // channel to off-load ops to when received peer flood
}

// FloodBlock - flood blocks to and from peers
func (peer *Peer) FloodBlock(stream api.Peer_FloodBlockServer) error {
	for {
		request, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Printf("[peer-grpc]: error occured when receiving block from peer")
			continue
		}
		peer.blocksBuffer <- *request
		if err := stream.Send(&api.FloodResponse{RequestId: request.RequestId}); err != nil {
			log.Printf("[peer-grpc]: error occured when responding to peer")
			continue
		}
	}
}

// FloodChain - flood chains to and from peers
func (peer *Peer) FloodChain(stream api.Peer_FloodChainServer) error {
	// TODO
	// I'm thinking that all the channels for buffering can be reused
	// i.e. I can just reuse the block channels for blockchains
	// it's just that when merging with local blockchain it will be different (the difference is only in blockchain.go)
	return status.Errorf(codes.Unimplemented, "method FloodChain not implemented")
}

// FloodOp - flood ops to and from peers
func (peer *Peer) FloodOp(stream api.Peer_FloodOpServer) error {
	for {
		request, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Printf("[peer-grpc]: error occured when receiving op from peer")
			continue
		}
		peer.opsBuffer <- *request
		if err := stream.Send(&api.FloodResponse{RequestId: request.RequestId}); err != nil {
			log.Printf("[peer-grpc]: error occured when responding to peer")
			continue
		}
	}
}

// HeartBeat - return heartbeat response to confirm liveness
func (peer *Peer) HeartBeat(ctx context.Context, req *api.HeartBeatRequest) (*api.HeartBeatResponse, error) {
	return &api.HeartBeatResponse{
		FromMiner: req.FromMiner,
		ToMiner:   req.ToMiner,
	}, nil
}

// RegisterMiner - registers miner peer grpc
func RegisterMiner(s *grpc.Server, blocks chan<- api.FloodBlockRequest, ops chan<- api.FloodOpRequest) {
	api.RegisterPeerServer(s, &Peer{
		blocksBuffer: blocks,
		opsBuffer:    ops,
	})
}
