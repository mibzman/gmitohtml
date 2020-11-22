package gmitohtml

import (
	"io"
	"net/http"
	"os"
	"time"
)

var modTime = time.Now()

type inMemoryFS map[string]http.File

func (fs inMemoryFS) Open(name string) (http.File, error) {
	if f, ok := fs[name]; ok {
		f.Seek(0, io.SeekStart)
		return f, nil
	}
	panic("No file")
}

type inMemoryFile struct {
	at   int64
	Name string
	data []byte
	fs   inMemoryFS
}

func loadFile(name string, val string, fs inMemoryFS) *inMemoryFile {
	return &inMemoryFile{at: 0,
		Name: name,
		data: []byte(val),
		fs:   fs}
}

func (f *inMemoryFile) Close() error {
	return nil
}
func (f *inMemoryFile) Stat() (os.FileInfo, error) {
	return &inMemoryFileInfo{f}, nil
}
func (f *inMemoryFile) Readdir(count int) ([]os.FileInfo, error) {
	res := make([]os.FileInfo, len(f.fs))
	i := 0
	for _, file := range f.fs {
		res[i], _ = file.Stat()
		i++
	}
	return res, nil
}
func (f *inMemoryFile) Read(b []byte) (int, error) {
	i := 0
	for f.at < int64(len(f.data)) && i < len(b) {
		b[i] = f.data[f.at]
		i++
		f.at++
	}
	return i, nil
}
func (f *inMemoryFile) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case 0:
		f.at = offset
	case 1:
		f.at += offset
	case 2:
		f.at = int64(len(f.data)) + offset
	}
	return f.at, nil
}

type inMemoryFileInfo struct {
	file *inMemoryFile
}

func (s *inMemoryFileInfo) Name() string       { return s.file.Name }
func (s *inMemoryFileInfo) Size() int64        { return int64(len(s.file.data)) }
func (s *inMemoryFileInfo) Mode() os.FileMode  { return os.ModeTemporary }
func (s *inMemoryFileInfo) ModTime() time.Time { return modTime }
func (s *inMemoryFileInfo) IsDir() bool        { return false }
func (s *inMemoryFileInfo) Sys() interface{}   { return nil }
