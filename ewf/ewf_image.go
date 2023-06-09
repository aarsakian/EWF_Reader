package ewf

import (
	"crypto/md5"
	"fmt"
	"time"
)

var CHUNK_SIZE int = 32768

type EWF_Image struct {
	ewf_files  EWF_files
	chuncksize uint32
	nofChunks  uint32
}

func (ewf_image EWF_Image) ShowInfo() {
	fmt.Println(ewf_image.ewf_files[0].GetChunckInfo())
}

func (ewf_image EWF_Image) ReadAt(offset int64, len uint64) {
	segment := offset / int64(ewf_image.nofChunks*ewf_image.chuncksize)
	base_offset := ewf_image.nofChunks * ewf_image.chuncksize
	if offset > int64(base_offset) {
		ewf_image.ewf_files[segment].ReadAt(offset-int64(base_offset), len)
	} else {
		chunck_id := offset / int64(ewf_image.chuncksize)

		chunck := ewf_image.ewf_files[segment].GetChunck(int(chunck_id))
		ewf_image.ewf_files[segment].ReadAt(int64(chunck.DataOffset), len)
	}

}

func (ewf_image EWF_Image) VerifyHash() bool {
	var data []byte
	for _, ewf_file := range ewf_image.ewf_files {
		data = ewf_file.VerifyHash(data)

	}
	calculated_md5 := fmt.Sprintf("%x", md5.Sum(data))
	return calculated_md5 == ewf_image.GetHash()

}

func (ewf_image EWF_Image) Verify() bool {

	for _, ewf_file := range ewf_image.ewf_files {
		if !ewf_file.Verify() {
			return false
		}
	}
	return true
}

func (ewf_image *EWF_Image) SetChunckInfo(chunkCount uint64, nofSectorPerChunk uint64, nofBytesPerSector uint64) {
	ewf_image.chuncksize = uint32(nofBytesPerSector) * uint32(nofSectorPerChunk)
	ewf_image.nofChunks = uint32(chunkCount)
}

func (ewf_image *EWF_Image) ParseEvidence(filenames []string) {
	var ewf_files EWF_files

	for _, filename := range filenames {
		start := time.Now()

		ewf_file := EWF_file{Name: filename}

		ewf_file.CreateHandler()

		ewf_file.ParseHeader()

		if !ewf_file.IsValid() {
			fmt.Println(filename, "not valid header")
			continue
		}

		ewf_file.ParseSegment()

		if ewf_file.IsFirst() {
			chunkCount, nofSectorPerChunk, nofBytesPerSector, _, _ := ewf_file.GetChunckInfo()
			ewf_image.SetChunckInfo(chunkCount, nofSectorPerChunk, nofBytesPerSector)

		}

		elapsed := time.Since(start)
		fmt.Printf("Parsed Evidence %s in %s\n ", filename, elapsed)

		ewf_files = append(ewf_files, ewf_file)

		ewf_file.CloseHandler()

		if ewf_file.IsLast() {
			break
		}

	}
	ewf_image.ewf_files = ewf_files

}

func (ewf_image EWF_Image) GetHash() string {
	// last file has hash info
	ewf_file := ewf_image.ewf_files[len(ewf_image.ewf_files)-1] // hash section always in last segment
	return ewf_file.GetHash()
}
