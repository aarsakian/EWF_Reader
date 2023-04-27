package sections

import (
	"time"

	"github.com/aarsakian/EWF_Reader/ewf/utils"
)

type EWF_Done_Section struct {
}

func (done_section *EWF_Done_Section) Parse(buf []byte) {

	defer utils.TimeTrack(time.Now(), "Parsing")

}

func (done_section *EWF_Done_Section) GetAttr(string) interface{} {
	return ""
}
