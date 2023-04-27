package sections

import (
	"time"

	"github.com/aarsakian/EWF_Reader/ewf/utils"
)

type EWF_Data_Section struct {
	body     []byte
	checksum [4]uint8 "adler32"
}

func (ewf_data_section *EWF_Data_Section) GetAttr(string) interface{} {

	return ""
}

func (ewf_data_section *EWF_Data_Section) Parse(buf []byte) {
	defer utils.TimeTrack(time.Now(), "Parsing")

	/*ewf_data_section.body = new(Volume_Data)
	s := reflect.ValueOf(ewf_data_section.body).Elem()
	for i := 0; i < s.NumField(); i++ {
		//parse struct attributes
		utils.Parse(r, s.Field(i).Addr().Interface())

	}
	/*  fmt.Println("Data section",ewf_data_section.body.NofSectorPerChunk, ewf_data_section.body.PALM,
	    ewf_data_section.body.ChunkCount, ewf_data_section.body.CompressionLevel, ewf_data_section.body.NofBytesPerSector)
	*/
}
