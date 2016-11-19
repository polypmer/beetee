package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
)

// Peer is the basic unit of other.
type Peer struct {
	ip   string
	port uint16
	id   string
	addr string
	// Connection
	conn net.Conn
	// Status
	//	stopping   chan struct{}
	alive      bool
	interested bool
	choked     bool
	choking    bool
	choke      sync.WaitGroup // NOTE: Use this?
	// Messages
	sendChan chan []byte
	recvChan chan []byte
	// Piece Data
	bitfield map[int]bool
}

// parsePeers is a http response gotten from
// the tracker; parse the peers byte message
// and put to global Peers slice.
func (r *TrackerResponse) parsePeers() {
	var start int
	for idx, val := range r.Peers {
		if val == ':' {
			start = idx + 1
			break
		}
	}
	p := r.Peers[start:]
	// A peer is represented in six bytes
	// four for ip and two for port
	for i := 0; i < len(p); i = i + 6 {
		ip := net.IPv4(p[i], p[i+1], p[i+2], p[i+3])
		port := (uint16(p[i+4]) << 8) | uint16(p[i+5])
		peer := Peer{
			ip:      ip.String(),
			port:    port,
			addr:    fmt.Sprintf("%s:%d", p.Ip, p.Port),
			choking: true,
			choked:  true,
		}
		Peers = append(Peers, &peer)
	}
}

func (p *Peer) ConnectPeer() error {
	log.Printf("Connecting to %s", p.addr)
	// Connect to address
	conn, err := net.Dial("tcp", p.addr)
	if err != nil {
		return err
	}
	p.conn = conn

	// NOTE: Does io.Readfull Block?
	err = p.sendHandShake()
	if err != nil {
		return err
	}
	p.alive = true
	logger.Printf("Connected to %s at %s", p.id, p.addr)
	return nil
}

// ListenPeer reads from socket.
func (p *Peer) ListenPeer() {
	for {
		length := make([]byte, 4)
		_, err := io.ReadFull(p.conn, length)
		if err != nil {
			// EOF
			debugger.Printf("Error %s with %s", err, p.id)
			p.alive = false
			p.conn.Close()
			return
		}
		payload := make([]byte, binary.BigEndian.Uint32(length))
		_, err = io.ReadFull(p.conn, payload)
		if err != nil {
			debugger.Printf("Error %s with %s", err, p.id)
			p.alive = false
			p.conn.Close()
			return
		}
		p.recvChan <- payload
	}
}
