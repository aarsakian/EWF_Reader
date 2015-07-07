package done
import (
    "bytes"
    "ewf/parseutil"
    "reflect"
    "time"
)


type EWF_Done_Section struct{
    
}

func (done_section *EWF_Done_Section)  Parse(r *bytes.Reader){
    
    defer parseutil.TimeTrack(time.Now(), "Parsing")
  
    
    s := reflect.ValueOf(done_section).Elem()
    for i := 0; i < s.NumField(); i++ {
    //parse struct attributes
        parseutil.Parse(r, s.Field(i).Addr().Interface())
       
    }
}

func (done_section *EWF_Done_Section) GetAttr() (interface{}) {
    return ""
}