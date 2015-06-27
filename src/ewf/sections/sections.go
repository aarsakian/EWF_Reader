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
    "ewf/parseutil"
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
    Section_header EWF_Section_Header
    Type string
    P Parser
    PC ParserCollector
   
}

type EWF_Section_Header struct {
    //after header of segment a section starts
    //size 76 bytes
    Header [16]uint8
    NextSectionOffs uint64 //from the start of the segment 
    SectionSize uint64
    Padding [40] uint8
    Checksum [4] uint8
    
}


func (section *Section) Parse() {
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
            section.PC =  new(table2.EWF_Table2_Section)
        case "table":
            section.PC = new(table2.EWF_Table_Section)
        case "next":
           section.P = new(next.EWF_Next_Section)
        case "data":
            section.P= new(data.EWF_Data_Section)
        case "volume":
          section.P = new(volume.EWF_Volume_Section)
    }
    fmt.Println("SECTION ", section.Type )
 
}


func (ewf_section_header *EWF_Section_Header) Parse(buf *bytes.Reader) {

    defer parseutil.TimeTrack(time.Now(), "Parsing") //header of each section
    s := reflect.ValueOf(ewf_section_header).Elem()
    for i := 0; i < s.NumField(); i++ {
    //parse struct attributes
        parseutil.Parse(buf, s.Field(i).Addr().Interface())
       
    }
  
}

func (section *Section) findType() {
    section.Type = parseutil.Stringify(section.Section_header.Header[:])
}

