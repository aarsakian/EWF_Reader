package ewf

import (
	"bytes"
	"errors"
	"fmt"
	"hash/adler32"
	"io"
	"os"
	"strings"

	"github.com/aarsakian/EWF_Reader/ewf/sections"
	Utils "github.com/aarsakian/EWF_Reader/ewf/utils"
	"github.com/aarsakian/EWF_Reader/logger"
)

const EWF_Section_Header_s = 76
const EWF_Header_s = 13
const NofSections = 200

type EWF_files []EWF_file

type EWF_file struct {
	Name         string
	Handler      *os.File
	Size         int64
	hasNext      bool
	isLast       bool
	Header       *EWF_Header
	Sections     *Sections
	FirstchunkId int
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
		nofTable_entries := int(table_section.GetAttr("Table_header").(*sections.EWF_Table_Section_Header_EnCase).NofEntries)
		for idx, chunk := range table_entries {

			if idx == nofTable_entries-1 { // last entry

				to = uint64(table_section.prev.Descriptor.NextSectionOffs)

			} else {
				to = uint64(table_entries[idx+1].DataOffset)
			}

			from := uint64(chunk.DataOffset)

			if to < from { //reached end of ewf_file
				logger.EWF_Readerlogger.Info(fmt.Sprintf("Reading at %d len %d chunk pos %d",
					chunk.DataOffset, buffer.Len(), idx))
				buf = ewf_file.ReadAt(int64(chunk.DataOffset), uint64(buffer.Len()))
			} else {
				logger.EWF_Readerlogger.Info(fmt.Sprintf("Reading at %d len %d chunk pos %d",
					chunk.DataOffset, to-from, idx))
				buf = ewf_file.ReadAt(int64(chunk.DataOffset), uint64(to-from))
			}

			if chunk.IsCompressed {

				buf = Utils.Decompress(buf)
			} else {
				buf = buf[:len(buf)-4] //last 4 bytes checksum
			}
			logger.EWF_Readerlogger.Info(fmt.Sprintf("Writing to buffer %d ", len(buf)))
			buffer.Write(buf)

		}

	}

}

func (ewf_file EWF_file) Verify(chunk_size int) bool {
	fmt.Printf("Verifying %s\n", ewf_file.Name)
	table_sections := ewf_file.Sections.Filter("table")

	ewf_file.CreateHandler()
	defer ewf_file.CloseHandler()
	var to, from uint64
	var buf []byte
	var deflated_data []byte
	for _, table_section := range table_sections {

		table_entries := table_section.GetAttr("Table_entries").(sections.Table_Entries)
		for idx, chunk := range table_entries {
			nofTable_entries := int(table_section.GetAttr("Table_header").(*sections.EWF_Table_Section_Header_EnCase).NofEntries)
			if idx == nofTable_entries-1 { // last entry
				//working with previous section since offsets refer to sectors section

				to = uint64(table_section.prev.Descriptor.NextSectionOffs)
			} else {
				to = uint64(table_entries[idx+1].DataOffset)

			}
			from = uint64(chunk.DataOffset)

			if to < from { //reached end of ewf_file
				logger.EWF_Readerlogger.Info(fmt.Sprintf("Reading at %d len %d",
					chunk.DataOffset, chunk_size))
				buf = ewf_file.ReadAt(int64(chunk.DataOffset), uint64(chunk_size))
			} else {
				logger.EWF_Readerlogger.Info(fmt.Sprintf("Reading at %d len %d",
					chunk.DataOffset, to-from))
				buf = ewf_file.ReadAt(int64(chunk.DataOffset), uint64(to-from))
			}

			if !chunk.IsCompressed {
				continue
			}
			deflated_data = Utils.Decompress(buf)

			if Utils.ReadEndianB(buf[len(buf)-4:]) != adler32.Checksum(deflated_data) {
				fmt.Println("problematic chunk", idx, to, from, ewf_file.Name, chunk.DataOffset, adler32.Checksum(deflated_data), Utils.ReadEndianB(buf[len(buf)-4:]))
				return false
			}

		}

	}
	return true
}

