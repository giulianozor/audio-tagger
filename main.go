package main

import (
	"flag"
	"fmt"
	"github.com/dhowden/tag"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	i := flag.String("in", "", "Input directory")
	o := flag.String("out", "", "Output directory")
	flag.Parse()

	if *i == "" {
		log.Fatal("Input directory not set")
	}
	if *o == "" {
		log.Fatal("Output directory not set")
	}

	iterate(*i, *o)
}

func iterate(path string, o string) {
	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Fatalf(err.Error())
		}
		if info.IsDir() == false {
			switch filepath.Ext(path) {
			case ".mp3", ".mp4", ".m4b":
				err := processfile(path, o)
				if err != nil {
					fmt.Println(err)
					return err
				}
			default:
				// do nothing
			}
		}
		return nil
	})
}

func processfile(f string, o string) error {
	fmt.Println("Processing " + f)
	ext := filepath.Ext(f)
	file, err := os.Open(f)
	if err != nil {
		return err
	}
	defer file.Close()

	m, err := tag.ReadFrom(file)
	if err != nil {
		return err
	}

	err = os.MkdirAll(strings.ReplaceAll(fmt.Sprintf("%s/%s/%s", o, m.Artist(), m.Album()), ":", ""), 0777)
	if err != nil {
		return err
	}

	err = os.Rename(f, strings.ReplaceAll(fmt.Sprintf("%s/%s/%s/%s%s", o, m.Artist(), m.Album(), m.Title(), ext), ":", ""))
	if err != nil {
		return err
	}
	return nil
}
