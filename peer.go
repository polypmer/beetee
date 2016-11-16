package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
)

type Peer struct {
	meta *TorrentMeta // Whence it came

	PeerId string `bencode:"peer id"` // Bencoding not being used
	Ip     string `bencode:"ip"`
	Port   uint16 `bencode:"port"`
	// Buffer
	// TODO: make writer?
	// After connection
	Shaken bool
	Conn   net.Conn
	Id     string
	// Peer Status
	Alive      bool
	Interested bool
	Choked     bool
	Stop       chan bool // TODO: What?
	ChokeWg    sync.WaitGroup
	// What the Peer Has, index wise
	Bitfield []bool
	has      map[uint32]bool
}

func (p *Peer) ConnectToPeer() error {
	addr := fmt.Sprintf("%s:%d", p.Ip, p.Port)
	logger.Println("Connecting to Peer: ", addr)

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	p.Conn = conn
	p.has = make(map[uint32]bool)
	err = p.ShakeHands()
	if err != nil {
		return err
	}
	p.Alive = true
	p.ChokeWg.Add(1)
	logger.Println("Connected to Peer: ", p.Id)
	// TODO: Keep alive loop in goroutine
	return nil

}

// ListenToPeer handshakes with peer and sends
// messages to the decoder
// Connect first
func (p *Peer) ListenToPeer() error {
	// Handshake
	logger.Printf("Peer %s : starting to Listen\n", p.Id)
	// Listen Loop
	go func() {
		for {
			length := make([]byte, 4)
			_, err := io.ReadFull(p.Conn, length)
			//debugger.Println(length)
			if err != nil {
				debugger.Printf("Error Reading Length %s, Stopping: %s", p.Id, err)
				p.Alive = false
				p.Conn.Close()
				return
			}
			payload := make([]byte, binary.BigEndian.Uint32(length))
			_, err = io.ReadFull(p.Conn, payload)
			if err != nil {
				debugger.Printf("Error Reading Payload %s, Stopping: %s", p.Id, err)
				// TODO: Stop connection
				//p.Stop <- true
				p.Alive = false
				p.Conn.Close()
				return
			}
			go p.decodeMessage(payload)
		}
	}()
	return nil
}
