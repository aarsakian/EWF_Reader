package ewf

import (
	"fmt"
	"hash/adler32"
	"time"

	"github.com/aarsakian/EWF_Reader/ewf/utils"
)

var CHUNK_SIZE int = 32768

type EWF_Image struct {
	ewf_files    EWF_files
	chunkOffsets map[int]bool
}

func (ewf_image *EWF_Image) GetChunckOffsets() {
	chunkOffsets := make(map[int]bool)
	for _, ewf_file := range ewf_image.ewf_files {
		chunkOffsets = ewf_file.GetChunckOffsets(chunkOffsets)
	}
	ewf_image.chunkOffsets = chunkOffsets
}

func (ewf_image EWF_Image) Verify() bool {
	var deflated_data []byte
	from := uint64(0)
	for chunckOffset, isCompressed := range ewf_image.chunkOffsets {
		if from == 0 {
			from = uint64(chunckOffset)
			continue

		}
		buf := ewf_image.ewf_files[0].ReadAt(int64(chunckOffset), uint64(chunckOffset)-from)
		if isCompressed {
			deflated_data = utils.Decompress(buf)
		}

		if utils.ReadEndianB(buf[len(buf)-4:]) != adler32.Checksum(deflated_data) {
			return false
		}
		from = uint64(chunckOffset)

	}
	return true
}

func (ewf_image EWF_Image) ReadAt(offset int, len uint64) []byte {
	var deflated_data []byte
	requested_data := make([]byte, len)
	from := uint64(0)
	idx := 0
	for chunckOffset, isCompressed := range ewf_image.chunkOffsets {
		if idx*CHUNK_SIZE >= offset {

			buf := ewf_image.ewf_files[0].ReadAt(int64(chunckOffset),
				uint64(chunckOffset)-from)
			if isCompressed {
				deflated_data = utils.Decompress(buf)
			}
			from = uint64(chunckOffset)

			break
		}
		idx++
	}
	if CHUNK_SIZE < offset {
		copy(requested_data, deflated_data[offset-CHUNK_SIZE:])
	} else {
		copy(requested_data, deflated_data[offset:])
	}

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

func (ewf_image EWF_Image) GetHash() string {
	// last file has hash info
	ewf_file := ewf_image.ewf_files[len(ewf_image.ewf_files)-1]
	return ewf_file.GetHash()
}
