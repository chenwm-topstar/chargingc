package lg

import (
	"io"
	"sync"

	"cchome-admin-topstar/utils/slices"

	"github.com/glycerine/rbuf"
)

var _ io.Writer = (*Terminal)(nil)

type Terminal struct {
	buf        *rbuf.FixedSizeRingBuf
	ttys       []*tty
	idSeq      int
	maxBufSize int
	lock       sync.Mutex
}

func NewTerminal(maxBufSize int) *Terminal {
	return &Terminal{
		buf:        rbuf.NewFixedSizeRingBuf(maxBufSize),
		maxBufSize: maxBufSize,
	}
}

func (t *Terminal) Write(p []byte) (int, error) {
	t.lock.Lock()
	defer t.lock.Unlock()

	n, err := t.buf.WriteAndMaybeOverwriteOldestData(p)
	for _, tty := range t.ttys {
		tty.buf.WriteAndMaybeOverwriteOldestData(p)
	}
	return n, err
}

func (t *Terminal) ForkTTY(tailBytes int) *tty {
	t.lock.Lock()
	defer t.lock.Unlock()

	if tailBytes > t.maxBufSize || tailBytes < 0 {
		tailBytes = t.maxBufSize
	}

	t.idSeq++
	tty := &tty{
		buf: rbuf.NewFixedSizeRingBuf(t.maxBufSize),
		id:  t.idSeq,
	}
	// Clone ring buffer to new tty
	a, b := t.buf.BytesTwo(false)
	tty.buf.Write(a)
	tty.buf.Write(b)
	if len(a)+len(b) > tailBytes {
		tty.buf.Advance(len(a) + len(b) - tailBytes)
	}

	t.ttys = append(t.ttys, tty)
	// Remove tty from terminal
	tty.close = func() error {
		t.lock.Lock()
		defer t.lock.Unlock()
		n := slices.Filter(t.ttys, func(i int) bool {
			return t.ttys[i].id != tty.id
		})
		t.ttys = t.ttys[:n]
		return nil
	}
	return tty
}

type tty struct {
	buf   *rbuf.FixedSizeRingBuf
	id    int
	close func() error
}

var _ io.ReadCloser = (*tty)(nil)

func (t *tty) Read(p []byte) (int, error) {
	return t.buf.Read(p)
}

func (t *tty) Close() error {
	return t.close()
}
