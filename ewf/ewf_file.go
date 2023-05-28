package ewf

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/aarsakian/EWF_Reader/ewf/sections"
	"github.com/aarsakian/EWF_Reader/ewf/utils"
)

const EWF_Section_Header_s = 76
const EWF_Header_s = 13
const NofSections = 200

type EWF_files []EWF_file

type EWF_file struct {
	Name       string
	Handler    *os.File
	Size       int64
	hasNext    bool
	isLast     bool
	SegmentNum uint
	Entries    []uint32
	Header     *EWF_Header
	Sections   *Sections
}

type EWF_Header struct {
	//header size 13bytes of each segment (file)
	Signature     [8]uint8 "EVF\x0d\x0a\xff\x00"
	StartofFields uint8    //
	SegNum        uint16   //
	EOF           uint16
}

func (ewf_file EWF_file) GetHash() string {
	section := ewf_file.Sections.GetSectionPtr("hash")
	if section != nil {
		return section.GetAttr("MD5").(string)
	} else {
		return "error"
	}

}

func (ewf_file EWF_file) GetVolInfo() string {
	section := ewf_file.Sections.GetSectionPtr("volume")
	if section != nil {
		chunkCount := section.GetAttr("ChunkCount").(uint64)
		nofSectorPerChunk := section.GetAttr("NofSectorPerChunk").(uint64)
		nofBytesPerSector := section.GetAttr("NofBytesPerSector").(uint64)
		nofSectors := section.GetAttr("NofSectors").(uint64)
		return fmt.Sprintf("chunck count %d nof Sectors Per chunck %d nof Bytes Per Sector %d nof Sectors %d",
			chunkCount, nofSectorPerChunk, nofBytesPerSector, nofSectors)
	} else {
		return "error"
	}
}
func (ewf_file EWF_file) GetChunckOffsets(chunkOffsets sections.Table_Entries) sections.Table_Entries {
	tableSections := ewf_file.Sections.Filter("table")
	for _, section := range tableSections {
		chunkOffsets = append(chunkOffsets, section.GetAttr("Table_entries").(sections.Table_Entries)...)

	}

	return chunkOffsets
}

func (ewf_file *EWF_file) ParseHeader() {
	defer utils.TimeTrack(time.Now(), "Parsing Segment Header")
	var ewf_header *EWF_Header = new(EWF_Header)

	buf := ewf_file.ReadAt(0, EWF_Header_s)
	utils.Unmarshal(buf, ewf_header)

	sig := utils.Stringify(ewf_header.Signature[:])

	if !strings.Contains(sig, "EVF") {
		os.Exit(0)
	}

	ewf_file.Header = ewf_header

}

func (ewf_file *EWF_file) ParseSegment() {

	var ewf_sections Sections

	//var m runtime.MemStats
	//  var data interface{}
	//var sectors_offs uint64
	var buf []byte
	cur_offset := int64(EWF_Header_s)
	var prev_section *Section
	for cur_offset < ewf_file.Size {
		//   parsing section headers
		var section *Section = new(Section)

		buf = ewf_file.ReadAt(cur_offset, EWF_Section_Header_s) //read section header

		var s_descriptor *Section_Descriptor = new(Section_Descriptor)
		utils.Unmarshal(buf, s_descriptor)

		section.Descriptor = s_descriptor

		section.Type = s_descriptor.GetType()

		buf = ewf_file.ReadAt(cur_offset+EWF_Section_Header_s,
			s_descriptor.SectionSize-EWF_Section_Header_s) //read section body
		if section.Type != "sectors" {
			section.ParseBody(buf)

		}

		if prev_section != nil {

			prev_section.next = section
			section.prev = prev_section

		}

		cur_offset = s_descriptor.NextSectionOffs

		if section.Type == "done" {
			ewf_sections.tail = section
			break
		} else if section.Type == "header" {
			ewf_sections.head = section
		}

		prev_section = section

	}
	ewf_file.Sections = &ewf_sections
	/*

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

			ewf_file.Entries = utils.Append(ewf_file.Entries, e)

		}
		Sections[i].GetAttr("MD5_value")

		fmt.Printf("Asked %d,Allocated %d,unused %d, released %d,round %d\n", m.HeapSys, m.HeapAlloc,
			m.HeapIdle, m.HeapReleased, i)

	}*/
	//disk section and sectors section

}

func (ewf_file *EWF_file) CreateHandler() {
	var err error

	var file *os.File
	file, err = os.Open(ewf_file.Name)
	ewf_file.Handler = file
	if err != nil {
		fmt.Println("Error opening  file:", err)
	}

	fs, err := file.Stat() //file descriptor
	if err != nil {
		fmt.Println("Error stat file:", err)
	}
	ewf_file.Size = fs.Size()

}

func (ewf_file *EWF_file) CloseHandler() {
	ewf_file.Handler.Close()
	ewf_file.Handler = nil
	ewf_file.Size = -1

}

func (ewf_file EWF_file) ReadAt(off int64, length uint64) []byte {
	//cast to struct respecting endianess
	//defer utils.TimeTrack(time.Now(), "reading")
	var err error

	buff := make([]byte, length)

	_, err = ewf_file.Handler.ReadAt(buff, off)
	if err == io.EOF {
		fmt.Println("Error reading file:", err)

	} else if err != nil {
		fmt.Println(err)
		panic(err)
	}

	return buff
}
