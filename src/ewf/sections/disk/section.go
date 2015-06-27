
package disk

import "bytes"

type EWF_Disk_Section struct {
    body []byte
}
func (ewf_disk_section *EWF_Disk_Section) GetAttr() (interface{}) {
    return &ewf_disk_section.body
}


func (ewf_disk_section *EWF_Disk_Section) Parse(buf *bytes.Reader){
    
}

