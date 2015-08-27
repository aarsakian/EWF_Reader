
package ewf

import (
        "bytes"
        "time"
        "os"
        "reflect"
        "ewf/parseutil"
        "ewf/sections"
        "strings"
        "io"
        "fmt"
        "log"
        "runtime"
        )

const EWF_Section_Header_s uint64 = 76
const EWF_Header_s uint64 = 13
const NofSections = 32

type EWF_file struct {
    File *os.File
    Size uint64
    hasNext bool
    isLast bool
    SegmentNum uint
}


type EWF_Header struct {
    //header size 13bytes of each segment (file)
    Signature [8] uint8 "EVF\x0d\x0a\xff\x00"
    StartofFields uint8 //
    SegNum uint16 //
    EOF uint16
    
    
}




func (ewf_header *EWF_Header) Parse(buf *bytes.Reader) {
    //parse struct attributes
    //iterate through the fields of the struct
    defer parseutil.TimeTrack(time.Now(), "Parsing")
    s := reflect.ValueOf(ewf_header).Elem()
    for i := 0; i < s.NumField(); i++ {
        parseutil.Parse(buf, s.Field(i).Addr().Interface())
    }
}


func (ewf_file* EWF_file)  Verify() {
    
}
func (ewf_file* EWF_file) ParseHeader(cur_offset *uint64) {
    defer parseutil.TimeTrack(time.Now(), "Parsing Segment Header")
    buf := ewf_file.ReadAt(EWF_Header_s, *cur_offset)//producer
    *cur_offset+=EWF_Header_s
    ewf_header :=new(EWF_Header)//ewf_header acts as a pointer

   
    
    ewf_header.Parse(buf)//consume
    sig := parseutil.Stringify(ewf_header.Signature[:])
   
    if !strings.Contains(sig, "EVF") {
        os.Exit(0)
    }
  
    
    
}
func (ewf_file* EWF_file) ParseSegment() {
    
   cur_offset := uint64(0)
   // var parser Parser
   
   ewf_file.ParseHeader(&cur_offset)
   
   
    
    var Sections[NofSections] *sections.Section
 
    var m runtime.MemStats  
  //  var data interface{}
    //var sectors_offs uint64
    var buf *bytes.Reader
    for i := 0; i <= NofSections; i++  {
     //   parsing section headers
       
        buf = ewf_file.ReadAt(EWF_Section_Header_s, cur_offset)//read section header
     
        cur_offset+=EWF_Section_Header_s
        Sections[i] = new(sections.Section)//create section
        Sections[i].ParseHeader(buf)
        Sections[i].Dispatch()//object factory section body creation
        if Sections[i].Type == "next" {
            
            ewf_file.hasNext = true
            break
        } else if Sections[i].Type == "done" {
            
            ewf_file.isLast = true
            break
        }
     
        if Sections[i].Type != "sectors" {
            buf = ewf_file.ReadAt(Sections[i].BodyOffset-cur_offset, cur_offset)//read section body
     
        }
        
      
       
      
      
        Sections[i].ParseBody(buf)
        fmt.Println("finished ",Sections[i].Type,
                 "KB Remaining",
                     (ewf_file.Size-cur_offset)/1024, "OFFS", cur_offset, " length KB",
                     (Sections[i].BodyOffset-cur_offset)/1024)
        cur_offset = Sections[i].BodyOffset
        runtime.ReadMemStats(&m)
        Sections[i].GetAttr("table_entries")
        md5 := Sections[i].GetAttr("MD5_value")
        fmt.Println("MD5",md5)
       
        fmt.Printf("Asked %d,Allocated %d,unused %d, released %d,round %d\n", m.HeapSys, m.HeapAlloc,
            m.HeapIdle, m.HeapReleased, i)
    
           
        
       
    }
    //disk section and sectors section
}


func (ewf_file *EWF_file) ReadAt(length uint64, off uint64)  ( *bytes.Reader) {
    //cast to struct respecting endianess
    defer parseutil.TimeTrack(time.Now(), "reading")
    buff := make([]byte, length)
    var err error
    var n int
 //read 100KB chunks
   STEP := uint64(1000*1024)
    rem:=length
    if length < STEP {
        _, err := ewf_file.File.ReadAt(buff, int64(off))
        if err == io.EOF {
        	fmt.Println("Error reading file:", err)
            
        }
    } else {
        for i :=uint64(0); i <= length; i += STEP {
            if (rem<STEP) {//final read
                n, err = ewf_file.File.ReadAt(buff[i:length], int64(off))
            } else {
                n, err = ewf_file.File.ReadAt(buff[i:i+STEP], int64(off))
            }
            off += uint64(n)
            rem -= uint64(n)
            if err != nil {
            	fmt.Println("Error reading file:", err)
                log.Fatal(err)
            }
        }
    }
   
   
    return bytes.NewReader(buff)
}


