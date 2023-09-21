package Utils

import (
	"bytes"
	"compress/flate"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
	// "io/ioutil"
)

type NoNull string

type Queue struct { //ring buffer
	Elements map[int]int
	tail     int
	head     int
	Capacity int
}

func (queue *Queue) EnQueue(element int) {
	queue.Elements[queue.tail] = element
	queue.tail = (queue.tail + 1) % queue.Capacity

}

func (queue Queue) IsEmpty() bool {
	return queue.head == queue.tail
}

func (queue Queue) IsFull() bool {
	return queue.head == (queue.tail+1)%queue.Capacity
}
func (queue *Queue) DeQueue() int {

	if !queue.IsEmpty() {
		element := queue.Elements[queue.head]
		queue.head = (queue.head + 1) % queue.Capacity

		return element
	} else {
		return 0
	}

}

func Parse(buf *bytes.Reader, val interface{}) {
	//target is val
	err := binary.Read(buf, binary.LittleEndian, val)

	if err != nil {
		fmt.Println("binary.Read failed: General", err)
	}

}

func Unmarshal(data []byte, v interface{}) error {
	idx := 0
	structValPtr := reflect.ValueOf(v)
	structType := reflect.TypeOf(v)
	if structType.Elem().Kind() != reflect.Struct {
		return errors.New("must be a struct")
	}
	for i := 0; i < structValPtr.Elem().NumField(); i++ {
		field := structValPtr.Elem().Field(i) //StructField type
		switch field.Kind() {
		case reflect.String:
			name := structType.Elem().Field(i).Name
			if name == "Signature" || name == "CollationSortingRule" {
				field.SetString(string(data[idx : idx+4]))
				idx += 4
			}
		case reflect.Uint8:
			var temp uint8
			binary.Read(bytes.NewBuffer(data[idx:idx+1]), binary.LittleEndian, &temp)
			field.SetUint(uint64(temp))
			idx += 1
		case reflect.Uint16:
			var temp uint16
			binary.Read(bytes.NewBuffer(data[idx:idx+2]), binary.LittleEndian, &temp)
			field.SetUint(uint64(temp))
			idx += 2
		case reflect.Int32:
			var temp int32
			binary.Read(bytes.NewBuffer(data[idx:idx+4]), binary.LittleEndian, &temp)
			field.SetInt(int64(temp))
			idx += 4
		case reflect.Uint32:
			var temp uint32
			binary.Read(bytes.NewBuffer(data[idx:idx+4]), binary.LittleEndian, &temp)
			field.SetUint(uint64(temp))
			idx += 4
		case reflect.Int64:
			var temp int64
			binary.Read(bytes.NewBuffer(data[idx:idx+8]), binary.LittleEndian, &temp)
			field.SetInt(temp)
			idx += 8
		case reflect.Uint64:
			var temp uint64
			name := structType.Elem().Field(i).Name
			if name == "ParRef" {
				buf := make([]byte, 8)
				copy(buf, data[idx:idx+6])
				binary.Read(bytes.NewBuffer(buf), binary.LittleEndian, &temp)
				idx += 6
			} else if name == "ChildVCN" {
				len := structValPtr.Elem().FieldByName("Len").Uint()
				binary.Read(bytes.NewBuffer(data[len-8:len]), binary.LittleEndian, &temp)

			} else {
				binary.Read(bytes.NewBuffer(data[idx:idx+8]), binary.LittleEndian, &temp)
				idx += 8
			}
			field.SetUint(temp)
		case reflect.Bool:
			field.SetBool(false)
			idx += 1
		case reflect.Array:
			arrT := reflect.ArrayOf(field.Len(), reflect.TypeOf(data[0])) //create array type to hold the slice
			arr := reflect.New(arrT).Elem()                               //initialize and access array
			var end int
			if idx+field.Len() > len(data) { //determine end
				end = len(data)
			} else {
				end = idx + field.Len()
			}
			for idx, val := range data[idx:end] {

				arr.Index(idx).Set(reflect.ValueOf(val))
			}

			field.Set(arr)
			idx += field.Len()

		}

	}
	return nil
}

