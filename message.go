package main

import (
	"encoding/binary"
	//"time"
)

const (
	ChokeMsg = iota
	UnchokeMsg
	InterestedMsg
	NotInterestedMsg
	HaveMsg
	BitFieldMsg
	RequestMsg
	BlockMsg // rather than PieceMsg
	CancelMsg
	PortMsg
)

/*###################################################
Recieving Messages
######################################################*/

// DecodePieceMessage takes the message without the
// length prefix
func DecodePieceMessage(msg []byte) *Block {
	if len(msg[9:]) < 1 {
		return nil
	}
	index := binary.BigEndian.Uint32(msg[1:5]) // NOTE: The piece in question
	begin := binary.BigEndian.Uint32(msg[5:9])
	data := msg[9:]
	var blocksize int
	if int(index) == len(d.Pieces)-1 { // NOTE: last piece
		if d.Pieces[index].size < BLOCKSIZE {
			blocksize = int(d.Pieces[index].size)
		} else {
			blocksize = BLOCKSIZE
		}
	} else {
		blocksize = BLOCKSIZE
	}

	return &Block{
		index:  index,
		offset: begin,
		data:   data,
		size:   blocksize}
}

// 19 bytes
func DecodeHaveMessage(msg []byte) uint32 {
	return binary.BigEndian.Uint32(msg[1:])
}

// DecodeBitfieldMessage returns a bool slice.
// NOTE: The bitfield will be sent with padding if the size is
// not divisible by eight.
// Thank you Tulva RC bittorent client for this algorithm
// github.com/jtakkala/tulva/
func DecodeBitfieldMessage(msg []byte) []bool {
	result := make([]bool, len(d.Pieces))
	bitfield := msg[1:]
	// For each byte, look at the bits
	// NOTE: that is 8 * 8
	for i := 0; i < len(bitfield); i++ {
		for j := 0; j < 8; j++ {
			index := i*8 + j
			if index >= len(d.Pieces) {
				break // Hit padding bits
			}
			byte := bitfield[i]              // Within bytes
			bit := (byte >> uint32(7-j)) & 1 // some shifting
			result[index] = bit == 1         // if bit is true
		}
	}
	return result
}

func DecodeRequestMessage(msg []byte) {
}

func DecodeCancelMessage(msg []byte) {
}

func DecodePortMessage(msg []byte) {
}

/*###################################################
Sending Messages
######################################################*/

//<pstrlen><pstr><reserved><info_hash><peer_id>
// 68 bytes long.
func HandShake(info *TorrentMeta) [68]byte {
	//h := make([]byte)
	var h [68]byte
	h[0] = pstrlen
	copy(h[1:20], pstr[:])
	copy(h[28:48], info.InfoHash[:])
	copy(h[48:], info.PeerId[:])

	return h
}

// StatusMessage sends the status message to peer.
// If sent -1 then a Keep alive message is sent.
func StatusMessage(status int) []byte {
	//<len=0001><id=1>
	msg := make([]byte, 5)
	length := make([]byte, 4)
	if status == -1 {
		binary.BigEndian.PutUint32(length, 0)
		return length // NOTE: Keep alive message
	} else {
		binary.BigEndian.PutUint32(length, 1)
	}

	copy(msg[:4], length)
	msg[4] = byte(status)

	return msg
}

// sendRequestMessage pass in the index of the piece your looking for,
// and the offset of the piece (it's offset index * BLOCKSIZE
func RequestMessage(idx uint32, offset int) []byte {
	//4-byte message length,1-byte message ID, and payload:
	// <len=0013><id=6><index><begin><length>
	msg := make([]byte, 17)
	// Message prefix
	binary.BigEndian.PutUint32(msg[:4], 13)
	msg[4] = byte(RequestMsg)
	// Payload
	binary.BigEndian.PutUint32(msg[5:9], idx)
	binary.BigEndian.PutUint32(msg[9:13], uint32(offset))
	binary.BigEndian.PutUint32(msg[13:], uint32(BLOCKSIZE))

	return msg
}

// PieceMessage send a block of a piece
// and the offset of the piece (it's offset index * BLOCKSIZE
func PieceMessage(idx uint32, offset int, data []byte) []byte {
	// 4-byte message length,1-byte message ID, and payload:
	// <len=0009+X><id=7><index><begin><block>
	msg := make([]byte, 13+len(data))
	// Message prefix
	binary.BigEndian.PutUint32(msg[:4], uint32(len(data)+9))
	msg[4] = byte(BlockMsg)
	binary.BigEndian.PutUint32(msg[5:9], idx)
	binary.BigEndian.PutUint32(msg[9:13], uint32(offset))
	copy(msg[13:], data)

	return msg
}

func requestPiece(piece int) [][]byte {
	blocksPerPiece := int(d.Torrent.Info.PieceLength) / BLOCKSIZE
	msgs := make([][]byte, 0)
	for offset := 0; offset < blocksPerPiece; offset++ {
		msgs = append(msgs, RequestMessage(uint32(piece), offset*BLOCKSIZE))
	}

	return msgs
}
