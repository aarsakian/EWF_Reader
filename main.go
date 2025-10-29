package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/aarsakian/EWF_Reader/ewf"
	Utils "github.com/aarsakian/EWF_Reader/ewf/utils"
	"github.com/aarsakian/EWF_Reader/logger"
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
	VerifyHash := flag.Bool("verifyhash", false, "verify stored hash in evidence")
	showImageInfo := flag.Bool("showinfo", false, "show evidence information")
	showHash := flag.Bool("showhash", false, "show stored hash value")
	offset := flag.Int64("offset", 0, "offset to read data from the evidence")
	length := flag.Int64("len", math.MaxInt64, "number of bytes to read from offset in the evidence")
	profile := flag.Bool("profile", false, "profile performance")
	logactive := flag.Bool("log", false, "log activity")
	out := flag.String("out", "", "filename to write raw data")

	flag.Parse()

	if *evidencePath == "" {
		fmt.Println("Evidence filename needed")
		os.Exit(0)
	}

	if *logactive {
		now := time.Now()
		logfilename := "logs" + now.Format("2006-01-02T15_04_05") + ".txt"
		logger.InitializeLogger(*logactive, logfilename)

	}

	defer Utils.TimeTrack(time.Now(), "finished")
	filenames := Utils.FindEvidenceFiles(*evidencePath)

	ewf_image := ewf.EWF_Image{Profiling: *profile}

	fmt.Printf("Parsing %d evidence files \n", len(filenames))
	ewf_image.ParseEvidenceCH(filenames)

	if *showImageInfo {
		ewf_image.ShowInfo()
	}

	if *showHash {
		hash := ewf_image.GetHash()
		fmt.Println(hash)
	}

	if *offset > int64(ewf_image.NofChunks)*int64(ewf_image.Chunksize) {
		panic("offset exceeds size of data")
	}

	if *length != math.MaxInt64 && *offset+*length > int64(ewf_image.NofChunks)*int64(ewf_image.Chunksize) {
		panic("len exceeds remaing data area")
	}

	now := time.Now()

	if *length == math.MaxInt64 && *out != "" {
		ewf_image.WriteRawFile(*out)
	} else if *length != math.MaxInt64 {
		fmt.Printf("data to read %d MB\n", *length/1024/1000)
		data := ewf_image.RetrieveData(*offset, *length)
		if *out != "" {
			f, _ := os.Create(*out)

			f.Write(data)
			f.Close()
		}
	}

	fmt.Printf("data wrote in %f secs \n", time.Since(now).Seconds())

	if *verify {
		fmt.Println("Verifying adler32 checksums")
		verified := ewf_image.Verify()
		fmt.Println("verified ", verified)
	}

	if *VerifyHash {
		verified1 := ewf_image.VerifyHash()
		if verified1 {
			fmt.Println("MD5 hash verified successfully", verified1, ewf_image.GetHash()) // buf)
		} else {
			fmt.Println("MD5 hash verification failed")
		}

	}

}
