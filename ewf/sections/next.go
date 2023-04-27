
package next

import "bytes"

type EWF_Next_Section struct {//last section of a segment file
    body []byte
}
func (ewf_next_section *EWF_Next_Section) GetAttr(string) (interface{}) {
    return &ewf_next_section.body
}


func (ewf_next_section* EWF_Next_Section) Parse(*bytes.Reader) {
    
}
