package main

import (
	"fmt"
	"net"
	"os"

	"github.com/miekg/dns"
)

func main() {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: 5354})
	if err != nil {
		fmt.Printf("listen: %v\n", err)
		os.Exit(1)
	}

	for {
		buf := make([]byte, 1232) // DNS Flag Day 2020
		size, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Printf("receive: %v\n", err)
			continue
		}

		msg := new(dns.Msg)
		msg.Unpack(buf[0:size])
		fmt.Printf("%s sent %#v\n", addr, msg)

		if !msg.MsgHdr.Response {
			msg.Response = true
			if msg.Question[0].Name == "alvo.me." {
				msg.Answer = []dns.RR{&dns.A{
					Hdr: dns.RR_Header{
						Name: "alvo.me.", Rrtype: dns.TypeA,
						Class: dns.ClassINET, Ttl: 3600,
					},
					A: net.IPv4(1, 2, 3, 4),
				}}
			}

			bytes, _ := msg.Pack()
			conn.WriteToUDP(bytes, addr)
		}
	}
}
