package main

import (
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

	if *showImageInfo {
		ewf_image.ShowInfo()
	}

	if *showHash {
		hash := ewf_image.GetHash()
		fmt.Println(hash)
	}

	if *offset != -1 && *len != 0 {
		ewf_image.ReadAt(*offset, *len)
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
