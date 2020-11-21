package fastfifo

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWriteCrossBoundary(t *testing.T) {
	var n int
	buf := make([]byte, 6)

	n = writeCrossBoundary([]byte("123456"), buf, 0)
	assert.Equal(t, 6, n)
	assert.Equal(t, []byte("123456"), buf)

	n = writeCrossBoundary([]byte("123456"), buf, 1)
	assert.Equal(t, 6, n)
	assert.Equal(t, []byte("612345"), buf)

	n = writeCrossBoundary([]byte("123456"), buf, 4)
	assert.Equal(t, 6, n)
	assert.Equal(t, []byte("345612"), buf)

	n = writeCrossBoundary([]byte("123456"), buf, 5)
	assert.Equal(t, 6, n)
	assert.Equal(t, []byte("234561"), buf)

	n = writeCrossBoundary([]byte("123456"), buf, 6)
	assert.Equal(t, 6, n)
	assert.Equal(t, []byte("123456"), buf)

	n = writeCrossBoundary([]byte("123456"), buf, 7)
	assert.Equal(t, 6, n)
	assert.Equal(t, []byte("612345"), buf)

	n = writeCrossBoundary([]byte("---"), buf, 0)
	assert.Equal(t, 3, n)
	assert.Equal(t, []byte("---345"), buf)

	n = writeCrossBoundary([]byte("xx"), buf, 5)
	assert.Equal(t, 2, n)
	assert.Equal(t, []byte("x--34x"), buf)
}

func TestReadCrossBoundary(t *testing.T) {
	var n int
	buf := make([]byte, 6)

	n = readCrossBoundary([]byte("123456"), 0, 6, buf)
	assert.Equal(t, 6, n)
	assert.Equal(t, []byte("123456"), buf)

	n = readCrossBoundary([]byte("123456"), 1, 6, buf)
	assert.Equal(t, 6, n)
	assert.Equal(t, []byte("234561"), buf)

	n = readCrossBoundary([]byte("123456"), 5, 6, buf)
	assert.Equal(t, 6, n)
	assert.Equal(t, []byte("612345"), buf)

	n = readCrossBoundary([]byte("123456"), 6, 6, buf)
	assert.Equal(t, 6, n)
	assert.Equal(t, []byte("123456"), buf)

	n = readCrossBoundary([]byte("123456"), 7, 6, buf)
	assert.Equal(t, 6, n)
	assert.Equal(t, []byte("234561"), buf)
}
