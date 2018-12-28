package flooding

import (
	"log"
	"net"
	"time"
)

func (flooder *Flooder) confirmPeerConnection(conn *net.TCPConn, fdRespondseIPPort string) (MinerInitMeta, error) {
	// first send your own init data
	flooder.bcm.RwLock.RLock()
	root := flooder.bcm.BlockRoot
	json, merr := marshalToJSONStr(NewMinerInitMeta(flooder.minerID, fdRespondseIPPort, root))
	flooder.bcm.RwLock.RUnlock()
	if merr != nil {
		return MinerInitMeta{}, merr
	}
	_, werr := writeToConnection(conn, json)
	if werr != nil {
		return MinerInitMeta{}, werr
	}

	// then wait for peers init data
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	jsonData, rerr := readFromConnection(conn)
	if rerr != nil {
		return MinerInitMeta{}, rerr
	}
	minerInitData := &MinerInitMeta{}
	umerr := unmarshalFromJSONStr(jsonData, minerInitData)
	if umerr != nil {
		return MinerInitMeta{}, rerr
	} else {
		return *minerInitData, nil
	}
}

func (flooder *Flooder) connectToPeerMiner(PeerMinersAddrs []string, OutgoingMinersIP string, fdResponseIPPort string) {
	flooder.rwLockConnectedMiners.Lock()
	defer flooder.rwLockConnectedMiners.Unlock()
	for _, peerMinerAddr := range PeerMinersAddrs {
		rtcpAddr, rtcpAddrErr := net.ResolveTCPAddr("tcp", peerMinerAddr)
		if rtcpAddrErr != nil {
			log.Printf("FLOODER: Could not connect to peer %s due to raddr error %s\n", peerMinerAddr, rtcpAddrErr)
			continue
		}
		ltcpAddr, ltcpAddrErr := net.ResolveTCPAddr("tcp", OutgoingMinersIP+":"+GetNewUnusedPort())
		if ltcpAddrErr != nil {
			log.Printf("FLOODER: Could not connect to peer %s due to laddr error %s\n", peerMinerAddr, ltcpAddrErr)
			continue
		}

		tcpConn, derr := net.DialTCP("tcp", ltcpAddr, rtcpAddr)
		if derr != nil {
			log.Printf("FLOODER: Could not connect to peer %s due to dial error %s\n", peerMinerAddr, derr)
			continue
		}

		minerInitMeta, cerr := flooder.confirmPeerConnection(tcpConn, fdResponseIPPort)
		if cerr != nil {
			log.Printf("FLOODER: Could not connect to peer %s due to confirmation failure %s\n", peerMinerAddr, cerr)
			continue
		}
		newMiner := NewConnectedMiner(minerInitMeta.MinerID, tcpConn, minerInitMeta.FDResponseIPPort)
		fderr := flooder.fd.AddMonitor(flooder.OutgoingMinersIP+":"+GetNewUnusedPort(), newMiner.fdRespondingIPPort, flooder.lostMsgThresh)
		if fderr != nil {
			log.Printf("Miner %s FLOODER: Could not connect to peer %s due to fd error %s\n", flooder.minerID, peerMinerAddr, fderr)
			continue
		}

		flooder.bcm.RwLock.Lock()
		flooder.bcm.MergeChain(flooder.bcm.BlockRoot.Hash, minerInitMeta.BlockChain)
		flooder.bcm.RwLock.Unlock()
		flooder.connectedMiners = append(flooder.connectedMiners, &newMiner)
		log.Printf("Miner %s FLOODER: connected to %s", flooder.minerID, newMiner.minerID)
	}
}

func (flooder *Flooder) floodPacket(pkt FloodPkt) {
	flooder.rwLockConnectedMiners.RLock()
	defer flooder.rwLockConnectedMiners.RUnlock()

	// get list of peers to flood
	targetPeers := make([]*ConnectedMiner, 0)
	for _, peer := range flooder.connectedMiners {
		if !contains(pkt.Recved, peer.minerID) {
			targetPeers = append(targetPeers, peer)
		}
	}
	log.Printf("Miner %s FLOODER: flooding to %d peers\n", flooder.minerID, len(targetPeers))
	// update list of receivers
	pkt.Recved = append(pkt.Recved, flooder.minerID)
	for _, peer := range targetPeers {
		pkt.Recved = append(pkt.Recved, peer.minerID)
	}

	// flood pkt
	for _, peer := range targetPeers {
		ferr := floodPacketHelper(pkt, peer)
		if ferr != nil {
			log.Printf("Miner %s FLOODER: WARNING. Failed to flood peer miner %s due to err: %s\n", flooder.minerID, peer.minerID, ferr)
		} else {
			log.Printf("Miner %s FLOODER: successfully flooded to peer miner %s\n", flooder.minerID, peer.minerID)
		}
	}
	log.Printf("Miner %s FLOODER: Finished flooding\n", flooder.minerID)
}

func floodPacketHelper(pkt FloodPkt, peer *ConnectedMiner) error {
	peer.lock.Lock()
	defer peer.lock.Unlock()

	str, merr := marshalToJSONStr(pkt)
	if merr != nil {
		return merr
	}
	_, werr := writeToConnection(peer.conn, str)
	if werr != nil {
		return werr
	}
	return nil
}

func (flooder *Flooder) processReceivedPacket(pkt FloodPkt) {
	flooder.floodPacket(pkt)
	if pkt.FloodType == FLOOD_OP {
		flooder.opsQueueInputChann <- pkt.Op
		log.Printf("Miner %s FLOODER: Forwared op with id: %s to ops queue\n", flooder.minerID, pkt.Op.OpID)
	} else if pkt.FloodType == FLOOD_BLOCK {
		flooder.bcmInputChann <- pkt.Block
		log.Printf("Miner %s FLOODER: Forwared block with hash: %s to bcm. BCM input chan now has length %d\n", flooder.minerID, pkt.Block.Hash, len(flooder.bcmInputChann))
	} else if pkt.FloodType == FLOOD_CHAIN {
		flooder.bcm.RwLock.Lock()
		localRootHash := flooder.bcm.BlockRoot.Hash
		flooder.bcm.MergeChain(localRootHash, pkt.Chain)
		flooder.bcm.RwLock.Unlock()
		log.Printf("Miner %s FLOODER: merged peer chain into local chain\n", flooder.minerID)
	}
}

func contains(slice []string, key string) bool {
	for _, s := range slice {
		if s == key {
			return true
		}
	}
	return false
}
