package grpc

import (
	"io"

	"github.com/omzmarlon/blockfs/pkg/api"
	pb "github.com/omzmarlon/blockfs/pkg/api"
	"github.com/omzmarlon/blockfs/pkg/util"
	"google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TODO
// - grpc service miner peer to peer communication

// grpc service for receiving flooded blocks, ops, and chains
type peer struct {
	T int
}

func (*peer) FloodBlock(stream api.Peer_FloodBlockServer) error {
	for {
		floodBlockReq, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		// TODO process received block
		grpcBlock := floodBlockReq.GetBlock()
		util.GrpcBlockToDomainBlock(*grpcBlock)
		// block := util.GrpcBlockToDomainBlock(*grpcBlock)
		if err := stream.Send(&pb.FloodResponse{}); err != nil {
			return err
		}
	}
}
func (*peer) FloodChain(srv api.Peer_FloodChainServer) error {
	// TODO
	return status.Errorf(codes.Unimplemented, "method FloodOp not implemented")
}
func (*peer) FloodOp(stream api.Peer_FloodOpServer) error {
	for {
		floodOpReq, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		// TODO process op
		floodOpReq.GetOp()
		// grpcOp := floodOpReq.GetOp()

		if err := stream.Send(&pb.FloodResponse{}); err != nil {
			return err
		}
	}
}

// RegisterMiner - registers miner peer grpc
func RegisterMiner(s *grpc.Server) {
	pb.RegisterPeerServer(s, &peer{})
}
