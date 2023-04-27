package digest

import (
	"bytes"
	"reflect"
	"time"

	"github.com/aarsakian/EWF_Reader/ewf/utils"

	"encoding/hex"
)

type EWF_Digest_Section struct {
	MD5_value [16]uint8 "MD5 hash of the media data"
	SHA1      [20]uint8 "SHA1 hash of the media data"
	Padding   [4]uint8  "0x00"
	Checksum  uint32    "Adler-32 of all the previous data within the additional digest section"
}

func (digest_section *EWF_Digest_Section) Parse(r *bytes.Reader) {

	defer utils.TimeTrack(time.Now(), "Parsing")

	s := reflect.ValueOf(digest_section).Elem()
	for i := 0; i < s.NumField(); i++ {
		//parse struct attributes
		utils.Parse(r, s.Field(i).Addr().Interface())

	}
}

func (digest_section *EWF_Digest_Section) GetAttr(attr string) interface{} {
	s := reflect.ValueOf(digest_section).Elem() //retrieve since it's a pointer

	sub_s := s.FieldByName(attr)
	if sub_s.IsValid() {

		switch v := sub_s.Interface().(type) {

		case []uint8:
			return hex.EncodeToString(v[:])

		default:
			return "unknown"

		}

	} else {
		return "Not Valid"
	}
}