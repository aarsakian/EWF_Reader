package ewf

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"runtime"
	"time"

	"github.com/aarsakian/EWF_Reader/ewf/sections"
	Utils "github.com/aarsakian/EWF_Reader/ewf/utils"
	"github.com/aarsakian/EWF_Reader/logger"
)

var CHUNK_SIZE int = 32768
var NOFchunkS int = 10000

type EWF_Image struct {
	ewf_files      EWF_files
	Chunksize      uint32
	NofChunks      uint32
	chunkOffsets   sections.Table_Entries
	QueuedchunkIds Utils.Queue

	Profiling bool
}

func (ewf_image *EWF_Image) RetrieveData(offset int64, length int64) []byte {
	if ewf_image.Profiling {
		defer Utils.TimeTrack(time.Now(), "Locating Data")
	}

	var buf bytes.Buffer
	buf.Grow(int(length))

	var chunksRequired int64
	relativeOffset := offset % int64(ewf_image.Chunksize) // start from this offset from first chunk
	chunkId := offset / int64(ewf_image.Chunksize)        // the start id with respect to asked offset

	if length+relativeOffset < int64(ewf_image.Chunksize) { // data less than chunk
		chunksRequired = length/int64(ewf_image.Chunksize) + 1 // how many chunks needed to retrieve data
	} else { // length exceeds chunk or when window of data is shifted so that it requires one more next chunk
		chunksRequired = length/int64(ewf_image.Chunksize) + 2 // how many chunks needed to retri
	}

	ewf_filesMap := ewf_image.LocateSegments(chunkId, chunksRequired) // the files that contains the asked data
	chunks := ewf_image.Getchunks(int(chunkId), int(chunksRequired))
	firstchunkId := int64(0)
	for ewf_file, ewf_file_Nofchunks := range ewf_filesMap {
		ewf_file_chunks := chunks[firstchunkId : ewf_file_Nofchunks+1]

		logger.EWF_Readerlogger.Info(fmt.Sprintf("File %s ", ewf_file.Name))

		ewf_file.LocateData(ewf_file_chunks, relativeOffset, int(length), &buf, int(ewf_image.Chunksize))

		firstchunkId = ewf_file_Nofchunks + 1

	}
	ewf_image.CacheIt(int(chunkId), int(chunksRequired), int(relativeOffset), &buf)

	return buf.Bytes()
}

func (ewf_image *EWF_Image) ShowInfo() {
	chunkCount, nofSectorPerChunk, nofBytesPerSector, nofSectors, toolInfo, _ := ewf_image.ewf_files[0].GetchunkInfo()
	fmt.Println("number of chunks", chunkCount)
	fmt.Println("sectors per chunk", nofSectorPerChunk)
	fmt.Println("bytes per sector", nofBytesPerSector)
	fmt.Println("number of sectors", nofSectors)
	fmt.Println("toolInfo", toolInfo)
}