func (ewf_file EWF_file) GetchunkInfo() (uint64, uint64, uint64, uint64, string, error) {
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

		section = ewf_file.Sections.GetSectionPtr("header")
		version := section.GetAttr("Version").(string)
		section = ewf_file.Sections.GetSectionPtr("header2")
		software := ""
		if section != nil {
			software = section.GetAttr("AN unknown").(string)
		}

		return chunkCount, nofSectorPerChunk, nofBytesPerSector, nofSectors, fmt.Sprintf("%s %s", software, version), nil
	} else {
		return 0, 0, 0, 0, "-", errors.New("section not found")
	}

}

func (ewf_file EWF_file) Getchunk(chunk_id int) sections.EWF_Table_Section_Entry {
	tableSections := ewf_file.Sections.Filter("table")

	chunk_cnt := 0
	for _, table := range tableSections {
		for _, chunk := range table.GetAttr("Table_entries").(sections.Table_Entries) {
			if chunk_id == chunk_cnt {
				return chunk
			}
			chunk_cnt += 1
		}
	}
	return sections.EWF_Table_Section_Entry{}
}

func (ewf_file EWF_file) PopulatechunkOffsets(chunkOffsetsPtrs sections.Table_EntriesPtrs, pos int) int {
	tableSections := ewf_file.Sections.Filter("table")
	//fmt.Printf("nof chunks: \n")

	for _, section := range tableSections {
		chunks := section.GetAttr("Table_entries").(sections.Table_Entries)
		for _, chunk := range chunks {
			current := chunk // change pointer address
			chunkOffsetsPtrs[pos] = &current
			pos++
		}
		//	fmt.Printf("%d \t", len(chunks))

	}
	//	fmt.Printf("nof table sections %d curPos %d \n", len(tableSections), pos)
	return pos

}

func (ewf_file EWF_file) GetTotalNofchunks() []int64 {
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
	sig := Utils.Stringify(ewf_file.Header.Signature[:])

	return strings.Contains(sig, "EVF")

}

func (ewf_file EWF_file) IsFirst() bool {
	return ewf_file.Header.SegNum == 1
}

