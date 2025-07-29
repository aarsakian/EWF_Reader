package sections

import (
	"bytes"
	"fmt"
	"hash/adler32"
	"reflect"

	Utils "github.com/aarsakian/EWF_Reader/ewf/utils"
)

/*
resides after volume in the first segment or after the file header in other segments
*/

const Chunk_Size uint32 = 64 * 512

/*
	DataOffset points to  Encase2-5 from beginning of the segment

EnCase 6 DataOffset points from beginning of the table base offset.
*/
type EWF_Table_Section_Entry struct {
	DataOffset   uint64 "MSB indicates if chunk data is (un)compressed (0)/1 offset relative to the start of the fileit is located in the preseding sectors section "
	IsCompressed bool   "1 -> Compressed"
	IsCached     bool
	To           uint64
	DataChuck    *DataChuck
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
	Table_header       *EWF_Table_Section_Header_EnCase
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
	NofEntries      uint32   "Number of Entries 0x01"
	Padding1        [4]uint8 "contains 0x00"
	TableBaseOffset uint64   "Adler32"
	Padding2        [4]uint8 "contains 0x00"
	Checksum        [4]uint8 "Adler32"
}

func (table_header *EWF_Table_Section_Header) Parse(buf []byte) {

	Utils.Unmarshal(buf, table_header)

}

func (table_header_encase *EWF_Table_Section_Header_EnCase) Parse(buf []byte) {

	Utils.Unmarshal(buf, table_header_encase)

}

func (table_entry *EWF_Table_Section_Entry) Parse(buf []byte, table_base_offset uint64) {

	table_entry.IsCompressed = buf[3]&0x80 == 0x80
	table_entry.IsCached = false
	buf[3] &= 0x7F //exlude MSB
	table_entry.DataOffset = uint64(Utils.ReadEndian(buf).(uint32)) + table_base_offset

}

func (table_footer *EWF_Table_Section_Footer) Parse(buf []byte) {

	//parse struct attributes

	Utils.Unmarshal(buf, table_footer)

}

func (ewf_table_section EWF_Table_Section) Verify() bool {
	return ewf_table_section.calculatedChecksum == uint32(Utils.ReadEndian(ewf_table_section.Table_footer.Checksum[:]).(uint32))
}

func (ewf_table_section *EWF_Table_Section) Parse(buf []byte) {

	//defer Utils.TimeTrack(time.Now(), "Parsing Table Section")
	var table_header_encase *EWF_Table_Section_Header_EnCase = new(EWF_Table_Section_Header_EnCase)
	table_header_encase.Parse(buf[:24])

	ewf_table_section.Table_header = table_header_encase

	var table_footer *EWF_Table_Section_Footer = new(EWF_Table_Section_Footer)
	table_footer.Parse(buf[len(buf)-4:])
	ewf_table_section.Table_footer = table_footer

	buf = buf[24 : len(buf)-4]
	ewf_table_section.calculatedChecksum = adler32.Checksum(buf)

	ewf_table_section.Table_entries = make([]EWF_Table_Section_Entry, ewf_table_section.Table_header.NofEntries)
	for i := uint32(0); i < ewf_table_section.Table_header.NofEntries; i += 1 {
		var ewf_table_section_entry *EWF_Table_Section_Entry = new(EWF_Table_Section_Entry)
		ewf_table_section_entry.Parse(buf[4*i:4+4*i], ewf_table_section.Table_header.TableBaseOffset)

		ewf_table_section.Table_entries[i] = *ewf_table_section_entry

	}

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

		data = sectors_buf[entry.DataOffset-sectors_offs : entry.DataOffset-sectors_offs+uint64(Chunk_Size)]

		if bytes.HasPrefix(data, zlib_header) {
			Utils.Decompress(data)
			fmt.Println("IDX", idx)
			/*sectors_buf[entry.ChunkDataOffset-uint32(sectors_offs):entry.ChunkDataOffset-uint32(sectors_offs)+5],
			  "REM",uint32(len(sectors_buf))-entry.ChunkDataOffset-uint32(sectors_offs), "CompresseD?",entry.IsCompressed)*/
			//  Utils.DecompressF(data)
		}

	}
	//last data chunk maybe less than 32K size
	last_entry := ewf_table_section.Table_entries[len(ewf_table_section.Table_entries)-1]
	data = sectors_buf[last_entry.DataOffset-sectors_offs : last_entry.DataOffset-sectors_offs+
		uint64(len(sectors_buf))-last_entry.DataOffset-sectors_offs]
	if bytes.HasSuffix(data, zlib_header) {
		Utils.DecompressF(data)
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
