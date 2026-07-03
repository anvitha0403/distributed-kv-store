package kvstore

import (
	"bufio"
	"fmt"
)

type DataFileWriter struct {
	writer *bufio.Writer
}

func (dfw *DataFileWriter) Append(data record) (int64, error) {
	record := data.String()
	var bytes int
	var err error
	if bytes, err = fmt.Fprintln(dfw.writer, record); err != nil {
		return int64(bytes), err
	}
	return int64(bytes), nil
}
 
func (dfw *DataFileWriter) Flush() error {
	return dfw.writer.Flush()
}
