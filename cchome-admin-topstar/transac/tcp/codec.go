package tcp

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/funny/link"
	"github.com/funny/slab"
)

type protocol struct {
	pool          slab.Pool
	maxPacketSize int
}

func (p *protocol) alloc(size int) []byte {
	return p.pool.Alloc(size)
}

func (p *protocol) free(msg []byte) {
	p.pool.Free(msg)
}

func (p *protocol) sendv(session *link.Session, buffers [][]byte) error {
	err := session.Send(buffers)
	if err != nil {
		session.Close()
	}
	return err
}

func (p *protocol) send(session *link.Session, msg []byte) error {
	err := session.Send(msg)
	if err != nil {
		session.Close()
	}
	return err
}

var _ = (link.Codec)((*codec)(nil))

var headFlag interface{}
var sizeofLen = 4
var sizeofOffset = 0
var endian binary.ByteOrder = binary.BigEndian

func SetHeadFlag(hf interface{}) {
	headFlag = hf
}

func SetEndian(e binary.ByteOrder) {
	endian = e
}

func SetLenFieldIndex(offset, size int) {
	sizeofOffset = offset
	sizeofLen = size
}

var ErrTooLargePacket = errors.New("too large packet")

type codec struct {
	*protocol
	conn    net.Conn
	reader  *bufio.Reader
	headBuf []byte
	mark    string
}

func (p *protocol) newCodec(conn net.Conn, bufferSize int) *codec {
	c := &codec{
		protocol: p,
		conn:     conn,
		reader:   bufio.NewReaderSize(conn, bufferSize),
	}
	c.headBuf = make([]byte, sizeofLen+sizeofOffset)
	return c
}

func (c *codec) Receive() (interface{}, error) {
	headBuf := make([]byte, 2)
	if _, err := io.ReadFull(c.reader, headBuf); err != nil {
		return nil, err
	}
	if headBuf[0] != 0x68 {
		return nil, fmt.Errorf("pack head[%x] error", headBuf)
	}

	length := headBuf[1]
	if length > 255 {
		return nil, fmt.Errorf("pack length [%d] overlength 0xff", length)
	}
	buffer := c.alloc(int(length) + 2)
	copy(buffer, headBuf)
	if _, err := io.ReadFull(c.reader, buffer[2:]); err != nil {
		c.free(buffer)
		return nil, err
	}
	return &buffer, nil
}

func (c *codec) Send(msg interface{}) error {
	if buffers, ok := (msg.([][]byte)); ok {
		netBuf := net.Buffers(buffers)
		_, err := netBuf.WriteTo(c.conn)
		return err
	}
	_, err := c.conn.Write(msg.([]byte))
	return err
}

func (c *codec) Close() error {
	return c.conn.Close()
}
