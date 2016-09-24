package psi

import (
	"encoding/binary"
	"fmt"
)

var FunctionTables = map[uint]func(buffer []byte) (interface{}, error){}

type (
	Common struct {
		TableId                byte
		SectionSyntaxIndicator bool
		reservedFlag           bool
		reserved               byte
		SectionLength          uint
	}
)

const COMMON_FILED_LENGTH = 3

func init() {
	FunctionTables = map[uint](func(buffer []byte) (interface{}, error)){
		0x00: ParsePat,
		0x12: ParseEit,
		0x26: ParseEit,
		0x27: ParseEit,
	}
}

func ParseCommon(buffer []byte, common *Common) (err error) {
	if COMMON_FILED_LENGTH != len(buffer) {
		err = fmt.Errorf("Invalid buffer size '%d' for PSI Header.", len(buffer))
		return err
	}

	common.TableId = buffer[0]
	common.SectionSyntaxIndicator = ((buffer[1] & 0x80) >> 7) == 1
	common.reservedFlag = ((buffer[1] & 0x40) >> 6) == 1
	common.reserved = (buffer[1] & 0x30) >> 4
	common.SectionLength = uint(binary.BigEndian.Uint16([]byte{buffer[1] & 0x0F, buffer[2]}))
	return nil
}
