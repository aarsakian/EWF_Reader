package sections

//  "fmt"

//     "hash/adler32"

type DataChucks []DataChuck

type DataChuck struct {
	Data []byte
}

type EWF_Sectors_Section struct {
	DataChucks DataChucks
}

func (ewf_sectors_section *EWF_Sectors_Section) GetAttr(string) interface{} {
	return ewf_sectors_section.DataChucks
}

/*func (ewf_sectors_section *EWF_Sectors_Section) Verify() bool {
   fmt.Println("CHLKSUM", ewf_sectors_section.checksum,  adler32.Checksum(ewf_sectors_section.data))
   return ewf_sectors_section.checksum == adler32.Checksum(ewf_sectors_section.data)
}*/

func (ewf_sectors_section *EWF_Sectors_Section) Parse(buf []byte) {

	ewf_sectors_section.DataChucks = append(ewf_sectors_section.DataChucks, DataChuck{Data: buf})
}
