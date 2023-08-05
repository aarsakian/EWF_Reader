package sections

import (
	"bytes"
	"fmt"
	"hash/adler32"
	"reflect"
	"time"

	"github.com/aarsakian/EWF_Reader/ewf/utils"
)

const Chunk_Size uint32 = 64 * 512

type EWF_Table_Section_Entry struct {
	DataOffset   uint32 "MSB indicates if chunk data is (un)compressed (0)/1 offset relative to the start of the fileit is located in the preseding sectors section "
	IsCompressed bool   "1 -> Compressed"
}

type EWF_Table_Section_Footer struct {
	Checksum [4]uint8
}

// not used in EnCase
type EWF_Table_Section_Data struct {
	Data []byte "chunk data compressed under deflate as well as its Checksum"
}

// resides right after table section
type EWF_Table2_Section struct {
	Table_section EWF_Table_Section
}

type Table_Entries []EWF_Table_Section_Entry

type Table_EntriesPtrs []*EWF_Table_Section_Entry

// table section  identifier
type EWF_Table_Section struct {
	Table_header       *EWF_Table_Section_Header
	Table_entries      Table_Entries
	Table_footer       *EWF_Table_Section_Footer
	calculatedChecksum uint32
}

type EWF_Table_Section_Header struct { //24 bytes
	NofEntries uint32    "Number of Entries 0x01"
	Padding    [16]uint8 "contains 0x00"
	Checksum   [4]uint8  "Adler32"
}

// EnCase 6-7
type EWF_Table_Section_Header_EnCase struct { //24bytes
	nofEntries      uint32   "Number of Entries 0x01"
	Padding1        [4]uint8 "contains 0x00"
	TableBaseOffset uint64   "Adler32"
	Padding2        [4]uint8 "contains 0x00"
	Checksum        [4]uint8 "Adler32"
}

func (table_header *EWF_Table_Section_Header) Parse(buf []byte) {

	utils.Unmarshal(buf, table_header)

}

func (table_entry *EWF_Table_Section_Entry) Parse(buf []byte) {

	IsCompressed := buf[3]&0x80 == 0x80
	buf[3] &= 0x7F //exlude MSB
	utils.Unmarshal(buf, table_entry)
	table_entry.IsCompressed = IsCompressed
}

func (table_footer *EWF_Table_Section_Footer) Parse(buf []byte) {

	//parse struct attributes

	utils.Unmarshal(buf, table_footer)

}

func (ewf_table_section EWF_Table_Section) Verify() bool {
	return ewf_table_section.calculatedChecksum == uint32(utils.ReadEndian(ewf_table_section.Table_footer.Checksum[:]).(uint32))
}

func (ewf_table_section *EWF_Table_Section) Parse(buf []byte) {

	defer utils.TimeTrack(time.Now(), "Parsing")
	var table_header *EWF_Table_Section_Header = new(EWF_Table_Section_Header)
	table_header.Parse(buf[:24])

	ewf_table_section.Table_header = table_header

	var table_footer *EWF_Table_Section_Footer = new(EWF_Table_Section_Footer)
	table_footer.Parse(buf[len(buf)-4:])
	ewf_table_section.Table_footer = table_footer

	buf = buf[24 : len(buf)-4]
	ewf_table_section.calculatedChecksum = adler32.Checksum(buf)

	var ewf_table_section_entries []EWF_Table_Section_Entry
	for i := uint32(0); i < ewf_table_section.Table_header.NofEntries; i += 1 {
		var ewf_table_section_entry *EWF_Table_Section_Entry = new(EWF_Table_Section_Entry)
		ewf_table_section_entry.Parse(buf[4*i : 4+4*i])

		ewf_table_section_entries = append(ewf_table_section_entries, *ewf_table_section_entry)

	}

	ewf_table_section.Table_entries = ewf_table_section_entries

}

func (ewf_table2_section *EWF_Table2_Section) Parse(buf []byte) {

}

func (ewf_table2_section *EWF_Table2_Section) Collect([]byte, uint64) {

}

func (ewf_table_section *EWF_Table_Section) Collect(sectors_buf []byte, sectors_offs uint64) {
	fmt.Println("NODF entries",
		len(ewf_table_section.Table_entries), ewf_table_section.Table_header.NofEntries)
	zlib_header := []byte{72, 13}
	var data []byte
	for idx, entry := range ewf_table_section.Table_entries[:len(ewf_table_section.Table_entries)-1] {

		data = sectors_buf[entry.DataOffset-uint32(sectors_offs) : entry.DataOffset-uint32(sectors_offs)+Chunk_Size]

		if bytes.HasPrefix(data, zlib_header) {
			utils.Decompress(data)
			fmt.Println("IDX", idx)
			/*sectors_buf[entry.ChunkDataOffset-uint32(sectors_offs):entry.ChunkDataOffset-uint32(sectors_offs)+5],
			  "REM",uint32(len(sectors_buf))-entry.ChunkDataOffset-uint32(sectors_offs), "CompresseD?",entry.IsCompressed)*/
			//  utils.DecompressF(data)
		}

	}
	//last data chunk maybe less than 32K size
	last_entry := ewf_table_section.Table_entries[len(ewf_table_section.Table_entries)-1]
	data = sectors_buf[last_entry.DataOffset-uint32(sectors_offs) : last_entry.DataOffset-uint32(sectors_offs)+
		uint32(len(sectors_buf))-last_entry.DataOffset-uint32(sectors_offs)]
	if bytes.HasSuffix(data, zlib_header) {
		utils.DecompressF(data)
	}
}

func (entry *EWF_Table_Section_Entry) GetAttr(attr string) interface{} {

	return reflect.ValueOf(entry).Elem().FieldByName(attr).Interface()
}

func (ewf_table_section *EWF_Table_Section) GetAttr(attr string) interface{} {
	s := reflect.ValueOf(ewf_table_section).Elem() //retrieve since it's a pointer

	return s.FieldByName(attr).Interface()

	/*	if attr == "Table_entries" {
			data_offsets := make([]uint32, sub_s.Len())
			for entry_number := 0; entry_number < sub_s.Len(); entry_number++ {
				s_inner := sub_s.Index(entry_number).Addr()
				// get_attr_f := s_inner.MethodByName("GetAttr")
				// fmt.Println("OFFSET",s_inner,get_attr_f.Call( [] reflect.Value{reflect.ValueOf("ChunkDataOffset")}))

				data_offsets[entry_number] = s_inner.Elem().FieldByName("ChunkDataOffset").Interface().(uint32)
			}
			return data_offsets
		} else {
			return ""
		}*/

}

func (ewf_table2_section *EWF_Table2_Section) GetAttr(string) interface{} {
	return ""
}
