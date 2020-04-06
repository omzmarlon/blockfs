package util

import (
	"github.com/omzmarlon/blockfs/pkg/api"
	"github.com/omzmarlon/blockfs/pkg/domain"
)

// GrpcBlockToDomainBlock - converts grpc block type to domain block type
func GrpcBlockToDomainBlock(block api.Block) domain.Block {
	ops := make([]domain.Op, len(block.GetOps()))
	for i := 0; i < len(ops); i++ {
		ops[i] = GrpcOpToDomainOp(*block.GetOps()[i])
	}
	children := make([]domain.Block, len(block.GetChildren()))
	for i := 0; i < len(children); i++ {
		children[i] = GrpcBlockToDomainBlock(*block.GetChildren()[i])
	}
	return domain.Block{
		Hash:     block.GetHash(),
		PrevHash: block.GetPrevHash(),
		MinerID:  block.GetMinerId(),
		Ops:      &ops,
		Nonce:    block.GetNounce(),
		Children: &children,
	}
}

// DomainBlockToGrpcBlock - converts domain block type to grpc block type
func DomainBlockToGrpcBlock(block domain.Block) api.Block {
	ops := make([]*api.Op, len(*block.Ops))
	for i := 0; i < len(*block.Ops); i++ {
		op := DomainOpToGrpcOp((*block.Ops)[i])
		ops[i] = &op
	}
	children := make([]*api.Block, len(*block.Children))
	for i := 0; i < len(*block.Children); i++ {
		child := DomainBlockToGrpcBlock((*block.Children)[i])
		children[i] = &child
	}
	return api.Block{
		Hash:     block.Hash,
		PrevHash: block.PrevHash,
		MinerId:  block.MinerID,
		Ops:      ops,
		Nounce:   block.Nonce,
		Children: children,
	}
}

// GrpcOpToDomainOp - converts grpc Op type to domain Op type
func GrpcOpToDomainOp(op api.Op) domain.Op {
	return domain.Op{
		OpID:     op.GetOpId(),
		MinerID:  op.GetMinerId(),
		OpAction: GrpcOpActionToDomainOpAction(op.GetOpAction()),
		Filename: op.GetFileName(),
		Record:   op.GetBytes(),
	}
}

// DomainOpToGrpcOp - converts domain Op type to grpc Op type
func DomainOpToGrpcOp(op domain.Op) api.Op {
	return api.Op{
		OpId:     op.OpID,
		MinerId:  op.MinerID,
		OpAction: DomainOpActionToGrpcOpAction(op.OpAction),
		FileName: op.Filename,
		Bytes:    op.Record,
	}
}

// GrpcOpActionToDomainOpAction - converts grpc op action to domain op action
func GrpcOpActionToDomainOpAction(action api.Op_OpAction) domain.OpAction {
	switch int(action) {
	case int(domain.OpUNKNOWN):
		return domain.OpUNKNOWN
	case int(domain.OpCREATE):
		return domain.OpCREATE
	case int(domain.OpAPPEND):
		return domain.OpAPPEND
	default:
		return domain.OpUNKNOWN
	}
}

// DomainOpActionToGrpcOpAction - converts domain op action to grpc op action
func DomainOpActionToGrpcOpAction(action domain.OpAction) api.Op_OpAction {
	switch int(action) {
	case int(domain.OpUNKNOWN):
		return api.Op_UNKNOWN
	case int(domain.OpCREATE):
		return api.Op_CREATE
	case int(domain.OpAPPEND):
		return api.Op_APPEND
	default:
		return api.Op_UNKNOWN
	}
}
