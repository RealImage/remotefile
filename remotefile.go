package remotefile

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"time"
)

var ErrOutOfBounds = errors.New("seek offset out of bounds")

// RemoteFile implements `fs.File`, `fs.FileInfo`, `io.ReadSeekCloser`, `io.ReaderAt`
type RemoteFile struct {
	FileName     string
	URL          *url.URL
	Length       int64
	LastModified time.Time

	offset int64
}

func (rf *RemoteFile) Read(p []byte) (n int, err error) {
	start, end := rf.calcRange(rf.offset, int64(len(p)))

	n, err = rf.send(p, start, end)

	rf.offset += int64(n)
	if rf.offset >= rf.Length-1 {
		err = io.EOF
	}

	return n, err
}

func (rf RemoteFile) ReadAt(p []byte, off int64) (n int, err error) {
	start, end := rf.calcRange(off, int64(len(p)))

	n, err = rf.send(p, start, end)

	if off+end-start >= rf.Length-1 {
		err = io.EOF
	}

	return n, err
}

func (rf *RemoteFile) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekCurrent:
		if rf.offset+offset >= rf.Length {
			return rf.offset, ErrOutOfBounds
		}
		rf.offset += offset

	case io.SeekEnd:
		if rf.Length-offset <= 0 {
			return rf.offset, ErrOutOfBounds
		}
		rf.offset = rf.Length - offset

	default:
		if offset >= rf.Length {
			return rf.offset, ErrOutOfBounds
		}
		rf.offset = offset

	}
	return rf.offset, nil
}

func (rf *RemoteFile) Close() error {
	rf.offset = 0
	return nil
}

func (rf RemoteFile) Stat() (fs.FileInfo, error) {
	return rf, nil
}

func (rf RemoteFile) Name() string {
	return rf.FileName
}

func (rf RemoteFile) Size() int64 {
	return rf.Length
}

func (rf RemoteFile) Mode() fs.FileMode {
	return fs.ModeSymlink
}

func (rf RemoteFile) ModTime() time.Time {
	return rf.LastModified
}

func (rf RemoteFile) IsDir() bool {
	return false
}

func (rf RemoteFile) Sys() any {
	return nil
}

func (rf *RemoteFile) calcRange(offset int64, length int64) (int64, int64) {
	start := offset
	end := offset + length
	if end >= rf.Length-1 {
		end = rf.Length - 1
	}
	return start, end
}

func (rf *RemoteFile) send(p []byte, start int64, end int64) (n int, err error) {
	req, err := http.NewRequest("GET", rf.URL.String(), nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%v-%v", start, end))

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer r.Body.Close()

	n, err = io.ReadAtLeast(r.Body, p, int(end-start))

	return n, err
}
