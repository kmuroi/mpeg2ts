package psi

import (
	"encoding/binary"
)

type (
	PATField struct {
		Common
		TransportStreamId    uint
		reserved2            byte
		VersionNumber        byte
		CurrentNextIndicator bool
		SectionNumber        byte
		LastSectionNumber    byte

		ProgramAssociations []ProgramAssociationField

		Crc []byte
	}

	ProgramAssociationField struct {
		ProgramNumber uint
		Reserve       byte
		Pid           uint //NetworkPID or ProgramMapPID
	}
)

const PAT_FIELD_LENGTH = 5
const PROGRAM_ASSOCIATION_LENGTH = 4

func ParsePat(buffer []byte) (interface{}, error) {
	pointerField := buffer[0]
	commonTail := (pointerField + 1) + COMMON_FILED_LENGTH // 1 is pointerField size

	pat := &PATField{}
	err := ParseCommon(buffer[(pointerField+1):commonTail], &pat.Common)
	if nil != err {
		return nil, err
	}

	patBuffer := buffer[commonTail:]
	pat.TransportStreamId = uint(binary.BigEndian.Uint16(patBuffer[0:2]))
	pat.reserved2 = patBuffer[2] & 0xC0 >> 6
	pat.VersionNumber = patBuffer[2] & 0x3E >> 1
	pat.CurrentNextIndicator = patBuffer[2]&0x01 > 0
	pat.SectionNumber = patBuffer[3]
	pat.LastSectionNumber = patBuffer[4]

	// TODO check PA Field is over buffer
	paNumber := (pat.SectionLength - (PAT_FIELD_LENGTH + 4)) / PROGRAM_ASSOCIATION_LENGTH
	pat.ProgramAssociations = make([]ProgramAssociationField, paNumber)
	for idx := uint(0); idx < paNumber; idx++ {
		headIndex := 5 + PROGRAM_ASSOCIATION_LENGTH*idx
		pa := &pat.ProgramAssociations[idx]
		pa.ProgramNumber = uint(binary.BigEndian.Uint16(patBuffer[headIndex : headIndex+2]))
		pa.Reserve = patBuffer[headIndex+2] & 0xE0 >> 5
		pa.Pid = uint(binary.BigEndian.Uint16([]byte{patBuffer[headIndex+2] & 0x1F, patBuffer[headIndex+3]}))

		// Add PMT's pid to function table
		_, ok := FunctionTables[pa.Pid]
		if !ok {
			FunctionTables[pa.Pid] = ParsePmt
		}
	}

	crcHead := 5 + PROGRAM_ASSOCIATION_LENGTH*paNumber
	pat.Crc = patBuffer[crcHead : crcHead+4]
	// TODO crc check

	return pat, nil
}
