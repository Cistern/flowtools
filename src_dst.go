package main

import (
	"github.com/PreetamJinka/ethernetdecode"
	"github.com/PreetamJinka/sflow-go"

	"fmt"
	"net"
)

func main() {
	udpAddr, _ := net.ResolveUDPAddr("udp", ":6343")
	conn, _ := net.ListenUDP("udp", udpAddr)

	buf := make([]byte, 65535)

	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err == nil {
			datagram := sflow.Decode(buf[0:n])
			for _, sample := range datagram.Samples {
				switch sample.SampleType() {
				case sflow.TypeFlowSample:
					fs := sample.(sflow.FlowSample)
					for _, record := range fs.Records {
						if record.RecordType() == sflow.TypeRawPacketFlow {
							r := record.(sflow.RawPacketFlowRecord)
							_, ipHdr, _ := ethernetdecode.Decode(r.Header)
							if ipHdr != nil && ipHdr.IpVersion() == 4 {
								ipv4 := ipHdr.(ethernetdecode.Ipv4Header)
								fmt.Printf("src: %v => dst: %v\n", net.IP(ipv4.Source[:]), net.IP(ipv4.Destination[:]))
							}
						}
					}
				default:
				}
			}
		} else {
			fmt.Println(err)
		}
	}
}
