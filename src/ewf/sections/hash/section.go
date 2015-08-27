package hash
import (
    "bytes"
    "ewf/parseutil"
    "reflect"
    "time"
    
    "encoding/hex"
)


type EWF_Hash_Section struct{
    MD5_value [16]uint8
    Unknown uint8
    Checksum uint32 //adler32
}

func (hash_section *EWF_Hash_Section)  Parse(r *bytes.Reader){
    
    defer parseutil.TimeTrack(time.Now(), "Parsing")
  
    
    s := reflect.ValueOf(hash_section).Elem()
    for i := 0; i < s.NumField(); i++ {
    //parse struct attributes
        parseutil.Parse(r, s.Field(i).Addr().Interface())
       
    }
}

func (hash_section *EWF_Hash_Section) GetAttr(attr string) (interface{}) {
    s := reflect.ValueOf(hash_section).Elem()//retrieve since it's a pointer
  
    
    sub_s := s.FieldByName(attr)
    if sub_s.IsValid() {
      
                switch v := sub_s.Interface().(type) {
            
                    case uint32, int8:
                            return v
                          
                    case [16]uint8:
                           return hex.EncodeToString(v[:])
                            
                    default:
                            return "unknown"
                           
            }

    } else {
        return "Not Valid"
    }
}
   