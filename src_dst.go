package main

import (
	"github.com/PreetamJinka/ethernetdecode"
	"github.com/PreetamJinka/sflow-go"

	"fmt"
	"net"
)

func main() {
	opened := 0
	closed := 0

	udpAddr, _ := net.ResolveUDPAddr("udp", ":6343")
	conn, err := net.ListenUDP("udp", udpAddr)

	fmt.Println(err)

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
							_, ipHdr, protoHdr := ethernetdecode.Decode(r.Header)
							if ipHdr != nil && ipHdr.IpVersion() == 4 {
								ipv4 := ipHdr.(ethernetdecode.Ipv4Header)
								switch protoHdr.Protocol() {
								case ethernetdecode.ProtocolTcp:
									tcp := protoHdr.(ethernetdecode.TcpHeader)

									// SYN+ACK flags
									if tcp.Flags&3 == 2 {
										opened++
									}

									// FIN+ACK flags
									if tcp.Flags&17 != 0 {
										closed++
									}
								case ethernetdecode.ProtocolUdp:

								}
								fmt.Printf("src: %v => dst: %v\n", net.IP(ipv4.Source[:]), net.IP(ipv4.Destination[:]))
								fmt.Printf("TCP connections opened: %d, closed: %d\n", opened, closed)
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
