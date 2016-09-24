package mpeg2ts

import (
	"bufio"
	"fmt"
	"os"

	"mpeg2ts/psi"
)

type (
	Parser struct {
		PayloadBuffers map[uint]([]byte)
	}
)

func NewParser() (p *Parser) {
	p = &Parser{
		PayloadBuffers: map[uint]([]byte){},
	}
	return p
}

func (p *Parser) Parse(tsPath string) error {
	fp, err := os.OpenFile(tsPath, os.O_RDONLY, 0600)
	if nil != err {
		return err
	}
	reader := bufio.NewReader(fp)

	for true {
		// Read one packet
		packetBuffer := make([]byte, PACKET_SIZE)
		size, readErr := reader.Read(packetBuffer)
		if nil != readErr {
			if "EOF" == readErr.Error() {
				break
			}
			return readErr
		}
		if size == 0 {
			// Just read finish when read previous packet.
			break
		}
		if size != PACKET_SIZE {
			// NOTE bufio.Reader's readable size is not full file size(usually, maybe about 4k byte.).
			// So, if read size is not enought array size, we need once retry reading.
			remainBytes := packetBuffer[size:PACKET_SIZE]
			remainSize, err := reader.Read(remainBytes)
			if nil != err {
				return err
			}
			if PACKET_SIZE != remainSize+size {
				return fmt.Errorf("Not enought packet data readed (%dbyte :expect %dbyte)", remainSize+size, PACKET_SIZE)
			}
		}

		packet, parseErr := ParseTsHeader(packetBuffer)
		if nil != parseErr {
			return parseErr
		}

		if packet.PayloadUnitStartIndicator {
			buffer, ok := p.PayloadBuffers[packet.Pid]
			if ok {
				// Parse each type of Pid
				f, ok := psi.FunctionTables[packet.Pid]
				if ok {
					// TODO Manage data
					_, funcErr := f(buffer)
					if nil != funcErr {
						panic(funcErr)
					}
				}
			}

			// Set new packet
			p.PayloadBuffers[packet.Pid] = packet.Payload
		} else {
			buffer, ok := p.PayloadBuffers[packet.Pid]
			if ok {
				p.PayloadBuffers[packet.Pid] = append(buffer, packet.Payload...)
			} else {
				//return fmt.Errorf("0x%X packet is not first, but buffer is not found.", packet.Pid)
			}
		}
	}
	return nil
}
