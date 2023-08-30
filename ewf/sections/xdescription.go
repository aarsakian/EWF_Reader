package sections

import Utils "github.com/aarsakian/EWF_Reader/ewf/utils"

type XDescription struct {
	Description string
}

func (xdescr *XDescription) Parse(buf []byte) {
	xdescr.Description = Utils.Stringify(buf)
}

func (xdescr XDescription) GetAttr(attr string) interface{} {
	return ""
}
