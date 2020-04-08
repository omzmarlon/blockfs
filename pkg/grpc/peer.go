package grpc

import (
	"context"

	"github.com/omzmarlon/blockfs/pkg/api"
	"google.golang.org/grpc"
)

// Peer is grpc service for receiving flooded blocks, ops, and chains
type Peer struct {
	blocksBuffer chan<- api.FloodBlockRequest // channel to off-load blocks to when received peer flood
	opsBuffer    chan<- api.FloodOpRequest    // channel to off-load ops to when received peer flood
}

// FloodBlock - processes blocks flooded from peers
func (peer *Peer) FloodBlock(ctx context.Context, request *api.FloodBlockRequest) (*api.FloodResponse, error) {
	peer.blocksBuffer <- *request
	return &api.FloodResponse{}, nil
}

// FloodChain - processes chains flooded from peers
func (peer *Peer) FloodChain(ctx context.Context, request *api.FloodChainRequest) (*api.FloodResponse, error) {
	// TODO
	// I'm thinking that all the channels for buffering can be reused
	// i.e. I can just reuse the block channels for blockchains
	// it's just that when merging with local blockchain it will be different (the difference is only in blockchain.go)
	return nil, nil
}

// FloodOp - processes ops flooded from peers
func (peer *Peer) FloodOp(ctx context.Context, request *api.FloodOpRequest) (*api.FloodResponse, error) {
	peer.opsBuffer <- *request
	return &api.FloodResponse{}, nil
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
