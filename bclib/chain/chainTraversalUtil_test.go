package chain

import (
	"cs416/P1-v3d0b-q4d0b/bclib"
	"fmt"
	"testing"
)

func CreateTestingChain() bclib.Block {
	myOp := bclib.NewROp("1", "123", bclib.CREATE, "myfile", nil)
	myAppend := bclib.NewROp("1", "123", bclib.APPEND, "myfile", nil)
	badOps := []bclib.ROp{myOp, myOp, myOp, myOp, myOp, myOp, myOp, myOp, myOp, myOp, myOp, myOp, myOp, myOp, myOp}
	block1 := bclib.NewBlock("1", "", "", nil, 123)
	block2 := bclib.NewBlock("2", "1", "123", badOps, 123)
	block3 := bclib.NewBlock("3", "1", "123", nil, 123)
	block4 := bclib.NewBlock("4", "1", "123", nil, 123)
	block5 := bclib.NewBlock("5", "2", "123", badOps, 123)
	block6 := bclib.NewBlock("6", "2", "123", badOps, 123)
	block7 := bclib.NewBlock("7", "3", "123", badOps, 123)
	block8 := bclib.NewBlock("8", "3", "123", badOps, 123)
	block9 := bclib.NewBlock("9", "3", "123", nil, 123)
	block10 := bclib.NewBlock("10", "4", "123", badOps, 123)
	block11 := bclib.NewBlock("11", "4", "456", []bclib.ROp{myOp, myOp}, 123)
	block12 := bclib.NewBlock("12", "10", "123", nil, 123)
	block13 := bclib.NewBlock("13", "11", "123", nil, 123)
	block14 := bclib.NewBlock("14", "13", "456", nil, 123)
	block15 := bclib.NewBlock("15", "14", "123", []bclib.ROp{myOp, myAppend}, 123)

	block14.Children = append(block14.Children, block15)
	block13.Children = append(block13.Children, block14)
	block10.Children = append(block10.Children, block12)
	block11.Children = append(block11.Children, block13)
	block4.Children = append(append(block4.Children, block10), block11)
	block3.Children = append(append(append(block3.Children, block7), block8), block9)
	block2.Children = append(append(block2.Children, block5), block6)
	block1.Children = append(append(append(block1.Children, block2), block3), block4)

	return block1

}

func CreateTestingChainWithNoOps() bclib.Block {
	block1 := bclib.NewBlock("1", "0", "", nil, 123)
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

	block14.Children = append(block14.Children, block15)
	block13.Children = append(block13.Children, block14)
	block10.Children = append(block10.Children, block12)
	block11.Children = append(block11.Children, block13)
	block4.Children = append(append(block4.Children, block10), block11)
	block3.Children = append(append(append(block3.Children, block7), block8), block9)
	block2.Children = append(append(block2.Children, block5), block6)
	block1.Children = append(append(append(block1.Children, block2), block3), block4)

	return block1
}

func CreateMultipleLongestChains() bclib.Block {
	block1 := bclib.NewBlock("1", "", "", nil, 123)
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
	block16 := bclib.NewBlock("16", "15", "123", nil, 123)
	block17 := bclib.NewBlock("17", "16", "123", nil, 123)

	block16.Children = append(block16.Children, block17)
	block14.Children = append(block14.Children, block15)
	block13.Children = append(block13.Children, block14)
	block12.Children = append(block12.Children, block16)
	block10.Children = append(block10.Children, block12)
	block11.Children = append(block11.Children, block13)
	block4.Children = append(append(block4.Children, block10), block11)
	block3.Children = append(append(append(block3.Children, block7), block8), block9)
	block2.Children = append(append(block2.Children, block5), block6)
	block1.Children = append(append(append(block1.Children, block2), block3), block4)

	return block1
}

func CreateSingleChildOpChain() bclib.Block {
	create1 := bclib.NewROp("1", "123", bclib.CREATE, "1", []byte{1})
	append1 := bclib.NewROp("1", "123", bclib.APPEND, "1", []byte{2})
	create2 := bclib.NewROp("1", "123", bclib.CREATE, "2", []byte{1})
	create3 := bclib.NewROp("1", "123", bclib.CREATE, "3", []byte{1})
	append1_1 := bclib.NewROp("1", "123", bclib.APPEND, "1", []byte{3})
	append3 := bclib.NewROp("1", "123", bclib.APPEND, "3", []byte{2})
	create4 := bclib.NewROp("1", "123", bclib.CREATE, "4", []byte{1})
	create5 := bclib.NewROp("1", "123", bclib.CREATE, "5", []byte{1})
	append3_1 := bclib.NewROp("1", "123", bclib.APPEND, "3", []byte{4})
	create6 := bclib.NewROp("1", "123", bclib.CREATE, "6", []byte{1})

	appendfail1 := bclib.NewROp("1", "123", bclib.APPEND, "1", []byte{69})
	appendfail2 := bclib.NewROp("1", "123", bclib.APPEND, "2", []byte{69})
	appendfail3 := bclib.NewROp("1", "123", bclib.APPEND, "3", []byte{69})
	appendfail4 := bclib.NewROp("1", "123", bclib.APPEND, "4", []byte{69})
	appendfail5 := bclib.NewROp("1", "123", bclib.APPEND, "5", []byte{69})
	appendfail6 := bclib.NewROp("1", "123", bclib.APPEND, "6", []byte{69})

	b1ops := []bclib.ROp{create1}
	b2ops := []bclib.ROp{append1}
	b3ops := []bclib.ROp{create2, create3}
	b4ops := []bclib.ROp{append1_1, append3}
	b5ops := []bclib.ROp{create4, create5, append3_1}
	b6ops := []bclib.ROp{create6, appendfail1, appendfail2, appendfail3, appendfail4, appendfail5, appendfail6}

	block6 := bclib.NewBlock("6", "5", "123", b6ops, 123)
	block5 := bclib.NewBlock("5", "4", "123", b5ops, 123)
	block4 := bclib.NewBlock("4", "3", "123", b4ops, 123)
	block3 := bclib.NewBlock("3", "2", "123", b3ops, 123)
	block2 := bclib.NewBlock("2", "1", "123", b2ops, 123)
	block1 := bclib.NewBlock("1", "og", "123", b1ops, 123)
	ogBlock := bclib.NewBlock("og", "", "123", nil, 123)

	block5.Children = append(block5.Children, block6)
	block4.Children = append(block4.Children, block5)
	block3.Children = append(block3.Children, block4)
	block2.Children = append(block2.Children, block3)
	block1.Children = append(block1.Children, block2)
	ogBlock.Children = append(ogBlock.Children, block1)
	return ogBlock

}

