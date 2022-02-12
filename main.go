package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

const PORT = "8000"

func main() {

	// router //
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Hello"))
		})
		r.Get("/gen", func(w http.ResponseWriter, r *http.Request) {
			overlay := ffmpeg.Input("./data/overlay.png").Filter("scale", ffmpeg.Args{"64:-1"})
			err := ffmpeg.Filter(
				[]*ffmpeg.Stream{
					ffmpeg.Input("./data/in1.mp4"),
					overlay,
				}, "overlay", ffmpeg.Args{"10:10"}, ffmpeg.KwArgs{"enable": "gte(t,1)"}).
				Output("./data/out1.mp4").OverWriteOutput().ErrorToStdOut().Run()
			if err != nil {
				fmt.Printf("%v\n", err)
			}
			w.Write([]byte("generated."))
		})
	})

	log.Printf("Listening on port %s", PORT)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", PORT), r))
}
