package sections

import "github.com/aarsakian/EWF_Reader/ewf/utils"

type XDescription struct {
	Description string
}

func (xdescr *XDescription) Parse(buf []byte) {
	xdescr.Description = utils.Stringify(buf)
}

func (xdescr XDescription) GetAttr(attr string) interface{} {
	return ""
}
