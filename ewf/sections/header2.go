package header2

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/aarsakian/EWF_Reader/ewf/utils"
)

const (
	EWF_HEADER_VALUES_INDEX_DESCRIPTION = iota
	EWF_HEADER_VALUES_INDEX_CASE_NUMBER
	EWF_HEADER_VALUES_INDEX_EXAMINER_NAME
	EWF_HEADER_VALUES_INDEX_EVIDENCE_NUMBER
	EWF_HEADER_VALUES_INDEX_NOTES
	EWF_HEADER_VALUES_INDEX_ACQUIRY_SOFTWARE_VERSION
	EWF_HEADER_VALUES_INDEX_ACQUIRY_OPERATING_SYSTEM
	EWF_HEADER_VALUES_INDEX_ACQUIRY_DATE
	EWF_HEADER_VALUES_INDEX_SYSTEM_DATE
	EWF_HEADER_VALUES_INDEX_PASSWORD
	EWF_HEADER_VALUES_INDEX_PROCESS_IDENTIFIER
	EWF_HEADER_VALUES_INDEX_UNKNOWN_DC
	EWF_HEADER_VALUES_INDEX_EXTENTS
	EWF_HEADER_VALUES_INDEX_COMPRESSION_TYPE

	EWF_HEADER_VALUES_INDEX_MODEL
	EWF_HEADER_VALUES_INDEX_SERIAL_NUMBER
	EWF_HEADER_VALUES_INDEX_DEVICE_LABEL

	/* Value to indicate the default number of header values
	 */
	EWF_HEADER_VALUES_DEFAULT_AMOUNT
)

type EWF_Header_Section struct {
}

type EWF_Header2_Section struct {
	BOM           [2]uint8 //declares endianess 0xff0xfe (little) vice versa
	NofCategories uint16
	CategoryName  []uint8
	a             string    "Description"
	c             string    "Case Number"
	n             string    "Evidence Number"
	e             string    "Examiner Name"
	t             string    "Notes"
	av            string    "Version"
	ov            string    "Platform"
	m             time.Time "Acquired Date"
	u             time.Time "System Date"
	p             string    "Password Hash"
	pid           string    "Process Identifiers"
	dc            string    "Unknown"
	ext           string    "Extents"
}

func (ewf_header_section *EWF_Header_Section) Parse(buf *bytes.Reader) {

}

func (ewf_h2_section *EWF_Header2_Section) Parse(buf *bytes.Reader) {
	//0x09 tab 0x0a new line delimiter
	//function to parse header2 section attributes
	//to do take into account endianess
	val := make([]byte, buf.Len())
	buf.Read(val)
	val = utils.Decompress(val)

	defer utils.TimeTrack(time.Now(), "Parsing")
	line_del, _ := hex.DecodeString("0a")
	tab_del, err := hex.DecodeString("09")
	if err != nil {
		log.Fatal(err)
	}
	var b *bytes.Reader

	for line_number, line := range bytes.Split(val, line_del) {
		for id_num, attr := range bytes.Split(line, tab_del) {
			b = bytes.NewReader(attr)
			if line_number == 0 {
				utils.Parse(b, &ewf_h2_section.BOM)
				utils.Parse(b, &ewf_h2_section.NofCategories)

			} else if line_number == 1 {
				utils.Parse(b, &ewf_h2_section.CategoryName)
			} else if line_number == 2 {

			} else if line_number == 3 {
				if id_num == EWF_HEADER_VALUES_INDEX_DESCRIPTION {
					ewf_h2_section.a = string(attr)
					fmt.Println("TIME", ewf_h2_section.a)
				} else if id_num == EWF_HEADER_VALUES_INDEX_CASE_NUMBER {
					ewf_h2_section.c = string(attr)

				} else if id_num == EWF_HEADER_VALUES_INDEX_EXAMINER_NAME {
					ewf_h2_section.n = string(attr)
				} else if id_num == EWF_HEADER_VALUES_INDEX_EVIDENCE_NUMBER {
					ewf_h2_section.e = string(attr)
				} else if id_num == EWF_HEADER_VALUES_INDEX_NOTES {
					ewf_h2_section.t = string(attr)
				} else if id_num == EWF_HEADER_VALUES_INDEX_ACQUIRY_SOFTWARE_VERSION {
					ewf_h2_section.av = string(attr)
				} else if id_num == EWF_HEADER_VALUES_INDEX_ACQUIRY_OPERATING_SYSTEM {
					ewf_h2_section.ov = string(attr)
				} else if id_num == EWF_HEADER_VALUES_INDEX_ACQUIRY_DATE {
					ewf_h2_section.m = utils.SetTime(attr)

				} else if id_num == EWF_HEADER_VALUES_INDEX_SYSTEM_DATE {
					ewf_h2_section.u = utils.SetTime(attr)

				} else if id_num == EWF_HEADER_VALUES_INDEX_PASSWORD {
					ewf_h2_section.p = string(attr)
				} else if id_num == EWF_HEADER_VALUES_INDEX_PROCESS_IDENTIFIER {
					ewf_h2_section.pid = string(attr)

				}

			}
		}
	}

}

func NeedsDecompress(section_obj interface{}) bool {
	var needsDecompress bool
	switch section_obj.(type) {

	case *EWF_Header2_Section:
		needsDecompress = true
	case *EWF_Header_Section:
		needsDecompress = true
	default:
		needsDecompress = false
	}
	return needsDecompress
}

func (ewf_h2_section *EWF_Header2_Section) GetAttr(string) interface{} {
	return ""
}

func (ewf_h_section *EWF_Header_Section) GetAttr(string) interface{} {
	return ""
}
