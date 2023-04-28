package sections

import (
	"time"

	"github.com/aarsakian/EWF_Reader/ewf/utils"
)

type EWF_Data_Section struct {
	MediaType         uint8
	Unknown1          [3]uint8
	ChunkCount        uint32
	NofSectorPerChunk uint32 "Number of Sectors per Chunk default 64"
	NofBytesPerSector uint32 "default 512"
	NofSectors        uint64 "Number of Sectors within all segment files"
	NofCylindersCHS   uint32 "Number of cylinders of the C:H:S usually empty"
	NofHeadesCHS      uint32 "Number of cylinders of the C:H:S usually empty"
	NofSectorsCHS     uint32 "Number of Sectors of the C:H:S usually empty"
	MediaFlags        uint8
	Uknown2           [3]uint8
	PALM              uint32 "Volume start sector"
	Unkown3           [4]uint8
	SMART             uint32 "start sector offset relative from the end of the media"
	CompressionLevel  uint8
	Unknown4          [3]uint8
	SectorErrorGr     [4]uint8 "Sector error granularity"
	Unknown5          [4]uint8
	GUID              [16]uint8 "identify uniquely a set of segment files"
	Signature         [5]uint8  "contains 0x00"
	CheckSum          [4]uint8  "adler32 of all the previous data within the additional volume section data 1048"
}

func (ewf_data_section *EWF_Data_Section) GetAttr(string) interface{} {

	return ""
}

func (ewf_data_section *EWF_Data_Section) Parse(buf []byte) {
	defer utils.TimeTrack(time.Now(), "Parsing")
	utils.Unmarshal(buf, ewf_data_section)

	/*  fmt.Println("Data section",ewf_data_section.body.NofSectorPerChunk, ewf_data_section.body.PALM,
	    ewf_data_section.body.ChunkCount, ewf_data_section.body.CompressionLevel, ewf_data_section.body.NofBytesPerSector)
	*/
}
