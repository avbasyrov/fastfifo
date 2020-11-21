package fastfifo

import (
	"encoding/binary"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func BenchmarkFastFifo_MemoryAllocations(b *testing.B) {
	var err error
	q := New(300, 10485760)
	b.ReportAllocs()
	b.ResetTimer()

	msg := []byte("some msg")
	dst := make([]byte, len(msg))

	i := 0
	for i < b.N {
		for ; i < b.N; i++ {
			err = q.Push(msg)
			if err == ErrNoSpaceLeft {
				break
			}
			if err != nil {
				b.Fatal(err)
			}
		}
		for ; i < b.N; i++ {
			_, err = q.Pop(dst)
			if err == ErrNoMoreData {
				break
			}
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}

func TestCalcMsgHeaderSize(t *testing.T) {
	assert.Equal(t, 1, calcMsgHeaderSize(255))
	assert.Equal(t, 2, calcMsgHeaderSize(256))
	assert.Equal(t, 2, calcMsgHeaderSize(65535))
	assert.Equal(t, 3, calcMsgHeaderSize(65536))
}

func TestFastFifo(t *testing.T) {
	var n int
	var err error
	buf := make([]byte, 11)
	bufTooSmall := make([]byte, 10)
	msgTooLong := make([]byte, 261)

	f := New(260, 13) // for 2 bytes header

	_, err = f.Pop(buf)
	require.Equal(t, ErrNoMoreData, err)

	err = f.Push([]byte("0123456789A")) // queue: "0123456789A"
	require.NoError(t, err)

	_, err = f.Pop(bufTooSmall)
	require.Error(t, err, ErrBufferTooSmall)

	n, err = f.Pop(buf) // queue: -
	require.NoError(t, err)
	assert.Equal(t, []byte("0123456789A"), buf[:n])

	err = f.Push([]byte("001234567890")) // need 2+12 bytes > capacity
	require.Equal(t, ErrNoSpaceLeft, err)

	err = f.Push(msgTooLong) // need 2+12 bytes > capacity
	require.Equal(t, ErrMessageTooLong, err)

	err = f.Push([]byte("abcd")) // taken 2+4 = 6 bytes, queue: "abcd"
	require.NoError(t, err)

	err = f.Push([]byte("1234")) // another 6 bytes, queue: "abcd", "1234"
	require.NoError(t, err)

	err = f.Push([]byte(".")) // only 1 byte left, but need 2+1
	assert.Equal(t, ErrNoSpaceLeft, err)

	n, err = f.Pop(buf) // queue: "1234"
	require.NoError(t, err)
	assert.Equal(t, []byte("abcd"), buf[:n])

	err = f.Push([]byte("abcde")) // 1st header byte will be at buf end, 2nd at buf start, queue: "1234", "abcde"
	require.NoError(t, err)

	err = f.Push([]byte("."))
	assert.Equal(t, ErrNoSpaceLeft, err)

	n, err = f.Pop(buf) // queue: "abcde"
	require.NoError(t, err)
	assert.Equal(t, []byte("1234"), buf[:n])

	n, err = f.Pop(buf) // queue: -
	require.NoError(t, err)
	assert.Equal(t, []byte("abcde"), buf[:n])

	err = f.Push([]byte("0123456789A")) // queue: "0123456789A"
	require.NoError(t, err)

	n, err = f.Pop(buf) // queue: -
	require.NoError(t, err)
	assert.Equal(t, []byte("0123456789A"), buf[:n])
}

func TestHeaderSize(t *testing.T) {
	var f *FastFifo

	// 1 byte header (uint8)
	f = New(255, 500)

	f.putMsgLength(254, f.headerBuffer)
	assert.Equal(t, uint(254), uint(f.headerBuffer[0]))
	assert.Equal(t, uint(254), f.getMsgLength())

	// 2 bytes header (uint16)
	f = New(257, 500)

	f.putMsgLength(255, f.headerBuffer)
	assert.Equal(t, uint(255), uint(binary.LittleEndian.Uint16(f.headerBuffer)))
	assert.Equal(t, uint(255), f.getMsgLength())

	f.putMsgLength(300, f.headerBuffer)
	assert.Equal(t, uint(300), uint(binary.LittleEndian.Uint16(f.headerBuffer)))
	assert.Equal(t, uint(300), f.getMsgLength())

	// 4 bytes header (uint32)
	f = New(256*256*256+1, 256*256*256+100)

	f.putMsgLength(255, f.headerBuffer)
	assert.Equal(t, uint(255), uint(binary.LittleEndian.Uint32(f.headerBuffer)))
	assert.Equal(t, uint(255), f.getMsgLength())

	f.putMsgLength(300, f.headerBuffer)
	assert.Equal(t, uint(300), uint(binary.LittleEndian.Uint32(f.headerBuffer)))
	assert.Equal(t, uint(300), f.getMsgLength())

	f.putMsgLength(70000, f.headerBuffer)
	assert.Equal(t, uint(70000), uint(binary.LittleEndian.Uint32(f.headerBuffer)))
	assert.Equal(t, uint(70000), f.getMsgLength())

	f.putMsgLength(256*256*256-1, f.headerBuffer)
	assert.Equal(t, uint(256*256*256-1), uint(binary.LittleEndian.Uint32(f.headerBuffer)))
	assert.Equal(t, uint(256*256*256-1), f.getMsgLength())
}
