
package parseutil

import (
    "bytes"
    "fmt"
    "encoding/binary"
    "time"
    "log"
    "compress/zlib"
    "compress/flate"
    "io"
    "strconv"
    "runtime"
   // "io/ioutil"
)

func Parse(buf *bytes.Reader, val interface{}) {
    //target is val
    err := binary.Read(buf, binary.LittleEndian, val)
  
    if err != nil {
        fmt.Println("binary.Read failed: General", err)
    }
    
}


func TimeTrack(start time.Time, name string) {
  
    elapsed := time.Since(start)
   
    log.Printf("%s took %s ", name, elapsed)
}


func SetTime(attr []byte) (time.Time) {
  //  fmt.Println("TIME LEN",string(attr),attr,len(attr))
    timestamp :=string(bytes.Join(bytes.Split(attr, []byte{00}),[]byte{}))//crazy since original byte seq not in ASCII
    if len(attr) == 21 {
        t , err:=strconv.ParseInt(timestamp, 10, 32)
    
        if (err != nil) {
            //ln("CONV",string(t))
            log.Fatal("ERRR",err)
      //  year int, month Month, day, hour, min, sec, nsec int, loc *Location) Time
       // return time.Date(attr[0:3],attr[3:4], attr[4:5], attr[5:6], attr[6:7], attr[7:8], 0 , time.UTC)
        }
        return time.Unix(t, 0)
    }
    
    return time.Now()
        
}


func  Decompress(val []byte) ([]byte) {
    defer TimeTrack(time.Now(), "decompressing")
    b := bytes.NewReader(val)
	r, err := zlib.NewReader(b)
    var buf bytes.Buffer// buffer needs no initilization pointer
	if err != nil {
		panic(err)
	}
 
	io.Copy(&buf, r)
    fmt.Printf("data %s %d \n",buf.Bytes(), len(buf.Bytes()))
	r.Close()
  
    return buf.Bytes()
}

func  DecompressF(val []byte) ([]byte) {
    var m runtime.MemStats 
    defer TimeTrack(time.Now(), "decompressing")
    fmt.Printf("Decompressing %x:\n", val[0:5])
    b := bytes.NewReader(val)
	r := flate.NewReader(b)
    var buf bytes.Buffer// buffer needs no initilization pointer
   
  
	io.Copy(&buf, r)
   
	r.Close()
    runtime.ReadMemStats(&m)
    fmt.Printf("Asked %d,Allocated %d,unused %d, released %d,round %d\n", m.HeapSys, m.HeapAlloc,
            m.HeapIdle, m.HeapReleased)
  
    return buf.Bytes()
}

func Stringify(val []uint8) (string) {
    var str string
    for _, elem :=range val{
        if elem != 0 {
            str += string(elem)
        }else{
            str = fmt.Sprintf("%s", str)
            break
        }
            
    }
    return str
}

