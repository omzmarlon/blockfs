package chain

import (
	"cs416/P1-v3d0b-q4d0b/bclib"
	"testing"
	"time"
)

func TestBCUpdateAndBlockValidation(t *testing.T) {

	//COMMENT OUT POW WHEN TESTING THIS
	newOpChan := make(chan bclib.Block)

	bcm := InitBlockChainManager("123", "", 2, 2, 2,2, 2, 2, 2, newOpChan)

	bc := CreateTestingChain()
	bcm.SetBlockRoot(bc)

	if len(FindLongestChain(bcm.GetBlockRoot())) != 5 {
		t.Log("FAILED")
	}
	////// TESTING ADDING TO BLOCK CHAIN////////
	newBlock := bclib.NewBlock("16", "15", "123", nil, 12345)

	bcm.UpdateBlockChain(newBlock)
	longest := FindLongestChain(bcm.GetBlockRoot())
	if len(longest) != 6 {
		t.Log("FAILED")
	}

	////// TESTING VALIDATION OF NEW BLOCK COMING IN////////
	newBlock0 := bclib.NewBlock("16", "200", "123", nil, 12345)

	if bcm.isValidBlock(newBlock0) {
		t.Log("FAILED")
	}

	//checks prev hash block exists
	newBlock1 := bclib.NewBlock("16", "200", "123", nil, 12345)
	if bcm.isValidBlock(newBlock1) {
		t.Log("FAILED")
	}

	//check person has sufficient coin
	newBlock3 := bclib.NewBlock("16", "15", "123", nil, 12345)
	if !bcm.isValidBlock(newBlock3) {
		t.Log("FAILED")
	}

	newOp := bclib.NewROp("1", "123", bclib.CREATE, "123", nil)
	newBlock4 := bclib.NewBlock("16", "15", "123", []bclib.ROp{newOp}, 12345)
	if bcm.isValidBlock(newBlock4) {
		t.Log("FAILED")
	}

	newNoop1 := bclib.NewBlock("17", "16", "123", nil, 12345)
	newNoop2 := bclib.NewBlock("18", "17", "123", nil, 12345)
	newNoop3 := bclib.NewBlock("19", "18", "123", nil, 12345)
	newNoop4 := bclib.NewBlock("20", "19", "123", nil, 12345)

	bcm.UpdateBlockChain(newNoop1)
	bcm.UpdateBlockChain(newNoop2)
	bcm.UpdateBlockChain(newNoop3)
	bcm.UpdateBlockChain(newNoop4)

	newBlock5 := bclib.NewBlock("21", "20", "123", []bclib.ROp{newOp}, 12345)
	if !bcm.isValidBlock(newBlock5) {
		t.Log("FAILED")
	}

	//CHECKING RSF SEMANTICS
	newOp1 := bclib.NewROp("10", "123", bclib.CREATE, "myfile", nil)
	newBlock6 := bclib.NewBlock("21", "20", "123", []bclib.ROp{newOp1}, 12345)

	if bcm.isValidBlock(newBlock6) {
		t.Log("FAILED")
	}

	newOp2 := bclib.NewROp("10", "123", bclib.APPEND, "WRGJHERKGJH", nil)
	newBlock7 := bclib.NewBlock("21", "20", "123", []bclib.ROp{newOp2}, 12345)

	if bcm.isValidBlock(newBlock7) {
		t.Log("FAILED")
	}
	//sanity check
	newBlock8 := bclib.NewBlock("21", "20", "123", nil, 12345)

	if !bcm.isValidBlock(newBlock8) {
		t.Log("FAILED")
	}

}

func TestNewBlockAndChanStuff(t *testing.T) {
	failCounter := 0
	for i := 0; i < 100; i++ {
		newOpChan := make(chan bclib.Block)

		bcm := InitBlockChainManager("123", "1", 2, 2, 2, 2, 2,2, 2, newOpChan)

		block2 := bclib.NewBlock("2", "1", "123", nil, 123)
		block3 := bclib.NewBlock("3", "1", "123", nil, 123)
		block4 := bclib.NewBlock("4", "1", "123", nil, 123)
		block5 := bclib.NewBlock("5", "2", "123", nil, 123)
		block6 := bclib.NewBlock("6", "2", "123", nil, 123)
		block7 := bclib.NewBlock("7", "3", "123", nil, 123)
		block8 := bclib.NewBlock("8", "3", "123", nil, 123)
		block9 := bclib.NewBlock("9", "3", "123", nil, 123)
		block10 := bclib.NewBlock("10", "4", "123", nil, 123)
		block11 := bclib.NewBlock("11", "4", "456", nil, 123)
		block12 := bclib.NewBlock("12", "10", "123", nil, 123)
		block13 := bclib.NewBlock("13", "11", "123", nil, 123)
		block14 := bclib.NewBlock("14", "13", "456", nil, 123)
		block15 := bclib.NewBlock("15", "14", "123", nil, 123)

		newOpChan <- block2
		newOpChan <- block3
		newOpChan <- block4
		newOpChan <- block5
		newOpChan <- block6
		newOpChan <- block7
		newOpChan <- block8
		newOpChan <- block9
		newOpChan <- block10
		newOpChan <- block11
		newOpChan <- block12
		newOpChan <- block13
		newOpChan <- block14
		newOpChan <- block15

		time.Sleep(100 * time.Millisecond)
		root := bcm.GetBlockRoot()

		longestChain := FindLongestChain(root)

		if len(longestChain) != 5 {
			failCounter++
			t.Log(longestChain)
			t.Log(len(longestChain))
			t.Log("FAILED")
		}
	}
	t.Log(failCounter)

}

/*
func TestNewBlockAndChanRequest(t *testing.T) {
	newOpChan := make(chan bclib.Block)

	myChain := CreateTestingChain()

	bcm := InitBlockChainManager("123", "1", newOpChan)

	bcm.SetBlockRoot(myChain)

	newBlock := bclib.NewBlock("16", "15", "123", nil, 123)

	newOpChan <- newBlock

	longestChain := FindLongestChain(bcm.GetBlockRoot())

	if len(longestChain)!= 6 {
		t.Log(len(longestChain))
		t.Log("FAILED")
	}

}*/