func TimeTrack(start time.Time, name string) {

	elapsed := time.Since(start)

	log.Printf("%s took %s ", name, elapsed)
}

func Append(src []uint32, data []uint32) []uint32 {
	m := len(src)
	n := m + len(data)
	if n > cap(src) { //reallocated
		dst := make([]uint32, (n+1)*2)
		copy(dst, src)
		src = dst
	}
	src = src[0:n]
	copy(src[m:n], data)
	return src
}

func GetTime(attr []byte) time.Time {
	//  fmt.Println("TIME LEN",string(attr),attr,len(attr))
	timestamp := string(bytes.Join(bytes.Split(attr, []byte{00}), []byte{})) //crazy since original byte seq not in ASCII
	if len(attr) == 21 {
		t, err := strconv.ParseInt(timestamp, 10, 32)

		if err != nil {
			//ln("CONV",string(t))
			//log.Fatal("ERRR", err)
			fmt.Printf("error converting time %s", err)
			//  year int, month Month, day, hour, min, sec, nsec int, loc *Location) Time
			// return time.Date(attr[0:3],attr[3:4], attr[4:5], attr[5:6], attr[6:7], attr[7:8], 0 , time.UTC)
		}
		return time.Unix(t, 0)
	}

	return time.Now()

}

func ByteReverse(barray []byte) []byte { //work with indexes
	//  fmt.Println("before",barray)
	for i, j := 0, len(barray)-1; i < j; i, j = i+1, j-1 {

		barray[i], barray[j] = barray[j], barray[i]

	}
	return barray

}

func RemoveNulls(barray []byte) []byte {
	var nonullbarray []byte
	for _, v := range barray {
		if v == 0x00 {
			continue
		}

		nonullbarray = append(nonullbarray, v)

	}
	return nonullbarray
}

/*CMF|FLG  0x78|  (FLG|CM)
CM 0-3 Compression method  8=deflate
CINFO 4-7 Compression info 7=32K window size only when CM=8
FLG 0-4 FCHECK  = CMF*256 + FLG multiple of 31 = 120*256==x mod 31 => x=156
5 FDICT 1=> DICT follows (DICT is the Adler-32 checksum  of this sequence of bytes )
6-7 FLEVEL compression level 0-3
9c = 1001 1100
FLEVEL 10
FDICT 0
FCHECK 12
ADLER32  algorithm is a 32-bit extension and improvement of the Fletcher algorithm,
A compliant decompressor must check CMF, FLG, and ADLER32,
*/

func Decompress(val []byte) []byte {
	//	defer TimeTrack(time.Now(), "decompressing")
	var r io.ReadCloser
	var b *bytes.Reader
	var bytesRead, lent int64
	var dst bytes.Buffer
	var err error

	b = bytes.NewReader(val)

	r, err = zlib.NewReader(b)
	if err != nil {
		if err == io.EOF {
			fmt.Println(err)

		}

		log.Fatal(err)
	}

	defer r.Close()

	lent, err = dst.ReadFrom(r)
	bytesRead += lent
	//	fmt.Println(":EM", bytesRead, len(val), int(bytesRead) > len(val))
	if err != nil {
		//fmt.Println(err)
		log.Fatal(err)

	}

	//var buf bytes.Buffer // buffer needs no initilization pointer
	if err != nil {
		panic(err)
	}

	//io.Copy(&buf, r)

	if err != nil {
		panic(err)
	}

	//   fmt.Printf("data  %d \n", len(buf.Bytes()))

	return dst.Bytes()
}

func DecompressF(val []byte) []byte {
	var m runtime.MemStats
	defer TimeTrack(time.Now(), "decompressing")
	fmt.Printf("Decompressing %x:\n", val[0:5])
	buf := bytes.NewReader(val)
	r := flate.NewReader(buf)

	p, err := io.ReadAll(r)

	if err != nil {
		panic(err)
	}
	r.Close()
	runtime.ReadMemStats(&m)
	fmt.Printf("Asked %d,Allocated %d,unused %d, released %d,round %d\n", m.HeapSys, m.HeapAlloc,
		m.HeapIdle, m.HeapReleased)

	return p
}

