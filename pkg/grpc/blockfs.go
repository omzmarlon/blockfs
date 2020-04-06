package grpc

import (
	"context"

	pb "github.com/omzmarlon/blockfs/pkg/api"
	status "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
)

// TODO:
// - actual implentation of the blockfs service

// blockfs is the grpc server for service the block file system
type blockfs struct {
}

// CreateFile - handle file create in the blockfs
func (*blockfs) CreateFile(ctx context.Context, req *pb.CreateFileRequest) (*pb.CreateFileResponse, error) {
	return &pb.CreateFileResponse{
		Status: &status.Status{Code: int32(codes.OK)},
	}, nil
}

// ListFiles - list all files existing in the blockfs
func (*blockfs) ListFiles(ctx context.Context, req *pb.ListFilesRequest) (*pb.ListFilesResponse, error) {
	return &pb.ListFilesResponse{
		Status:    &status.Status{Code: int32(codes.OK)},
		FileNames: make([]string, 0),
	}, nil
}

// TotalRecs - total number of records in a file
func (*blockfs) TotalRecs(ctx context.Context, req *pb.TotalRecsRequest) (*pb.TotalRecsResponse, error) {
	return &pb.TotalRecsResponse{
		Status:  &status.Status{Code: int32(codes.OK)},
		NumRecs: 0,
	}, nil
}

// ReadRec - read a specific record in a file
func (*blockfs) ReadRec(ctx context.Context, req *pb.ReadRecRequest) (*pb.ReadRecResponse, error) {
	return &pb.ReadRecResponse{
		Status: &status.Status{Code: int32(codes.OK)},
		Record: &pb.Record{Bytes: make([]byte, 0)},
	}, nil
}

// AppendRec - append a record to a file
func (*blockfs) AppendRec(ctx context.Context, req *pb.AppendRecRequest) (*pb.AppendRecResponse, error) {
	return &pb.AppendRecResponse{
		Status:    &status.Status{Code: int32(codes.OK)},
		RecordNum: 0,
	}, nil
}

// RegisterBlockFS - registers the blockfs rpc service
func RegisterBlockFS(s *grpc.Server) {
	pb.RegisterBlockFSServer(s, &blockfs{})
}
