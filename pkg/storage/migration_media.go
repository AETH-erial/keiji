package storage

import (
	"bytes"
	"net/http"
)

// detectMediaType trusts bytes rather than the filename or multipart MIME header.
func detectMediaType(data []byte) string {
	switch http.DetectContentType(data) {
	case "image/jpeg", "image/png", "image/gif", "image/webp":
		return http.DetectContentType(data)
	}
	if isWebM(data) {
		return "video/webm"
	}
	return ""
}

// isWebM parses the EBML header and requires an explicit WebM DocType. Merely
// finding the word "webm" or an EBML magic number also admits Matroska files.
func isWebM(data []byte) bool {
	if len(data) < 5 || !bytes.Equal(data[:4], []byte{0x1a, 0x45, 0xdf, 0xa3}) {
		return false
	}
	size, n, ok := ebmlSize(data[4:])
	if !ok || size > uint64(len(data)-4-n) {
		return false
	}
	header := data[4+n : 4+n+int(size)]
	found := false
	for len(header) > 0 {
		// EBML element IDs retain their leading marker bit.
		_, idLen, valid := ebmlSize(header)
		if !valid || idLen > 4 || len(header) <= idLen {
			return false
		}
		id := header[:idLen]
		length, lenLen, valid := ebmlSize(header[idLen:])
		if !valid || length > uint64(len(header)-idLen-lenLen) {
			return false
		}
		value := header[idLen+lenLen : idLen+lenLen+int(length)]
		if bytes.Equal(id, []byte{0x42, 0x82}) {
			if found || !bytes.Equal(value, []byte("webm")) {
				return false
			}
			found = true
		}
		header = header[idLen+lenLen+int(length):]
	}
	return found
}

// ebmlSize reads a bounded EBML variable-length integer. Unknown sizes are
// rejected because the header and its children must have finite lengths.
func ebmlSize(b []byte) (uint64, int, bool) {
	if len(b) == 0 || b[0] == 0 {
		return 0, 0, false
	}
	mask, n := byte(0x80), 1
	for b[0]&mask == 0 {
		mask >>= 1
		n++
	}
	if n > len(b) {
		return 0, 0, false
	}
	value := uint64(b[0] & (mask - 1))
	for _, v := range b[1:n] {
		value = value<<8 | uint64(v)
	}
	if value == (uint64(1)<<(7*n))-1 {
		return 0, 0, false
	}
	return value, n, true
}
