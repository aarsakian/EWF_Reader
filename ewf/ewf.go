package ewf

import (
	"bytes"

	"github.com/aarsakian/EWF_Reader/ewf/sections"

	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aarsakian/EWF_Reader/ewf/utils"
)

const EWF_Section_Header_s int64 = 76
const EWF_Header_s int64 = 13
const NofSections = 200

type EWF_file struct {
	File       *os.File
	Size       uint64
	hasNext    bool
	isLast     bool
	SegmentNum uint
	Entries    []uint32
	Header     *EWF_Header
}

type EWF_Header struct {
	//header size 13bytes of each segment (file)
	Signature     [8]uint8 "EVF\x0d\x0a\xff\x00"
	StartofFields uint8    //
	SegNum        uint16   //
	EOF           uint16
}

func (ewf_file *EWF_file) Verify() {

}
func (ewf_file *EWF_file) ParseHeader(cur_offset *uint64) {
	defer utils.TimeTrack(time.Now(), "Parsing Segment Header")
	var ewf_header *EWF_Header = new(EWF_Header)

	buf := make([]byte, EWF_Header_s)
	ewf_file.File.ReadAt(buf, 0)
	utils.Umarshal(buf, ewf_header)

	sig := utils.Stringify(ewf_header.Signature[:])

	if !strings.Contains(sig, "EVF") {
		os.Exit(0)
	}

	ewf_file.Header = ewf_header

}

func (ewf_file *EWF_file) ParseSegment() {

	var Sections sections.Sections

	//var m runtime.MemStats
	//  var data interface{}
	//var sectors_offs uint64
	var buf []byte
	cur_offset := EWF_Header_s
	for cur_offset < int64(ewf_file.Size) {
		//   parsing section headers
		var section sections.Section = new(sections.Section)

		buf = make([]byte, EWF_Section_Header_s)
		ewf_file.File.ReadAt(buf, cur_offset) //read section header

		var s_descriptor *sections.Section_Descriptor = new(sections.Section_Descriptor)
		utils.Unmarshal(buf, s_descriptor)

		section.Descriptor = s_descriptor

		cur_offset = s_descriptor.NextSectionOffs

		Sections = append(Sections, section)
	}
	/*
		Sections[i] = new(sections.Section) //create section
		Sections[i].ParseHeader(buf)
		Sections[i].Dispatch() //object factory section body creation
		if Sections[i].Type == "next" {

			ewf_file.hasNext = true
			break
		} else if Sections[i].Type == "done" {

			ewf_file.isLast = true
			break
		}

		if Sections[i].Type != "sectors" {
			buf = ewf_file.ReadAt(Sections[i].BodyOffset-cur_offset, cur_offset) //read section body

		}

		Sections[i].ParseBody(buf)
		fmt.Println("finished ", Sections[i].Type,
			"KB Remaining",
			(ewf_file.Size-cur_offset)/1024, "OFFS", cur_offset, " length KB",
			(Sections[i].BodyOffset-cur_offset)/1024)
		cur_offset = Sections[i].BodyOffset
		runtime.ReadMemStats(&m)
		if Sections[i].Type == "table" {
			e := Sections[i].GetAttr("Table_entries").([]uint32)[:]

			ewf_file.Entries = parseutil.Append(ewf_file.Entries, e)

		}
		Sections[i].GetAttr("MD5_value")

		fmt.Printf("Asked %d,Allocated %d,unused %d, released %d,round %d\n", m.HeapSys, m.HeapAlloc,
			m.HeapIdle, m.HeapReleased, i)

	}*/
	//disk section and sectors section

}

func (ewf_file *EWF_file) ReadAt(length uint64, off uint64) *bytes.Reader {
	//cast to struct respecting endianess
	defer parseutil.TimeTrack(time.Now(), "reading")
	buff := make([]byte, length)
	var err error
	var n int
	//read 100KB chunks
	STEP := uint64(1000 * 1024)
	rem := length
	if length < STEP {
		_, err := ewf_file.File.ReadAt(buff, int64(off))
		if err == io.EOF {
			fmt.Println("Error reading file:", err)

		}
	} else {
		for i := uint64(0); i <= length; i += STEP {
			if rem < STEP { //final read
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
