package main

import (
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/xfrr/goffmpeg/transcoder"
)

const (
	inputPath       = "./data/in.mp4"
	overlayFilePath = "./data/overlay.png"
	outputDirPath   = "./data/goffmpeg-cli"
)

func main() {
	transcode(fmt.Sprintf("%s/out.mp4", outputDirPath))
	transcodeByPipe(fmt.Sprintf("%s/out-pipe.mp4", outputDirPath))
}

func transcode(outputPath string) {
	// Create new instance of transcoder
	trans := new(transcoder.Transcoder)

	// Initialize transcoder passing the input file path and output file path
	err := trans.Initialize(inputPath, outputPath)
	if err != nil {
		panic(err)
	}

	// Start transcoder process without checking progress
	done := trans.Run(false)

	// This channel is used to wait for the process to end
	err = <-done
	// Handle error...
	if err != nil {
		panic(err)
	}
}

func transcodeByPipe(outputPath string) {

	// Create new instance of transcoder
	trans := new(transcoder.Transcoder)

	// Initialize an empty transcoder
	err := trans.InitializeEmptyTranscoder()
	// Handle error...
	if err != nil {
		panic(err)
	}

	// set input file
	trans.SetInputPath(inputPath)

	// Create an output pipe to read from, which will return *io.PipeReader.
	// Must also specify the output container format
	r, err := trans.CreateOutputPipe("mp4")

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer r.Close()
		defer wg.Done()

		// Read data from output pipe
		data, err := ioutil.ReadAll(r)
		if err != nil {
			panic(err)
		}
		// write file
		err = ioutil.WriteFile(outputPath, data, 0644)
		if err != nil {
			panic(err)
		}
	}()

	// Start transcoder process without checking progress
	fmt.Println(trans.GetCommand())
	done := trans.Run(true)

	// This channel is used to wait for the transcoding process to end
	err = <-done
	if err != nil {
		panic(err)
	}
	wg.Wait()

}
