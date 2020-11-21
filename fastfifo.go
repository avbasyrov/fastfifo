package fastfifo

import (
	"sync"
)

type constError string

func (err constError) Error() string {
	return string(err)
}

const (
	// ErrNoMoreData is the error returned by Pop() when no more items is available in queue.
	ErrNoMoreData = constError("no more data in queue")
	// ErrNoSpaceLeft is the error returned by Push() when not enough space in queue for pushed item.
	ErrNoSpaceLeft = constError("no space left")

	// ErrMessageTooLong is the error returned by Push() when pushed item exceeds maxMessageSize.
	ErrMessageTooLong = constError("message too long")
	// ErrBufferTooSmall is the error returned by Pop() when destination buffer too small.
	ErrBufferTooSmall = constError("destination buffer too small")
)

// FastFifo is FIFO queue with fixed size ring (cycled) buffer
// and no memory allocations during Push / Pop.
// Thread safe.
type FastFifo struct {
	buffer         []byte
	headerBuffer   []byte
	maxMessageSize int
	data           cycledRange
	freeSpace      cycledRange

	mu *sync.Mutex
}

type cycledRange struct {
	start  int
	length int
}

// New returns a FastFifo with given capacity and maxMessageSize.
func New(maxMessageSize int, capacity int) *FastFifo {
	return &FastFifo{
		headerBuffer:   make([]byte, calcMsgHeaderSize(maxMessageSize)),
		maxMessageSize: maxMessageSize,
		buffer:         make([]byte, capacity),
		data: cycledRange{
			start:  0,
			length: 0,
		},
		freeSpace: cycledRange{
			start:  0,
			length: capacity,
		},
		mu: &sync.Mutex{},
	}
}

// Push item to queue. Thread safe.
// Returns ErrNoSpaceLeft when no more space in queue.
// Returns ErrMessageTooLong when msg exceeds maxMessageSize limit.
func (f *FastFifo) Push(msg []byte) error {
	if len(msg) > f.maxMessageSize {
		return ErrMessageTooLong
	}

	totalLength := len(f.headerBuffer) + len(msg)

	f.mu.Lock()
	defer f.mu.Unlock()

	if totalLength > f.freeSpace.length {
		return ErrNoSpaceLeft
	}

	if len(f.buffer)-f.freeSpace.start >= totalLength {
		f.putMsgLength(uint(len(msg)), f.buffer[f.freeSpace.start:])
		copy(f.buffer[f.freeSpace.start+len(f.headerBuffer):], msg)
		f.data.length += totalLength
		f.freeSpace.start += totalLength
		if f.freeSpace.start >= len(f.buffer) {
			f.freeSpace.start -= len(f.buffer)
		}
		f.freeSpace.length -= totalLength
		return nil
	}

	if len(f.buffer)-f.freeSpace.start >= len(f.headerBuffer) {
		f.putMsgLength(uint(len(msg)), f.buffer[f.freeSpace.start:])
	} else {
		f.putMsgLength(uint(len(msg)), f.headerBuffer)
		writeCrossBoundary(f.headerBuffer, f.buffer, f.freeSpace.start)
	}

	writeCrossBoundary(msg, f.buffer, f.freeSpace.start+len(f.headerBuffer))
	f.data.length += totalLength
	f.freeSpace.start += totalLength
	if f.freeSpace.start >= len(f.buffer) {
		f.freeSpace.start -= len(f.buffer)
	}
	f.freeSpace.length -= totalLength

	return nil
}

// Pop get and remove item from queue. Thread safe.
// Returns ErrNoMoreData when no more items in queue.
// Returns ErrBufferTooSmall when msg exceeds size of dst buffer.
func (f *FastFifo) Pop(dst []byte) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.data.length == 0 {
		return 0, ErrNoMoreData
	}

	readCrossBoundary(f.buffer, f.data.start, len(f.headerBuffer), f.headerBuffer)
	length := int(f.getMsgLength())

	if cap(dst) < length {
		return 0, ErrBufferTooSmall
	}

	readCrossBoundary(f.buffer, f.data.start+len(f.headerBuffer), length, dst)

	totalLength := len(f.headerBuffer) + length
	f.freeSpace.length += totalLength
	f.data.start += totalLength
	if f.data.start >= len(f.buffer) {
		f.data.start -= len(f.buffer)
	}
	f.data.length -= totalLength

	return length, nil
}

func calcMsgHeaderSize(maxMessageSize int) int {
	size := 0

	for i := maxMessageSize; i > 0; i /= 256 {
		size++
	}

	return size
}

func (f *FastFifo) putMsgLength(size uint, dst []byte) {
	dst[0] = byte(size)

	for i := 1; i < len(f.headerBuffer); i++ {
		dst[i] = byte(size >> (i * 8))
	}
}

func (f *FastFifo) getMsgLength() uint {
	size := uint(f.headerBuffer[0])

	for i := 1; i < len(f.headerBuffer); i++ {
		size |= uint(uint(f.headerBuffer[i]) << (i * 8))
	}

	return size
}
