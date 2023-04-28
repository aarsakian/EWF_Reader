package sections

import (

	//  "fmt"

	"github.com/aarsakian/EWF_Reader/ewf/utils"
	//     "hash/adler32"
)

type DataChuck struct {
	data     []byte
	Checksum [4]uint8 //adler-32
}

type EWF_Sectors_Section struct {
	DataChucks []DataChuck
}

func (ewf_sectors_section *EWF_Sectors_Section) GetAttr(string) interface{} {
	return ewf_sectors_section.DataChucks
}

/*func (ewf_sectors_section *EWF_Sectors_Section) Verify() bool {
   fmt.Println("CHLKSUM", ewf_sectors_section.checksum,  adler32.Checksum(ewf_sectors_section.data))
   return ewf_sectors_section.checksum == adler32.Checksum(ewf_sectors_section.data)
}*/

func (ewf_sectors_section *EWF_Sectors_Section) Parse(buf []byte) {

	utils.Unmarshal(buf, ewf_sectors_section)

	//  fmt.Println("ERR FREE",ewf_sectors_section.Verify())
}
