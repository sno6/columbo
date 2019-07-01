package main

import (
	"flag"
	"log"
	"os"

	"github.com/sno6/columbo/web/server"
)

func main() {
	// TODO: Move this to use Cobra as in the Gitlab project.
	web := flag.Bool("w", false, "Spawn a web server for audio detection.")
	dir := flag.String("d", "", "Search for audio files to fingerprint in this directory.")
	file := flag.String("f", "", "A portion of an audio file you wish to detect.")
	flag.Parse()

	if flag.NFlag() == 0 || *dir == "" {
		usage()
	}

	switch {
	case *file != "":
		if err := runCLI(*file, *dir); err != nil {
			log.Fatal(err)
		}
	case *web:
		if err := runWeb(*dir); err != nil {
			log.Fatal(err)
		}
	default:
		usage()
	}
}

func usage() {
	flag.Usage()
	os.Exit(1)
}

func runCLI(file, dir string) error {
	// TODO (sno6): Can leave CLI as a TODO for now until after release with web app.
	return nil
}

func runWeb(dir string) error {
	// store := fingerprint.NewMemoryStore()
	// if err := store.Process(dir); err != nil {
	// 	return err
	// }
	s, err := server.New(nil)
	if err != nil {
		return err
	}
	return s.Start()
}
