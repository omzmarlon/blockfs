package domain

////////////////////////////////////////////////////////////////////////////////
// <BLOCKCHAIN DEFINITIONS>

// Block is the fundamental unit of a blockchain
type Block struct {
	Hash     string
	PrevHash string
	MinerID  string
	Ops      []Op
	Nonce    uint32
	// pointer to slice but not element of slice because we may append new
	// child (modifying slice) but not modify the individual element, so we they are
	// returned we return by value as a copy (although it's still safer to make a copy)
	// for the ops above, it should never change once created so no pointer to slice
	Children *[]Block
}

// NewBlock constructor
func NewBlock(hash string, prevHash string, minerID string, ops []Op, nonce uint32) Block {
	children := make([]Block, 0)
	return Block{
		Hash:     hash,
		PrevHash: prevHash,
		MinerID:  minerID,
		Ops:      ops,
		Nonce:    nonce,
		Children: &children,
	}
}

// </BLOCKCHAIN DEFINITIONS>
////////////////////////////////////////////////////////////////////////////////

// Op stands for rfs operations that requires talking to *R*emote miners
type Op struct {
	OpID     string // uniquely identify an op in the entire blockchain network
	MinerID  string // the miner who submits this op
	OpAction OpAction
	Filename string
	Record   []byte
}

// NewOp constructor
func NewOp(opID string, minerID string, opAction OpAction, filename string, record []byte) Op {
	return Op{
		OpID:     opID,
		MinerID:  minerID,
		OpAction: opAction,
		Filename: filename,
		Record:   record,
	}
}

// OpAction enum
type OpAction int

const (
	OpUNKNOWN OpAction = 0
	OpCREATE  OpAction = 1
	OpAPPEND  OpAction = 2
)
