package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/aarsakian/EWF_Reader/ewf"
	Utils "github.com/aarsakian/EWF_Reader/ewf/utils"
	EWF_Readerlogger "github.com/aarsakian/EWF_Reader/logger"
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

func main() {
	evidencePath := flag.String("evidence", "", "path to evidence")
	verify := flag.Bool("verify", false, "verify data of evidence using adler32 checksum")
	VerifyHash := flag.Bool("verifyHash", false, "verify stored hash in evidence")
	showImageInfo := flag.Bool("showInfo", false, "show evidence information")
	showHash := flag.Bool("showHash", false, "show stored hash value")
	offset := flag.Int64("offset", -1, "offset to read data from the evidence")
	length := flag.Int64("len", 0, "number of bytes to read from offset in the evidence")
	profile := flag.Bool("profile", false, "profile performance")
	logactive := flag.Bool("log", false, "log activity")

	flag.Parse()

	if *evidencePath == "" {
		fmt.Println("Evidence filename needed")
		os.Exit(0)
	}

	if *logactive {
		now := time.Now()
		logfilename := "logs" + now.Format("2006-01-02T15_04_05") + ".txt"
		EWF_Readerlogger.InitializeLogger(*logactive, logfilename)

	}

	defer Utils.TimeTrack(time.Now(), "finished")
	filenames := Utils.FindEvidenceFiles(*evidencePath)
	ewf_image := ewf.EWF_Image{Profiling: *profile}

	ewf_image.ParseEvidence(filenames)

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

	if *offset+*length > int64(ewf_image.NofChunks)*int64(ewf_image.Chuncksize) {
		panic("len exceeds remaing data area")
	}

	if *offset != -1 && *length != 0 {
		fmt.Printf("data to read %d\n", *length)
		data := ewf_image.RetrieveData(*offset, *length)

		fmt.Printf("%x\n", data[0:1])
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
