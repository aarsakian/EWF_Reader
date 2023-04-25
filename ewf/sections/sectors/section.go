package sectors

import (
	"bytes"

	//  "fmt"

	"github.com/aarsakian/EWF_Reader/ewf/utils"
	//     "hash/adler32"
)

type EWF_Sectors_Section struct {
	data []byte
}

func (ewf_sectors_section *EWF_Sectors_Section) GetAttr(string) interface{} {
	return ewf_sectors_section.data[:]
}

/*func (ewf_sectors_section *EWF_Sectors_Section) Verify() bool {
   fmt.Println("CHLKSUM", ewf_sectors_section.checksum,  adler32.Checksum(ewf_sectors_section.data))
   return ewf_sectors_section.checksum == adler32.Checksum(ewf_sectors_section.data)
}*/

func (ewf_sectors_section *EWF_Sectors_Section) Parse(buf *bytes.Reader) {

	ewf_sectors_section.data = make([]byte, buf.Len())
	utils.Parse(buf, &ewf_sectors_section.data)

	//  fmt.Println("ERR FREE",ewf_sectors_section.Verify())
}
