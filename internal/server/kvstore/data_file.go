package kvstore

import (
	"bufio"
	"errors"
	"io"
	"os"
	"path/filepath"
)

const (
	PRIMARY_FILENAME = "my.db"
)

type DataFile struct {
	dir               string
	fullpath          string
	file              *os.File
	bytesWrittenSoFar int64 // Track the total bytes written so far
}

// NewDataFile creates a new DataFile instance by opening or creating the primary data file
// at the specified directory path. It returns a pointer to the DataFile and an error if
// the file cannot be opened or created.
//
// Parameters:
//
//	path - The directory path where the data file should be located.
//
// Returns:
//
//	*DataFile - Pointer to the created DataFile instance.
//	error     - Error encountered during file opening or creation, or nil if successful.
func NewDataFile(path string) (*DataFile, error) {
	fullPath := filepath.Join(path, PRIMARY_FILENAME)

	f, err := os.OpenFile(fullPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	return &DataFile{
		dir:      path,
		fullpath: fullPath,
		file:     f,
	}, nil
}

// GetIterator returns a new FileIterator starting at the specified offset within the DataFile.
// It provides sequential access to the file's contents from the given position.
// If the iterator cannot be created, an error is returned.
func (df *DataFile) GetIterator(offset int64) (*FileIterator, error) {
	return newFileIterator(df.file, offset)
}

// Append writes the record to the file, flushes, and returns the starting byte offset.
func (df *DataFile) Append(data record) (int64, error) {
	writer, err := df.Writer()
	if err != nil {
		return 0, err
	}
	bytesWritten, err := writer.Append(data)
	if err != nil {
		return 0, err
	}
	err = writer.Flush()
	if err != nil {
		return 0, err
	}
	startingOffset := df.bytesWrittenSoFar
	df.bytesWrittenSoFar += bytesWritten
	return startingOffset, nil
}

// Writer returns a new DataFileWriter instance associated with the DataFile.
// The DataFileWriter uses a buffered writer for efficient writing to the underlying file.
// It returns the DataFileWriter and any error encountered during creation.
func (df *DataFile) Writer() (*DataFileWriter, error) {
	_, err := df.file.Seek(0, io.SeekEnd) // Ensure the file pointer is at the end of the file before writing new records
	if err != nil {
		return nil, err
	}
	return &DataFileWriter{
		writer: bufio.NewWriter(df.file),
	}, nil
}

// Close closes the underlying file associated with the DataFile, releasing any
// resources held by it. It returns any error produced by the underlying file's
// Close operation. The DataFile should not be used after Close has been called.
func (df *DataFile) Close() error {
	return df.file.Close()
}

// ReadRecordAt reads a single record (one line) starting at the given byte offset,
// parses it with record.FromString, and returns it. The file pointer is restored to
// the start. Returns an error if seeking, scanning, parsing fails, or no record is found.
func (df *DataFile) ReadRecordAt(offset int64) (*record, error) {
	_, err := df.file.Seek(int64(offset), io.SeekStart)
	if err != nil {
		return nil, err
	}
	defer df.file.Seek(0, io.SeekStart) // Reset the file pointer to the start after reading

	scanner := bufio.NewScanner(df.file)
	if scanner.Scan() {
		line := scanner.Text()
		rec := new(record)
		rec.FromString(line)
		return rec, nil
	}
	if scanner.Err() == nil { // If no error but no lines were read, return nil
		return nil, errors.New("no record found at the specified offset")
	}
	return nil, scanner.Err()
}
