package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

const (
	inputPath       = "./data/in.mp4"
	overlayFilePath = "./data/overlay.png"
	outputDirPath   = "./data/ffmpeg-go-cli"
)

func main() {
	// gen(fmt.Sprintf("%s/out.mp4", outputDirPath))
	genByPipe(fmt.Sprintf("%s/out-pipe.mp4", outputDirPath))
}

func gen(outputPath string) {
	overlay := ffmpeg.Input(overlayFilePath).Filter("scale", ffmpeg.Args{"64:-1"})
	err := ffmpeg.Filter(
		[]*ffmpeg.Stream{
			ffmpeg.Input(inputPath),
			overlay,
		}, "overlay", ffmpeg.Args{"10:10"}, ffmpeg.KwArgs{"enable": "gte(t,1)"}).
		Output(outputPath).OverWriteOutput().ErrorToStdOut().Run()
	if err != nil {
		panic(err)
	}
}

func genByPipe(outputPath string) {
	pr, pw := io.Pipe()
	done := make(chan error)
	go func() {
		err := ffmpeg.Input(inputPath).
			Drawtext("hogehoge", 10, 10, false, map[string]interface{}{
				"enable":   "gte(t,2)",
				"fontsize": "10",
				"fontfile": "./data/ttf/ipaexg.ttf",
			}).
			Drawtext("PIYO.", 10, 50, false, map[string]interface{}{
				"enable":   "gte(t,2)",
				"fontsize": "20",
			}).
			Output("pipe:", ffmpeg.KwArgs{
				"f":        "mp4",
				"an":       "",
				"movflags": "frag_keyframe",
			}).
			WithOutput(pw, os.Stdout).
			Run()
		_ = pw.Close()
		done <- err
		close(done)
	}()

	bytes, err := ioutil.ReadAll(pr)
	if err != nil {
		panic(err)
	}
	err = <-done
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(outputPath, bytes, 0644)
	if err != nil {
		panic(err)
	}
}
