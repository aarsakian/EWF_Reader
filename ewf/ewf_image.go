package ewf

import (
	"fmt"
	"time"

	"github.com/aarsakian/EWF_Reader/ewf/utils"
)

var CHUNK_SIZE int = 32768

type EWF_Image struct {
	ewf_files    EWF_files
	chunkOffsets []int
}

func (ewf_image *EWF_Image) GetChunckOffsets() {
	var chunkOffsets []int
	for _, ewf_file := range ewf_image.ewf_files {
		chunkOffsets = ewf_file.GetChunckOffsets(chunkOffsets)
	}
	ewf_image.chunkOffsets = chunkOffsets
}

func (ewf_image EWF_Image) ReadAt(offset int, len uint64) []byte {
	var deflated_data []byte
	requested_data := make([]byte, len)

	for idx := range ewf_image.chunkOffsets {
		if idx*CHUNK_SIZE >= offset {
			from := ewf_image.chunkOffsets[idx]
			to := ewf_image.chunkOffsets[idx+1]

			buf := ewf_image.ewf_files[0].ReadAt(int64(ewf_image.chunkOffsets[idx]),
				uint64(to-from))

			deflated_data = utils.Decompress(buf)
			break
		}
	}
	copy(requested_data, deflated_data[offset:])
	return requested_data
}

func (ewf_image *EWF_Image) ParseEvidence(filenames []string) {
	ewf_files := make(EWF_files, len(filenames))
	for idx, filename := range filenames {
		start := time.Now()

		ewf_file := EWF_file{Name: filename, SegmentNum: uint(idx)}

		ewf_file.ParseHeader()

		ewf_file.ParseSegment()

		elapsed := time.Since(start)
		fmt.Printf("Parsed Evidence %s in %s\n ", filename, elapsed)

		ewf_files[idx] = ewf_file

	}
	ewf_image.ewf_files = ewf_files

}
