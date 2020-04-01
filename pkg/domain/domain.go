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
	Children []Block
}

// NewBlock constructor
func NewBlock(hash string, prevHash string, minerID string, ops []Op, nonce uint32) Block {
	return Block{
		Hash:     hash,
		PrevHash: prevHash,
		MinerID:  minerID,
		Ops:      ops,
		Nonce:    nonce,
		Children: []Block{},
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
	CREATE OpAction = 0
	APPEND OpAction = 1
)
