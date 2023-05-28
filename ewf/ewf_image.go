package ewf

import (
	"crypto/md5"
	"fmt"
	"hash/adler32"
	"time"

	"github.com/aarsakian/EWF_Reader/ewf/sections"
	"github.com/aarsakian/EWF_Reader/ewf/utils"
)

var CHUNK_SIZE int = 32768

type EWF_Image struct {
	ewf_files    EWF_files
	chunkOffsets sections.Table_Entries
}

func (ewf_image *EWF_Image) GetChunckOffsets() {
	var chunkOffsets sections.Table_Entries
	for _, ewf_file := range ewf_image.ewf_files {
		chunkOffsets = ewf_file.GetChunckOffsets(chunkOffsets)
	}
	//sort.Sort(chunkOffsets)
	ewf_image.chunkOffsets = chunkOffsets
}

func (ewf_image EWF_Image) ShowInfo() {
	fmt.Println(ewf_image.ewf_files[0].GetVolInfo())
}

func (ewf_image EWF_Image) VerifyHash() bool {

	var data []byte
	ewf_image.ewf_files[0].CreateHandler()
	defer ewf_image.ewf_files[0].CloseHandler()
	for idx, chunck := range ewf_image.chunkOffsets {
		if idx == len(ewf_image.chunkOffsets)-1 {

			break

		}
		from := uint64(chunck.DataOffset)
		to := uint64(ewf_image.chunkOffsets[idx+1].DataOffset)
		buf := ewf_image.ewf_files[0].ReadAt(int64(chunck.DataOffset), to-from)
		if !chunck.IsCompressed {
			continue
		}

		data = append(data, utils.Decompress(buf)...)

	}
	calculated_md5 := fmt.Sprintf("%x", md5.Sum(data))
	return calculated_md5 == ewf_image.GetHash()

}

func (ewf_image EWF_Image) Verify() bool {
	var deflated_data []byte
	ewf_image.ewf_files[0].CreateHandler()
	defer ewf_image.ewf_files[0].CloseHandler()
	for idx, chunck := range ewf_image.chunkOffsets {
		if idx == len(ewf_image.chunkOffsets)-1 {

			break

		}
		from := uint64(chunck.DataOffset)
		to := uint64(ewf_image.chunkOffsets[idx+1].DataOffset)
		buf := ewf_image.ewf_files[0].ReadAt(int64(chunck.DataOffset), to-from)
		if !chunck.IsCompressed {
			continue
		}

		deflated_data = utils.Decompress(buf)
		if utils.ReadEndianB(buf[len(buf)-4:]) != adler32.Checksum(deflated_data) {
			fmt.Println("problematic chunck", idx, chunck.DataOffset)

		}

	}
	return true
}

func (ewf_image EWF_Image) ReadAt(offset int, len uint64) []byte {
	var deflated_data []byte
	requested_data := make([]byte, len)

	for idx, chunck := range ewf_image.chunkOffsets {

		if idx > 0 && idx*CHUNK_SIZE >= offset {
			ewf_image.ewf_files[0].CreateHandler()

			from := uint64(ewf_image.chunkOffsets[idx-1].DataOffset)
			to := uint64(chunck.DataOffset)

			buf := ewf_image.ewf_files[0].ReadAt(int64(from), to-from)
			if chunck.IsCompressed {
				deflated_data = utils.Decompress(buf)
				copy(requested_data, deflated_data[offset-(idx-1)*CHUNK_SIZE:])
			} else {
				copy(requested_data, buf[offset-(idx-1)*CHUNK_SIZE:])
			}
			ewf_image.ewf_files[0].CloseHandler()
			break
		}

	}

	return requested_data
}

func (ewf_image *EWF_Image) ParseEvidence(filenames []string) {
	ewf_files := make(EWF_files, len(filenames))
	for idx, filename := range filenames {
		start := time.Now()

		ewf_file := EWF_file{Name: filename, SegmentNum: uint(idx)}

		ewf_file.CreateHandler()

		ewf_file.ParseHeader()

		ewf_file.ParseSegment()

		elapsed := time.Since(start)
		fmt.Printf("Parsed Evidence %s in %s\n ", filename, elapsed)

		ewf_files[idx] = ewf_file

		ewf_file.CloseHandler()

	}
	ewf_image.ewf_files = ewf_files

}

func (ewf_image EWF_Image) GetHash() string {
	// last file has hash info
	ewf_file := ewf_image.ewf_files[len(ewf_image.ewf_files)-1]
	return ewf_file.GetHash()
}
