package character

import (
	"bytes"
	"io/ioutil"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

const (
	// ControlSets
	LS1 = 0x0E
	LS0 = 0x0F
	SS2 = 0x19
	ESC = 0x1B
	SS3 = 0x1D
	SP  = 0x20

	LS2  = 0x6E
	LS3  = 0x6F
	LS1R = 0x7E
	LS2R = 0x7D
	LS3R = 0x7C

	G0    = 0x28
	G1    = 0x29
	G2    = 0x2A
	G3    = 0x2B
	DBYTE = 0x24
	DRCS  = 0x20

	// Final Byte of GraphicSet
	KANJI                = 0x42
	ALNUM                = 0x4A
	HIRAGANA             = 0x30
	KATAKANA             = 0x31
	MOSAIC_A             = 0x32
	MOSAIC_B             = 0x33
	MOSAIC_C             = 0x34
	MOSAIC_D             = 0x35
	PROPOSIONAL_ALNUM    = 0x36
	PROPOSIONAL_HIRAGANA = 0x37
	PROPOSIONAL_KATAKANA = 0x38
	JIS_X0201_KATAKANA   = 0x49
	JIS_COMPATI_KANJI_1  = 0x39
	JIS_COMPATI_KANJI_2  = 0x3A
	ADDITIONAL_SYMBOL    = 0x3B
)

const (
	WORD_CONTROL = iota
	WORD_SPECIAL_SYMBOL
	WORD_GL_SYMBOL
	WORD_GR_SYMBOL
)

type (
	// NOTE EBit means 8bit
	EBitCharacterDecorder struct {
		g0 byte
		g1 byte
		g2 byte
		g3 byte
		gl GraphicSetElement
		gr GraphicSetElement
	}

	GraphicSetElement struct {
		set       *byte
		singleSet *byte
		Single    bool
	}
)

// TODO import set library and use
var finalBytes = map[byte]struct{}{
	KANJI:                struct{}{},
	ALNUM:                struct{}{},
	HIRAGANA:             struct{}{},
	KATAKANA:             struct{}{},
	MOSAIC_A:             struct{}{},
	MOSAIC_B:             struct{}{},
	MOSAIC_C:             struct{}{},
	MOSAIC_D:             struct{}{},
	PROPOSIONAL_ALNUM:    struct{}{},
	PROPOSIONAL_HIRAGANA: struct{}{},
	PROPOSIONAL_KATAKANA: struct{}{},
	JIS_X0201_KATAKANA:   struct{}{},
	JIS_COMPATI_KANJI_1:  struct{}{},
	JIS_COMPATI_KANJI_2:  struct{}{},
	ADDITIONAL_SYMBOL:    struct{}{},
}

// ISO-2022-JP Escape Bytes
var ESC_KANJI = []byte{0x1B, 0x24, 0x42}
var ESC_ASCII = []byte{0x1B, 0x28, 0x42}

func NewEBitCharacterDecorder() (decorder EBitCharacterDecorder) {
	// REF ARIB STD-B24 Chapter8 Table8-2 Initial status
	decorder = EBitCharacterDecorder{
		g0: KANJI,
		g1: ALNUM,
		g2: HIRAGANA,
		g3: KATAKANA,
	}
	decorder.gl.set = &(decorder.g0)
	decorder.gr.set = &(decorder.g2)
	return decorder
}

// Decode 8bit-character.
// NOTE 8bit-character define below ARIB document.
// 	    8bit-character is based on ISO/IEC2022
// REF ARIB STD-B10 第2部 付録A
// REF ARIB STD-B24 第一編 第2部
// REF https://ja.wikipedia.org/wiki/ARIB%E5%A4%96%E5%AD%97
func (decorder EBitCharacterDecorder) Decode(buffer []byte) string {
	decodedStr := ""
	for idx := 0; idx < len(buffer); idx++ {
		word := buffer[idx]
		switch getWordType(word) {
		case WORD_CONTROL:
			readSize := decorder.control(buffer[idx:])
			idx += (readSize - 1)
		case WORD_GL_SYMBOL:
			str, readByte := decorder.decodeGl(buffer[idx:])
			idx += (readByte - 1)
			decodedStr += str
		case WORD_GR_SYMBOL:
			str, readByte := decorder.decodeGr(buffer[idx:])
			idx += (readByte - 1)
			decodedStr += str
		}
	}

	return decodedStr
}

func (decorder *EBitCharacterDecorder) control(buffer []byte) int {
	word := buffer[0]
	readByte := 1
	switch word {
	case LS1:
		decorder.gl.set = &(decorder.g0)
		decorder.gl.Single = false
	case LS0:
		decorder.gl.set = &(decorder.g1)
		decorder.gl.Single = false
	case SS2:
		decorder.gl.singleSet = &(decorder.g2)
		decorder.gl.Single = true
	case ESC:
		readByte += decorder.escControl(word, buffer)
	case SS3:
		decorder.gl.set = &(decorder.g3)
		decorder.gl.Single = true
	case SP:
		// T.B.D.
	default:
		// T.B.D.
	}

	return readByte
}

func (decorder *EBitCharacterDecorder) escControl(word byte, buffer []byte) int {
	nextWord := buffer[1]
	readByte := 1 // read size in this function
	switch nextWord {
	case LS2:
		decorder.gl.set = &(decorder.g2)
		decorder.gl.Single = false
	case LS3:
		decorder.gl.set = &(decorder.g3)
		decorder.gl.Single = false
	case LS1R:
		decorder.gr.set = &(decorder.g1)
		decorder.gr.Single = false
	case LS2R:
		decorder.gr.set = &(decorder.g2)
		decorder.gr.Single = false
	case LS3R:
		decorder.gr.set = &(decorder.g3)
		decorder.gr.Single = false
	// Graphic Set
	case G0, G1, G2, G3:
		thirdByte := buffer[2]
		_, ok := finalBytes[thirdByte]
		if ok {
			decorder.setGraphicSet(nextWord, thirdByte)
			readByte = 2
		} else if DRCS == thirdByte {
			// T.B.D.
			readByte = 3
		}
	case DBYTE:
		thirdByte := buffer[2]
		_, ok := finalBytes[thirdByte]
		if ok {
			decorder.setGraphicSet(G0, thirdByte)
			readByte = 2
		} else if thirdByte == G1 || thirdByte == G2 || thirdByte == G3 {
			decorder.setGraphicSet(thirdByte, buffer[3])
			readByte = 3
		} else if DRCS == thirdByte {
			// T.B.D.
			readByte = 4
		}
	default:
	}

	return readByte
}

func (decorder *EBitCharacterDecorder) setGraphicSet(positionByte byte, finalByte byte) {
	switch positionByte {
	case G0:
		decorder.g0 = finalByte
	case G1:
		decorder.g1 = finalByte
	case G2:
		decorder.g2 = finalByte
	case G3:
		decorder.g3 = finalByte
	}
}

func (decorder *EBitCharacterDecorder) decodeGl(buffer []byte) (string, int) {
	if WORD_GL_SYMBOL != getWordType(buffer[0]) {
		panic("Not GL buffer.")
	}

	// check gl area
	for idx, v := range buffer {
		switch getWordType(v) {
		case WORD_CONTROL, WORD_GR_SYMBOL:
			return decorder.decode(*decorder.gl.set, buffer[:idx]), idx
		case WORD_SPECIAL_SYMBOL:
			// T.B.D.
			return decorder.decode(*decorder.gl.set, buffer[:idx]), idx
		}
	}
	return decorder.decode(*decorder.gl.set, buffer), len(buffer)
}

func (decorder *EBitCharacterDecorder) decodeGr(buffer []byte) (string, int) {
	if WORD_GR_SYMBOL != getWordType(buffer[0]) {
		panic("Not GL buffer.")
	}

	grBuffer := []byte{}
	// check gr area
	for idx, v := range buffer {
		switch getWordType(v) {
		case WORD_CONTROL, WORD_GL_SYMBOL:
			return decorder.decode(*decorder.gr.set, grBuffer), idx
		case WORD_SPECIAL_SYMBOL:
			// T.B.D.
			return decorder.decode(*decorder.gr.set, grBuffer), idx
		}
		grBuffer = append(grBuffer, v&0x7F)
	}
	return decorder.decode(*decorder.gr.set, grBuffer), len(grBuffer)
}

func (decorder *EBitCharacterDecorder) decode(decodeType byte, buffer []byte) string {
	switch decodeType {
	case KANJI:
		return decodeKanji(append(ESC_KANJI, buffer...))
	case ALNUM:
		return decodeKanji(append(ESC_ASCII, buffer...))
	case HIRAGANA:
		return decodeHiragana(buffer)
	case KATAKANA:
		return decodeKatakana(buffer)
	}
	return ""
}

func decodeKanji(buffer []byte) string {
	reader := transform.NewReader(bytes.NewBuffer(buffer), japanese.ISO2022JP.NewDecoder())
	decodedBytes, _ := ioutil.ReadAll(reader)
	str := string(decodedBytes)
	return str
}

// NOTE adjust to ISO-2022-JP's HIRAGANA code
// JIS 0x24
func decodeHiragana(buffer []byte) string {
	jisBuffer := make([]byte, len(buffer)*2)
	for idx, word := range buffer {
		jisBuffer[idx*2] = byte(0x24)
		jisBuffer[idx*2+1] = word
	}
	return decodeKanji(append(ESC_KANJI, jisBuffer...))
}

// NOTE adjust to ISO-2022-JP's KATAKANA code
// JIS 0x25
func decodeKatakana(buffer []byte) string {
	jisBuffer := make([]byte, len(buffer)*2)
	for idx, word := range buffer {
		jisBuffer[idx*2] = byte(0x25)
		jisBuffer[idx*2+1] = word
	}

	return decodeKanji(append(ESC_KANJI, jisBuffer...))
}

func getWordType(word byte) byte {
	tmp := word & 0x7F
	if 0x00 <= tmp && 0x1F >= tmp {
		return WORD_CONTROL
	} else if 0x20 == tmp || 0x7F == tmp {
		return WORD_SPECIAL_SYMBOL
	} else {
		if (word & 0x80) == 0 {
			return WORD_GL_SYMBOL
		} else {
			return WORD_GR_SYMBOL
		}
	}
}
