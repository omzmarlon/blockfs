package flooding

import (
	"cs416/P1-v3d0b-q4d0b/bclib"
	"cs416/P1-v3d0b-q4d0b/bclib/chain"
	"fmt"
	"testing"
	"time"
)

func Test_FullyConnected3Flooder_LastPeerFlood1Pkt(t *testing.T) {
	// setup
	bcm := chain.InitBlockChainManager("a", "a", 1, 1, 1, 1, 1, 1, 1, make(chan bclib.Block, 1))
	f1QueueOpsInput := make(chan bclib.ROp, 10)
	f2QueueOpsInput := make(chan bclib.ROp, 10)
	f3QueueOpsInput := make(chan bclib.ROp, 10)
	_, err1 := InitFloodingProtocol(FlooderConfig{
		MinerID:            "miner1",
		IncomingMinersAddr: "127.0.0.1:3001",
		OutgoingMinersIP:   "127.0.0.1",
		PeerMinersAddrs:    []string{},
		FDResponseIPPort:   "127.0.0.1:5001",
		LostMsgThresh:      uint8(20),
	}, f1QueueOpsInput, make(chan bclib.Block), bcm)
	_, err2 := InitFloodingProtocol(FlooderConfig{
		MinerID:            "miner2",
		IncomingMinersAddr: "127.0.0.1:3002",
		OutgoingMinersIP:   "127.0.0.1",
		PeerMinersAddrs:    []string{"127.0.0.1:3001"},
		FDResponseIPPort:   "127.0.0.1:5002",
		LostMsgThresh:      uint8(20),
	}, f2QueueOpsInput, make(chan bclib.Block), bcm)

	flooder3, err3 := InitFloodingProtocol(FlooderConfig{
		MinerID:            "miner3",
		IncomingMinersAddr: "127.0.0.1:3003",
		OutgoingMinersIP:   "127.0.0.1",
		PeerMinersAddrs:    []string{"127.0.0.1:3001", "127.0.0.1:3002"},
		FDResponseIPPort:   "127.0.0.1:5003",
		LostMsgThresh:      uint8(20),
	}, f3QueueOpsInput, make(chan bclib.Block), bcm)
	if err1 != nil {
		t.Errorf("Failed to create f1: %s", err1)
		t.Fail()
	}
	if err2 != nil {
		t.Errorf("Failed to create f2: %s", err2)
		t.Fail()
	}
	if err3 != nil {
		t.Errorf("Failed to create f3: %s", err3)
		t.Fail()
	}
	f3OpsInput := flooder3.GetFloodingOpsInputChann()

	time.Sleep(1 * time.Second)
	// test
	f3OpsInput <- bclib.NewROp("opid1", "miner1", bclib.CREATE, "", []byte{})
	time.Sleep(1 * time.Second)

	// verify
	if getNumberOfOps(f1QueueOpsInput) != 1 {
		t.Fail()
	}

	if getNumberOfOps(f2QueueOpsInput) != 1 {
		fmt.Println(1234)
		t.Fail()
	}

	if getNumberOfOps(f3QueueOpsInput) != 0 {
		fmt.Println(1235)
		t.Fail()
	}
}

