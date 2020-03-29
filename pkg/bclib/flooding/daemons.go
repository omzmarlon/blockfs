package flooding

import (
	"io"
	"log"
	"net"
	"time"

	"github.com/omzmarlon/blockfs/pkg/bclib"
	"github.com/omzmarlon/blockfs/pkg/fdlib"
)

func (flooder *Flooder) minerFailureDetectionDaemon(notifyCh <-chan fdlib.FailureDetected) {
	for {
		select {
		case minerFailure := <-notifyCh:
			remoteIPPort := minerFailure.UDPIpPort
			flooder.fd.RemoveMonitor(remoteIPPort)

			// remove failed peer miner
			flooder.rwLockConnectedMiners.Lock()
			var deleteIndex int
			var deletePeerID string
			for i, peer := range flooder.connectedMiners {
				if peer.fdRespondingIPPort == remoteIPPort {
					deleteIndex = i
					deletePeerID = peer.minerID
				}
			}
			if deleteIndex < len(flooder.connectedMiners) {
				flooder.connectedMiners = append(flooder.connectedMiners[:deleteIndex], flooder.connectedMiners[deleteIndex+1:]...)
				log.Printf("Miner %s FLOODER: removed failing peer miner: %s\n", flooder.minerID, deletePeerID)
			}
			flooder.rwLockConnectedMiners.Unlock()
		}
	}
}

func (flooder *Flooder) floodingInputDaemon() {
	for {
		select {
		case opInput := <-flooder.opsInput:
			newPkt := prepareFloodingPkt(FLOOD_OP, opInput, bclib.Block{}, bclib.Block{})
			log.Printf("Miner %s FLOODER received new local OP: %s to flood. Creating new flooding pkt id: %d\n", flooder.minerID, opInput.OpID, newPkt.ID)
			flooder.floodPacket(newPkt)
		case blockInput := <-flooder.blocksInput:
			newPkt := prepareFloodingPkt(
				FLOOD_BLOCK,
				bclib.ROp{},
				bclib.NewBlock(blockInput.Hash, blockInput.PrevHash, blockInput.MinerID, blockInput.Ops, blockInput.Nonce),
				bclib.Block{})
			log.Printf("Miner %s FLOODER received new local BLOCK: %s to flood. Creating new flooding pkt id: %d\n", flooder.minerID, blockInput.Hash, newPkt.ID)
			flooder.floodPacket(newPkt)
		}
	}
}

func (flooder *Flooder) acceptMinerDaemon(fdResponseIPPort string) {
	for {
		newMinerConn, aerr := flooder.minerConnListener.AcceptTCP()
		if aerr != nil {
			log.Printf("FLOODER: Could not accept peer miner connection due to AcceptTCP error: %s\n", aerr)
			continue
		}
		minerInitMeta, cerr := flooder.confirmPeerConnection(newMinerConn, fdResponseIPPort)
		if cerr != nil {
			log.Printf("FLOODER: Could not accept peer miner connection due to peer confirmation error: %s\n", cerr)
			continue
		}

		newMiner := NewConnectedMiner(minerInitMeta.MinerID, newMinerConn, minerInitMeta.FDResponseIPPort)
		fderr := flooder.fd.AddMonitor(flooder.OutgoingMinersIP+":"+GetNewUnusedPort(), newMiner.fdRespondingIPPort, flooder.lostMsgThresh)
		if fderr != nil {
			log.Printf("Miner %s FLOODER: Could not connect to peer %s due to fd error %s\n", flooder.minerID, newMiner.minerID, fderr)
			continue
		}

		flooder.bcm.RwLock.Lock()
		flooder.bcm.MergeChain(flooder.bcm.BlockRoot.Hash, minerInitMeta.BlockChain)
		flooder.bcm.RwLock.Unlock()

		log.Printf("Miner %s FLOODER: connected to %s", flooder.minerID, newMiner.minerID)
		flooder.rwLockConnectedMiners.Lock()
		flooder.connectedMiners = append(flooder.connectedMiners, &newMiner)
		flooder.rwLockConnectedMiners.Unlock()
	}
}

func (flooder *Flooder) receivePacketDaemon() {
	for {
		flooder.rwLockConnectedMiners.RLock()
		for _, peer := range flooder.connectedMiners {
			peer.lock.Lock()
			rddlerr := peer.conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
			if rddlerr != nil {
				log.Printf("FLOODER: INFO. could not set read deadline for %s passed. Moving to next peer\n", peer.minerID)
				peer.lock.Unlock()
				continue
			}
			jsonData, rerr := readFromConnection(peer.conn)
			if e, ok := rerr.(net.Error); ok && e.Timeout() {
				// ignore timeout errors
				peer.lock.Unlock()
				continue
			}
			if rerr == io.EOF {
				// ignore eof errors
				peer.lock.Unlock()
				continue
			}
			if rerr != nil {
				log.Printf("Miner %s FLOODER: Missed packet from %s due to read error: %s\n", flooder.minerID, peer.minerID, rerr)
				peer.lock.Unlock()
				continue
			}
			floodPkt := &FloodPkt{}
			umerr := unmarshalFromJSONStr(jsonData, floodPkt)
			if umerr != nil {
				log.Printf("Miner %s FLOODER: Missed packet from %s due to unmarshall error: %s\n", flooder.minerID, peer.minerID, umerr)
				peer.lock.Unlock()
				continue
			}
			peer.lock.Unlock()

			// *** Printing ***
			dataIDToPrint := ""
			if floodPkt.FloodType == FLOOD_OP {
				dataIDToPrint = floodPkt.Op.OpID
			}
			if floodPkt.FloodType == FLOOD_BLOCK {
				dataIDToPrint = floodPkt.Block.Hash
			}

			if floodPkt.FloodType == FLOOD_CHAIN {
				dataIDToPrint = "New Chain"
			}

			log.Printf("Miner %s FLOODER: received pkt id: %d, of type %d, with data id: %s\n", flooder.minerID, floodPkt.ID, floodPkt.FloodType, dataIDToPrint)
			// *** Printing ***
			flooder.processReceivedPacket(*floodPkt)
		}
		flooder.rwLockConnectedMiners.RUnlock()
	}
}
