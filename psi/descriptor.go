package psi

import (
//"mpeg2ts/character"
)

type (
	DescriptorCommon struct {
		Tag    byte
		Length byte
	}

	// Descriptor Tag Number : 0x4D
	EventDescriptor struct {
		DescriptorCommon
		LanguageCode []byte
		Article
	}

	// Descriptor Tag Number : 0x4E
	ExtendEventDescriptor struct {
		DescriptorCommon
		Number        byte
		LastNumber    byte
		LanguageCode  []byte
		ArticleLength byte

		Articles []Article

		ExtendLength     byte
		ExtendDescriptor []byte
	}

	// Parts of descriptors
	Article struct {
		NameLength           byte
		Name                 []byte
		NameDescriptorLength byte
		NameDescriptor       []byte
	}
)

const (
	EventTag       = 0x4D
	ExtendEventTag = 0x4E

	LANGUAGE_JPN = "jpn"
)

func ParseDescriptor(buffer []byte) (interface{}, uint) {
	switch buffer[0] {
	case EventTag:
		return ParseEventDescriptor(buffer)
	case ExtendEventTag:
		return ParseExtendEventDescriptor(buffer)
	default:
		common := ParseDescriptorCommon(buffer)
		return nil, uint(common.Length + 2)
	}
}

func ParseEventDescriptor(buffer []byte) (EventDescriptor, uint) {
	ed := EventDescriptor{
		DescriptorCommon: ParseDescriptorCommon(buffer),
		LanguageCode:     buffer[2:5],
	}
	size := uint(ed.Length + 2)
	ed.Article, _ = ParseArticle(buffer[5:size])
	return ed, size
}

func ParseExtendEventDescriptor(buffer []byte) (ExtendEventDescriptor, uint) {
	eed := ExtendEventDescriptor{
		DescriptorCommon: ParseDescriptorCommon(buffer),
		Number:           buffer[2] & 0xF0 >> 4,
		LastNumber:       buffer[2] & 0x0F,
		LanguageCode:     buffer[3:6],
		ArticleLength:    buffer[7],
	}
	size := uint(eed.Length + 2)
	return eed, size
}

func ParseArticle(buffer []byte) (Article, uint) {
	article := Article{
		NameLength: buffer[0],
	}
	article.Name = buffer[1 : article.NameLength+1]
	article.NameDescriptorLength = buffer[article.NameLength+1]
	tail := article.NameDescriptorLength + article.NameLength + 2
	article.NameDescriptor = buffer[article.NameLength+2 : tail]
	return article, uint(tail)
}

func ParseDescriptorCommon(buffer []byte) DescriptorCommon {
	common := DescriptorCommon{
		Tag:    buffer[0],
		Length: buffer[1],
	}
	return common
}
