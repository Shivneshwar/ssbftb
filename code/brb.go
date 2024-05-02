package main

import (
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var number, identifier int
var mutex sync.Mutex = sync.Mutex{}
var peerMap = make(map[string]*Peer)
var msg = make([][][]string, 0)
var wasDelivered = make([]bool, 0)
var cur, nxt, txlbl, rxlbl = make([]int64, 0), make([]int64, 0), make([]int64, 0), make([]int64, 0)

func main() {

	lines := readConfigFile(os.Args[1])
	num := len(lines)
	t := num / 3
	if num%3 == 0 {
		t = t - 1
	}
	id, err := strconv.Atoi(os.Args[2])
	if err != nil {
		log.Fatal(err)
		return
	}
	identifier, number = id, num
	myAddr, err := net.ResolveUDPAddr("udp", lines[id])
	if err != nil {
		log.Fatal(err)
		return
	}
	msg = make([][][]string, num)
	wasDelivered = make([]bool, num)
	cur, nxt, txlbl, rxlbl = make([]int64, num), make([]int64, num), make([]int64, num), make([]int64, num)
	peerMap = make(map[string]*Peer, num-1)

	for i, l := range lines {
		if i != id {
			theirAddr, err := net.ResolveUDPAddr("udp", l)
			if err != nil {
				log.Print("Resolve server address failed.")
				log.Fatal(err)
				return
			}
			peerMap[theirAddr.IP.String()+":"+strconv.Itoa(theirAddr.Port)] = makePeer(i, theirAddr, num)
		}
		msg[i] = make([][]string, 3)
		cur[i], nxt[i], txlbl[i], rxlbl[i] = -1, -1, 0, 0
		for j := Init; j <= Ready; j++ {
			msg[i][j] = make([]string, num)
		}
	}

	conn, err := net.ListenUDP("udp", myAddr)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer conn.Close()

	myPeer := makePeer(id, myAddr, num)
	myPeer.setMessage(msgs(id))
	go comm(conn, peerMap, myPeer)
	go testTime(peerMap)

	for {
		mutex.Lock()
		if toRecycle() {
			log.Print("Illegitimate state")
			recycle(identifier)
		}
		for i := range msg {
			if CheckIllegalInitState(i) {
				log.Print("Illegitimate init state")
				recycle(i)
			}
			for _, mg := range msg[i][Init] {
				if mg != "" {
					msg[i][Echo][id] = mg
				}
			}
			mmap := make(map[string]int)
			for _, m := range msg[i][Echo] {
				if m != "" {
					mmap[m]++
				}
			}
			for key, value := range mmap {
				if value > (num+t)/2 {
					// if msg[i][Ready][id] == "" {
					// 	log.Print("Adding to Ready1 ", msg)
					// }
					msg[i][Ready][id] = key
				}
			}
			mmap = make(map[string]int)
			for _, m := range msg[i][Ready] {
				if m != "" {
					mmap[m]++
				}
			}
			for key, value := range mmap {
				if value >= t+1 {
					// if msg[i][Ready][id] == "" {
					// 	log.Print("Adding to Ready2 ", msg)
					// }
					msg[i][Ready][id] = key
				}
			}
		}

		myPeer.setMessage(msgs(id))
		for _, value := range peerMap {
			if value.channelAvailable() {
				mrg(value.getMessage(), value.id)
			}
		}
		mutex.Unlock()
	}

}

func recycle(k int) {
	for _, value := range peerMap {
		value.channelReset()
	}
	for j := Init; j <= Ready; j++ {
		msg[k][j] = make([]string, number)
	}
	wasDelivered[k] = false
}

func msgs(id int) [][][]string {
	m := make2D()
	m[id][Init] = msg[id][Init]
	for j := Echo; j <= Ready; j++ {
		for i := 0; i < len(msg); i++ {
			for idn, mg := range msg[i][j] {
				if idn == id {
					m[i][j][id] = mg
				}
			}
		}
	}
	return m
}

func mrg(mj [][][]string, j int) {
	msg[j][Init] = mj[j][Init]
	for y := Echo; y <= Ready; y++ {
		for x := 0; x < len(msg); x++ {
			msg[x][y][j] = ""
			if mj[x][y][j] != "" {
				msg[x][y][j] = mj[x][y][j]
			}
			// for i, mg := range mj[x][y] {
			// 	if mg != "" {
			// 		msg[x][y][i] = mg
			// 	}
			// }
		}
	}
}

func brbBroadcast(v string) {
	if txAvailable() {
		//log.Print("Recycling when broadcasting a new message ", v)
		recycle(identifier)
		msg[identifier][Init][identifier] = v
	}
}

func brbDeliver(k int) string {
	t := number / 3
	if number%3 == 0 {
		t = t - 1
	}
	mmap := make(map[string]int)
	count := 0
	mg := ""
	for _, m := range msg[k][Ready] {
		if m == "" {
			continue
		}
		mmap[m]++
		if count < mmap[m] {
			count = mmap[m]
			mg = m
		}
	}
	if count >= number-t && rxAvailable(k) {
		wasDelivered[k] = mg != ""
		return mg
	}
	return ""
}

func toRecycle() bool {
	t := number / 3
	if number%3 == 0 {
		t = t - 1
	}
	for _, mg := range msg[identifier][Echo] {
		if mg == "" {
			continue
		}
		found := false
		for _, mgi := range msg[identifier][Init] {
			if mg == mgi {
				found = true
				break
			}
		}
		if !found {
			log.Print("B ", msg)
			return true
		}
	}

	mg := msg[identifier][Ready][identifier]
	if mg == "" {
		return false
	}
	counter1 := 0
	counter2 := 0
	for _, mgi := range msg[identifier][Echo] {
		if mg == mgi {
			counter1++
		}
	}
	for _, mgi := range msg[identifier][Ready] {
		if mg == mgi {
			counter2++
		}
	}
	if !(counter1 > (number+t)/2 || counter2 >= t+1) {
		log.Print("C ", msg)
		return true
	}
	return false
}

func CheckIllegalInitState(k int) bool {
	counter := 0
	for _, mg := range msg[k][Init] {
		if mg != "" {
			counter++
		}
	}
	return counter > 1
}

func test(peerMap map[string]*Peer) {
	var messageNumber int64
	log.Print("Trying to broadcast")
	mutex.Lock()
	brbBroadcast(strconv.FormatInt(messageNumber, 10) + "-" + strconv.Itoa(identifier))
	mutex.Unlock()
	for {
		mutex.Lock()
		if wasDelivered[identifier] {
			log.Print("Trying to broadcast")
			brbBroadcast(strconv.FormatInt(messageNumber, 10) + "-" + strconv.Itoa(identifier))
		} else {
			ret := brbDeliver(identifier)
			if ret != "" {
				log.Print("Delivered message", " ", ret)
				messageNumber++
			}
		}
		for _, value := range peerMap {
			ret := brbDeliver(value.id)
			if ret != "" {
				log.Print("Delivered message", " ", ret)
				if value.id == identifier {
					messageNumber++
				}
			}
		}
		log.Print(msg)
		//log.Print(cur, nxt, txlbl, rxlbl)
		mutex.Unlock()
		time.Sleep(2 * time.Second)
	}
}

func test0(peerMap map[string]*Peer) {
	var messageNumber int64
	if identifier == 0 {
		log.Print("Trying to broadcast")
		mutex.Lock()
		brbBroadcast(strconv.FormatInt(messageNumber, 10) + "-" + strconv.Itoa(identifier))
		mutex.Unlock()
	}
	for {
		mutex.Lock()
		if wasDelivered[identifier] && identifier == 0 {
			log.Print("Trying to broadcast")
			brbBroadcast(strconv.FormatInt(messageNumber, 10) + "-" + strconv.Itoa(identifier))
		} else {
			ret := brbDeliver(0)
			if ret != "" {
				log.Print("Delivered message", " ", ret)
				messageNumber++
			}
		}
		log.Print(msg)
		//log.Print(cur, nxt, txlbl, rxlbl)
		mutex.Unlock()
		time.Sleep(2 * time.Second)
	}
}

func testTime0(peerMap map[string]*Peer) {
	last := 10
	if identifier < last {
		mutex.Lock()
		brbBroadcast(time.Now().Format(time.RFC3339Nano))
		mutex.Unlock()
	}
	for {
		mutex.Lock()
		if wasDelivered[identifier] && identifier < last {
			brbBroadcast(time.Now().Format(time.RFC3339Nano))
		} else {
			for i := 0; i < last; i++ {
				ret := brbDeliver(i)
				if ret != "" {
					sentTime, err := time.Parse(time.RFC3339Nano, ret)
					if err != nil {
						log.Print("Delivered msg", " ", ret)
					} else {
						log.Print("Delivered message", " ", time.Since(sentTime).Seconds())
					}
				}
			}
		}
		mutex.Unlock()
	}
}

func testTime(peerMap map[string]*Peer) {
	last := 9
	ll := 2
	chs := time.Now().Format(time.RFC3339Nano)
	for i := 1; i < ll; i++ {
		chs = chs + "h"
	}
	if identifier < last {
		mutex.Lock()
		brbBroadcast(chs)
		mutex.Unlock()
	}
	for {
		mutex.Lock()
		if wasDelivered[identifier] && identifier < last {
			brbBroadcast(chs)
		} else {
			for i := 0; i < last; i++ {
				ret := brbDeliver(i)
				index := strings.IndexRune(ret, 'h')
				if index != -1 {
					sentTime, err := time.Parse(time.RFC3339Nano, ret[:index])
					if err != nil {
						log.Print("Delivered msg", " ", ret, err)
					} else {
						log.Print("Delivered message", " ", time.Since(sentTime).Seconds())
					}
				} else {
					//fmt.Println("Character not found in string ", ret)
				}
			}
		}
		mutex.Unlock()
	}
}
