package ewf

import (
	"fmt"

	"github.com/aarsakian/EWF_Reader/ewf/sections"
	Utils "github.com/aarsakian/EWF_Reader/ewf/utils"
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
	Descriptor                   *Section_Descriptor
	DescriptorCalculatedChecksum uint32
	Type                         string
	BodyOffset                   uint64
	body                         Body
	next                         *Section
	prev                         *Section
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

func (sections Sections) Filter(sectionNames []string) []Section {
	var filteredSections []Section

	for _, sectionName := range sectionNames {
		section := sections.head
		for section != nil {

			if section.Type == sectionName {
				filteredSections = append(filteredSections, *section)

			}
			section = section.next

		}
	}

	return filteredSections
}

func (sections Sections) GetSectionPtr(sectionName string) *Section {
	section := sections.head
	for section != nil {
		if section.Type != sectionName {
			section = section.next
			continue
		}
		return section

	}

	return nil
}

func (descriptor Section_Descriptor) IsBodyEmpty() bool {
	return descriptor.SectionSize == 0
}

func (descriptor Section_Descriptor) GetType() string {
	return Utils.Stringify(descriptor.Header[:])
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
	case "x_description":
		section.body = new(sections.XDescription)
	case "x_hash":
		section.body = new(sections.XHash)
	case "x_statistics":
		section.body = new(sections.XStatistics)
	case "error2":
		section.body = new(sections.Error2_Section)
	default:
		fmt.Println("uknown section", section.Type)
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

	defer Utils.TimeTrack(time.Now(), "Parsing") //header of each section

	s := reflect.ValueOf(section_header).Elem()
	for i := 0; i < s.NumField(); i++ {
		//parse struct attributes
		Utils.Parse(buf, s.Field(i).Addr().Interface())

	}

}

func (section_header *Section_Header) Verify(datar *bytes.Reader) bool {
	var buf []byte

	datar.Read(buf)
	fmt.Println(section_header.Checksum, len(buf))
	return section_header.Checksum == adler32.Checksum(buf[:72])

}
*/
