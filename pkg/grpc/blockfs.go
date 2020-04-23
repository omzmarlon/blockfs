package grpc

import (
	"context"

	guuid "github.com/google/uuid"
	pb "github.com/omzmarlon/blockfs/pkg/api"
	"github.com/omzmarlon/blockfs/pkg/blockchain"
	"github.com/omzmarlon/blockfs/pkg/domain"
	status "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
)

// TODO:
// - actual implentation of the blockfs service

// blockfs is the grpc server for service the block file system
type blockfs struct {
	conf          BlockFSConf
	opsProcessing chan<- *domain.Op
	opsFlooding   chan<- *domain.Op
	blockchain    *blockchain.Blockchain
}

type BlockFSConf struct {
	MinerID       string
	ConfirmLength int
}

// CreateFile - handle file create in the blockfs
func (fs *blockfs) CreateFile(ctx context.Context, req *pb.CreateFileRequest) (*pb.CreateFileResponse, error) {
	op := &domain.Op{
		OpID:     guuid.New().String(),
		MinerID:  fs.conf.MinerID,
		OpAction: domain.OpCREATE,
		Filename: req.FileName,
		Record:   make([]byte, 0),
	}
	fs.opsProcessing <- op
	fs.opsFlooding <- op
	fs.waitUntilConfirmed(op.OpID)
	return &pb.CreateFileResponse{
		Status: &status.Status{Code: int32(codes.OK)},
	}, nil
}

// ListFiles - list all files existing in the blockfs
func (fs *blockfs) ListFiles(ctx context.Context, req *pb.ListFilesRequest) (*pb.ListFilesResponse, error) {
	data := fs.blockchain.GetData()
	files := make([]string, 0)
	for _, ops := range data {
		for _, op := range ops {
			if op.OpAction == domain.OpCREATE {
				files = append(files, op.Filename)
			}
		}
	}
	return &pb.ListFilesResponse{
		Status:    &status.Status{Code: int32(codes.OK)},
		FileNames: files,
	}, nil
}

// TotalRecs - total number of records in a file
func (fs *blockfs) TotalRecs(ctx context.Context, req *pb.TotalRecsRequest) (*pb.TotalRecsResponse, error) {
	count := uint32(0)
	data := fs.blockchain.GetData()
	for _, ops := range data {
		for _, op := range ops {
			if op.Filename == req.FileName && op.OpAction == domain.OpAPPEND {
				count++
			}
		}
	}
	return &pb.TotalRecsResponse{
		Status:  &status.Status{Code: int32(codes.OK)},
		NumRecs: count,
	}, nil
}

// ReadRec - read a specific record in a file
func (fs *blockfs) ReadRec(ctx context.Context, req *pb.ReadRecRequest) (*pb.ReadRecResponse, error) {
	data := fs.blockchain.GetData()
	index := uint32(0)
	for _, ops := range data {
		for _, op := range ops {
			if op.Filename == req.FileName && op.OpAction == domain.OpAPPEND {
				if index == req.RecordNum {
					return &pb.ReadRecResponse{
						Status: &status.Status{Code: int32(codes.OK)},
						Record: &pb.Record{Bytes: op.Record},
					}, nil
				} else {
					index++
				}
			}
		}
	}
	return &pb.ReadRecResponse{
		Status: &status.Status{Code: int32(codes.NotFound)},
		Record: &pb.Record{Bytes: make([]byte, 0)},
	}, nil
}

// AppendRec - append a record to a file
func (fs *blockfs) AppendRec(ctx context.Context, req *pb.AppendRecRequest) (*pb.AppendRecResponse, error) {
	op := &domain.Op{
		OpID:     guuid.New().String(),
		MinerID:  "",
		OpAction: domain.OpAPPEND,
		Filename: req.FileName,
		Record:   req.Record.Bytes,
	}
	fs.opsProcessing <- op
	fs.opsFlooding <- op
	fs.waitUntilConfirmed(op.OpID)
	return &pb.AppendRecResponse{
		Status:    &status.Status{Code: int32(codes.OK)},
		RecordNum: 0,
	}, nil
}

func (fs *blockfs) waitUntilConfirmed(targetID string) {
	confirmed := false
	for !confirmed {
		// wait until submitted op is confirmed by blockchain
		chain := fs.blockchain.CloneLongestChain()
		for {
			for _, chainOp := range chain.Ops {
				if chainOp.OpID == targetID && chain.Metadata.LongestChainLength >= uint64(fs.conf.ConfirmLength) {
					confirmed = true
				}
			}
			if confirmed || len(*chain.Children) == 0 {
				break
			}
			chain = *((*chain.Children)[0])
		}
	}
}

// RegisterBlockFS - registers the blockfs rpc service
func RegisterBlockFS(s *grpc.Server, conf BlockFSConf, opsProcessing chan<- *domain.Op, opsFlooding chan<- *domain.Op, blockchain *blockchain.Blockchain) {
	pb.RegisterBlockFSServer(s, &blockfs{
		conf:          conf,
		opsProcessing: opsProcessing,
		blockchain:    blockchain,
		opsFlooding:   opsFlooding,
	})
}