func TestFindingLongestChain(t *testing.T) {
	bc := CreateTestingChain()
	longestChain := FindLongestChain(bc)

	if len(longestChain) != 5 {
		t.Log("FAILED")
	}
}

func TestFindingLongestChainWithRoot(t *testing.T) {
	block1 := bclib.NewBlock("1", "", "123", nil, 123)
	longestChain := FindLongestChain(block1)

	if len(longestChain) != 0 {
		t.Log("FAILED")
	}
}

func TestCheckBlockExists(t *testing.T) {
	bc := CreateTestingChain()

	if !CheckBlockExists(bc, "1") {
		t.Log("FAILED")
	}

	if !CheckBlockExists(bc, "2") {
		t.Log("FAILED")
	}
	if !CheckBlockExists(bc, "3") {
		t.Log("FAILED")
	}
	if !CheckBlockExists(bc, "4") {
		t.Log("FAILED")
	}
	if !CheckBlockExists(bc, "5") {
		t.Log("FAILED")
	}
	if !CheckBlockExists(bc, "6") {
		t.Log("FAILED")
	}
	if !CheckBlockExists(bc, "7") {
		t.Log("FAILED")
	}
	if !CheckBlockExists(bc, "8") {
		t.Log("FAILED")
	}
	if !CheckBlockExists(bc, "9") {
		t.Log("FAILED")
	}
	if !CheckBlockExists(bc, "10") {
		t.Log("FAILED")
	}
	if !CheckBlockExists(bc, "11") {
		t.Log("FAILED")
	}
	if !CheckBlockExists(bc, "12") {
		t.Log("FAILED")
	}
	if !CheckBlockExists(bc, "13") {
		t.Log("FAILED")
	}
	if !CheckBlockExists(bc, "14") {
		t.Log("FAILED")
	}
	if !CheckBlockExists(bc, "15") {
		t.Log("FAILED")
	}
	if CheckBlockExists(bc, "16") {
		t.Log("FAILED")
	}
}

func TestCountCoins(t *testing.T) {
	bc := CreateTestingChain()
	coins := CountCoins(bc, 9, 5, 11, "123")
	if coins != -15 {
		t.Log(coins)
		t.Log("FAILED")
	}

	coins2 := CountCoins(bc, 9, 5, 11, "456")
	if coins2 != 14 {
		t.Log(coins2)
		t.Log("FAILED")
	}

	coins3 := CountCoins(bc, 9, 5, 11, "789")
	if coins3 != 0 {
		t.Log(coins3)
		t.Log("FAILED")
	}

}

func TestGetAllFileNamesOnLongestChain(t *testing.T) {
	bcNoOp := CreateTestingChainWithNoOps()
	longestChain0 := FindLongestChain(bcNoOp)
	filesNoOp := GetAllFileNamesOnChain(bcNoOp, longestChain0)

	if len(filesNoOp) != 0 {
		t.Log("FAILED")
	}

	bc := CreateTestingChain()
	longestChain := FindLongestChain(bc)
	files := GetAllFileNamesOnChain(bc, longestChain)

	if len(files) != 1 {
		t.Log("FAILED")
	}

}

func TestPrintBC(t *testing.T) {
	bc := CreateTestingChainWithNoOps()
	PrintBlockChain("miner1", bc)
}

func TestFindLongestChains(t *testing.T) {
	bc := CreateMultipleLongestChains()
	fmt.Println(FindLongestChains(bc))
}

func TestGetAllFileNamesOnChain(t *testing.T) {
	bc := CreateSingleChildOpChain()
	longestChain := FindLongestChain(bc)
	allfiles := GetAllFileNamesOnChain(bc, longestChain)
	if len(allfiles) != 3 {
		t.Log("FAILED")
	}

	numRecs1 := CountRecNumWithFname(bc, "1", 2, 1)
	if numRecs1 != 3 {
		t.Log(numRecs1)
		t.Log("FAILED num recs 1")
	}

	numRecs2 := CountRecNumWithFname(bc, "2", 2, 1)
	if numRecs2 != 1 {
		t.Log(numRecs2)
		t.Log("FAILED num recs 2")
	}

	numRecs3 := CountRecNumWithFname(bc, "3", 2, 1)
	if numRecs3 != 3 {
		t.Log(numRecs3)
		t.Log("FAILED num recs 3")
	}

	numRecs4 := CountRecNumWithFname(bc, "1", 4, 4)
	if numRecs4 != 2 {
		t.Log(numRecs4)
		t.Log("FAILED num recs 4")
	}
}

func TestReadRec(t *testing.T) {
	bc := CreateSingleChildOpChain()
	longestChain := FindLongestChain(bc)

	test1, _ := ReadRec(bc, longestChain, "1", 2, 1)
	fmt.Println(test1)
}
