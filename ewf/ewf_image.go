package ewf

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"time"

	"github.com/aarsakian/EWF_Reader/ewf/sections"
)

var CHUNK_SIZE int = 32768

type EWF_Image struct {
	ewf_files     EWF_files
	Chuncksize    uint32
	NofChunks     uint32
	ChunckOffsets sections.Table_EntriesPtrs
	CachedChuncks [][]byte
}

func (ewf_image EWF_Image) ShowInfo() {
	chunkCount, nofSectorPerChunk, nofBytesPerSector, nofSectors, _ := ewf_image.ewf_files[0].GetChunckInfo()
	fmt.Println("number of chuncks", chunkCount)
	fmt.Println("sectors per chunck", nofSectorPerChunk)
	fmt.Println("bytes per sector", nofBytesPerSector)
	fmt.Println("number of sectors", nofSectors)
}

func (ewf_image EWF_Image) LocateSegments(chunck_id int64, nofRequestedChunks int64) EWF_files {
	var ewf_files EWF_files
	remainingChunks := nofRequestedChunks
	startChunckId := chunck_id
	for _, ewf_file := range ewf_image.ewf_files {
		for {
			if startChunckId >= int64(ewf_file.FirstChunckId) &&
				startChunckId < int64(ewf_file.FirstChunckId)+int64(ewf_file.NumberOfChuncks) {
				ewf_files = append(ewf_files, ewf_file)

				remainingChunks -= int64(ewf_file.NumberOfChuncks) - startChunckId
				startChunckId += int64(ewf_file.NumberOfChuncks)

			}
			if remainingChunks <= 0 {
				return ewf_files
			}
		}

	}
	return ewf_files
}

func (ewf_image EWF_Image) VerifyHash() bool {

	var buf bytes.Buffer
	buf.Grow(int(ewf_image.NofChunks * ewf_image.Chuncksize))

	for _, ewf_file := range ewf_image.ewf_files {

		ewf_file.CollectData(&buf)

	}
	calculated_md5 := fmt.Sprintf("%x", md5.Sum(buf.Bytes()))
	return calculated_md5 == ewf_image.GetHash()

}

func (ewf_image EWF_Image) Verify() bool {

	for _, ewf_file := range ewf_image.ewf_files {
		if !ewf_file.Verify(int(ewf_image.Chuncksize)) {
			return false
		}
	}
	return true
}

func (ewf_image *EWF_Image) SetChunckInfo(chunkCount uint64, nofSectorPerChunk uint64, nofBytesPerSector uint64) {
	ewf_image.Chuncksize = uint32(nofBytesPerSector) * uint32(nofSectorPerChunk)
	ewf_image.NofChunks = uint32(chunkCount)
}

func (ewf_image EWF_Image) GetChuncks(chunckId int, chuncksRequired int) sections.Table_EntriesPtrs {
	return ewf_image.ChunckOffsets[chunckId : chunckId+chuncksRequired+1] // add one for boundary
}

func (ewf_image *EWF_Image) PopulateChunckOffsets() {

	offsets := make(sections.Table_EntriesPtrs, ewf_image.NofChunks)
	chuncksProcessed := 0
	for idx, ewf_file := range ewf_image.ewf_files {
		ewf_image.ewf_files[idx].FirstChunckId = chuncksProcessed
		chuncksProcessed += ewf_file.PopulateChunckOffsets(offsets, chuncksProcessed)

		ewf_image.ewf_files[idx].NumberOfChuncks = uint32(chuncksProcessed)
		fmt.Printf("finished segment %s processed chuncks %d\n", ewf_file.Name, chuncksProcessed)
	}
	ewf_image.ChunckOffsets = offsets
}

func (ewf_image EWF_Image) IsCached(chunckId int, chuncksRequired int) bool {
	for id := 0; id < chuncksRequired; id++ {
		if ewf_image.CachedChuncks[chunckId+id] == nil {
			return false
		}
	}
	return true
}

func (ewf_image EWF_Image) RetrieveFromCache(chunckId int, chuncksRequired int, buf *bytes.Buffer) {
	for id := 0; id < chuncksRequired; id++ {
		buf.Write(ewf_image.CachedChuncks[chunckId+id])
	}
}

func (ewf_image *EWF_Image) CacheIt(chunckId int, chuncksRequired int, buf bytes.Buffer) {
	for id := 0; id < chuncksRequired; id++ {
		ewf_image.CachedChuncks[chunckId+id] = buf.Next(int(ewf_image.Chuncksize))
	}
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
