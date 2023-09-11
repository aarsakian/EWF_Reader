package sections

import (
	"bytes"
	"encoding/hex"
	"log"

	Utils "github.com/aarsakian/EWF_Reader/ewf/utils"
)

type EWF_Header2_Section struct {
	NofCategories     string
	CategoryName      string
	AcquiredMediaInfo map[string]string
}

func (ewf_h2_section *EWF_Header2_Section) Parse(buf []byte) {
	//0x09 tab 0x0a new line delimiter
	//function to parse header2 section attributes
	//to do take into account endianess

	val := Utils.Decompress(buf)

	//defer Utils.TimeTrack(time.Now(), "Parsing Header2 section")
	line_del, _ := hex.DecodeString("000a")
	tab_del, err := hex.DecodeString("0009")
	if err != nil {
		log.Fatal(err)
	}

	var identifiers []string
	var values []string
	var time_ids []int // save ids of m & u
	for line_number, line := range bytes.Split(val, line_del) {
		for id_num, attr := range bytes.Split(line, tab_del) {
			attr = Utils.RemoveNulls(attr)
			if line_number == 0 {
				ewf_h2_section.NofCategories = string(attr[0])

			} else if line_number == 1 {
				ewf_h2_section.CategoryName = string(attr)
			} else if line_number == 2 {
				identifier := AcquiredMediaIdentifiers[string(attr)]
				if identifier == "Acquired Date" || identifier == "System Date" {
					time_ids = append(time_ids, id_num)
				}
				identifiers = append(identifiers, identifier)
			} else if line_number == 3 {
				if len(time_ids) == 2 && (id_num == time_ids[0] || id_num == time_ids[1]) {
					values = append(values, Utils.GetTime(attr).Format("2006-01-02T15:04:05"))

				} else {
					values = append(values, string(attr))
				}

			}
		}
	}
	ewf_h2_section.AcquiredMediaInfo = Utils.ToMap(identifiers, values)

}

func (ewf_h2_section EWF_Header2_Section) GetAttr(requestedInfo string) interface{} {

	return ewf_h2_section.AcquiredMediaInfo[requestedInfo]

}
