package kvstore

import (
	"bufio"
	"io"
	"os"
)

type FileIterator struct {
	scanner    *bufio.Scanner
	curOffset  int64
	openedfile *os.File
}

func newFileIterator(openedfile *os.File, offset int64) (*FileIterator, error) {
	_, err := openedfile.Seek(offset, io.SeekStart) // Reset file pointer to the given offset (relative to the start of the file)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(openedfile)
	fileIterator := &FileIterator{
		scanner:    scanner,
		curOffset:  offset,
		openedfile: openedfile,
	}
	return fileIterator, nil
}

func (fi *FileIterator) HasNext() bool {
	return fi.scanner.Scan()
}

// Get returns the current record and its starting offset in the file.
func (fi *FileIterator) Get() (record, int64) {
	var data record
	line := fi.scanner.Text()
	data.FromString(line)
	bytesRead := int64(len(line)) + 1          // +1 for the newline character
	fi.curOffset += bytesRead                  // Update the current offset (including newline character)
	startingOffset := fi.curOffset - bytesRead // Calculate the starting offset of the record
	return data, startingOffset
}