func Test_FullyConnected3Flooder_FirstPeerFlood1Pkt(t *testing.T) {
	bcm := chain.InitBlockChainManager("a", "a", 1, 1, 1, 1, 1, 1, 1, make(chan bclib.Block, 1))
	// setup
	f1QueueOpsInput := make(chan bclib.ROp, 10)
	f2QueueOpsInput := make(chan bclib.ROp, 10)
	f3QueueOpsInput := make(chan bclib.ROp, 10)
	flooder1, err1 := InitFloodingProtocol(FlooderConfig{
		MinerID:            "miner1",
		IncomingMinersAddr: "127.0.0.1:3011",
		OutgoingMinersIP:   "127.0.0.1",
		PeerMinersAddrs:    []string{},
		FDResponseIPPort:   "127.0.0.1:5011",
		LostMsgThresh:      uint8(20),
	}, f1QueueOpsInput, make(chan bclib.Block), bcm)
	_, err2 := InitFloodingProtocol(FlooderConfig{
		MinerID:            "miner2",
		IncomingMinersAddr: "127.0.0.1:3012",
		OutgoingMinersIP:   "127.0.0.1",
		PeerMinersAddrs:    []string{"127.0.0.1:3011"},
		FDResponseIPPort:   "127.0.0.1:5012",
		LostMsgThresh:      uint8(20),
	}, f2QueueOpsInput, make(chan bclib.Block), bcm)

	_, err3 := InitFloodingProtocol(FlooderConfig{
		MinerID:            "miner3",
		IncomingMinersAddr: "127.0.0.1:3013",
		OutgoingMinersIP:   "127.0.0.1",
		PeerMinersAddrs:    []string{"127.0.0.1:3011", "127.0.0.1:3012"},
		FDResponseIPPort:   "127.0.0.1:5013",
		LostMsgThresh:      uint8(20),
	}, f3QueueOpsInput, make(chan bclib.Block), bcm)
	if err1 != nil {
		t.Errorf("Failed to create f1: %s", err1)
		t.Fail()
	}
	if err2 != nil {
		t.Errorf("Failed to create f2: %s", err2)
		t.Fail()
	}
	if err3 != nil {
		t.Errorf("Failed to create f3: %s", err3)
		t.Fail()
	}

	// test
	f1OpsInput := flooder1.GetFloodingOpsInputChann()
	time.Sleep(1 * time.Second)
	f1OpsInput <- bclib.NewROp("opid1", "miner1", bclib.CREATE, "", []byte{})
	time.Sleep(1 * time.Second)

	// verify
	if getNumberOfOps(f1QueueOpsInput) != 0 {
		t.Fail()
	}

	if getNumberOfOps(f2QueueOpsInput) != 1 {
		t.Fail()
	}

	if getNumberOfOps(f3QueueOpsInput) != 1 {
		t.Fail()
	}

}

func Test_FullyConnected3Flooder_FirstPeerFloodMultiplePkt(t *testing.T) {
	bcm := chain.InitBlockChainManager("a", "a", 1, 1, 1, 1, 1, 1, 1, make(chan bclib.Block, 1))
	// setup
	f1QueueOpsInput := make(chan bclib.ROp, 10)
	f2QueueOpsInput := make(chan bclib.ROp, 10)
	f3QueueOpsInput := make(chan bclib.ROp, 10)
	flooder1, err1 := InitFloodingProtocol(FlooderConfig{
		MinerID:            "miner1",
		IncomingMinersAddr: "127.0.0.1:3021",
		OutgoingMinersIP:   "127.0.0.1",
		PeerMinersAddrs:    []string{},
		FDResponseIPPort:   "127.0.0.1:5021",
		LostMsgThresh:      uint8(20),
	}, f1QueueOpsInput, make(chan bclib.Block), bcm)
	_, err2 := InitFloodingProtocol(FlooderConfig{
		MinerID:            "miner2",
		IncomingMinersAddr: "127.0.0.1:3022",
		OutgoingMinersIP:   "127.0.0.1",
		PeerMinersAddrs:    []string{"127.0.0.1:3021"},
		FDResponseIPPort:   "127.0.0.1:5022",
		LostMsgThresh:      uint8(20),
	}, f2QueueOpsInput, make(chan bclib.Block), bcm)

	_, err3 := InitFloodingProtocol(FlooderConfig{
		MinerID:            "miner3",
		IncomingMinersAddr: "127.0.0.1:3023",
		OutgoingMinersIP:   "127.0.0.1",
		PeerMinersAddrs:    []string{"127.0.0.1:3021", "127.0.0.1:3022"},
		FDResponseIPPort:   "127.0.0.1:5023",
		LostMsgThresh:      uint8(20),
	}, f3QueueOpsInput, make(chan bclib.Block), bcm)
	if err1 != nil {
		t.Errorf("Failed to create f1: %s", err1)
		t.Fail()
	}
	if err2 != nil {
		t.Errorf("Failed to create f2: %s", err2)
		t.Fail()
	}
	if err3 != nil {
		t.Errorf("Failed to create f3: %s", err3)
		t.Fail()
	}

	// test
	f1OpsInput := flooder1.GetFloodingOpsInputChann()
	time.Sleep(1 * time.Second)
	f1OpsInput <- bclib.NewROp("opid1", "miner1", bclib.CREATE, "", []byte{})
	f1OpsInput <- bclib.NewROp("opid2", "miner1", bclib.CREATE, "", []byte{})
	f1OpsInput <- bclib.NewROp("opid3", "miner1", bclib.CREATE, "", []byte{})
	time.Sleep(1 * time.Second)

	// verify
	if getNumberOfOps(f1QueueOpsInput) != 0 {
		t.Fail()
	}

	if getNumberOfOps(f2QueueOpsInput) != 3 {
		t.Fail()
	}

	if getNumberOfOps(f3QueueOpsInput) != 3 {
		t.Fail()
	}

}

