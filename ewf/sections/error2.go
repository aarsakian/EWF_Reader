package sections

import (
	Utils "github.com/aarsakian/EWF_Reader/ewf/utils"
)

type Error2_Entry struct {
	CodeID       uint32
	SectorOffset uint64
	SectorCount  uint32
	Description  []byte
}

type Error2_Section struct {
	Error2_Entries []Error2_Entry
}

func (error2_section *Error2_Section) Parse(data []byte) {
	curOffset := 0

	for curOffset < len(data) {
		error2_entry := new(Error2_Entry)
		err := Utils.Unmarshal(data[curOffset:], error2_entry)
		if err != nil {
			break
		}
		curOffset += 16
		for pos, val := range data[curOffset:] {
			if val == 0 {
				copy(error2_entry.Description, data[curOffset:curOffset+pos])
				curOffset += pos
				break

			}

		}

	}

}

func (error2_section Error2_Section) GetAttr(attr string) interface{} {
	return ""
}

func (error2_entry Error2_Entry) ToString() string {
	return string(error2_entry.Description)
}
