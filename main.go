package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/aarsakian/EWF_Reader/ewf"
)

var MediaTypes = map[uint]string{0x00: "Removable Storage Media",
	0x01: "Fixed Storage Media", 0x03: "Optical Disc", 0x0e: "Logical Evidence File", 0x10: "Physical Memory RAM",
}

var MediaFlags = map[uint]string{0x01: "Image File",
	0x02: "Physical Device", 0x04: "Fast Writer blocker used", 0x08: "Tableau writer used",
}

var CompressionLevel = map[uint]string{0x00: "no compression",
	0x01: "good compression", 0x02: "best compression",
}

type Decompressor interface {
	Decompress([]byte)
}

func FindEvidenceFiles(path_ string) []string {

	basePath := filepath.Dir(path_)

	_, fname := filepath.Split(path_)

	Files, err := ioutil.ReadDir(basePath)
	if err != nil {
		log.Fatal("ERR", err)
	}
	k := 0
	filenames := make([]string, len(Files))

	for _, finfo := range Files {

		if !finfo.IsDir() {

			if strings.HasPrefix(finfo.Name(), strings.Split(fname, ".")[0]) {

				filenames[k] = filepath.Join(basePath, finfo.Name()) //supply channel
				//fmt.Println("INFO", basePath+finfo.Name(), strings.Split(fname, ".")[0])
				k += 1
			}

		}
	}
	filenames = filenames[:k]
	fmt.Println(filenames)
	return filenames

}

func main() {
	evidencePath := flag.String("evidence", "", "path to evidence")
	verify := flag.Bool("verify", false, "verify data of evidence using adler32 checksum")
	VerifyHash := flag.Bool("verifyHash", false, "verify stored hash in evidence")
	showImageInfo := flag.Bool("showInfo", false, "show evidence information")
	showHash := flag.Bool("showHash", false, "show stored hash value")
	offset := flag.Int64("offset", -1, "offset to read data from the evidence")
	len := flag.Int64("len", 0, "number of bytes to read from offset in the evidence")
	flag.Parse()

	if *evidencePath == "" {
		fmt.Println("Evidence filename needed")
		os.Exit(0)
	}

	filenames := FindEvidenceFiles(*evidencePath)
	var ewf_image ewf.EWF_Image
	ewf_image.ParseEvidence(filenames)
	ewf_image.PopulateChunckOffsets()

	ewf_image.CachedChuncks = make([][]byte, ewf_image.NofChunks)

	if *showImageInfo {
		ewf_image.ShowInfo()
	}

	if *showHash {
		hash := ewf_image.GetHash()
		fmt.Println(hash)
	}

	if *offset > int64(ewf_image.NofChunks)*int64(ewf_image.Chuncksize) {
		panic("offset exceeds size of data")
	}

	if *offset+*len > int64(ewf_image.NofChunks)*int64(ewf_image.Chuncksize) {
		panic("len exceeds remaing data area")
	}

	if *offset != -1 && *len != 0 {

		var buf bytes.Buffer
		buf.Grow(int(*len))

		chunckId := *offset / int64(ewf_image.Chuncksize)       // the start id with respect to asked offset
		chuncksRequired := *len/int64(ewf_image.Chuncksize) + 1 // how many chuncks needed to retrieve data
		if ewf_image.IsCached(int(chunckId), int(chuncksRequired)) {
			ewf_image.RetrieveFromCache(int(chunckId), int(chuncksRequired), &buf)

		} else {
			ewf_files := ewf_image.LocateSegments(chunckId, chuncksRequired)
			chuncks := ewf_image.GetChuncks(int(chunckId), int(chuncksRequired))
			for _, ewf_file := range ewf_files {
				relativeOffset := *offset % int64(ewf_image.Chuncksize)

				ewf_file.LocateData(chuncks, relativeOffset, &buf)

				ewf_image.CacheIt(int(chunckId), int(chuncksRequired), buf)
			}

		}

		fmt.Printf("%s\n", buf.Bytes())
	}

	if *verify {
		verified := ewf_image.Verify()
		fmt.Println("verified ", verified)
	}

	if *VerifyHash {
		verified1 := ewf_image.VerifyHash()
		fmt.Println("Verified hash", verified1) // buf)
	}

}
