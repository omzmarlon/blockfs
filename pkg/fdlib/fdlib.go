/*

This package specifies the API to the failure detector library to be
used in assignment 1 of UBC CS 416 2018W1.

You are *not* allowed to change the API below. For example, you can
modify this file by adding an implementation to Initialize, but you
cannot change its API.

*/

package fdlib

import (
	"bytes"
	"encoding/gob"
	"errors"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"
)

const prod = true

//////////////////////////////////////////////////////
// Define the message types fdlib has to use to communicate to other
// fdlib instances. We use Go's type declarations for this:
// https://golang.org/ref/spec#Type_declarations

// Heartbeat message.
type HBeatMessage struct {
	EpochNonce uint64 // Identifies this fdlib instance/epoch.
	SeqNum     uint64 // Unique for each heartbeat in an epoch.
}

// An ack message; response to a heartbeat.
type AckMessage struct {
	HBEatEpochNonce uint64 // Copy of what was received in the heartbeat.
	HBEatSeqNum     uint64 // Copy of what was received in the heartbeat.
}

// Notification of a failure, signal back to the client using this
// library.
type FailureDetected struct {
	UDPIpPort string    // The RemoteIP:RemotePort of the failed node.
	Timestamp time.Time // The time when the failure was detected.
}

//////////////////////////////////////////////////////

// An FD interface represents an instance of the fd
// library. Interfaces are everywhere in Go:
// https://gobyexample.com/interfaces
type FD interface {
	// Tells the library to start responding to heartbeat messages on
	// a local UDP IP:port. Can return an error that is related to the
	// underlying UDP connection.
	StartResponding(LocalIpPort string) (err error)

	// Tells the library to stop responding to heartbeat
	// messages. Always succeeds.
	StopResponding()

	// Tells the library to start monitoring a particular UDP IP:port
	// with a specific lost messages threshold. Can return an error
	// that is related to the underlying UDP connection.
	AddMonitor(LocalIpPort string, RemoteIpPort string, LostMsgThresh uint8) (err error)

	// Tells the library to stop monitoring a particular remote UDP
	// IP:port. Always succeeds.
	RemoveMonitor(RemoteIpPort string)

	// Tells the library to stop monitoring all nodes.
	StopMonitoring()
}

// The constructor for a new FD object instance. Note that notifyCh
// can only be received on by the client that receives it from
// initialize:
// https://www.golang-book.com/books/intro/10
func Initialize(EpochNonce uint64, ChCapacity uint8) (fd FD, notifyCh <-chan FailureDetected, err error) {

	fNotify := make(chan FailureDetected, ChCapacity)
	failureDetector := &FailureDetector{
		EpochNonce:     EpochNonce,
		fNotify:        fNotify,
		resConn:        nil,
		resControlChan: nil,
		stopConfirm:    make(chan int),
		timeoutCache:   make(map[string]time.Duration),
		monitors:       make(map[string]*NodeData),
		monitorsRwLock: &sync.RWMutex{},
	}
	if !prod {
		log.Printf("Initialized failure detector: %+v\n", failureDetector)
	}
	return failureDetector, fNotify, nil
}

// FailureDetector - Type for failure detection lib
type FailureDetector struct {
	// fd
	EpochNonce uint64
	fNotify    chan FailureDetected
	// ack responses
	resConn        *net.UDPConn
	resControlChan chan int
	stopConfirm    chan int
	// monitoring:
	monitors       map[string]*NodeData
	monitorsRwLock *sync.RWMutex
	timeoutCache   map[string]time.Duration
}

// NodeData - contains data about a monitored node
type NodeData struct {
	conn          *net.UDPConn
	LostMsgThresh uint8
	Timeout       time.Duration
	LostCount     uint8
	ctrl          chan int
	onWait        map[uint64]*HBMeta
}

// HBMeta holds information about a sent heartbeat
type HBMeta struct {
	Timestamp time.Time
}

// StartResponding - Implementation for start responding to heart beat messages
func (fd *FailureDetector) StartResponding(LocalIpPort string) (err error) {
	// check existing responder
	if fd.resConn != nil {
		return errors.New("already responding to heatbeat")
	}

	// set up responder connection
	lAddr, err := net.ResolveUDPAddr("udp", LocalIpPort)
	if err != nil {
		return errors.New("Could not resolve local udp address")
	}
	conn, err := net.ListenUDP("udp", lAddr)
	if err != nil {
		return errors.New("Could not listen on udp connection")
	}
	fd.resConn = conn

	// set up responder control channel
	fd.resControlChan = make(chan int)

	// launch heartbeat responder
	go fd.heartBeatResponder()
	return nil
}

