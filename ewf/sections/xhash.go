package sections

import (
	"encoding/hex"

	"github.com/aarsakian/EWF_Reader/ewf/utils"
)

type XHash struct {
	Type    uint16 // hash type
	Unknown uint16
	Value   []uint8
}

func (xhash *XHash) Parse(buf []byte) {
	utils.Unmarshal(buf, xhash)
	xhash.Value = buf[4:]
}

func (xhash *XHash) GetAttr(attr string) interface{} {
	return hex.EncodeToString(xhash.Value)
}
