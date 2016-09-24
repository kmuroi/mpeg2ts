package psi

import (
	"encoding/binary"
	"fmt"
	"math"
	"time"
)

type (
	EITField struct {
		Common
		ServiceId                uint
		reserved2                byte
		Version                  byte
		NextIndicator            bool
		SectionNumber            byte
		LastSectionNumber        byte
		TransportStreamId        uint
		OriginalNetworkId        uint
		SegmentLastSectionNumber byte
		LastTableId              byte

		Events []EITEvent

		Crc []byte
	}

	EITEvent struct {
		EventId               uint16
		StartTime             time.Time
		Duration              time.Duration
		RunningStatus         byte
		FreeCAMode            bool
		DescriptorsLoopLength uint16

		Descriptors []interface{}
	}

	EITDescriptor struct {
		Tag    byte
		Length byte
		Data   []byte
	}
)

const (
	TIME_LAYOUT = "2006/01/02 15:04:05 MST"
)

var EIT_FIELD_LENGTH = 11
var EIT_EVENT_FIELD_LENGTH = 12

func ParseEit(buffer []byte) (interface{}, error) {
	pointerField := buffer[0]
	commonTail := (pointerField + 1) + COMMON_FILED_LENGTH // 1 is pointerField size

	eit := &EITField{}
	err := ParseCommon(buffer[(pointerField+1):commonTail], &eit.Common)
	if nil != err {
		return nil, err
	}

	eitBuffer := buffer[commonTail:]
	eit.ServiceId = uint(binary.BigEndian.Uint16(eitBuffer[0:2]))
	eit.reserved2 = eitBuffer[2] & 0xC0 >> 6
	eit.Version = eitBuffer[2] & 0x3E >> 1
	eit.NextIndicator = (eitBuffer[2] & 0x01) > 0
	eit.SectionNumber = eitBuffer[3]
	eit.LastSectionNumber = eitBuffer[4]
	eit.TransportStreamId = uint(binary.BigEndian.Uint16(eitBuffer[5:7]))
	eit.OriginalNetworkId = uint(binary.BigEndian.Uint16(eitBuffer[7:9]))
	eit.SegmentLastSectionNumber = eitBuffer[9]
	eit.LastTableId = eitBuffer[10]

	eventBuffer := eitBuffer[EIT_FIELD_LENGTH : eit.SectionLength-4]
	for idx := uint(0); uint(len(eventBuffer)) > idx; {
		event := EITEvent{
			EventId:               binary.BigEndian.Uint16(eventBuffer[idx : idx+2]),
			StartTime:             decodeTime(eventBuffer[idx+2 : idx+7]),
			Duration:              decodeDuration(eventBuffer[idx+7 : idx+10]),
			RunningStatus:         eventBuffer[idx+10] & 0xE0 >> 5,
			FreeCAMode:            eventBuffer[idx+10]&0x10 > 0,
			DescriptorsLoopLength: binary.BigEndian.Uint16([]byte{eventBuffer[idx+10] & 0x0F, eventBuffer[idx+11]}),
		}

		descriptorHead := idx + uint(EIT_EVENT_FIELD_LENGTH)
		descriptorTail := descriptorHead + uint(event.DescriptorsLoopLength)
		descriptorBuffer := eventBuffer[descriptorHead:descriptorTail]
		for readSize := uint(0); uint(len(descriptorBuffer)) > readSize; {
			// TODO store each Descriptors
			_, descriptorSize := ParseDescriptor(descriptorBuffer[readSize:])
			readSize += descriptorSize
		}
		idx = descriptorTail
	}

	eit.Crc = eitBuffer[eit.SectionLength-4 : eit.SectionLength]
	// TODO check CRC

	return eit, nil
}

func decodeTime(buffer []byte) time.Time {
	if 5 != len(buffer) {
		panic("")
	}

	// MJD to JTC
	mjd := binary.BigEndian.Uint32(append([]byte{0x0, 0x0}, buffer[:2]...))
	mjdF := float64(mjd)
	tmpY := math.Trunc((mjdF - 15078.2) / 365.25)
	tmp := tmpY * 365.25
	tmpM := math.Trunc(mjdF-14956.1-math.Trunc(tmp)) / 30.6001
	day := mjd - 14956 - uint32(tmp) - uint32(tmpM*30.6001)
	k := uint32(0)
	if tmpM > 13 {
		k = 1
	}
	year := 1900 + uint32(tmpY) + k
	month := uint32(tmpM) - 1 - k*12

	// TODO enable to change location
	hour, minute, second := decodeHMS(buffer[2:5])
	timeStr := fmt.Sprintf("%d/%02d/%02d %02d:%02d:%02d JST",
		year,
		month,
		day,
		hour,
		minute,
		second)
	decoded, err := time.Parse(TIME_LAYOUT, timeStr)
	if nil != err {
		panic(err)
	}
	return decoded
}

func decodeDuration(buffer []byte) time.Duration {
	hour, minute, second := decodeHMS(buffer)
	duration, err := time.ParseDuration(fmt.Sprintf("%dh%dm%ds", hour, minute, second))
	if nil != err {
		panic(err)
	}
	return duration
}

func decodeHMS(buffer []byte) (hour int, minute int, second int) {
	if 3 != len(buffer) {
		panic("")
	}

	// unknown duration fill in 1 all bits
	if 0xFF == buffer[0] && 0xFF == buffer[1] && 0xFF == buffer[2] {
		return 0, 0, 0
	}

	hour = int(buffer[0]&0xF0>>4*10 + buffer[0]&0x0F)
	minute = int(buffer[1]&0xF0>>4*10 + buffer[1]&0x0F)
	second = int(buffer[2]&0xF0>>4*10 + buffer[2]&0x0F)
	return hour, minute, second
}
