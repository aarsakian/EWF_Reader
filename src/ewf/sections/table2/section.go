
package table2

import (
    "bytes"
    "time"
    "ewf/parseutil"
    "fmt"
    "reflect"
   

    
)

const Chunk_Size uint32 = 64*512

type EWF_Table_Section_Entry struct{
    ChunkDataOffset uint32 "MSB indicates if chunk data is (un)compressed (0)/1 offset relative to the start of the fileit is located in the preseding sectors section "
    IsCompressed uint8 "1 -> Compressed"
}

type EWF_Table_Section_Footer struct {
    Checksum [4]uint8
}
//not used in EnCase
type  EWF_Table_Section_Data struct {
    Data []byte "chunk data compressed under deflate as well as its Checksum"
}

//resides right after table section
type EWF_Table2_Section struct {
    table_section EWF_Table_Section 
}


//table section  identifier 
type EWF_Table_Section struct {
    table_header EWF_Table_Section_Header
    table_entries []EWF_Table_Section_Entry
    table_footer EWF_Table_Section_Footer
}

type EWF_Table_Section_Header struct { //24 bytes
    nofEntries uint32 "Number of Entries 0x01"
    Padding [16]uint8 "contains 0x00"
    Checksum [4]uint8 "Adler32"
}


//EnCase 6-7
type EWF_Table_Section_Header_EnCase struct { //24bytes
    nofEntries  uint32 "Number of Entries 0x01"
    Padding1 [4]uint8 "contains 0x00"
    TableBaseOffset uint64 "Adler32"
    Padding2 [4]uint8 "contains 0x00"
    Checksum [4]uint8 "Adler32"
}


func  (table_header *EWF_Table_Section_Header) Parse(buf *bytes.Reader) {
   
    //parse struct attributes
   
    parseutil.Parse(buf, &table_header.nofEntries)
    parseutil.Parse(buf, &table_header.Padding)
    parseutil.Parse(buf, &table_header.Checksum)
    
  
}



func  (table_entry *EWF_Table_Section_Entry) Parse(buf *bytes.Reader){
    var b *bytes.Reader
   
    val :=  make([]byte, int64(buf.Len()))
       
      
    buf.Read(val)
    //parse struct attributes
    table_entry.IsCompressed = val[3]<<1&1
    val[3] &= 0x7F //exlude MSB
    
    b = bytes.NewReader(val)
    
    parseutil.Parse(b, &table_entry.ChunkDataOffset)
}


func  (table_footer *EWF_Table_Section_Footer) Parse(buf *bytes.Reader) {
   
  
    //parse struct attributes
   
    
    parseutil.Parse(buf, &table_footer.Checksum)
   
    
}


func (ewf_table_section *EWF_Table_Section) Parse(buf *bytes.Reader) {
    
    defer parseutil.TimeTrack(time.Now(), "Parsing")
    val :=  make([]byte, int64(buf.Len()))
       
      
    buf.Read(val)
    
    ewf_table_section.table_header.Parse(bytes.NewReader(val[0:24]))
    ewf_table_section.table_footer.Parse(bytes.NewReader(val[len(val)-4:len(val)]))
    val = val[24:len(val)-4]
    k:=0
    ewf_table_section.table_entries = make([]EWF_Table_Section_Entry ,ewf_table_section.table_header.nofEntries)
    for i :=uint32(0); i<ewf_table_section.table_header.nofEntries; i+=1  {
        
        ewf_table_section.table_entries[i].Parse(bytes.NewReader(val[0+k:4+k]))
      //  fmt.Println("EFW in by",i,
        //       ewf_table_section.table_entries[i].IsCompressed,ewf_table_section.table_entries[i].ChunkDataOffset)
        k+=4
           
    }
    
  
  
}


func (ewf_table2_section *EWF_Table2_Section) Parse(buf *bytes.Reader){
    
}


func (ewf_table2_section *EWF_Table2_Section) Collect([]byte, uint64){
    
}


func (ewf_table_section *EWF_Table_Section) Collect(sectors_buf []byte, sectors_offs uint64) {
    fmt.Println("NODF entries",len(ewf_table_section.table_entries),ewf_table_section.table_header.nofEntries)
    zlib_header := []byte{72, 13}
    var data  []byte
    for  idx, entry := range ewf_table_section.table_entries[:len(ewf_table_section.table_entries)-1] {
          
           
            data = sectors_buf[entry.ChunkDataOffset-uint32(sectors_offs):entry.ChunkDataOffset-uint32(sectors_offs)+Chunk_Size]
          
            if bytes.HasPrefix(data, zlib_header) {
                parseutil.Decompress(data)
                 fmt.Println("IDX", idx)
                    /*sectors_buf[entry.ChunkDataOffset-uint32(sectors_offs):entry.ChunkDataOffset-uint32(sectors_offs)+5],
                    "REM",uint32(len(sectors_buf))-entry.ChunkDataOffset-uint32(sectors_offs), "CompresseD?",entry.IsCompressed)*/
              //  parseutil.DecompressF(data)
            }
           
    
    }
    //last data chunk maybe less than 32K size
    last_entry := ewf_table_section.table_entries[len(ewf_table_section.table_entries)-1]
    data = sectors_buf[last_entry.ChunkDataOffset-uint32(sectors_offs):
                       last_entry.ChunkDataOffset-uint32(sectors_offs)+
                       uint32(len(sectors_buf))-last_entry.ChunkDataOffset-uint32(sectors_offs)]
    if bytes.HasSuffix(data, zlib_header) {
        parseutil.DecompressF(data)
    }
}



func (ewf_table_section *EWF_Table_Section) GetAttr(attr string) (interface{}) {
    s := reflect.ValueOf(ewf_table_section).Elem()//retrieve since it's a pointer
  
    
    sub_s := s.FieldByName(attr)
    
    if (attr == "table_entries") {
            for entry_number := 0; entry_number < sub_s.Len(); entry_number++ {
               
                 s_inner :=  sub_s.Index(entry_number)
               //  fmt.Println("ADDR",s_inner)
                for inner_idx :=0; inner_idx < s_inner.NumField(); inner_idx++ {
                   // fmt.Println("ENTRY",s_inner.Field(inner_idx))
                }
            }
      
    }
    return ""
}


func (ewf_table2_section *EWF_Table2_Section) GetAttr(string) (interface{}) {
    return ""
}
