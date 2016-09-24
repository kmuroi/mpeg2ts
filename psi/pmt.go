package psi

import (
	"encoding/binary"
)

type (
	PMTField struct {
		Common
		ProgramNumber        uint16
		reserved2            byte
		VersionNumber        byte
		CurrentNextIndicator bool
		SectionNumber        byte
		LastSectionNumber    byte
		reserved3            byte
		PCRPid               uint16
		reserved4            byte
		ProgramInfoLength    uint16

		Descriptors []PMTDescriptor
		Streams     []PMTStream

		Crc []byte
	}

	PMTDescriptor struct {
		Tag    byte
		Length byte
		Data   []byte
	}

	PMTStream struct {
		StreamType    byte
		reserved      byte
		ElementaryPid uint16
		reserved2     byte
		ESInfoLength  uint16
		ESInfo        []byte
	}
)

const PMT_FIELD_LENGTH = 9

func ParsePmt(buffer []byte) (interface{}, error) {
	pointerField := buffer[0]
	commonTail := (pointerField + 1) + COMMON_FILED_LENGTH // 1 is pointerField size

	pmt := &PMTField{}
	err := ParseCommon(buffer[(pointerField+1):commonTail], &pmt.Common)
	if nil != err {
		return nil, err
	}

	pmtBuffer := buffer[commonTail:]
	pmt.ProgramNumber = binary.BigEndian.Uint16(pmtBuffer[0:2])
	pmt.reserved2 = pmtBuffer[2] & 0xC0 >> 6
	pmt.VersionNumber = pmtBuffer[2] & 0x3E >> 1
	pmt.CurrentNextIndicator = pmtBuffer[2]&0x01 > 0
	pmt.SectionNumber = pmtBuffer[3]
	pmt.LastSectionNumber = pmtBuffer[4]
	pmt.reserved3 = pmtBuffer[5] & 0xE0 >> 5
	pmt.PCRPid = binary.BigEndian.Uint16([]byte{pmtBuffer[5] & 0x1F, pmtBuffer[6]})
	pmt.reserved4 = pmtBuffer[7] & 0xF0 >> 4
	pmt.ProgramInfoLength = binary.BigEndian.Uint16([]byte{pmtBuffer[7] & 0x0F, pmtBuffer[8]})

	descriptorTail := PMT_FIELD_LENGTH + pmt.ProgramInfoLength
	descriptorBuffer := pmtBuffer[PMT_FIELD_LENGTH:descriptorTail]
	idx := 0
	for true {
		descriptor := PMTDescriptor{
			Tag:    descriptorBuffer[idx],
			Length: descriptorBuffer[idx+1],
		}
		fullSize := idx + 2 + int(descriptor.Length)
		descriptor.Data = descriptorBuffer[idx+2 : fullSize]
		idx += fullSize
		pmt.Descriptors = append(pmt.Descriptors, descriptor)
		if len(descriptorBuffer) <= idx {
			break
		}
	}

	streamTail := pmt.SectionLength - 4
	// TODO maybe wrong, so need fix
	//streamSize := int(streamTail - uint(descriptorTail))
	//streamBuffer := pmtBuffer[descriptorTail:streamTail]
	//idx = 0
	//for true {
	//	stream := PMTStream{
	//		StreamType:    streamBuffer[idx],
	//		reserved:      streamBuffer[idx+1] & 0xC0 >> 5,
	//		ElementaryPid: binary.BigEndian.Uint16([]byte{streamBuffer[idx+1] & 0x1F, streamBuffer[idx+2]}),
	//		reserved2:     streamBuffer[idx+3] & 0xF0 >> 4,
	//		ESInfoLength:  binary.BigEndian.Uint16([]byte{streamBuffer[idx+3] & 0x0F, streamBuffer[idx+4]}),
	//	}
	//	stream.ESInfo = pmtBuffer[idx+5 : idx+5+int(stream.ESInfoLength)]
	//	idx += (5 + int(stream.ESInfoLength))

	//	if streamSize <= idx {
	//		break
	//	}
	//}

	// TODO crc check
	pmt.Crc = pmtBuffer[streamTail:pmt.SectionLength]
	return pmt, nil
}
