package main

import (
	"bufio"
	"log"
	"net"
	"os"
	"time"
)

type BrbMessage int

const (
	Init BrbMessage = iota
	Echo
	Ready
)

func readConfigFile(path string) []string {
	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	var lines []string
	fileScanner := bufio.NewScanner(file)
	fileScanner.Split(bufio.ScanLines)
	for fileScanner.Scan() {
		lines = append(lines, fileScanner.Text())
	}
	file.Close()
	return lines
}

// sends a UDP message to address "addr" with content "message"
func sendMessage(connection *net.UDPConn, addr *net.UDPAddr, message []byte) {
	_, err := connection.WriteToUDP(message, addr)
	if err != nil {
		log.Fatal(err)
	}
}

func make2D() [][][]string {
	m := make([][][]string, number)
	for x := 0; x < number; x++ {
		m[x] = make([][]string, 3)
		for y := Init; y <= Ready; y++ {
			m[x][y] = make([]string, number)
		}
	}
	return m
}

// function that returns the minimum between the two time.Duration
func minDuration(x, y time.Duration) time.Duration {
	if x < y {
		return x
	}
	return y
}

// function that returns the maximum between the two time.Duration
func maxDuration(x, y time.Duration) time.Duration {
	if x > y {
		return x
	}
	return y
}

// function that returns the minimum between the two int64
func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

// function that returns the maximum between the two int64
func max(a, b int64) int64 {
	if a < b {
		return b
	}
	return a
}
