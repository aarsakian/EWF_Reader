
package volume

import "bytes"

type EWF_Volume_Section struct {
    Ukwnown [4]uint8
    ChunkCount uint32 "Number of Chunks per Segment"
    NofSectorPerChunk uint32 "Number of Sectors per Chunk default 64"
    NofBytesPerSector uint32 "default 512"
    NofSectors uint32 "Number of Sectors within all segment files"
    Reserved [20]uint8 "default 0x00"
    Padding [45]uint8
    Singature [5]uint8 "EWF signature"
    CheckSum [4]uint8
    Adler32 string "adler-32 of all the previous data within the additional section volume data"
    Vol_Data Volume_Data
}

type Volume_Data struct {
    MediaType uint8
    Unknown1 [3]uint8
    ChunkCount uint32
    NofSectorPerChunk uint32 "Number of Sectors per Chunk default 64"
    NofBytesPerSector uint32 "default 512"
    NofSectors uint64 "Number of Sectors within all segment files"
    NofCylindersCHS uint32 "Number of cylinders of the C:H:S usually empty"
    NofHeadesCHS uint32 "Number of cylinders of the C:H:S usually empty"
    NofSectorsCHS uint32 "Number of Sectors of the C:H:S usually empty"
    MediaFlags uint8
    Uknown2 [3]uint8
    PALM uint32 "Volume start sector"
    Unkown3 [4]uint8
    SMART uint32 "start sector offset relative from the end of the media"
    CompressionLevel uint8
    Unknown4 [3]uint8
    SectorErrorGr [4]uint8 "Sector error granularity"
    Unknown5 [4]uint8
    GUID [16]uint8 "identify uniquely a set of segment files"
    Signature [5]uint8 "contains 0x00"
    CheckSum [4]uint8 "adler32 of all the previous data within the additional volume section data"
}


func (ewf_volume_section *EWF_Volume_Section) Parse(buf *bytes.Reader){
    
}


func (ewf_volume_section *EWF_Volume_Section) GetAttr(string) (interface{}) {
    return ""
}