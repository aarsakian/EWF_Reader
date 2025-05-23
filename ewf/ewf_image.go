package ewf

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"time"

	"github.com/aarsakian/EWF_Reader/ewf/sections"
	Utils "github.com/aarsakian/EWF_Reader/ewf/utils"
	"github.com/aarsakian/EWF_Reader/logger"
)

var CHUNK_SIZE int = 32768
var NOFCHUNCKS int = 10000

type EWF_Image struct {
	ewf_files       EWF_files
	Chuncksize      uint32
	NofChunks       uint32
	ChunckOffsets   sections.Table_EntriesPtrs
	QueuedChunckIds Utils.Queue

	Profiling bool
}

func (ewf_image *EWF_Image) RetrieveData(offset int64, length int64) []byte {
	if ewf_image.Profiling {
		defer Utils.TimeTrack(time.Now(), "Locating Data")
	}

	var buf bytes.Buffer
	buf.Grow(int(length))

	var chuncksRequired int64
	relativeOffset := offset % int64(ewf_image.Chuncksize) // start from this offset from first chunck
	chunckId := offset / int64(ewf_image.Chuncksize)       // the start id with respect to asked offset

	if length+relativeOffset < int64(ewf_image.Chuncksize) { // data less than chunck
		chuncksRequired = length/int64(ewf_image.Chuncksize) + 1 // how many chuncks needed to retrieve data
	} else { // length exceeds chunck or when window of data is shifted so that it requires one more next chunck
		chuncksRequired = length/int64(ewf_image.Chuncksize) + 2 // how many chuncks needed to retri
	}

	ewf_filesMap := ewf_image.LocateSegments(chunckId, chuncksRequired) // the files that contains the asked data
	chuncks := ewf_image.GetChuncks(int(chunckId), int(chuncksRequired))
	firstChunckId := int64(0)
	for ewf_file, ewf_file_Nofchuncks := range ewf_filesMap {
		ewf_file_chuncks := chuncks[firstChunckId : ewf_file_Nofchuncks+1]

		logger.EWF_Readerlogger.Info(fmt.Sprintf("File %s ", ewf_file.Name))

		ewf_file.LocateData(ewf_file_chuncks, relativeOffset, int(length), &buf, int(ewf_image.Chuncksize))

		firstChunckId = ewf_file_Nofchuncks + 1

	}
	ewf_image.CacheIt(int(chunckId), int(chuncksRequired), int(relativeOffset), &buf)

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

func (ewf_image EWF_Image) LocateSegments(chunck_id int64, nofRequestedChunks int64) map[EWF_file]int64 {
	if ewf_image.Profiling {
		defer Utils.TimeTrack(time.Now(), "Locating Segments")
	}
	ewf_filesMap := map[EWF_file]int64{}

	remainingChunks := nofRequestedChunks
	startChunckId := chunck_id
	for idx, ewf_file := range ewf_image.ewf_files {

		for {
			if idx == len(ewf_image.ewf_files)-1 && startChunckId >= int64(ewf_image.ewf_files[idx].FirstChunckId) ||
				idx < len(ewf_image.ewf_files) && startChunckId >= int64(ewf_image.ewf_files[idx].FirstChunckId) &&
					startChunckId < int64(ewf_image.ewf_files[idx+1].FirstChunckId) { //located in this segment
				// workaround to keep unique values
				ewf_filesMap[ewf_file] = nofRequestedChunks - remainingChunks
				remainingChunks -= 1
				startChunckId += 1 //advance to the next chunck

			} else {

				break
			}
			if remainingChunks == 0 {
				ewf_filesMap[ewf_file] = nofRequestedChunks - remainingChunks
				return ewf_filesMap
			}
		}

	}
	return ewf_filesMap
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

func (ewf_image EWF_Image) IsImageEncase6Type() bool {
	return ewf_image.ewf_files[0].Sections.head.body.GetAttr("An unknown") == "EnCase" || ewf_image.ewf_files[0].Sections.head.body.GetAttr("Version") == "20201230"
}

func (ewf_image *EWF_Image) populateChunckOffsets() {
	if ewf_image.Profiling {
		defer Utils.TimeTrack(time.Now(), "populating chuncks map")
	}
	offsets := make(sections.Table_EntriesPtrs, ewf_image.NofChunks)
	chuncksProcessed := 0

	for idx := range ewf_image.ewf_files {
		ewf_image.ewf_files[idx].FirstChunckId = chuncksProcessed
		chuncksProcessed = ewf_image.ewf_files[idx].PopulateChunckOffsets(offsets, chuncksProcessed)

		//fmt.Printf("finished segment %s processed chuncks %d\n", ewf_file.Name, chuncksProcessed)
	}
	ewf_image.ChunckOffsets = offsets
}

func (ewf_image EWF_Image) IsChunckCached(chunckId int) bool {

	return ewf_image.ChunckOffsets[chunckId].IsCached

}

func (ewf_image *EWF_Image) CacheIt(chunckId int, chuncksRequired int, relivativeOffset int, buf *bytes.Buffer) {
	data := buf.Bytes()
	for id := 0; id < chuncksRequired; id++ {
		if ewf_image.IsChunckCached(chunckId + id) {
			continue
		}

		if buf.Len() < int(ewf_image.Chuncksize) { //cache only when buffer equals the chunck size
			continue
		}
		//last chunck is used as end offset skip it or next exceeds available buffer
		if id == chuncksRequired-1 || (id+1)*int(ewf_image.Chuncksize) > len(data) {
			break
		}

		if id == 0 && relivativeOffset != 0 { // first chunck not complete skip it
			continue
		}
		//
		if ewf_image.QueuedChunckIds.IsFull() {
			cachedChunckId := ewf_image.QueuedChunckIds.DeQueue()
			ewf_image.ChunckOffsets[cachedChunckId].IsCached = false
			ewf_image.ChunckOffsets[cachedChunckId].DataChuck = nil
		}

		ewf_image.QueuedChunckIds.EnQueue(chunckId + id)
		ewf_image.ChunckOffsets[chunckId+id].IsCached = true

		dataChunck := sections.DataChuck{Data: data[id*int(ewf_image.Chuncksize) : (id+1)*int(ewf_image.Chuncksize)]}
		ewf_image.ChunckOffsets[chunckId+id].DataChuck = &dataChunck

	}
}

func (ewf_image *EWF_Image) ParseEvidence(filenames []string) {
	ewf_files := make(EWF_files, len(filenames))
	if ewf_image.Profiling {
		Utils.TimeTrack(time.Now(), fmt.Sprintf("Parsed segments  %d in", len(filenames)))
	}
	for idx, filename := range filenames {

		ewf_file := EWF_file{Name: filename}

		ewf_file.CreateHandler()

		ewf_file.ParseHeader()

		if !ewf_file.IsValid() {
			fmt.Println(filename, "not valid header")
			continue
		}

		ewf_file.ParseSegment()
		fmt.Printf("Parsed segment %s\n", filename)

		if ewf_file.IsFirst() {
			chunkCount, nofSectorPerChunk, nofBytesPerSector, _, _, _ := ewf_file.GetChunckInfo()
			ewf_image.SetChunckInfo(chunkCount, nofSectorPerChunk, nofBytesPerSector)

		}
		ewf_files[idx] = ewf_file
		ewf_file.CloseHandler()

		if ewf_file.IsLast() {
			break
		}

	}
	elements := make(map[int]int, NOFCHUNCKS)
	ewf_image.QueuedChunckIds = Utils.Queue{Elements: elements, Capacity: NOFCHUNCKS}
	ewf_image.ewf_files = ewf_files
	fmt.Printf("about to populate map of chuncks\n")
	ewf_image.populateChunckOffsets()

}

func (ewf_image EWF_Image) GetHash() string {
	// last file has hash info
	ewf_file := ewf_image.ewf_files[len(ewf_image.ewf_files)-1] // hash section always in last segment
	return ewf_file.GetHash()
}
