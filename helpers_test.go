package main

import (
	"bytes"
	"encoding/binary"
)

// makeID3v2 returns a minimal ID3v2.3 tag followed by zeroed audio frames,
// enough for the github.com/dhowden/tag library to read the title/artist/album.
func makeID3v2(id, text string) []byte {
	// ID3v2 header (10 bytes)
	buf := &bytes.Buffer{}
	buf.WriteString("ID3")
	buf.WriteByte(0x03) // version 2.3.0
	buf.WriteByte(0x00) // revision
	buf.WriteByte(0x00) // flags

	frames := &bytes.Buffer{}
	writeTextFrame(frames, "TIT2", "Test Title")
	writeTextFrame(frames, "TPE1", "Test Artist")
	writeTextFrame(frames, "TALB", "Test Album")

	// size is stored as synchsafe 32-bit int in bytes 6..10
	size := frames.Len()
	buf.Write(synchsafe(size))

	payload := append(buf.Bytes(), frames.Bytes()...)

	// Zero-fill to make it look like a real playable file (audio frames don't matter).
	pad := make([]byte, 64)
	payload = append(payload, pad...)
	return payload
}

// makeID3v2Title returns a minimal ID3v2.3 tag with a custom title.
func makeID3v2Title(title string) []byte {
	buf := &bytes.Buffer{}
	buf.WriteString("ID3")
	buf.WriteByte(0x03) // version 2.3.0
	buf.WriteByte(0x00) // revision
	buf.WriteByte(0x00) // flags

	frames := &bytes.Buffer{}
	writeTextFrame(frames, "TIT2", title)
	writeTextFrame(frames, "TPE1", "Test Artist")
	writeTextFrame(frames, "TALB", "Test Album")

	size := frames.Len()
	buf.Write(synchsafe(size))
	payload := append(buf.Bytes(), frames.Bytes()...)
	payload = append(payload, make([]byte, 64)...)
	return payload
}

func writeTextFrame(buf *bytes.Buffer, id, text string) {
	buf.WriteString(id)
	// encoding byte 0x00 = latin-1, then text bytes, then a NUL terminator
	content := append([]byte{0x00}, append([]byte(text), 0x00)...)
	buf.Write(binary.BigEndian.AppendUint32(nil, uint32(len(content))))
	buf.WriteByte(0x00) // flags
	buf.WriteByte(0x00) // flags
	buf.Write(content)
}

func synchsafe(n int) []byte {
	s := make([]byte, 4)
	s[3] = byte(n & 0x7f)
	s[2] = byte((n >> 7) & 0x7f)
	s[1] = byte((n >> 14) & 0x7f)
	s[0] = byte((n >> 21) & 0x7f)
	return s
}

func makeMP3Bytes() []byte {
	return makeID3v2("", "")
}

// makeEmptyTag returns a valid ID3v2.3 header with no frames, so metadata reads
// as empty (missing title/artist/album).
func makeEmptyTag() []byte {
	buf := &bytes.Buffer{}
	buf.WriteString("ID3")
	buf.WriteByte(0x03) // version 2.3.0
	buf.WriteByte(0x00) // revision
	buf.WriteByte(0x00) // flags
	buf.Write(synchsafe(0))
	return buf.Bytes()
}