func Test_FullyConnected3FlooderWith1Additional_FirstPeerFloodM1PktTheAdditiaonlFlood1Pkt(t *testing.T) {
	bcm := chain.InitBlockChainManager("a", "a", 1, 1, 1, 1, 1, 1, 1, make(chan bclib.Block, 1))
	// setup
	f1QueueOpsInput := make(chan bclib.ROp, 10)
	f2QueueOpsInput := make(chan bclib.ROp, 10)
	f3QueueOpsInput := make(chan bclib.ROp, 10)
	f4QueueOpsInput := make(chan bclib.ROp, 10)
	flooder1, err1 := InitFloodingProtocol(FlooderConfig{
		MinerID:            "miner1",
		IncomingMinersAddr: "127.0.0.1:3031",
		OutgoingMinersIP:   "127.0.0.1",
		PeerMinersAddrs:    []string{},
		FDResponseIPPort:   "127.0.0.1:5031",
		LostMsgThresh:      uint8(20),
	}, f1QueueOpsInput, make(chan bclib.Block), bcm)
	_, err2 := InitFloodingProtocol(FlooderConfig{
		MinerID:            "miner2",
		IncomingMinersAddr: "127.0.0.1:3032",
		OutgoingMinersIP:   "127.0.0.1",
		PeerMinersAddrs:    []string{"127.0.0.1:3031"},
		FDResponseIPPort:   "127.0.0.1:5032",
		LostMsgThresh:      uint8(20),
	}, f2QueueOpsInput, make(chan bclib.Block), bcm)

	_, err3 := InitFloodingProtocol(FlooderConfig{
		MinerID:            "miner3",
		IncomingMinersAddr: "127.0.0.1:3033",
		OutgoingMinersIP:   "127.0.0.1",
		PeerMinersAddrs:    []string{"127.0.0.1:3031", "127.0.0.1:3032"},
		FDResponseIPPort:   "127.0.0.1:5033",
		LostMsgThresh:      uint8(20),
	}, f3QueueOpsInput, make(chan bclib.Block), bcm)
	flooder4, err4 := InitFloodingProtocol(FlooderConfig{
		MinerID:            "miner4",
		IncomingMinersAddr: "127.0.0.1:3034",
		OutgoingMinersIP:   "127.0.0.1",
		PeerMinersAddrs:    []string{"127.0.0.1:3033"},
		FDResponseIPPort:   "127.0.0.1:5034",
		LostMsgThresh:      uint8(20),
	}, f4QueueOpsInput, make(chan bclib.Block), bcm)
	if err1 != nil {
		t.Errorf("Failed to create f1: %s", err1)
		t.Fail()
	}
	if err2 != nil {
		t.Errorf("Failed to create f2: %s", err2)
		t.Fail()
	}
	if err3 != nil {
		t.Errorf("Failed to create f3: %s", err3)
		t.Fail()
	}
	if err4 != nil {
		t.Errorf("Failed to create f4: %s", err4)
		t.Fail()
	}

	// test
	f1OpsInput := flooder1.GetFloodingOpsInputChann()
	f4OpsInput := flooder4.GetFloodingOpsInputChann()
	time.Sleep(1 * time.Second)
	f1OpsInput <- bclib.NewROp("opid1", "miner1", bclib.CREATE, "", []byte{})
	f4OpsInput <- bclib.NewROp("opid2", "miner1", bclib.CREATE, "", []byte{})
	time.Sleep(1 * time.Second)

	// verify
	if getNumberOfOps(f1QueueOpsInput) != 1 {
		t.Fail()
	}

	if getNumberOfOps(f2QueueOpsInput) != 2 {
		t.Fail()
	}

	if getNumberOfOps(f3QueueOpsInput) != 2 {
		t.Fail()
	}

	if getNumberOfOps(f4QueueOpsInput) != 1 {
		t.Fail()
	}

}

