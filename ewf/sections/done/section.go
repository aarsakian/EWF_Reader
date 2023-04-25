package done

import (
	"bytes"
	"reflect"
	"time"

	"github.com/aarsakian/EWF_Reader/ewf/utils"
)

type EWF_Done_Section struct {
}

func (done_section *EWF_Done_Section) Parse(r *bytes.Reader) {

	defer utils.TimeTrack(time.Now(), "Parsing")

	s := reflect.ValueOf(done_section).Elem()
	for i := 0; i < s.NumField(); i++ {
		//parse struct attributes
		utils.Parse(r, s.Field(i).Addr().Interface())

	}
}

func (done_section *EWF_Done_Section) GetAttr(string) interface{} {
	return ""
}
