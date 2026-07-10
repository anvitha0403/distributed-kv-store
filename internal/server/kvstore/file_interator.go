package kvstore

import (
	"encoding/binary"
	"io"
	"os"

	pb "github.com/anvitha0403/golog/internal/server/api/v1"
	"google.golang.org/protobuf/proto"
)

type FileIterator struct {
	curOffset  int64
	openedfile *os.File
}

func newFileIterator(openedfile *os.File, offset int64) (*FileIterator, error) {
	_, err := openedfile.Seek(offset, io.SeekStart) // Reset file pointer to the given offset (relative to the start of the file)
	if err != nil {
		return nil, err
	}
	fileIterator := &FileIterator{
		curOffset:  offset,
		openedfile: openedfile,
	}
	return fileIterator, nil
}

func (fi *FileIterator) HasNext() bool {
	// Peek length prefix
	buf := make([]byte, 4)
	n, err := fi.openedfile.Read(buf)
	if err != nil || n < 4 {
		return false
	}
	// Reset back so Get can read properly
	fi.openedfile.Seek(fi.curOffset, io.SeekStart)
	return true
}

// Get returns the current record and its starting offset in the file.
func (fi *FileIterator) Get() (*pb.Record, int64, error) {
	startingOffset := fi.curOffset

	lengthBuf := make([]byte, 4)
	if _, err := fi.openedfile.Read(lengthBuf); err != nil {
		return nil, startingOffset, err
	}
	length := binary.BigEndian.Uint32(lengthBuf)

	dataBuf := make([]byte, length)
	if _, err := fi.openedfile.Read(dataBuf); err != nil {
		return nil, startingOffset, err
	}

	rec := &pb.Record{}
	if err := proto.Unmarshal(dataBuf, rec); err != nil {
		return nil, startingOffset, err
	}

	fi.curOffset += int64(4 + length)
	return rec, startingOffset, nil
}
