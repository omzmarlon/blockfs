package reqhandler

import (
	"cs416/P1-v3d0b-q4d0b/bclib"
	"testing"
)

func createTestingChain() bclib.Block{
	block1 := bclib.NewBlock("1", "", "123", nil, 123)
	block2 := bclib.NewBlock("2", "1", "123", nil, 123)
	block3 := bclib.NewBlock("3", "1", "123", nil, 123)
	block4 := bclib.NewBlock("4", "1", "123", nil, 123)
	block5 := bclib.NewBlock("5", "2", "123", nil, 123)
	block6 := bclib.NewBlock("6", "2", "123", nil, 123)
	block7 := bclib.NewBlock("7", "3", "123", nil, 123)
	block8 := bclib.NewBlock("8", "3", "123", nil, 123)
	block9 := bclib.NewBlock("9", "3", "123", nil, 123)
	block10 := bclib.NewBlock("10", "4", "123", nil, 123)
	block11 := bclib.NewBlock("11", "4", "123", nil, 123)
	block12 := bclib.NewBlock("12", "10", "123", nil, 123)
	block13 := bclib.NewBlock("13", "11", "123", nil, 123)
	block14 := bclib.NewBlock("14", "13", "123", nil, 123)
	block15 := bclib.NewBlock("15", "14", "123", nil, 123)

	block14.Children = append(block14.Children, block15)
	block13.Children = append(block13.Children, block14)
	block10.Children = append(block10.Children, block12)
	block11.Children = append(block11.Children, block13)
	block4.Children = append(append(block4.Children, block10), block11)
	block3.Children = append(append(append(block3.Children, block7), block8), block9)
	block2.Children = append(append(block2.Children, block5), block6)
	block1.Children = append(append(append(block1.Children, block2), block2), block4)

	return block1

}

func TestCheckAndReturn (t *testing.T) {
}