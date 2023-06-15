package ewf

import (
	"bytes"
	"errors"
	"fmt"
	"hash/adler32"
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
	Name            string
	Handler         *os.File
	Size            int64
	hasNext         bool
	isLast          bool
	Header          *EWF_Header
	Sections        *Sections
	FirstChunckId   int
	NumberOfChuncks uint32
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
		return "error hash section not found"
	}

}

func (ewf_file EWF_file) CollectData(buffer *bytes.Buffer) {
	table_sections := ewf_file.Sections.Filter("table")
	ewf_file.CreateHandler()
	defer ewf_file.CloseHandler()
	var to uint64
	var buf []byte
	for _, table_section := range table_sections {
		table_entries := table_section.GetAttr("Table_entries").(sections.Table_Entries)
		nofTable_entries := int(table_section.GetAttr("Table_header").(*sections.EWF_Table_Section_Header).NofEntries)
		for idx, chunck := range table_entries {

			if idx == nofTable_entries-1 { // last entry

				to = uint64(table_section.prev.Descriptor.NextSectionOffs)

			} else {
				to = uint64(table_entries[idx+1].DataOffset)
			}

			from := uint64(chunck.DataOffset)

			buf = ewf_file.ReadAt(int64(chunck.DataOffset), to-from)
			if chunck.IsCompressed {

				buf = utils.Decompress(buf)
			} else {
				buf = buf[:len(buf)-4] //last 4 bytes checksum
			}

			buffer.Write(buf)

		}

	}

}

func (ewf_file EWF_file) Verify(chunk_size int) bool {
	table_sections := ewf_file.Sections.Filter("table")

	ewf_file.CreateHandler()
	defer ewf_file.CloseHandler()
	var to, from uint64
	var buf []byte
	deflated_data := make([]byte, chunk_size)
	for _, table_section := range table_sections {

		table_entries := table_section.GetAttr("Table_entries").(sections.Table_Entries)
		for idx, chunck := range table_entries {
			nofTable_entries := int(table_section.GetAttr("Table_header").(*sections.EWF_Table_Section_Header).NofEntries)
			if idx == nofTable_entries-1 { // last entry
				//working with previous section since offsets refer to sectors section

				to = uint64(table_section.prev.Descriptor.NextSectionOffs)
			} else {
				to = uint64(table_entries[idx+1].DataOffset)

			}
			from = uint64(chunck.DataOffset)

			buf = ewf_file.ReadAt(int64(chunck.DataOffset), to-from)

			if !chunck.IsCompressed {
				continue
			}
			deflated_data = utils.Decompress(buf)

			if utils.ReadEndianB(buf[len(buf)-4:]) != adler32.Checksum(deflated_data) {
				fmt.Println("problematic chunck", idx, to, from, ewf_file.Name, chunck.DataOffset, adler32.Checksum(deflated_data), utils.ReadEndianB(buf[len(buf)-4:]))
				return false
			}

		}

	}
	return true
}

func (ewf_file EWF_file) GetChunckInfo() (uint64, uint64, uint64, uint64, error) {
	var section *Section
	section = ewf_file.Sections.GetSectionPtr("volume")
	if section == nil { // alternative search for disk
		section = ewf_file.Sections.GetSectionPtr("disk")
	}

	if section != nil {
		chunkCount := section.GetAttr("ChunkCount").(uint64)

		nofSectorPerChunk := section.GetAttr("NofSectorPerChunk").(uint64)
		nofBytesPerSector := section.GetAttr("NofBytesPerSector").(uint64)
		nofSectors := section.GetAttr("NofSectors").(uint64)
		return chunkCount, nofSectorPerChunk, nofBytesPerSector, nofSectors, nil
	} else {
		return 0, 0, 0, 0, errors.New("section not found")
	}

}

func (ewf_file EWF_file) GetChunck(chunck_id int) sections.EWF_Table_Section_Entry {
	tableSections := ewf_file.Sections.Filter("table")

	chunck_cnt := 0
	for _, table := range tableSections {
		for _, chunck := range table.GetAttr("Table_entries").(sections.Table_Entries) {
			if chunck_id == chunck_cnt {
				return chunck
			}
			chunck_cnt += 1
		}
	}
	return sections.EWF_Table_Section_Entry{}
}

func (ewf_file EWF_file) PopulateChunckOffsets(chunckOffsetsPtrs sections.Table_EntriesPtrs) int {
	tableSections := ewf_file.Sections.Filter("table")
	pos := 0
	for _, section := range tableSections {
		chuncks := section.GetAttr("Table_entries").(sections.Table_Entries)
		for id := range chuncks {
			chunckOffsetsPtrs[pos] = &chuncks[id]
			pos++
		}

	}
	return pos
}

func (ewf_file EWF_file) GetTotalNofChuncks() []int64 {
	var lastOffsets []int64
	tableSections := ewf_file.Sections.Filter("table")
	for _, section := range tableSections {
		section.GetAttr("Table_entries")
	}
	return lastOffsets
}

func (ewf_file EWF_file) IsLast() bool {
	return ewf_file.Sections.tail.Type == "done"
}

func (ewf_file EWF_file) IsValid() bool {
	sig := utils.Stringify(ewf_file.Header.Signature[:])

	return strings.Contains(sig, "EVF")

}

func (ewf_file EWF_file) IsFirst() bool {
	return ewf_file.Header.SegNum == 1
}

func (ewf_file *EWF_file) ParseHeader() {
	defer utils.TimeTrack(time.Now(), "Parsing Segment Header")
	var ewf_header *EWF_Header = new(EWF_Header)

	buf := ewf_file.ReadAt(0, EWF_Header_s)
	utils.Unmarshal(buf, ewf_header)

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

		section.DescriptorCalculatedChecksum = adler32.Checksum(buf[:len(buf)-4])

		section.Descriptor = s_descriptor

		section.Type = s_descriptor.GetType()

		buf = ewf_file.ReadAt(cur_offset+EWF_Section_Header_s,
			s_descriptor.SectionSize-EWF_Section_Header_s) //read section body
		if section.Type != "sectors" {
			section.ParseBody(buf)

		}

		if ewf_sections.head == nil {
			ewf_sections.head = section
		}

		if prev_section != nil {

			prev_section.next = section
			section.prev = prev_section

		}

		cur_offset = s_descriptor.NextSectionOffs
		fmt.Println(section.Type, cur_offset)
		if section.Type == "done" || section.Type == "next" {
			ewf_sections.tail = section
			break
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

func (ewf_file EWF_file) LocateData(chuncksPtrs sections.Table_EntriesPtrs, buf *bytes.Buffer) {
	ewf_file.CreateHandler()
	defer ewf_file.CloseHandler()
	for idx, chunck := range chuncksPtrs {
		if idx == len(chuncksPtrs)-1 { // last chunck not part of asked chunck range
			break
		}
		to := chuncksPtrs[idx+1].DataOffset
		from := chunck.DataOffset
		data := ewf_file.ReadAt(int64(from), uint64(to-from))
		if chunck.IsCompressed {
			data = utils.Decompress(data)
		}

		remainingSpace := buf.Cap() - buf.Len()
		if len(data) > remainingSpace { // user asked a size less than the last chunck
			buf.Write(data[:remainingSpace])
			break
		}
		buf.Write(data)

	}

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
