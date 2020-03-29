package bclib

// All blockchain domain types go here...

var Done = make(chan int)

// Block - The fundamental unit of a blockchain
type Block struct {
	Hash     string
	PrevHash string
	MinerID  string
	Ops      []ROp
	Nonce    uint32
	Children []Block
}

// NewBlock constructor
func NewBlock(hash string, prevHash string, minerID string, ops []ROp, nonce uint32) Block {
	return Block{
		Hash:     hash,
		PrevHash: prevHash,
		MinerID:  minerID,
		Ops:      ops,
		Nonce:    nonce,
		Children: []Block{},
	}
}

// ROp stands for rfs operations that requires talking to *R*emote miners
type ROp struct {
	OpID     string // uniquely identify an op in the entire blockchain network
	MinerID  string // the miner who submits this op
	OpAction ROpAction
	Filename string
	Record   []byte
}

// NewROp constructor
func NewROp(opID string, minerID string, opAction ROpAction, filename string, record []byte) ROp {
	return ROp{
		OpID:     opID,
		MinerID:  minerID,
		OpAction: opAction,
		Filename: filename,
		Record:   record,
	}
}

// ROpAction enum
type ROpAction int

const (
	CREATE ROpAction = 0
	APPEND ROpAction = 1
)
