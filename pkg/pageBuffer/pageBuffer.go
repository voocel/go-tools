package pageBuffer

import (
	"errors"
	"io"
	"log"
)

type page struct {
	buf []byte
}

type PageBuffer struct {
	pages []*page

	length       int
	nextPageSize int
}

func NewPageBuffer(pageSize int) *PageBuffer {
	b := &PageBuffer{}
	b.pages = append(b.pages, &page{buf: make([]byte, 0, pageSize)})
	b.nextPageSize = pageSize * 2
	return b
}

func (b *PageBuffer) Write(data []byte) (int, error) {
	dataLen := len(data)
	for {
		cp := b.pages[len(b.pages)-1]

		n := copy(cp.buf[len(cp.buf):cap(cp.buf)], data)
		cp.buf = cp.buf[:len(cp.buf)+n]
		b.length += n

		if len(data) == n {
			break
		}
		data = data[n:]

		b.pages = append(b.pages, &page{buf: make([]byte, 0, b.nextPageSize)})
		b.nextPageSize *= 2
	}

	return dataLen, nil
}

func (b *PageBuffer) WriteByte(data byte) error {
	_, err := b.Write([]byte{data})
	return err
}

func (b *PageBuffer) Len() int {
	return b.length
}

func (b *PageBuffer) pageForOffset(offset int) (int, int) {
	AssertTrue(offset < b.length)

	var pageIdx, startIdx, sizeNow int
	for i := 0; i < len(b.pages); i++ {
		cp := b.pages[i]

		if sizeNow+len(cp.buf)-1 < offset {
			sizeNow += len(cp.buf)
		} else {
			pageIdx = i
			startIdx = offset - sizeNow
			break
		}
	}

	return pageIdx, startIdx
}

func (b *PageBuffer) Truncate(n int) {
	pageIdx, startIdx := b.pageForOffset(n)
	b.pages = b.pages[:pageIdx+1]
	cp := b.pages[len(b.pages)-1]
	cp.buf = cp.buf[:startIdx]
	b.length = n
}

func (b *PageBuffer) Bytes() []byte {
	buf := make([]byte, b.length)
	written := 0
	for i := 0; i < len(b.pages); i++ {
		written += copy(buf[written:], b.pages[i].buf)
	}

	return buf
}

func (b *PageBuffer) WriteTo(w io.Writer) (int64, error) {
	written := int64(0)
	for i := 0; i < len(b.pages); i++ {
		n, err := w.Write(b.pages[i].buf)
		written += int64(n)
		if err != nil {
			return written, err
		}
	}

	return written, nil
}

func (b *PageBuffer) NewReaderAt(offset int) *PageBufferReader {
	pageIdx, startIdx := b.pageForOffset(offset)

	return &PageBufferReader{
		buf:      b,
		pageIdx:  pageIdx,
		startIdx: startIdx,
	}
}

type PageBufferReader struct {
	buf      *PageBuffer
	pageIdx  int
	startIdx int
}

func (r *PageBufferReader) Read(p []byte) (int, error) {
	pc := len(r.buf.pages)

	read := 0
	for r.pageIdx < pc && read < len(p) {
		cp := r.buf.pages[r.pageIdx]
		endIdx := len(cp.buf)

		n := copy(p[read:], cp.buf[r.startIdx:endIdx])
		read += n
		r.startIdx += n

		if r.startIdx >= cap(cp.buf) {
			r.pageIdx++
			r.startIdx = 0
			continue
		}

		if r.pageIdx == pc-1 {
			break
		}
	}

	if read == 0 {
		return read, io.EOF
	}

	return read, nil
}

func AssertTrue(b bool) {
	if !b {
		log.Fatalf("%+v", errors.New("assert failed"))
	}
}
