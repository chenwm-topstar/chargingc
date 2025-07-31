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

// =======================================================================

var _ = (link.Codec)((*codec)(nil))

var headFlag interface{}                       // 头部标示
var sizeofLen = 4                              // 报文长度大小
var sizeofOffset = 0                           // 报文偏移
var endian binary.ByteOrder = binary.BigEndian // 字节序大小端

// SetHeadFlag 设置头部标示
func SetHeadFlag(hf interface{}) {
	headFlag = hf
}

// SetEndian 设置大小端
func SetEndian(e binary.ByteOrder) {
	endian = e
}

// SetLenFieldIndex 设置包长度字段所在位置
func SetLenFieldIndex(offset, size int) {
	sizeofOffset = offset
	sizeofLen = size
}

// ErrTooLargePacket 超长数据包
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

// Receive implements link/Codec.Receive() method.
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

// Send implements link/Codec.Send() method.
func (c *codec) Send(msg interface{}) error {
	if buffers, ok := (msg.([][]byte)); ok {
		netBuf := net.Buffers(buffers)
		_, err := netBuf.WriteTo(c.conn)
		return err
	}
	_, err := c.conn.Write(msg.([]byte))
	return err
}

// Close implements link/Codec.Close() method.
func (c *codec) Close() error {
	return c.conn.Close()
}