// StopResponding - Implementation for stop responding to heart beat messages
func (fd *FailureDetector) StopResponding() {
	if fd.resControlChan != nil {
		fd.resControlChan <- 0
		<-fd.stopConfirm
	}
}

// AddMonitor - Implementation for adding a monitor
func (fd *FailureDetector) AddMonitor(LocalIpPort string, RemoteIpPort string, LostMsgThresh uint8) (err error) {
	fd.monitorsRwLock.Lock()
	defer fd.monitorsRwLock.Unlock()

	// check existing monitor
	if monitor, ok := fd.monitors[RemoteIpPort]; ok {
		monitor.LostMsgThresh = LostMsgThresh
		return nil
	}

	// create udp connection
	lAddr, lErr := net.ResolveUDPAddr("udp", LocalIpPort)
	if lErr != nil {
		return errors.New("Could not resolve local udp address")
	}
	conn, connErr := net.ListenUDP("udp", lAddr)
	if connErr != nil {
		return errors.New("Could not setup udp conn")
	}
	// initialize monitor
	timeout := 3 * time.Second
	if toc, ok := fd.timeoutCache[RemoteIpPort]; ok {
		timeout = toc
	}
	monitor := &NodeData{
		conn:          conn,
		LostMsgThresh: LostMsgThresh,
		Timeout:       timeout,
		LostCount:     0,
		ctrl:          make(chan int),
		onWait:        make(map[uint64]*HBMeta),
	}
	fd.monitors[RemoteIpPort] = monitor

	// launch monitor
	go fd.heartbeatTransmitter(RemoteIpPort)
	return nil
}

// RemoveMonitor - Implementation for removing a monitor
func (fd *FailureDetector) RemoveMonitor(RemoteIpPort string) {
	fd.monitorsRwLock.RLock()
	defer fd.monitorsRwLock.RUnlock()
	if monitor, ok := fd.monitors[RemoteIpPort]; ok {
		monitor.ctrl <- 0
		<-fd.stopConfirm
	}
}

// StopMonitoring - Implementaton for stopping all monitoring
func (fd *FailureDetector) StopMonitoring() {
	fd.monitorsRwLock.RLock()
	defer fd.monitorsRwLock.RUnlock()
	for _, m := range fd.monitors {
		m.ctrl <- 0
		<-fd.stopConfirm
	}
}

// ************ monitoring logic ************

func (fd *FailureDetector) heartbeatTransmitter(RemoteIpPort string) {
	if fd.monitors[RemoteIpPort].LostMsgThresh == 0 {
		if !prod {
			log.Printf("Detected Failed node: %s, %+v\n", RemoteIpPort, fd.monitors[RemoteIpPort])
		}
		fd.notifyFailureHelper(RemoteIpPort)
		fd.cleanMonitorHelper(RemoteIpPort)
		return // stop monitoring this node
	}

	fd.sendHeartbeatHelper(RemoteIpPort) // initial send
	for {
		// get this monitor
		fd.monitorsRwLock.RLock()
		monitor := fd.monitors[RemoteIpPort]
		fd.monitorsRwLock.RUnlock()
		select {
		case c := <-monitor.ctrl:
			if c == 0 {
				log.Printf("Removing monitor %s, %+v\n", RemoteIpPort, monitor)
				fd.cleanMonitorHelper(RemoteIpPort)
				fd.stopConfirm <- 1 // confirm shutdown monitor
				return
			}
		default:
		}

		// prepare data
		conn := monitor.conn

		// waiting for ack
		ddlerr := conn.SetReadDeadline(time.Now().Add(monitor.Timeout))
		if ddlerr != nil {
			log.Printf("Unexpected SetReadDeadline error: %s\n", ddlerr)
			continue
		}
		recvBuff := make([]byte, 1024)
		n, _, err := conn.ReadFromUDP(recvBuff)
		if err != nil {
			monitor.LostCount++
			if monitor.LostCount >= monitor.LostMsgThresh {
				if !prod {
					log.Printf("Detected Failed node: %s, %+v\n", RemoteIpPort, monitor)
				}
				fd.notifyFailureHelper(RemoteIpPort)
				fd.cleanMonitorHelper(RemoteIpPort)
				return // stop monitoring this node
			}

		} else {
			// reading ack
			ack := &AckMessage{}
			unMErr := unMarshall(recvBuff, n, ack)
			if unMErr != nil {
				log.Printf("Unexpected unmarshall error: %s\n", unMErr)
				continue
			}
			if fd.EpochNonce != ack.HBEatEpochNonce {
				log.Printf("Received unmatching EpochNonce, this: %d, received: %d\n", fd.EpochNonce, ack.HBEatEpochNonce)
				continue
			}
			// retrieve heartbeat data
			hbMeta := monitor.onWait[ack.HBEatSeqNum]
			// compute RTT and update timeout
			duration := time.Now().Sub(hbMeta.Timestamp)
			oldRTT := monitor.Timeout
			nextRTT := (duration + oldRTT) / 2
			if !prod {
				log.Printf("Node: %s has new calculated new RTT: %s\n", RemoteIpPort, nextRTT)
			}
			monitor.Timeout = nextRTT
			// update lost count
			monitor.LostCount = 0
			// clear heartbeat data
			delete(monitor.onWait, ack.HBEatSeqNum)
		}
		fd.sendHeartbeatHelper(RemoteIpPort)
	}
}

