package main

import (
	"math"
)

const (
	B                     int64 = math.MaxInt64 - 1
	CHANNELCAPACITY       int64 = 4
	CONSISTENCYCHECKROUND int64 = 1
)

var LAMBDA int64 = int64(math.Pow(2, 60))

func rxMsg(j int, payload Payload) {
	//log.Print("First ", payload.Cur, cur[j])
	if behind(2, cur[identifier], payload.Nxt) && txlbl[j] == payload.RxLabel {
		//log.Print("Second ", cur[identifier], payload.Nxt)
		txlbl[j] = min(B, txlbl[j]+1)
	}
	if !behind(1, payload.Cur, cur[j]) {
		//log.Print("Receiver side legitimate recycle of ", j)
		cur[j] = payload.Cur
		recycle(j)
		//log.Print("After recycle ", msg)
	}
	rxlbl[j] = payload.TxLabel
}

func increment() int64 {
	if cur[identifier] != -1 {
		for i, c := range txlbl {
			if i != identifier && c <= 2*(CHANNELCAPACITY+1) {
				return -1
			}
		}
	}
	cur[identifier] = (cur[identifier] + 1) % (B + 1)
	txlbl = make([]int64, number)
	return cur[identifier]
}

func fetch(k int) int64 {
	if behind(1, cur[k], nxt[k]) {
		return -1
	} else {
		//log.Print("Inside fetch-else ", nxt[k], cur[k])
		nxt[k] = cur[k]
		return nxt[k]
	}
}

func txAvailable() bool {
	return increment() != -1
}

func rxAvailable(k int) bool {
	return fetch(k) != -1
}

func behind(d int64, s int64, c int64) bool {
	if d*LAMBDA <= c {
		if s <= c && s >= c-d*LAMBDA {
			return true
		}
	} else {
		if s <= c {
			return true
		}
		if s > B-(d*LAMBDA-c-1) {
			return true
		}
	}
	return false
}
