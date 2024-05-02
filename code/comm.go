package main

/*
Sender:
01 upon timeout
02 	send (counter )
03 upon message arrival
04 begin
05 	receive (MsgCounter)
06		if MsgCounter ≥ counter then
07		begin
08			counter := MsgCounter + 1
09			send (counter)
10		end
11		else send (counter)
12 end

Receiver:
13 upon message arrival
14 begin
15 	receive (MsgCounter)
16		if MsgCounter ≠ counter then // token arrived
17			counter := MsgCounter
18		send (counter) // send token
19 end

Every time Pi receives a token from Pj. Pi write the current value of Rij in the value of the token

Write operation of Pi into rij is implemented by locally writing into Rij

Read operation of Pi from rji is implemented by:
1. Pi receives the token from Pj
2. Pi receives the token from Pj. Return the value attached to this token

*/

import (
	"encoding/json"
	"log"
	"net"
	"strconv"
	"time"
)

func comm(conn *net.UDPConn, peerMap map[string]*Peer, myPeer *Peer) {
	for {
		timeoutChecker(conn, peerMap, myPeer)
		buf := make([]byte, 32768)
		// the below line is achieve timeout, the earliest timeout time for all available peers is chosen
		conn.SetReadDeadline(getMinTimeout(peerMap))
		n, addr, err := conn.ReadFromUDP(buf[0:])
		if err != nil {
			// When timeout occurs, continue the do forever loop
			if err.(*net.OpError).Timeout() {
				continue
			}
			log.Fatal(err)
		}
		receivedTime := time.Now()
		// Check if message is from known peer
		peer, contains := peerMap[addr.IP.String()+":"+strconv.Itoa(addr.Port)]
		if contains {
			var payload Payload
			err = json.Unmarshal(buf[:n], &payload)
			if err != nil {
				log.Fatal(err)
			}
			peer.addToRTTEstimator(receivedTime, payload.TheirCounter)
			// Sending message based on stop and wait protocol. Line 06-10 in sender alg.
			if payload.TheirCounter >= peer.myCounter {
				peer.setResetCounter(max(0, peer.getResetCounter()-1))
				peer.updateMyCounter(payload.TheirCounter + 1)
			}
			// Receiving message based on stop and wait protocol. Line 16-18 in receiver alg.
			if payload.MyCounter != peer.theirCounter {
				peer.theirCounter = payload.MyCounter
				mutex.Lock()
				rxMsg(peer.id, payload)
				mutex.Unlock()
				if peer.isFirstTokenArrived() {
					peer.setMessage(payload.PayloadContents)
				}
				peer.updateTokenArrival()
			}
			mutex.Lock()
			// Below part is the response we sent to the message just received. simualting send(counter).
			//log.Print("txMsg ", cur[identifier], nxt[peer.id], txlbl[peer.id], rxlbl[peer.id])
			payload = Payload{peer.myCounter, peer.theirCounter, myPeer.getMessage(),
				cur[identifier], nxt[peer.id], txlbl[peer.id], rxlbl[peer.id]}
			mutex.Unlock()
			payloadBytes := jsonSerialize(payload)
			sendMessage(conn, peer.addr, payloadBytes)
			timeNow := time.Now()
			peer.sentMessages[peer.myCounter] = timeNow
			peer.setTimeout(timeNow)
		}
	}
}

// Check if timeout has been reached for each peer and retransmits the message if it has. Line 1-2 in Sender.
func timeoutChecker(conn *net.UDPConn, peerMap map[string]*Peer, myPeer *Peer) {
	for _, peer := range peerMap {
		if peer.checkTimeout() {
			peer.updateRetransmissionCounter()
			mutex.Lock()
			payload := Payload{peer.myCounter, peer.theirCounter, myPeer.getMessage(),
				cur[identifier], nxt[peer.id], txlbl[peer.id], rxlbl[peer.id]}
			mutex.Unlock()
			payloadBytes := jsonSerialize(payload)
			//log.Print("Size of payload: ", len(payloadBytes))
			sendMessage(conn, peer.addr, payloadBytes)
			peer.setTimeout(time.Now())
		}
		//log.Print(peer.message, peer.isFresh())
	}
}
