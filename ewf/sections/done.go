package sections

import (
	"time"

	Utils "github.com/aarsakian/EWF_Reader/ewf/utils"
)

type EWF_Done_Section struct {
}

func (done_section *EWF_Done_Section) Parse(buf []byte) {

	defer Utils.TimeTrack(time.Now(), "Parsing")

}

func (done_section *EWF_Done_Section) GetAttr(string) interface{} {
	return ""
}