func (fd *FailureDetector) cleanMonitorHelper(RemoteIpPort string) {
	fd.monitorsRwLock.Lock()
	defer fd.monitorsRwLock.Unlock()
	monitor := fd.monitors[RemoteIpPort]
	fd.timeoutCache[RemoteIpPort] = monitor.Timeout // cache calculated timeout for this node
	monitor.conn.Close()
	delete(fd.monitors, RemoteIpPort)
}

func (fd *FailureDetector) notifyFailureHelper(RemoteIpPort string) {
	detection := FailureDetected{
		UDPIpPort: RemoteIpPort,
		Timestamp: time.Now(),
	}
	fd.fNotify <- detection
}

func (fd *FailureDetector) sendHeartbeatHelper(RemoteIpPort string) {
	fd.monitorsRwLock.RLock()
	monitor := fd.monitors[RemoteIpPort]
	fd.monitorsRwLock.RUnlock()

	conn := monitor.conn
	source := rand.NewSource(time.Now().UnixNano())
	rand := rand.New(source)
	hbSeq := rand.Uint64()
	heartbeat := HBeatMessage{fd.EpochNonce, hbSeq}
	hbBytes, merr := marshall(heartbeat)
	if merr != nil {
		log.Printf("Unexpected marshall error: %s\n", merr)
		return
	}

	// sending data
	monitor.onWait[hbSeq] = &HBMeta{
		Timestamp: time.Now(),
	}
	rAddr, rErr := net.ResolveUDPAddr("udp", RemoteIpPort)
	if rErr != nil {
		log.Printf("Unexpected udp addr resolving error: %s\n", rErr)
		return
	}
	_, err := conn.WriteToUDP(hbBytes, rAddr)
	if err != nil {
		delete(monitor.onWait, hbSeq)
		log.Printf("Unexpected send heartbeat error: %s\n", err)
		return
	}
}

// ************************

// ************ Heartbeat response logic ************
func (fd *FailureDetector) heartBeatResponder() {
	for {
		// checking shutdown signal
		select {
		case msg := <-fd.resControlChan:
			if msg == 0 {
				fd.resConn.Close()
				fd.resConn = nil
				close(fd.resControlChan)
				fd.resControlChan = nil
				fd.stopConfirm <- 1 // confirm stop responding
				return
			}
		default:
		}

		// set up connection read
		timeoutDuration := 1 * time.Second
		fd.resConn.SetReadDeadline(time.Now().Add(timeoutDuration))
		recvBuf := make([]byte, 1024)

		// read data from connection
		n, raddr, err := fd.resConn.ReadFromUDP(recvBuf)
		if err != nil {
			continue
		}

		// decoding data read
		heartBeat := &HBeatMessage{}
		unMarshallErr := unMarshall(recvBuf, n, heartBeat)
		if unMarshallErr != nil {
			log.Printf("Unexpected error in unmarshalling heartbeat messages: %+v\n", heartBeat)
			continue
		}

		// responding ack message
		ack := AckMessage{HBEatEpochNonce: heartBeat.EpochNonce, HBEatSeqNum: heartBeat.SeqNum}
		bytes, marshalErr := marshall(ack)
		if marshalErr != nil {
			log.Printf("Unexpected error in marshalling ack message: %s\n", marshalErr)
			continue
		}
		// responding to heartbeat message
		_, writeUDPErr := fd.resConn.WriteToUDP(bytes, raddr)
		if writeUDPErr != nil {
			log.Printf("Error in sending ack message for error: %s\n", writeUDPErr)
			continue
		}
	}
}

// ************************

// ************ marshalling helpers ************
func marshall(data interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(data)
	return buf.Bytes(), err
}

func unMarshall(data []byte, size int, e interface{}) error {
	var buffer bytes.Buffer
	dec := gob.NewDecoder(&buffer)
	buffer.Write(data[0:size])
	err := dec.Decode(e)
	return err
}

// ************************