func (ewf_image *EWF_Image) LocateSegments(chunk_id int64, nofRequestedChunks int64) map[EWF_file]int64 {
	if ewf_image.Profiling {
		defer Utils.TimeTrack(time.Now(), "Locating Segments")
	}
	ewf_filesMap := map[EWF_file]int64{}

	remainingChunks := nofRequestedChunks
	startchunkId := chunk_id
	for idx, ewf_file := range ewf_image.ewf_files {

		for {
			if idx == len(ewf_image.ewf_files)-1 && startchunkId >= int64(ewf_image.ewf_files[idx].FirstchunkId) ||
				idx < len(ewf_image.ewf_files) && startchunkId >= int64(ewf_image.ewf_files[idx].FirstchunkId) &&
					startchunkId < int64(ewf_image.ewf_files[idx+1].FirstchunkId) { //located in this segment
				// workaround to keep unique values
				ewf_filesMap[ewf_file] = nofRequestedChunks - remainingChunks
				remainingChunks -= 1
				startchunkId += 1 //advance to the next chunk

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

func (ewf_image *EWF_Image) VerifyHash() bool {

	var buf bytes.Buffer
	buf.Grow(int(ewf_image.NofChunks * ewf_image.Chunksize))

	for _, ewf_file := range ewf_image.ewf_files {

		ewf_file.CollectData(&buf)

	}
	calculated_md5 := fmt.Sprintf("%x", md5.Sum(buf.Bytes()))
	return calculated_md5 == ewf_image.GetHash()

}

func (ewf_image *EWF_Image) Verify() bool {

	for _, ewf_file := range ewf_image.ewf_files {
		if !ewf_file.Verify(int(ewf_image.Chunksize)) {
			return false
		}
	}
	return true
}

func (ewf_image *EWF_Image) SetchunkInfo(chunkCount uint64, nofSectorPerChunk uint64, nofBytesPerSector uint64) {
	ewf_image.Chunksize = uint32(nofBytesPerSector) * uint32(nofSectorPerChunk)
	ewf_image.NofChunks = uint32(chunkCount)
}

func (ewf_image *EWF_Image) Getchunks(chunkId int, chunksRequired int) sections.Table_Entries {
	return ewf_image.chunkOffsets[chunkId : chunkId+chunksRequired+1] // add one for boundary
}

func (ewf_image *EWF_Image) IsImageEncase6Type() bool {
	return ewf_image.ewf_files[0].Sections.head.body.GetAttr("An unknown") == "EnCase" || ewf_image.ewf_files[0].Sections.head.body.GetAttr("Version") == "20201230"
}

func (ewf_image *EWF_Image) populatechunkOffsets() {
	if ewf_image.Profiling {
		defer Utils.TimeTrack(time.Now(), "populating chunks map")
	}
	offsets := make(sections.Table_Entries, ewf_image.NofChunks)
	chunksProcessed := 0

	for idx := range ewf_image.ewf_files {
		ewf_image.ewf_files[idx].FirstchunkId = chunksProcessed
		chunksProcessed = ewf_image.ewf_files[idx].PopulatechunkOffsets(offsets, chunksProcessed)

		//fmt.Printf("finished segment %s processed chunks %d\n", ewf_file.Name, chunksProcessed)
	}
	ewf_image.chunkOffsets = offsets
}

func (ewf_image *EWF_Image) IschunkCached(chunkId int) bool {

	return ewf_image.chunkOffsets[chunkId].IsCached

}

func (ewf_image *EWF_Image) CacheIt(chunkId int, chunksRequired int, relivativeOffset int, buf *bytes.Buffer) {
	data := buf.Bytes()
	startId := 0
	for id := 0; id < chunksRequired; id++ {
		if ewf_image.IschunkCached(chunkId + id) {
			continue
		}

		//last chunk is used as end offset skip it or next exceeds available buffer
		if id == chunksRequired-1 || (id+1)*int(ewf_image.Chunksize) > len(data) {
			break
		}

		if id == 0 && relivativeOffset != 0 { // first chunk not complete skip it
			data = data[int(ewf_image.Chunksize)-relivativeOffset:]
			startId = 1
			continue
		}
		//
		if ewf_image.QueuedchunkIds.IsFull() {
			cachedchunkId := ewf_image.QueuedchunkIds.DeQueue()
			ewf_image.chunkOffsets[cachedchunkId].IsCached = false
			ewf_image.chunkOffsets[cachedchunkId].DataChuck = nil
		}

		ewf_image.QueuedchunkIds.EnQueue(chunkId + id)
		ewf_image.chunkOffsets[chunkId+id].IsCached = true

		ewf_image.chunkOffsets[chunkId+id].DataChuck = &sections.DataChuck{Data: data[(id-startId)*int(ewf_image.Chunksize) : (id+1-startId)*int(ewf_image.Chunksize)]}

	}
}

func (ewf_image *EWF_Image) ParseEvidenceCH(filenames []string) {
	now := time.Now()
	// IO Bound
	numWorker := 2 * runtime.NumCPU()
	ewf_files := make(EWF_files, len(filenames))

	if ewf_image.Profiling {
		Utils.TimeTrack(time.Now(), fmt.Sprintf("Parsed segments  %d in", len(filenames)))
	}

	done := make([]chan bool, numWorker)
	ewf_filesCH := make(chan EWF_file)

	filesPerWorker := len(filenames) / numWorker

	for i := 0; i < numWorker; i++ {
		start := i * filesPerWorker
		end := start + filesPerWorker
		if i == numWorker-1 { //last worker gets remaining files
			end = len(filenames)
		}
		done[i] = make(chan bool)
		go ewf_image.ParseEvidenceWorker(filenames[start:end], start, ewf_filesCH, done[i])
	}

	//sync point
	go func() {
		for i := 0; i < numWorker; i++ {
			<-done[i]
		}
		close(ewf_filesCH)
	}()
	for ewf_file := range ewf_filesCH {
		//fmt.Printf("Processed %s id %d\n", ewf_file.Name, ewf_file.Id)
		ewf_files[ewf_file.Id] = ewf_file
	}
	fmt.Printf("Parsed evidence %d files in %f secs\n", len(filenames), time.Since(now).Seconds())
	ewf_image.QueuedchunkIds = Utils.Queue{Capacity: NOFchunkS}
	ewf_image.ewf_files = ewf_files
	now = time.Now()

	ewf_image.populatechunkOffsets()
	fmt.Printf("populated map of chunks in %f secs\n", time.Since(now).Seconds())

}

func (ewf_image *EWF_Image) ParseEvidence(filenames []string) {
	now := time.Now()

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
		if ewf_image.Profiling {
			fmt.Printf("Parsed segment %s\n", filename)
		}

		if ewf_file.IsFirst() {
			chunkCount, nofSectorPerChunk, nofBytesPerSector, _, _, _ := ewf_file.GetchunkInfo()
			ewf_image.SetchunkInfo(chunkCount, nofSectorPerChunk, nofBytesPerSector)

		}
		ewf_files[idx] = ewf_file
		ewf_file.CloseHandler()

		if ewf_file.IsLast() {
			break
		}

	}

	fmt.Printf("Parsed evidence %d files in %f secs\n", len(filenames), time.Since(now).Seconds())

	ewf_image.QueuedchunkIds = Utils.Queue{Capacity: NOFchunkS}
	ewf_image.ewf_files = ewf_files

	now = time.Now()

	ewf_image.populatechunkOffsets()
	fmt.Printf("populated map of chunks in %f secs\n", time.Since(now).Seconds())

}

func (ewf_image *EWF_Image) ParseEvidenceWorker(filenames []string, start int, ewf_filesCH chan<- EWF_file, done chan<- bool) {
	for idx, filename := range filenames {
		ewf_file := EWF_file{Name: filename, Id: start + idx}
		ewf_file.CreateHandler()

		ewf_file.ParseHeader()

		if !ewf_file.IsValid() {
			fmt.Println(ewf_file.Name, "not valid header")
			break
		}

		ewf_file.ParseSegment()
		if ewf_image.Profiling {
			fmt.Printf("Parsed segment %s\n", ewf_file.Name)
		}

		if ewf_file.IsFirst() {
			chunkCount, nofSectorPerChunk, nofBytesPerSector, _, _, _ := ewf_file.GetchunkInfo()
			ewf_image.SetchunkInfo(chunkCount, nofSectorPerChunk, nofBytesPerSector)

		}
		ewf_file.CloseHandler()
		ewf_filesCH <- ewf_file
		if ewf_file.IsLast() {
			break
		}
	}

	done <- true
}

func (ewf_image *EWF_Image) GetHash() string {
	// last file has hash info
	ewf_file := ewf_image.ewf_files[len(ewf_image.ewf_files)-1] // hash section always in last segment
	return ewf_file.GetHash()
}
