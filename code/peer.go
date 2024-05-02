package main

import (
	"encoding/json"
	"log"
	"math"
	"net"
	"sync"
	"time"
)

// MinEstRTT = Minimum value allowed as timeout in the stop and wait protocol
// MaxEstRTT = Maximum value allowed as timeout in the stop and wait protocol
// DefaultRTTTimeout = Default value of timeout when unable to estimate RTT time
const (
	MinEstRTT         = time.Duration(200 * time.Millisecond)
	MaxEstRTT         = time.Duration(300 * time.Millisecond)
	DefaultRTTTimeout = time.Duration(300 * time.Millisecond)
)

// Socket is an IP:Port combination
type Socket struct {
	IP   net.IP
	Port int
}

// This is the payload being exchanged with other processes
// One counter for when acting as sender and another counter for when acting as receiver
type Payload struct {
	MyCounter       uint64
	TheirCounter    uint64
	PayloadContents [][][]string
	Cur             int64
	Nxt             int64
	TxLabel         int64
	RxLabel         int64
}

// Peer struct is used to encapsulate all the processes and it's relevant information
type Peer struct {
	id                  int
	addr                *net.UDPAddr
	mutex               sync.Mutex
	firstTokenArrived   bool
	sentMessages        map[uint64]time.Time
	timeoutTime         time.Time
	rttEstimator        MovingAverage
	myCounter           uint64
	theirCounter        uint64
	retransmissionCount uint8
	resetCounter        int64
	consistencyCheck    int64
	message             [][][]string
}

func makePeer(id int, peerAddr *net.UDPAddr, num int) *Peer {
	return &Peer{id, peerAddr, sync.Mutex{}, false, make(map[uint64]time.Time),
		time.Now().Add(DefaultRTTTimeout), NewMovingAverage(), 0, 0, 0,
		2*CHANNELCAPACITY + 1, 0, make2D()}
}

func (p *Peer) addToRTTEstimator(t time.Time, counter uint64) {
	if p.retransmissionCount != 0 {
		return
	}
	sentTime, found := p.sentMessages[counter]
	if found {
		p.rttEstimator.Add(float64(t.Sub(sentTime)))
		delete(p.sentMessages, counter)
	}

}

func (p *Peer) setTimeout(t time.Time) {
	if p.rttEstimator.Value() == 0 {
		p.timeoutTime = t.Add(DefaultRTTTimeout)
	} else {
		timeout := minDuration(maxDuration(time.Duration(
			p.rttEstimator.Value()*math.Pow(2, float64(p.retransmissionCount))), MaxEstRTT), MinEstRTT)
		p.timeoutTime = t.Add(timeout)
	}
}

func (p *Peer) updateMyCounter(newCounter uint64) {
	p.myCounter = newCounter
	p.retransmissionCount = 0
}

func (p *Peer) updateRetransmissionCounter() {
	if p.retransmissionCount < 10 {
		p.retransmissionCount++
	} else {
		p.rttEstimator.Reset()
	}
	delete(p.sentMessages, p.myCounter)
}

func (p *Peer) checkTimeout() bool {
	return time.Now().After(p.timeoutTime)
}

func (p *Peer) getMessage() [][][]string {
	var m [][][]string
	p.mutex.Lock()
	m = p.message
	p.mutex.Unlock()
	return m
}

func (p *Peer) setMessage(m [][][]string) {
	p.mutex.Lock()
	p.message = m
	p.mutex.Unlock()
}

func (p *Peer) channelAvailable() bool {
	return p.getResetCounter() == 0
}

func (p *Peer) channelReset() {
	p.mutex.Lock()
	if p.consistencyCheck == 0 {
		p.resetCounter = 2*CHANNELCAPACITY + 1
	} else {
		p.resetCounter = 2
	}
	p.consistencyCheck = (p.consistencyCheck + 1) % CONSISTENCYCHECKROUND
	p.mutex.Unlock()
}

func (p *Peer) getResetCounter() int64 {
	var m int64
	p.mutex.Lock()
	m = p.resetCounter
	p.mutex.Unlock()
	return m
}

func (p *Peer) setResetCounter(n int64) {
	p.mutex.Lock()
	p.resetCounter = n
	p.mutex.Unlock()
}

func (p *Peer) isFirstTokenArrived() bool {
	return p.firstTokenArrived
}

func (p *Peer) updateTokenArrival() {
	p.firstTokenArrived = !p.firstTokenArrived
}

func getMinTimeout(peerMap map[string]*Peer) time.Time {
	var min time.Time
	for _, v := range peerMap {
		min = v.timeoutTime
		break
	}
	for _, v := range peerMap {
		if v.timeoutTime.Before(min) {
			min = v.timeoutTime
		}
	}
	return min
}

func jsonSerialize(payload Payload) []byte {
	bytes, err := json.Marshal(&payload)
	if err != nil {
		log.Print("Marshal Register information failed.")
		log.Fatal(err)
	}
	return bytes
}
