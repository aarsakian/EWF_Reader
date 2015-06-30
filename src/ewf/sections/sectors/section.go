
package sectors

import (
        "bytes"
 
       // "fmt"
        
         "ewf/parseutil"
           "hash/adler32"
       )

type EWF_Sectors_Section struct {
    data []byte
    checksum uint32 //adler32
}

func (ewf_sectors_section *EWF_Sectors_Section) GetAttr() (interface{}) {
    return ewf_sectors_section.data[:]
}


func (ewf_sectors_section *EWF_Sectors_Section) ErrorFree() bool {
   return ewf_sectors_section.checksum == adler32.Checksum(ewf_sectors_section.data)
}

func (ewf_sectors_section *EWF_Sectors_Section) Parse(buf *bytes.Reader){
   
   
    ewf_sectors_section.data = make([]byte, buf.Len())
    parseutil.Parse(buf, &ewf_sectors_section.data)
}

