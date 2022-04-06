package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/mamoroom/go-ffmpeg-mock/infra/cloudstorage"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

const PORT = "8080"

const (
	LOCATION = "Asia/Tokyo"
)

func init() {
	loc, err := time.LoadLocation(LOCATION)
	if err != nil {
		loc = time.FixedZone(LOCATION, 9*60*60)
	}
	time.Local = loc
}

const (
	inputFilePath   = "./data/in.mp4"
	overlayFilePath = "./data/overlay.png"
	outputDirPath   = "./data/ffmpeg-go"
)

func main() {

	ctx := context.Background()
	storageClient := cloudstorage.New(ctx)
	// router //
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Hello"))
		})
		// gen h264 video
		r.Get("/gen", func(w http.ResponseWriter, r *http.Request) {
			err := ffmpeg.Input(inputFilePath).
				Output(fmt.Sprintf("%s/out.mp4", outputDirPath),
					ffmpeg.KwArgs{
						"c:v": "h264",
					}).OverWriteOutput().ErrorToStdOut().Run()
			if err != nil {
				fmt.Printf("%v\n", err)
			}
			w.Write([]byte("generated."))
		})
		// gen video which is overlayed
		r.Get("/gen-overlay", func(w http.ResponseWriter, r *http.Request) {
			overlay := ffmpeg.Input(overlayFilePath).Filter("scale", ffmpeg.Args{"64:-1"})
			err := ffmpeg.Filter(
				[]*ffmpeg.Stream{
					ffmpeg.Input(inputFilePath),
					overlay,
				}, "overlay", ffmpeg.Args{"10:10"}, ffmpeg.KwArgs{"enable": "gte(t,1)"}).
				Output(fmt.Sprintf("%s/out-overlay.mp4", outputDirPath)).OverWriteOutput().ErrorToStdOut().Run()
			if err != nil {
				fmt.Printf("%v\n", err)
			}
			w.Write([]byte("generated."))
		})
		// gen video with text by pipe
		r.Get("/gen-piped", func(w http.ResponseWriter, r *http.Request) {
			pr, pw := io.Pipe()
			done := make(chan error)
			go func() {
				err := ffmpeg.Input(inputFilePath).
					Drawtext("hogehoge", 50, 50, false, map[string]interface{}{
						"enable": "between(t,3,5)",
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
				fmt.Printf("%v\n", err)
				w.Write([]byte("error."))
				return
			}
			err = <-done
			if err != nil {
				fmt.Printf("%v\n", err)
				w.Write([]byte("error."))
				return
			}
			err = ioutil.WriteFile(fmt.Sprintf("%s/out-piped.mp4", outputDirPath), bytes, 0644)
			if err != nil {
				fmt.Printf("%v\n", err)
				w.Write([]byte("error."))
				return
			}
			w.Write([]byte("generated."))
		})
		r.Get("/gen2gcs", func(w http.ResponseWriter, r *http.Request) {
			pr, pw := io.Pipe()
			done := make(chan error)
			go func() {
				err := ffmpeg.Input(inputFilePath).
					Drawtext("hogehoge", 50, 50, false, map[string]interface{}{
						"enable": "between(t,3,5)",
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
				fmt.Printf("%v\n", err)
				w.Write([]byte("error."))
				return
			}
			err = <-done
			if err != nil {
				fmt.Printf("%v\n", err)
				w.Write([]byte("error."))
				return
			}

			storageClient.UploadByBinContent(ctx, bytes, "gen/out-piped.mp4", nil)
			if err != nil {
				fmt.Printf("%v\n", err)
				w.Write([]byte("error."))
				return
			}
			w.Write([]byte("generated."))
		})
	})

	fileServer(r, "/static", http.Dir("data"))

	log.Printf("Listening on port %s", PORT)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", PORT), r))
}

func fileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}
