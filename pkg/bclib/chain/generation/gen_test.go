package generation

// import (
// 	"cs416/P1-v3d0b-q4d0b/bclib"
// 	"cs416/P1-v3d0b-q4d0b/bclib/chain"
// 	"testing"
// )

// func TestGenerateNoOpBlock(t *testing.T) {
// 	// setup
// 	minerID := "abcd"
// 	prevHash := "prev hash"
// 	opD := uint8(2)
// 	noopD := uint8(2)

// 	// test
// 	blocksOutput := make(chan bclib.Block)
// 	prevHashInput, opsInput := InitBlockGenerator(minerID, opD, noopD, blocksOutput)
// 	opsInput <- []bclib.ROp{}
// 	t.Log("Sent ops")
// 	prevHashInput <- prevHash
// 	t.Log("Sent prevHash")

// 	// verify
// 	block := <-blocksOutput
// 	t.Logf("Generated block: %+v\n", block)
// 	t.Logf("Generated block with hash: %s\n", chain.ComputeHashString(block.PrevHash, block.MinerID, block.Nonce, block.Ops))
// }

// func TestGenerateOpBlock(t *testing.T) {
// 	// setup
// 	minerID := "abcd"
// 	prevHash := "prev hash"
// 	opD := uint8(1)
// 	noopD := uint8(2)

// 	// test
// 	blocksOutput := make(chan bclib.Block)
// 	prevHashInput, opsInput := InitBlockGenerator(minerID, opD, noopD, blocksOutput)
// 	opsInput <- []bclib.ROp{bclib.NewROp("op1", minerID, bclib.CREATE, "", []byte{1, 2, 3})}
// 	t.Log("Sent ops")
// 	prevHashInput <- prevHash
// 	t.Log("Sent prevHash")

// 	// verify
// 	block := <-blocksOutput
// 	t.Logf("Generated block: %+v\n", block)
// 	t.Logf("Generated block with hash: %s\n", chain.ComputeHashString(block.PrevHash, block.MinerID, block.Nonce, block.Ops))
// }