func (ewf_file *EWF_file) ParseHeader() {

	var ewf_header *EWF_Header = new(EWF_Header)

	buf := ewf_file.ReadAt(0, EWF_Header_s)
	Utils.Unmarshal(buf, ewf_header)

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

		buf = ewf_file.ReadAt(cur_offset, EWF_Section_Header_s) //read section header 76 bytes

		var s_descriptor *Section_Descriptor = new(Section_Descriptor)
		Utils.Unmarshal(buf, s_descriptor)

		section.DescriptorCalculatedChecksum = adler32.Checksum(buf[:len(buf)-4])

		section.Descriptor = s_descriptor

		section.Type = s_descriptor.GetType()
		//	fmt.Println(section.Type, cur_offset)
		if !s_descriptor.IsBodyEmpty() && section.Type != "sectors" {
			buf = ewf_file.ReadAt(cur_offset+EWF_Section_Header_s,
				s_descriptor.SectionSize-EWF_Section_Header_s) //read section body

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

			ewf_file.Entries = Utils.Append(ewf_file.Entries, e)

		}
		Sections[i].GetAttr("MD5_value")

		fmt.Printf("Asked %d,Allocated %d,unused %d, released %d,round %d", m.HeapSys, m.HeapAlloc,
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

func (ewf_file EWF_file) LocateDataAlt(chunks sections.Table_EntriesPtrs, from_offset int64,
	dataLen int, buf *bytes.Buffer, chunk_size int) {
	ewf_file.CreateHandler()
	defer ewf_file.CloseHandler()
	relativeOffset := int(from_offset)
	chunkDataCH := make(chan Utils.ChunkData, 10000)
	resultsCH := make(chan Utils.Result, 10000)
	done := make(chan bool)
	output := make([][]byte, len(chunks)-1)
	go func() {
		for idx, chunk := range chunks {
			if idx == len(chunks)-1 { //  last chunk not part of asked chunk range
				break
			}

			if chunk.IsCached {
				chunkDataCH <- Utils.ChunkData{Data: chunk.DataChuck.Data, IsCompressed: chunk.IsCompressed, Id: idx}
			} else {
				to := chunks[idx+1].DataOffset
				from := chunk.DataOffset

				if to < from { //reached end of ewf_file

					logger.EWF_Readerlogger.Info(fmt.Sprintf("Reading at %d len %d chunk pos %d chan len %d",
						chunk.DataOffset, chunk_size, idx, len(chunkDataCH)))

					chunkDataCH <- Utils.ChunkData{Data: ewf_file.ReadAt(int64(from), uint64(chunk_size)), IsCompressed: chunk.IsCompressed, Id: idx}
				} else {
					logger.EWF_Readerlogger.Info(fmt.Sprintf("Reading at %d len %d chunk pos %d chan len %d",
						chunk.DataOffset, to-from, idx, len(chunkDataCH)))

					chunkDataCH <- Utils.ChunkData{Data: ewf_file.ReadAt(int64(from), uint64(to-from)), IsCompressed: chunk.IsCompressed, Id: idx}
				}

			}
		}
		close(chunkDataCH)
	}()

	go func() {

		Utils.DecompressCH(chunkDataCH, resultsCH, done)

	}()

	go func() {

		for result := range resultsCH {
			output[result.Id] = result.Data
		}
		done <- true
	}()

	<-done

	for _, data := range output {

		remainingSpace := dataLen - buf.Len() // free buffer size
		logger.EWF_Readerlogger.Info(fmt.Sprintf("Remaining space %d", remainingSpace))

		data = data[:chunk_size]

		if len(data) > remainingSpace && remainingSpace+relativeOffset < len(data) { // user asked a size less than the last chunk
			buf.Write(data[relativeOffset : relativeOffset+remainingSpace])
			break
		}
		buf.Write(data[relativeOffset:])
		relativeOffset = 0
	}

}

func (ewf_file EWF_file) LocateData(chunks sections.Table_EntriesPtrs, from_offset int64,
	dataLen int, buf *bytes.Buffer, chunk_size int) {
	ewf_file.CreateHandler()
	defer ewf_file.CloseHandler()
	relativeOffset := int(from_offset)
	var data []byte
	for idx, chunk := range chunks {
		if idx == len(chunks)-1 { //  last chunk not part of asked chunk range
			break
		}

		if chunk.IsCached {
			data = chunk.DataChuck.Data
		} else {
			to := chunks[idx+1].DataOffset
			from := chunk.DataOffset

			if to < from { //reached end of ewf_file
				logger.EWF_Readerlogger.Info(fmt.Sprintf("Reading at %d len %d chunk pos %d",
					chunk.DataOffset, chunk_size, idx))
				data = ewf_file.ReadAt(int64(from), uint64(chunk_size))
			} else {
				logger.EWF_Readerlogger.Info(fmt.Sprintf("Reading at %d len %d chunk pos %d",
					chunk.DataOffset, to-from, idx))
				data = ewf_file.ReadAt(int64(from), uint64(to-from))
			}

			if chunk.IsCompressed {
				data = Utils.Decompress(data)
			}
			data = data[:chunk_size] // when checksum is included real size is chunk_size +4
		}

		//fmt.Printf("%s %d \t,", data[0:4], idx)
		remainingSpace := dataLen - buf.Len() // free buffer size
		logger.EWF_Readerlogger.Info(fmt.Sprintf("Remaining space %d", remainingSpace))
		if len(data) > remainingSpace && remainingSpace+relativeOffset < len(data) { // user asked a size less than the last chunk
			buf.Write(data[relativeOffset : relativeOffset+remainingSpace])
			break
		}
		buf.Write(data[relativeOffset:])
		relativeOffset = 0

	}

}

func (ewf_file EWF_file) ReadAt(off int64, length uint64) []byte {
	//cast to struct respecting endianess
	//defer Utils.TimeTrack(time.Now(), "reading")
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
