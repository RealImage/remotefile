# RemoteFile.go

`RemoteFile` is a wrapper around a URL that lets you use it as a file, including calls to `Read`, `ReadAt` and `Stat`. It can be constructed with a `FileName`, `URL` and `Length`, implements the following interfaces: 

```go
type RemoteFile struct {
	fs.File
	fs.FileInfo
	io.ReadSeekCloser
	io.ReaderAt

	FileName string
	URL      url.URL
	Length   int64
}
```
