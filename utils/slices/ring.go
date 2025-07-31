package slices

import (
	"errors"
	"sync"
)

// Ring is a Thread-safe string slice with fixed size.
// Only the last "capacity" strings are retained.
type Ring struct {
	Data []interface{}
	size int
	cap  int
	Lock sync.RWMutex
}

// NewRing returns a Ring object with capacity "cap".
func NewRing(cap int) *Ring {
	return &Ring{
		Data: make([]interface{}, cap),
		cap:  cap,
	}
}

// Size returns the logical size of the appended data.
// Only the [Size-Cap, Size) elements are retained.
func (rs *Ring) Size() int {
	rs.Lock.RLock()
	defer rs.Lock.RUnlock()
	return rs.size
}

// Len returns the actual number of elements stored in the ring.
func (rs *Ring) Len() int {
	rs.Lock.RLock()
	defer rs.Lock.RUnlock()
	if rs.size > rs.cap {
		return rs.cap
	}
	return rs.size
}

// Cap returns the capacity of the ring.
func (rs *Ring) Cap() int {
	return rs.cap
}

// Append adds elements to the tail of ring.
func (rs *Ring) Append(ele ...interface{}) {
	if len(ele) == 0 {
		return
	}
	rs.Lock.Lock()
	defer rs.Lock.Unlock()
	if len(ele) > rs.cap {
		// The input has already fill up the ring, get only the last "cap".
		rs.fill(ele[len(ele)-rs.cap:len(ele)], rs.size+len(ele)-rs.cap)
	} else {
		rs.fill(ele, rs.size)
	}
	rs.size += len(ele)
}

func (rs *Ring) fill(objs []interface{}, from int) {
	start := from % rs.cap
	if start+len(objs) > rs.cap {
		copy(rs.Data[start:], objs[:rs.cap-start])
		copy(rs.Data[:len(objs)-(rs.cap-start)], objs[rs.cap-start:])
	} else {
		copy(rs.Data[start:], objs)
	}
}

// Get returns the idx'th element.
func (rs *Ring) Get(idx int) (interface{}, error) {
	rs.Lock.RLock()
	defer rs.Lock.RUnlock()
	if idx >= rs.size || idx < rs.size-rs.cap || idx < 0 {
		return "", errors.New("Index out of range")
	}
	return rs.Data[idx%rs.cap], nil
}
