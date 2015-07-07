package sections

import (
    "bytes"
    "fmt"
    "reflect"
    "time"
    "ewf/sections/header2"
    "ewf/sections/next"
    "ewf/sections/sectors"
    "ewf/sections/table2"
    "ewf/sections/disk"
    "ewf/sections/volume"
    "ewf/sections/data"
    "ewf/sections/hash"
    "ewf/sections/done"
    "ewf/parseutil"
    "hash/adler32"
)




type Parser interface {
    Parse(*bytes.Reader) 
    Attr
   
}
type Attr interface{
    GetAttr() (interface{})
}
type Collector interface{
    Collect([]byte, uint64)
   
}

type ParserCollector interface{
    Collector
    Parser
}

type Section struct {
    SHeader Section_Header
    Type string
    BodyOffset uint64
    P Parser
   
}

type Section_Header struct {
    //after header of segment a section starts
    //size 76 bytes
    Header [16]uint8
    NextSectionOffs uint64 //from the start of the segment 
    SectionSize uint64
    Padding [40] uint8
    Checksum uint32
    
}


func (section *Section) ParseHeader(buf *bytes.Reader) {
    section.SHeader.Parse(buf)//parse header attributes
    section.BodyOffset = section.SHeader.NextSectionOffs

}

func (section *Section) ParseBody(buf *bytes.Reader) {
    if section.Type != "sectors" {
        section.P.Parse(buf)
    }
 
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


func (section *Section) Dispatch()  {
   
    section.findType()
    switch section.Type {
        case "header2":
            section.P = new(header2.EWF_Header2_Section)
        case "header":
            section.P =  new(header2.EWF_Header_Section)
        case "disk":
            section.P =  new(disk.EWF_Disk_Section)
        case "sectors":
            section.P =  new(sectors.EWF_Sectors_Section)
        case "table2":
            section.P =  new(table2.EWF_Table2_Section)
        case "table":
            section.P = new(table2.EWF_Table_Section)
        case "next":
           section.P = new(next.EWF_Next_Section)
        case "data":
            section.P= new(data.EWF_Data_Section)
        case "volume":
          section.P = new(volume.EWF_Volume_Section)
        case "Done":
           section.P = new(done.EWF_Done_Section)
        case "hash":
           section.P = new(hash.EWF_Hash_Section)
    }
    fmt.Println("SECTION ", section.Type )
 
}


func (section_header *Section_Header) Parse(buf *bytes.Reader) {

    defer parseutil.TimeTrack(time.Now(), "Parsing") //header of each section
      
   
    s := reflect.ValueOf(section_header).Elem()
    for i := 0; i < s.NumField(); i++ {
    //parse struct attributes
        parseutil.Parse(buf, s.Field(i).Addr().Interface())
       
    }
  
   
}

func (section_header *Section_Header) Verify(datar *bytes.Reader) bool {
    var buf []byte
   
    datar.Read(buf)
     fmt.Println(section_header.Checksum, len(buf))
    return section_header.Checksum == adler32.Checksum(buf[:72])
   
}


func (section *Section) findType() {
    section.Type = parseutil.Stringify(section.SHeader.Header[:])
}

