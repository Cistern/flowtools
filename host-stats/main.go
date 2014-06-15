package main

import (
	"github.com/PreetamJinka/sflow-go"

	"fmt"
	"net"
	"time"
)

func printDatagram(buf []byte) {
	dgram := sflow.Decode(buf)

	for _, sample := range dgram.Samples {
		switch sample.SampleType() {
		case sflow.TypeCounterSample:
			for _, record := range sample.(sflow.CounterSample).Records {
				printRecord(record)
			}
		default:
		}
	}

	fmt.Println("----")
}

func printRecord(record sflow.Record) {
	switch record.RecordType() {
	case sflow.TypeHostCpuCounter:
		r := record.(sflow.HostCpuCounters)
		fmt.Println("CPU")
		fmt.Printf("  Loads: %.02f %.02f %.02f, Processes running/total: %d/%d, Uptime: %v\n",
			r.Load1m, r.Load5m, r.Load15m,
			r.ProcsRunning, r.ProcsTotal, time.Duration(r.Uptime)*time.Second)
	case sflow.TypeHostMemoryCounter:
		r := record.(sflow.HostMemoryCounters)
		fmt.Println("Memory")
		fmt.Printf("  Free/Total: %dM/%dM, Swap Free/Total: %dM/%dM\n",
			r.Free/1024/1024, r.Total/1024/1024,
			r.SwapFree/1024/1024, r.SwapTotal/1024/1024)
	case sflow.TypeHostDiskCounter:
		r := record.(sflow.HostDiskCounters)
		fmt.Println("Disk")
		fmt.Printf("  Space Free/Total: %dG/%dG, Bytes Read/Written: %dM/%dM\n",
			r.Free/1024/1024/1024, r.Total/1024/1024/1024,
			r.BytesRead/1024/1024, r.BytesWritten/1024/1024)
	case sflow.TypeHostNetCounter:
		r := record.(sflow.HostNetCounters)
		fmt.Println("Network")
		fmt.Printf("  Packets In/Out: %d/%d, Bytes In/Out: %dM/%dM\n",
			r.PacketsIn, r.PacketsOut,
			r.BytesIn/1024/1024, r.BytesOut/1024/1024)
	}
}

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", ":6343")
	if err != nil {
		fmt.Println(err)
		return
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println(err)
		return
	}

	buf := make([]byte, 65535)
	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err == nil {
			printDatagram(buf[:n])
		}
	}
}
