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
	chunkCount, nofSectorPerChunk, nofBytesPerSector, nofSectors, _ := ewf_image.ewf_files[0].GetChunckInfo()
	fmt.Println("number of chuncks", chunkCount)
	fmt.Println("sectors per chunck", nofSectorPerChunk)
	fmt.Println("bytes per sector", nofBytesPerSector)
	fmt.Println("number of sectors", nofSectors)
}

func (ewf_image EWF_Image) ReadAt(offset int64, len uint64) []byte {

	segment_size := int64(ewf_image.nofChunks * ewf_image.chuncksize)
	segment_id := offset / segment_size
	var chunck_id int64

	if offset > segment_size {
		chunck_id = offset / int64(ewf_image.chuncksize)

	} else {
		chunck_id = offset / int64(ewf_image.chuncksize)

	}

	ewf_file := ewf_image.ewf_files[segment_id]
	chunck := ewf_file.GetChunck(int(chunck_id))

	ewf_file.CreateHandler()
	defer ewf_file.CloseHandler()

	return ewf_file.ReadAt(int64(chunck.DataOffset), len)
}

func (ewf_image EWF_Image) VerifyHash() bool {
	var data []byte
	for _, ewf_file := range ewf_image.ewf_files {

		data = ewf_file.CollectData(data)

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
