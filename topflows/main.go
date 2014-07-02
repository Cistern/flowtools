package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/PreetamJinka/ethernetdecode"
	"github.com/PreetamJinka/sflow-go"
	"github.com/PreetamJinka/udpchan"
)

var (
	sourceTalkersBytes      = map[string]uint64{}
	destinationTalkersBytes = map[string]uint64{}

	sourceTalkersPackets      = map[string]uint64{}
	destinationTalkersPackets = map[string]uint64{}

	lock = sync.Mutex{}

	window = time.Second * 30
)

func main() {
	listenAddr := flag.String("listen", ":6343", "Listening address for sFlow datagrams")
	flag.DurationVar(&window, "window", window, "How long to aggregate counters before resetting")

	flag.Parse()

	datagrams, err := udpchan.Listen(*listenAddr, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Listening on", *listenAddr)

	go func() {
		for _ = range time.Tick(window) {
			printTopTalkers()
			resetTopTalkers()
		}
	}()

	for datagram := range datagrams {
		decodedDgram := sflow.Decode(datagram)
		for _, sample := range decodedDgram.Samples {
			if sample.SampleType() == sflow.TypeFlowSample {
				analyzeFlowSample(sample.(sflow.FlowSample))
			}
		}
	}
}

func printTopTalkers() {
	lock.Lock()
	defer lock.Unlock()

	var (
		sortedKeys []string
		max        int
	)

	fmt.Println()
	fmt.Println("==================")

	fmt.Println()
	fmt.Println("by source bytes")
	fmt.Println("---------------")
	sortedKeys = sortMap(sourceTalkersBytes)
	if len(sortedKeys) < 10 {
		max = len(sortedKeys)
	} else {
		max = 10
	}
	for _, key := range sortedKeys[:max] {
		fmt.Printf("%-20s -- %.2f Bps\n", key, float32(sourceTalkersBytes[key])/float32(window.Seconds()))
	}

	fmt.Println()
	fmt.Println("by dest bytes")
	fmt.Println("---------------")
	sortedKeys = sortMap(destinationTalkersBytes)
	if len(sortedKeys) < 10 {
		max = len(sortedKeys)
	} else {
		max = 10
	}
	for _, key := range sortedKeys[:max] {
		fmt.Printf("%-20s -- %.2f Bps\n", key, float32(destinationTalkersBytes[key])/float32(window.Seconds()))
	}

	fmt.Println()
	fmt.Println("by source packets")
	fmt.Println("---------------")
	sortedKeys = sortMap(sourceTalkersPackets)
	if len(sortedKeys) < 10 {
		max = len(sortedKeys)
	} else {
		max = 10
	}
	for _, key := range sortedKeys[:max] {
		fmt.Printf("%-20s -- %.2f Pps\n", key, float32(sourceTalkersPackets[key])/float32(window.Seconds()))
	}

	fmt.Println()
	fmt.Println("by dest packets")
	fmt.Println("---------------")
	sortedKeys = sortMap(destinationTalkersPackets)
	if len(sortedKeys) < 10 {
		max = len(sortedKeys)
	} else {
		max = 10
	}
	for _, key := range sortedKeys[:max] {
		fmt.Printf("%-20s -- %.2f Pps\n", key, float32(destinationTalkersPackets[key])/float32(window.Seconds()))
	}
}

func resetTopTalkers() {
	lock.Lock()
	defer lock.Unlock()

	sourceTalkersBytes = map[string]uint64{}
	destinationTalkersBytes = map[string]uint64{}

	sourceTalkersPackets = map[string]uint64{}
	destinationTalkersPackets = map[string]uint64{}
}

func analyzeFlowSample(s sflow.FlowSample) {
	for _, record := range s.Records {
		if record.RecordType() == sflow.TypeRawPacketFlow {
			analyzeFlowRecord(record.(sflow.RawPacketFlowRecord))
		}
	}
}

func analyzeFlowRecord(r sflow.RawPacketFlowRecord) {
	_, ipHdr, _ := ethernetdecode.Decode(r.Header)
	if ipHdr == nil {
		return
	}

	switch ipHdr.IpVersion() {
	case 4:
		ipv4 := ipHdr.(ethernetdecode.Ipv4Header)
		analyzeIpv4Header(ipv4)
	case 6:
		ipv6 := ipHdr.(ethernetdecode.Ipv6Header)
		analyzeIpv6Header(ipv6)
	}
}

func analyzeIpv4Header(ipv4 ethernetdecode.Ipv4Header) {
	src := net.IP(ipv4.Source[:4]).String()
	dst := net.IP(ipv4.Destination[:4]).String()
	size := uint64(ipv4.Len)

	updateCounters(src, dst, size)
}

func analyzeIpv6Header(ipv6 ethernetdecode.Ipv6Header) {
	src := net.IP(ipv6.Source[:16]).String()
	dst := net.IP(ipv6.Destination[:16]).String()
	size := uint64(ipv6.PayloadLength)

	updateCounters(src, dst, size)
}

func updateCounters(source string, destination string, bytes uint64) {
	lock.Lock()
	defer lock.Unlock()

	if sourceBytes, present := sourceTalkersBytes[source+" => "+destination]; present {
		sourceTalkersBytes[source+" => "+destination] = sourceBytes + bytes
		sourceTalkersPackets[source+" => "+destination] += 1
	} else {
		sourceTalkersBytes[source+" => "+destination] = bytes
		sourceTalkersPackets[source+" => "+destination] = 1
	}

	if dstBytes, present := destinationTalkersBytes[destination+" <= "+source]; present {
		destinationTalkersBytes[destination+" <= "+source] = dstBytes + bytes
		destinationTalkersPackets[destination+" <= "+source] += 1
	} else {
		destinationTalkersBytes[destination+" <= "+source] = bytes
		destinationTalkersPackets[destination+" <= "+source] = 1
	}
}
