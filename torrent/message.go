package torrent

import "encoding/binary"

// Message is the interface that all Message types should adhere to. They should all have a Format
// function which returns the byte array to be written to a connection.
type Message interface {
	Format() []byte
}

// KeepAlive implements a keep-alive message.
type KeepAlive struct {
}

func (m KeepAlive) Format() []byte {
	return uint32ToByteSlice(0)
}

// Choke implements a choke message.
type Choke struct {
}

func (m Choke) Format() []byte {
	buf := uint32ToByteSlice(1)
	return append(buf, 0)
}

// Unchoke implements an unchoke message.
type Unchoke struct {
}

func (m Unchoke) Format() []byte {
	buf := uint32ToByteSlice(1)
	return append(buf, 1)
}

// Interested implements an interested message.
type Interested struct {
}

func (m Interested) Format() []byte {
	buf := uint32ToByteSlice(1)
	return append(buf, 2)
}

// NotInterested implements a not interested message.
type NotInterested struct {
}

func (m NotInterested) Format() []byte {
	buf := uint32ToByteSlice(1)
	return append(buf, 3)
}

// Have implements a have message.
type Have struct {
	PieceIndex uint32
}

func (m Have) Format() []byte {
	buf := uint32ToByteSlice(5)
	buf = append(buf, 4)
	return append(buf, uint32ToByteSlice(m.PieceIndex)...)
}

// Bitfield implements a bitfield message.
type Bitfield struct {
	Data []byte
}

func (m Bitfield) Format() []byte {
	buf := uint32ToByteSlice(uint32(1 + len(m.Data)))
	buf = append(buf, 5)
	return append(buf, m.Data...)

}

// Request implements a request message.
type Request struct {
	Index  uint32
	Begin  uint32
	Length uint32
}

func (m Request) Format() []byte {
	buf := uint32ToByteSlice(13)
	buf = append(buf, 6)
	buf = append(buf, uint32ToByteSlice(m.Index)...)
	buf = append(buf, uint32ToByteSlice(m.Begin)...)
	return append(buf, uint32ToByteSlice(m.Length)...)
}

// Piece implements a piece message.
type Piece struct {
	Index uint32
	Begin uint32
	Block []byte
}

func (m Piece) Format() []byte {
	buf := uint32ToByteSlice(uint32(9 + len(m.Block)))
	buf = append(buf, 7)
	buf = append(buf, uint32ToByteSlice(m.Index)...)
	buf = append(buf, uint32ToByteSlice(m.Begin)...)
	return append(buf, m.Block...)
}

// Cancel implements a cancel message.
type Cancel struct {
	Index  uint32
	Begin  uint32
	Length uint32
}

func (m Cancel) Format() []byte {
	buf := uint32ToByteSlice(13)
	buf = append(buf, 8)
	buf = append(buf, uint32ToByteSlice(m.Index)...)
	buf = append(buf, uint32ToByteSlice(m.Begin)...)
	return append(buf, uint32ToByteSlice(m.Length)...)
}

func uint32ToByteSlice(v uint32) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, v)
	return buf
}
