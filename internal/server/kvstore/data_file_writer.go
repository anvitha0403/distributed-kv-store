package kvstore

import (
	"bufio"
	"encoding/binary"

	pb "github.com/anvitha0403/golog/internal/server/api/v1"
	"google.golang.org/protobuf/proto"
)

type DataFileWriter struct {
	writer *bufio.Writer
}

func (df *DataFileWriter) Append(rec *pb.Record) (int64, error) {

	// Serialize
	bytes, err := proto.Marshal(rec)
	if err != nil {
		return 0, err
	}

	// Write length prefix
	length := make([]byte, 4)
	binary.BigEndian.PutUint32(length, uint32(len(bytes)))
	if _, err := df.writer.Write(length); err != nil {
		return 0, err
	}

	// Write record
	lengthData, err := df.writer.Write(bytes)
	if err != nil {
		return 0, err
	}

	return int64(lengthData) + 4, nil
}
func (dfw *DataFileWriter) Flush() error {
	return dfw.writer.Flush()
}
