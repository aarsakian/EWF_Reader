package sections

import (
	"bytes"
	"encoding/hex"
	"log"
	"time"

	Utils "github.com/aarsakian/EWF_Reader/ewf/utils"
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

var CompressionLevels = map[string]string{
	"b": "Best",
	"f": "Fastest",
	"n": "No compression",
}

var AcquiredMediaIdentifiers = map[string]string{
	"a":   "Description",
	"c":   "Case Number",
	"n":   "Evidence Number",
	"e":   "Examiner Name",
	"t":   "Notes",
	"av":  "Version",
	"ov":  "Platform",
	"m":   "Acquired Date",
	"u":   "System Date",
	"p":   "Password Hash",
	"pid": "Process Identifiers",
	"dc":  "Unknown",
	"ext": "Extents",
	"r":   "Compression level",
}

type EWF_Header_Section struct {
	NofCategories     string
	CategoryName      string
	AcquiredMediaInfo map[string]string
}

func (ewf_h_section *EWF_Header_Section) Parse(buf []byte) {
	//0x09 tab 0x0a new line delimiter
	//function to parse header2 section attributes
	//to do take into account endianess

	val := Utils.Decompress(buf)

	defer Utils.TimeTrack(time.Now(), "Parsing")
	line_del, _ := hex.DecodeString("0a")
	tab_del, err := hex.DecodeString("09")
	if err != nil {
		log.Fatal(err)
	}

	var identifiers []string
	var values []string

	for line_number, line := range bytes.Split(val, line_del) {
		for id_num, attr := range bytes.Split(line, tab_del) {

			if line_number == 0 {
				ewf_h_section.NofCategories = string(attr[0])

			} else if line_number == 1 {
				ewf_h_section.CategoryName = string(attr)
			} else if line_number == 2 {
				identifiers = append(identifiers, AcquiredMediaIdentifiers[string(attr)])
			} else if line_number == 3 {
				if id_num == EWF_HEADER_VALUES_INDEX_ACQUIRY_DATE || id_num == EWF_HEADER_VALUES_INDEX_SYSTEM_DATE {
					values = append(values, Utils.GetTime(attr).Format("2006-01-02T15:04:05"))

				} else {
					values = append(values, string(attr))
				}

			}
		}
	}
	ewf_h_section.AcquiredMediaInfo = Utils.ToMap(identifiers, values)

}

func (ewf_h_section EWF_Header_Section) GetAttr(string) interface{} {
	return ewf_h_section
}
