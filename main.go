package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/dhowden/tag"
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

	if err := iterate(*i, *o); err != nil {
		log.Fatal(err)
	}
}

func iterate(input, output string) error {
	root, err := filepath.Abs(input)
	if err != nil {
		return err
	}
	outDir, err := filepath.Abs(output)
	if err != nil {
		return err
	}
	inside := outDir == root || strings.HasPrefix(outDir, root+string(filepath.Separator))

	var failed []string
	walkErr := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// Skip the output directory (and contents) when it lives under input,
			// to avoid reprocessing already-moved files.
			if inside && (path == outDir || strings.HasPrefix(path, outDir+string(filepath.Separator))) {
				return filepath.SkipDir
			}
			return nil
		}
		switch filepath.Ext(path) {
		case ".mp3", ".mp4", ".m4b":
			if err := processFile(path, outDir); err != nil {
				log.Printf("skipping %s: %v", path, err)
				failed = append(failed, path)
			}
		}
		return nil
	})
	if walkErr != nil {
		return walkErr
	}
	if len(failed) > 0 {
		return fmt.Errorf("%d file(s) could not be processed", len(failed))
	}
	return nil
}

func sanitize(s string) string {
	s = strings.ReplaceAll(s, ":", "")
	s = strings.ReplaceAll(s, "/", "")
	s = strings.ReplaceAll(s, "\\", "")
	s = strings.ReplaceAll(s, "\x00", "")
	return strings.TrimSpace(s)
}

func processFile(f, outputDir string) error {
	fmt.Println("Processing " + f)
	file, err := os.Open(f)
	if err != nil {
		return err
	}

	m, err := tag.ReadFrom(file)
	file.Close()
	if err != nil {
		return err
	}

	if m.Artist() == "" || m.Album() == "" || m.Title() == "" {
		return fmt.Errorf("%s: missing artist, album or title metadata", f)
	}

	artist := sanitize(m.Artist())
	album := sanitize(m.Album())
	title := sanitize(m.Title())
	if artist == "" || album == "" || title == "" {
		return fmt.Errorf("%s: artist, album or title is empty after sanitizing", f)
	}

	destDir := filepath.Join(outputDir, artist, album)
	if err := os.MkdirAll(destDir, 0777); err != nil {
		return err
	}

	dest := filepath.Join(destDir, title+filepath.Ext(f))
	if f == dest {
		return nil
	}
	dest = uniquePath(dest)

	if err := os.Rename(f, dest); err != nil {
		// os.Rename fails across filesystems (EXDEV); fall back to copy+delete.
		return copyFile(f, dest)
	}
	return nil
}

// uniquePath returns a non-colliding path by appending " (n)" before the
// extension when a file already exists at the given path.
func uniquePath(p string) string {
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return p
	}
	ext := filepath.Ext(p)
	base := strings.TrimSuffix(p, ext)
	for i := 1; ; i++ {
		candidate := fmt.Sprintf("%s (%d)%s", base, i, ext)
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()
	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Remove(src)
}