func Stringify(val []uint8) string {
	var strBuilder strings.Builder
	strBuilder.Grow(len(val))

	for _, elem := range val {
		if elem == 0 {
			break
		}
		strBuilder.WriteString(string(elem))

	}
	return strBuilder.String()
}

func ReadEndianB(barray []byte) (val interface{}) {
	//conversion function
	//fmt.Println("before conversion----------------",barray)
	//fmt.Printf("len%d ",len(barray))

	switch len(barray) {
	case 8:
		var vale uint64
		binary.Read(bytes.NewBuffer(barray), binary.BigEndian, &vale)
		val = vale

	case 4:
		var vale uint32
		//   fmt.Println("barray",barray)
		binary.Read(bytes.NewBuffer(barray), binary.BigEndian, &vale)
		val = vale
		val = vale
	case 2:

		var vale uint16

		binary.Read(bytes.NewBuffer(barray), binary.BigEndian, &vale)
		//   fmt.Println("after conversion vale----------------",barray,vale)
		val = vale

	case 1:

		var vale uint8

		binary.Read(bytes.NewBuffer(barray), binary.BigEndian, &vale)
		//      fmt.Println("after conversion vale----------------",barray,vale)
		val = vale

	default: //best it would be nil
		var vale uint64

		binary.Read(bytes.NewBuffer(barray), binary.BigEndian, &vale)
		val = vale
	}
	return val
}

func ReadEndian(barray []byte) any {
	//conversion function
	//fmt.Println("before conversion----------------",barray)
	//fmt.Printf("len%d ",len(barray))

	switch len(barray) {
	case 8:
		var vale uint64
		binary.Read(bytes.NewBuffer(barray), binary.LittleEndian, &vale)
		return vale
	case 6:

		var vale uint32
		buf := make([]byte, 6)
		binary.Read(bytes.NewBuffer(barray[:4]), binary.LittleEndian, &vale)
		var vale1 uint16
		binary.Read(bytes.NewBuffer(barray[4:]), binary.LittleEndian, &vale1)
		binary.LittleEndian.PutUint32(buf[:4], vale)
		binary.LittleEndian.PutUint16(buf[4:], vale1)
		val, _ := binary.ReadUvarint(bytes.NewBuffer(buf))
		return val

	case 4:
		var vale uint32
		//   fmt.Println("barray",barray)
		binary.Read(bytes.NewBuffer(barray), binary.LittleEndian, &vale)
		return vale
	case 2:

		var vale uint16

		binary.Read(bytes.NewBuffer(barray), binary.LittleEndian, &vale)
		//   fmt.Println("after conversion vale----------------",barray,vale)
		return vale

	case 1:

		var vale uint8

		binary.Read(bytes.NewBuffer(barray), binary.LittleEndian, &vale)
		//      fmt.Println("after conversion vale----------------",barray,vale)
		return vale

	default: //best it would be nil
		var vale uint64

		binary.Read(bytes.NewBuffer(barray), binary.LittleEndian, &vale)
		return vale
	}
}

func ToMap[K comparable, T any](keys []K, vals []T) map[K]T {
	m := map[K]T{}

	for idx, key := range keys {
		m[key] = vals[idx]
	}
	return m
}

func FindEvidenceFiles(path_ string) []string {

	basePath := filepath.Dir(path_)

	Files, err := ioutil.ReadDir(basePath)
	if err != nil {
		log.Fatal("ERR", err)
	}

	var filenames []string
	r, _ := regexp.Compile("\\.[eE][a-zA-Z0-9]{1,2}$")
	for _, finfo := range Files {

		if finfo.IsDir() {

			continue
		}

		if r.MatchString(finfo.Name()) {

			filenames = append(filenames, filepath.Join(basePath, finfo.Name())) //supply channel
			//fmt.Println("INFO", basePath+finfo.Name(), strings.Split(fname, ".")[0])

		}
	}

	return filenames

}
