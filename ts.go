package mpeg2ts

import (
	"encoding/binary"
	"fmt"
)

type (
	Packet struct {
		Magic                      byte
		TransportErrorIndicator    bool
		PayloadUnitStartIndicator  bool
		TransportPriority          bool
		Pid                        uint
		TransportScramblingControl byte
		AdaptationFieldControl     byte
		ContinuityCounter          byte

		Adaptation *AdaptationField
		Payload    []byte
	}

	AdaptationField struct {
		FieldLength                       byte
		DiscontinuityIndicator            bool
		RandomAccessIndicato              bool
		ElementaryStreamPriorityIndicator bool
		PCRFlag                           bool
		OPCRFlag                          bool
		SplicingPointFlag                 bool
		TransportPrivateDataFlag          bool
		AdaptationFieldExtensionFlag      bool

		// T.B.D.
		// Optionals
		//PCR
		//OPCR
		//SpliceCountdown byte
		//TransportPrivateDataLength byte
		//TransportPrivateData	[]byte
		//AdaptationExtension []byte
		//StuffingBytes	byte
	}
)

const PACKET_SIZE = 188

func ParseTsHeader(buffer []byte) (packet *Packet, err error) {
	if PACKET_SIZE != len(buffer) {
		err = fmt.Errorf("Invalid buffer size for packet. (%d passed but explain %d)", len(buffer), PACKET_SIZE)
		return
	}

	packet = &Packet{
		Magic: buffer[0],
		TransportErrorIndicator:   (buffer[1] & 0x80) > 0,
		PayloadUnitStartIndicator: (buffer[1] & 0x40) > 0,
		TransportPriority:         (buffer[1] & 0x20) > 0,
		Pid:                       uint(binary.BigEndian.Uint16([]byte{buffer[1] & 0x1F, buffer[2]})),
		TransportScramblingControl: (buffer[3] & 0xc0) >> 6,
		AdaptationFieldControl:     (buffer[3] & 0x30) >> 4,
		ContinuityCounter:          (buffer[3] & 0x0f),
	}

	bufferHead := 4
	if packet.HaveAdaptation() {
		adaptation, readLength := parseAdaptationField(buffer[bufferHead:PACKET_SIZE])
		packet.Adaptation = adaptation
		bufferHead += readLength
	}

	if packet.HavePayload() {
		packet.Payload = buffer[bufferHead:PACKET_SIZE]
	}
	return packet, err
}

func parseAdaptationField(buffer []byte) (adaptation *AdaptationField, readLength int) {
	adaptation = &AdaptationField{
		FieldLength:                       buffer[0],
		DiscontinuityIndicator:            (buffer[1] & 0x80) > 0,
		RandomAccessIndicato:              (buffer[1] & 0x40) > 0,
		ElementaryStreamPriorityIndicator: (buffer[1] & 0x20) > 0,
		PCRFlag:                      (buffer[1] & 0x10) > 0,
		OPCRFlag:                     (buffer[1] & 0x08) > 0,
		SplicingPointFlag:            (buffer[1] & 0x04) > 0,
		TransportPrivateDataFlag:     (buffer[1] & 0x02) > 0,
		AdaptationFieldExtensionFlag: (buffer[1] & 0x01) > 0,
		// TODO implement
	}
	readLength = int(adaptation.FieldLength) + 1
	return adaptation, readLength
}

func (p Packet) HaveAdaptation() bool {
	return (p.AdaptationFieldControl & 0x02) > 0
}

func (p Packet) HavePayload() bool {
	return (p.AdaptationFieldControl & 0x01) > 0
}
