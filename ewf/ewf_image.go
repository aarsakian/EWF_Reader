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
	ChunckOffsets sections.Table_Entries
	CachedChuncks [][]byte
}

func (ewf_image EWF_Image) RetrieveData(offset int64, length int64) []byte {
	var buf bytes.Buffer
	buf.Grow(int(length))

	chunckId := offset / int64(ewf_image.Chuncksize)          // the start id with respect to asked offset
	chuncksRequired := length/int64(ewf_image.Chuncksize) + 1 // how many chuncks needed to retrieve data
	if ewf_image.IsCached(int(chunckId), int(chuncksRequired)) {
		ewf_image.RetrieveFromCache(int(chunckId), int(chuncksRequired), &buf)

	} else {
		ewf_files := ewf_image.LocateSegments(chunckId, chuncksRequired) // the files that contains the asked data
		chuncks := ewf_image.GetChuncks(int(chunckId), int(chuncksRequired))
		for _, ewf_file := range ewf_files {
			relativeOffset := offset % int64(ewf_image.Chuncksize)

			ewf_file.LocateData(chuncks, relativeOffset, int(length), &buf)

			ewf_image.CacheIt(int(chunckId), int(chuncksRequired), &buf)
		}

	}
	return buf.Bytes()
}

func (ewf_image EWF_Image) ShowInfo() {
	chunkCount, nofSectorPerChunk, nofBytesPerSector, nofSectors, toolInfo, _ := ewf_image.ewf_files[0].GetChunckInfo()
	fmt.Println("number of chuncks", chunkCount)
	fmt.Println("sectors per chunck", nofSectorPerChunk)
	fmt.Println("bytes per sector", nofBytesPerSector)
	fmt.Println("number of sectors", nofSectors)
	fmt.Println("toolInfo", toolInfo)
}

func (ewf_image EWF_Image) LocateSegments(chunck_id int64, nofRequestedChunks int64) EWF_files {
	ewf_filesMap := map[EWF_file]bool{}
	var ewf_files EWF_files
	remainingChunks := nofRequestedChunks
	startChunckId := chunck_id
	for _, ewf_file := range ewf_image.ewf_files {
		ewf_filesMap[ewf_file] = false
		for {
			if startChunckId >= int64(ewf_file.FirstChunckId) &&
				startChunckId < int64(ewf_file.NumberOfChuncks) { //located in this segment
				// workaround to keep unique values
				if !ewf_filesMap[ewf_file] {
					ewf_files = append(ewf_files, ewf_file)
					ewf_filesMap[ewf_file] = true
				}

				remainingChunks -= 1
				startChunckId += 1 //advance to the next chunck

			} else {
				break
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

func (ewf_image EWF_Image) GetChuncks(chunckId int, chuncksRequired int) sections.Table_Entries {
	return ewf_image.ChunckOffsets[chunckId : chunckId+chuncksRequired+1] // add one for boundary
}

func (ewf_image EWF_Image) IsImageEncase6Type() bool {
	return ewf_image.ewf_files[0].Sections.head.body.GetAttr("An unknown") == "EnCase" || ewf_image.ewf_files[0].Sections.head.body.GetAttr("Version") == "20201230"
}

func (ewf_image *EWF_Image) populateChunckOffsets() {

	offsets := make(sections.Table_Entries, ewf_image.NofChunks)
	chuncksProcessed := 0

	for idx, ewf_file := range ewf_image.ewf_files {
		ewf_image.ewf_files[idx].FirstChunckId = chuncksProcessed
		chuncksProcessed = ewf_file.PopulateChunckOffsets(offsets, chuncksProcessed)

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

func (ewf_image *EWF_Image) CacheIt(chunckId int, chuncksRequired int, buf *bytes.Buffer) {
	for id := 0; id < chuncksRequired; id++ {
		if buf.Len() < int(ewf_image.Chuncksize) { //cache only when buffer equals the chunck size
			continue
		}
		ewf_image.CachedChuncks[chunckId+id] = buf.Bytes()[:int(ewf_image.Chuncksize)]
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
			chunkCount, nofSectorPerChunk, nofBytesPerSector, _, _, _ := ewf_file.GetChunckInfo()
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

	ewf_image.populateChunckOffsets()

	ewf_image.CachedChuncks = make([][]byte, ewf_image.NofChunks)

}

func (ewf_image EWF_Image) GetHash() string {
	// last file has hash info
	ewf_file := ewf_image.ewf_files[len(ewf_image.ewf_files)-1] // hash section always in last segment
	return ewf_file.GetHash()
}
