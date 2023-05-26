package ewf

import (
	"github.com/aarsakian/EWF_Reader/ewf/sections"
	"github.com/aarsakian/EWF_Reader/ewf/utils"
)

type Sections struct {
	head *Section
	tail *Section
}

type Body interface {
	Parse([]byte)
	GetAttr(string) interface{}
}

type Collector interface {
	Collect([]byte, uint64)
}

type ParserCollector interface {
	Collector
	Body
}

type Section struct {
	Descriptor *Section_Descriptor
	Type       string
	BodyOffset uint64
	body       Body
	next       *Section
	prev       *Section
}

type Section_Descriptor struct {
	//after header of segment a section starts
	//size 76 bytes
	Header          [16]uint8
	NextSectionOffs int64 //from the start of the segment
	SectionSize     uint64
	Padding         [40]uint8
	Checksum        uint32
}

func (descriptor Section_Descriptor) GetType() string {
	return utils.Stringify(descriptor.Header[:])
}

func (section *Section) GetAttr(val string) any {
	return section.body.GetAttr(val)
}

/*
	func (section *Section) ParseHeader(buf *bytes.Reader) {
		section.Header.Parse(buf) //parse header attributes
		section.BodyOffset = section.SHeader.NextSectionOffs

}
*/
func (section *Section) ParseBody(buf []byte) {
	switch section.Type {
	case "header2":
		section.body = new(sections.EWF_Header2_Section)
	case "header":
		section.body = new(sections.EWF_Header_Section)
	case "disk":
		section.body = new(sections.EWF_Disk_Section)
	case "sectors":
		section.body = new(sections.EWF_Sectors_Section)
	case "table2":
		section.body = new(sections.EWF_Table2_Section)
	case "table":
		section.body = new(sections.EWF_Table_Section)
	case "next":
		section.body = new(sections.EWF_Next_Section)
	case "data":
		section.body = new(sections.EWF_Data_Section)
	case "volume":
		section.body = new(sections.EWF_Volume_Section)
	case "done":
		section.body = new(sections.EWF_Done_Section)
	case "hash":
		section.body = new(sections.EWF_Hash_Section)
	case "digest":
		section.body = new(sections.EWF_Digest_Section)
	}

	section.body.Parse(buf)

	/* if Sections[i].Type == "table2" || Sections[i].Type == "table" {
	       Sections[i].PC.Parse(buf)
	       Sections[i].PC.Collect(data.([]byte)[:], sectors_offs)
	   } else if Sections[i].Type == "sectors" {
	         Sections[i].P.Parse(buf)
	         data = Sections[i].P.GetAttr().([]byte)
	         sectors_offs = cur_offset
	         fmt.Println("TYPe",reflect.TypeOf(data))
	   } else {
	       Sections[i].P.Parse(buf)
	   }*/
}

/*


func (section_header *Section_Header) Parse(buf *bytes.Reader) {

	defer utils.TimeTrack(time.Now(), "Parsing") //header of each section

	s := reflect.ValueOf(section_header).Elem()
	for i := 0; i < s.NumField(); i++ {
		//parse struct attributes
		utils.Parse(buf, s.Field(i).Addr().Interface())

	}

}

func (section_header *Section_Header) Verify(datar *bytes.Reader) bool {
	var buf []byte

	datar.Read(buf)
	fmt.Println(section_header.Checksum, len(buf))
	return section_header.Checksum == adler32.Checksum(buf[:72])

}
*/
