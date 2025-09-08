package ewf

import (
	"bytes"
	"fmt"
	"os"
	"sync"
)

var RAW_CHUNK_SIZE int64 = 4 * 1024 * 1024

var bufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, RAW_CHUNK_SIZE)
	},
}

type Job struct {
	offset int64
	size   int64
}

func (ewf_image EWF_Image) WriteRawFile(outfile string) {
	offset := int64(0)
	diskSize := int64(ewf_image.GetDiskSize())
	fmt.Printf("about to write %d MB raw data to %s\n", diskSize/1024/1024, outfile)
	//_buf := BufferPool.Get().([]byte)

	f, _ := os.OpenFile(outfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()

	var buf bytes.Buffer
	buf.Grow(int(RAW_CHUNK_SIZE))

	for offset < diskSize {
		if diskSize-offset > int64(RAW_CHUNK_SIZE) {

			ewf_image.RetrieveDataPreAllocateBuffer(&buf, offset, RAW_CHUNK_SIZE)
		} else {

			ewf_image.RetrieveDataPreAllocateBuffer(&buf, offset, int64(diskSize)-offset)
		}

		f.Write(buf.Bytes())
		buf.Reset()
		offset += RAW_CHUNK_SIZE

	}
}

func (ewf_image EWF_Image) WriteParallelRawFile(outfile string) {
	offset := int64(0)
	diskSize := int64(ewf_image.GetDiskSize())

	f, _ := os.OpenFile(outfile, os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()

	var wg sync.WaitGroup
	numWorkers := 4

	jobs := make(chan Job, numWorkers)

	for i := 0; i < numWorkers; i++ {
		go func() {
			for job := range jobs {

				data := ewf_image.RetrieveData(job.offset, job.size)

				_, err := f.WriteAt(data, job.offset)
				if err != nil {
					fmt.Println(err)

				}

				wg.Done()
			}
		}()
	}

	for offset < diskSize {

		if diskSize-offset > int64(RAW_CHUNK_SIZE) {
			jobs <- Job{offset, RAW_CHUNK_SIZE}
		} else {
			jobs <- Job{offset, int64(diskSize) - offset}
		}
		//fmt.Println("sent to channel", offset)
		offset += RAW_CHUNK_SIZE
		wg.Add(1)
	}

	wg.Wait()
	close(jobs)

}
