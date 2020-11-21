package fastfifo

func readCrossBoundary(src []byte, since int, length int, dst []byte) int {
	for since > len(src) {
		since -= len(src)
	}

	untilBoundary := len(src) - since

	if untilBoundary < length {
		return copy(dst, src[since:since+untilBoundary]) + copy(dst[untilBoundary:], src[0:length-untilBoundary])
	}

	return copy(dst, src[since:since+length])
}

func writeCrossBoundary(src []byte, dst []byte, dstSince int) int {
	for dstSince >= len(dst) {
		dstSince -= len(dst)
	}

	left := len(dst) - dstSince

	if left < len(src) {
		return copy(dst[dstSince:], src[:left]) + copy(dst[0:], src[left:])
	}

	return copy(dst[dstSince:], src)
}
