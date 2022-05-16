# remotefile.go
Use the fs.FIle interface over remote files based on URL.

type RemoteFile struct {
	fs.File
	fs.FileInfo
	io.ReadSeekCloser
	io.ReaderAt

	FileName string
	URL      url.URL
	Length   int64
}
