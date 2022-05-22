package remotefile

import (
	"bytes"
	"crypto/rand"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestRemoteFile_Read(t *testing.T) {
	size := 1000000
	halfSize := size / 2
	data, u := buildServer(size)

	rf := &RemoteFile{
		FileName:     "t1",
		URL:          u,
		Length:       int64(size),
		LastModified: time.Now(),
	}
	buf := make([]byte, halfSize)
	n, err := io.ReadFull(rf, buf)
	crashIf(err)
	if n != halfSize {
		t.Error("should have read 5 bytes")
	}
	if !bytes.Equal(buf, data[0:halfSize]) {
		t.Error("first half is wrong")
	}
	n, err = io.ReadFull(rf, buf)
	crashIf(err)
	if n != halfSize {
		t.Error("should have read 5 bytes")
	}
	if !bytes.Equal(buf, data[halfSize:]) {
		t.Error("second half is wrong")
	}
}

func TestRemoteFile_Read2(t *testing.T) {
	size := 10000
	data, u := buildServer(size)
	rf := &RemoteFile{
		FileName:     "t1",
		URL:          u,
		Length:       int64(size),
		LastModified: time.Now(),
	}
	receivedData, err := io.ReadAll(rf)
	crashIf(err)
	if !bytes.Equal(data, receivedData) {
		t.Error("read problems.")
	}
}

func TestRemoteFile_ReadAt(t *testing.T) {
	size := 1000
	data, u := buildServer(size)
	rf := &RemoteFile{
		FileName:     "t1",
		URL:          u,
		Length:       int64(size),
		LastModified: time.Now(),
	}
	buf := make([]byte, 100)
	n, err := rf.ReadAt(buf, 0)
	crashIf(err)
	if n != 100 {
		t.Error("should have read 100 bytes")
	}
	if !bytes.Equal(buf, data[0:100]) {
		t.Error("bytes are wrong")
	}

	n, err = rf.ReadAt(buf, 950)
	if err != io.EOF {
		t.Error("expecting EOF")
	}
	if n != 50 {
		t.Error("should have read 50 bytes")
	}
	if !bytes.Equal(buf[0:n], data[950:]) {
		t.Error("bytes are wrong")
	}
}

func TestRemoteFile_Seek(t *testing.T) {
	size := 1000
	data, u := buildServer(size)
	rf := &RemoteFile{
		FileName:     "t1",
		URL:          u,
		Length:       int64(size),
		LastModified: time.Now(),
	}

	currentOffset, err := rf.Seek(10, io.SeekStart)
	crashIf(err)
	if currentOffset != 10 {
		t.Error("should have set offset to 0")
	}
	assertBytes(t, readByte(rf), data[10:11])

	currentOffset, err = rf.Seek(15, io.SeekCurrent)
	crashIf(err)
	if currentOffset != 26 {
		t.Error("should have set offset to 26 - the readByte call would have read one byte.")
	}

	currentOffset, err = rf.Seek(5, io.SeekEnd)
	crashIf(err)
	if currentOffset != 995 {
		t.Error("should set offset from the end")
	}

	currentOffset, err = rf.Seek(100, io.SeekCurrent)
	if err != ErrOffset {
		t.Error("should have been an error")
	}
	if currentOffset != 995 {
		t.Error("should not have changed the offset on error.")
	}

	currentOffset, err = rf.Seek(1001, io.SeekStart)
	if err != ErrOffset {
		t.Error("should have been an error")
	}
	if currentOffset != 995 {
		t.Error("should not have changed the offset on error.")
	}

	currentOffset, err = rf.Seek(1001, io.SeekEnd)
	if err != ErrOffset {
		t.Error("should have been an error")
	}
	if currentOffset != 995 {
		t.Error("should not have changed the offset on error.")
	}

	currentOffset, err = rf.Seek(1000, io.SeekStart)
	if err != ErrOffset {
		t.Error("should not allow setting offset outside bounds")
	}
}

func assertBytes(t *testing.T, buf []byte, data []byte) {
	if !bytes.Equal(buf, data) {
		t.Error("wrong bytes")
	}
}

func readByte(rf *RemoteFile) []byte {
	buf := make([]byte, 1)
	_, err := rf.Read(buf)
	crashIf(err)
	return buf
}

func buildServer(size int) ([]byte, *url.URL) {
	data := make([]byte, size)
	_, err := io.ReadFull(rand.Reader, data)
	crashIf(err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeContent(w, r, "t1", time.Now(), bytes.NewReader(data))
	}))

	u, err := url.Parse(server.URL)
	crashIf(err)
	return data, u
}

func crashIf(err error) {
	if err != nil {
		panic(err)
	}
}
