
package sectors

import (
        "bytes"
        "io"
       // "fmt"
          "compress/flate"
         "ewf/parseutil"
       )

type EWF_Sectors_Section struct {
    body []byte
}

func (ewf_sectors_section *EWF_Sectors_Section) GetAttr() (interface{}) {
    return ewf_sectors_section.body[:]
}


func (ewf_sectors_section *EWF_Sectors_Section) Parse(buf *bytes.Reader){
   
    ewf_sectors_section.body = make([]byte, buf.Len())
    parseutil.Parse(buf, &ewf_sectors_section.body)
}

func (sectors *EWF_Sectors_Section) Decompress() ([]byte) {
    b := bytes.NewReader(sectors.body)
	r := flate.NewReader(b)
    var buf bytes.Buffer// buffer needs no initilization pointer

	io.Copy(&buf, r)
	r.Close()
  
    return buf.Bytes()
}
