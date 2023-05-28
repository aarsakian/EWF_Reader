package main

import (
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

	if len(os.Args) < 2 {
		fmt.Println("Evidence filename needed")
		os.Exit(0)
	}

	filenames := FindEvidenceFiles(os.Args[1]) //producer
	var ewf_image ewf.EWF_Image
	ewf_image.ParseEvidence(filenames) //consumer
	ewf_image.GetChunckOffsets()
	ewf_image.ReadAt(48423, 512)
	ewf_image.ShowInfo()
	hash := ewf_image.GetHash()

	fmt.Println(hash)
	verified := ewf_image.Verify()
	ewf_image.VerifyHash()
	fmt.Println("Verified status", verified)

}