// func Test_FullyConnected3Flooder_LastPeerFlood1PktAndThenSecondFlooderFails(t *testing.T) {
// 	// setup
// 	f1QueueOpsInput := make(chan bclib.ROp, 10)
// 	f2QueueOpsInput := make(chan bclib.ROp, 10)
// 	f3QueueOpsInput := make(chan bclib.ROp, 10)
// 	flooder1, err1 := InitFloodingProtocol(FlooderConfig{
// 		MinerID:            "miner1",
// 		IncomingMinersAddr: "127.0.0.1:3051",
// 		OutgoingMinersIP:   "127.0.0.1",
// 		PeerMinersAddrs:    []string{},
// 		FDResponseIPPort:   "127.0.0.1:5051",
// 		LostMsgThresh:      uint8(20),
// 	}, f1QueueOpsInput)
// 	flooder2, err2 := InitFloodingProtocol(FlooderConfig{
// 		MinerID:            "miner2",
// 		IncomingMinersAddr: "127.0.0.1:3052",
// 		OutgoingMinersIP:   "127.0.0.1",
// 		PeerMinersAddrs:    []string{"127.0.0.1:3051"},
// 		FDResponseIPPort:   "127.0.0.1:5052",
// 		LostMsgThresh:      uint8(20),
// 	}, f2QueueOpsInput)

// 	flooder3, err3 := InitFloodingProtocol(FlooderConfig{
// 		MinerID:            "miner3",
// 		IncomingMinersAddr: "127.0.0.1:3053",
// 		OutgoingMinersIP:   "127.0.0.1",
// 		PeerMinersAddrs:    []string{"127.0.0.1:3051", "127.0.0.1:3052"},
// 		FDResponseIPPort:   "127.0.0.1:5053",
// 		LostMsgThresh:      uint8(20),
// 	}, f3QueueOpsInput)
// 	if err1 != nil {
// 		t.Errorf("Failed to create f1: %s", err1)
// 		t.Fail()
// 	}
// 	if err2 != nil {
// 		t.Errorf("Failed to create f2: %s", err2)
// 		t.Fail()
// 	}
// 	if err3 != nil {
// 		t.Errorf("Failed to create f3: %s", err3)
// 		t.Fail()
// 	}
// 	f1OpsInput := flooder1.GetFloodingOpsInputChann()
// 	f3OpsInput := flooder3.GetFloodingOpsInputChann()

// 	time.Sleep(1 * time.Second)
// 	// test
// 	// regular flooding
// 	f3OpsInput <- bclib.NewROp("opid1", "miner1", bclib.CREATE, "", []byte{})
// 	time.Sleep(1 * time.Second)

// 	// verify
// 	if getNumberOfOps(f1QueueOpsInput) != 1 {
// 		t.Fail()
// 	}

// 	if getNumberOfOps(f2QueueOpsInput) != 1 {
// 		t.Fail()
// 	}

// 	if getNumberOfOps(f3QueueOpsInput) != 0 {
// 		t.Fail()
// 	}

// 	// test
// 	// disconntect/fail flooder 2
// 	time.Sleep(2 * time.Second)
// 	flooder2.fd.StopResponding()
// 	time.Sleep(1 * time.Second)
// 	f3OpsInput <- bclib.NewROp("opid2", "miner1", bclib.CREATE, "", []byte{})
// 	f3OpsInput <- bclib.NewROp("opid3", "miner1", bclib.CREATE, "", []byte{})
// 	f1OpsInput <- bclib.NewROp("opid4", "miner1", bclib.CREATE, "", []byte{})
// 	f1OpsInput <- bclib.NewROp("opid5", "miner1", bclib.CREATE, "", []byte{})
// 	time.Sleep(1 * time.Second)
// 	// verify
// 	if getNumberOfOps(f1QueueOpsInput) != 2 {
// 		fmt.Println(123)
// 		t.Fail()
// 	}

// 	if getNumberOfOps(f2QueueOpsInput) != 0 {
// 		t.Fail()
// 	}

// 	if getNumberOfOps(f3QueueOpsInput) != 2 {
// 		fmt.Println(12366)
// 		t.Fail()
// 	}
// }

func getNumberOfOps(chann chan bclib.ROp) int {
	count := 0
	for {
		select {
		case <-chann:
			count++
		case <-time.After(500 * time.Millisecond):
			return count
		}
	}
}
