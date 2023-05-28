package sections

import (
	"reflect"
	"time"

	"github.com/aarsakian/EWF_Reader/ewf/utils"

	"encoding/hex"
)

type EWF_Hash_Section struct {
	MD5      [16]uint8
	Unknown  uint8
	Checksum uint32 //adler32
}

func (hash_section *EWF_Hash_Section) Parse(buf []byte) {

	defer utils.TimeTrack(time.Now(), "Parsing")

	utils.Unmarshal(buf, hash_section)
}

func (hash_section EWF_Hash_Section) GetAttr(attr string) interface{} {
	s := reflect.ValueOf(hash_section) //retrieve since it's a pointer

	sub_s := s.FieldByName(attr)
	if sub_s.IsValid() {

		switch v := sub_s.Interface().(type) {

		case uint32, int8:
			return v

		case [16]uint8:
			return hex.EncodeToString(v[:])

		default:
			return "unknown"

		}

	} else {
		return "Not Valid"
	}
}
