package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSanitize(t *testing.T) {
	cases := map[string]string{
		"":                     "",
		"Artist":               "Artist",
		"Artist: Live":         "Artist Live",
		"Rock: Roll: Best of:": "Rock Roll Best of",
		"with: colons: inside": "with colons inside",
		"no-colons-here":       "no-colons-here",
		"multiple::double":     "multipledouble",
		":leading":             "leading",
		"trailing:":            "trailing",
		"a/b/c":                "abc",
		`..\..\evil`:           "....evil",
		" spaced ":             "spaced",
	}
	for in, want := range cases {
		if got := sanitize(in); got != want {
			t.Errorf("sanitize(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestProcessFileCreatesTaggedDirectories(t *testing.T) {
	inputDir := t.TempDir()
	outputDir := t.TempDir()

	src := filepath.Join(inputDir, "input.mp3")
	if err := os.WriteFile(src, makeMP3Bytes(), 0644); err != nil {
		t.Fatalf("writing source file: %v", err)
	}

	if err := processFile(src, outputDir); err != nil {
		t.Fatalf("processFile: %v", err)
	}

	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Errorf("expected source file to be moved (renamed), still exists: %v", err)
	}

	expected := filepath.Join(outputDir, "Test Artist", "Test Album", "Test Title.mp3")
	if _, err := os.Stat(expected); err != nil {
		t.Errorf("expected moved file at %q, got error: %v", expected, err)
	}
}

func TestProcessFileAlreadyAtDestination(t *testing.T) {
	outputDir := t.TempDir()
	dest := filepath.Join(outputDir, "Test Artist", "Test Album")
	if err := os.MkdirAll(dest, 0777); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	src := filepath.Join(dest, "Test Title.mp3")
	if err := os.WriteFile(src, makeMP3Bytes(), 0644); err != nil {
		t.Fatalf("writing source file: %v", err)
	}

	if err := processFile(src, outputDir); err != nil {
		t.Fatalf("processFile: %v", err)
	}

	if _, err := os.Stat(src); err != nil {
		t.Errorf("expected file to remain at %q after rename to itself, got error: %v", src, err)
	}
}

func TestProcessFileAlreadyAtDestinationRelativeOutput(t *testing.T) {
	base := t.TempDir()
	// Chdir so a relative output path is meaningful.
	oldwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(base); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer os.Chdir(oldwd)

	if err := os.MkdirAll(filepath.Join("library", "Test Artist", "Test Album"), 0777); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	src := filepath.Join("library", "Test Artist", "Test Album", "Test Title.mp3")
	if err := os.WriteFile(src, makeMP3Bytes(), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	if err := processFile(src, "library"); err != nil {
		t.Fatalf("processFile: %v", err)
	}

	if _, err := os.Stat(src); err != nil {
		t.Errorf("file should remain at destination, not duplicated: %v", err)
	}
	if _, err := os.Stat(filepath.Join("library", "Test Artist", "Test Album", "Test Title (1).mp3")); !os.IsNotExist(err) {
		t.Error("file was needlessly duplicated despite being at destination")
	}
}

func TestProcessFileMissingInput(t *testing.T) {
	if err := processFile("/does/not/exist.mp3", t.TempDir()); err == nil {
		t.Error("expected error for missing input file, got nil")
	}
}

func TestProcessFileMissingMetadata(t *testing.T) {
	src := filepath.Join(t.TempDir(), "x.mp3")
	if err := os.WriteFile(src, makeEmptyTag(), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := processFile(src, t.TempDir()); err == nil {
		t.Error("expected error for missing metadata, got nil")
	}
}

func TestCopyFile(t *testing.T) {
	src := filepath.Join(t.TempDir(), "src.bin")
	dst := filepath.Join(t.TempDir(), "dst.bin")
	data := []byte("hello world copy")
	if err := os.WriteFile(src, data, 0644); err != nil {
		t.Fatalf("write src: %v", err)
	}
	if err := copyFile(src, dst); err != nil {
		t.Fatalf("copyFile: %v", err)
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if string(got) != string(data) {
		t.Errorf("copy mismatch: got %q want %q", got, data)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Errorf("source should be removed after copy, err=%v", err)
	}
}

func TestIterateContinuesPastFailedFile(t *testing.T) {
	inputDir := t.TempDir()
	outputDir := t.TempDir()

	// A good file that should be moved.
	good := filepath.Join(inputDir, "good.mp3")
	if err := os.WriteFile(good, makeMP3Bytes(), 0644); err != nil {
		t.Fatalf("write good: %v", err)
	}
	// A bad file (missing metadata) that should be skipped without aborting.
	bad := filepath.Join(inputDir, "bad.mp3")
	if err := os.WriteFile(bad, makeEmptyTag(), 0644); err != nil {
		t.Fatalf("write bad: %v", err)
	}

	// iterate should process the good file and report the failed one at the end.
	if err := iterate(inputDir, outputDir); err == nil {
		t.Error("expected error reporting failed file, got nil")
	}

	if _, err := os.Stat(filepath.Join(outputDir, "Test Artist", "Test Album", "Test Title.mp3")); err != nil {
		t.Errorf("good file should have been moved despite the bad one: %v", err)
	}
	if _, err := os.Stat(bad); err != nil {
		t.Errorf("bad file should have been left in place, got error: %v", err)
	}
}

func TestUniquePath(t *testing.T) {
	dir := t.TempDir()

	a := filepath.Join(dir, "track.mp3")
	if got := uniquePath(a); got != a {
		t.Errorf("uniquePath on missing file = %q, want %q", got, a)
	}
	if err := os.WriteFile(a, []byte("x"), 0644); err != nil {
		t.Fatalf("write a: %v", err)
	}

	b := uniquePath(a)
	wantB := filepath.Join(dir, "track (1).mp3")
	if b != wantB {
		t.Errorf("uniquePath after collision = %q, want %q", b, wantB)
	}
	if err := os.WriteFile(b, []byte("x"), 0644); err != nil {
		t.Fatalf("write b: %v", err)
	}

	c := uniquePath(a)
	wantC := filepath.Join(dir, "track (2).mp3")
	if c != wantC {
		t.Errorf("uniquePath after second collision = %q, want %q", c, wantC)
	}
}

func TestIterateSkipsDirectoriesAndUnsupportedExtensions(t *testing.T) {
	inputDir := t.TempDir()
	outputDir := t.TempDir()

	sub := filepath.Join(inputDir, "subdir")
	if err := os.MkdirAll(sub, 0777); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Unsupported extension should not be moved.
	txt := filepath.Join(inputDir, "notes.txt")
	if err := os.WriteFile(txt, []byte("hello"), 0644); err != nil {
		t.Fatalf("writing txt: %v", err)
	}
	// Supported extension should be moved.
	mp3 := filepath.Join(inputDir, "move.mp3")
	if err := os.WriteFile(mp3, makeMP3Bytes(), 0644); err != nil {
		t.Fatalf("writing mp3: %v", err)
	}

	if err := iterate(inputDir, outputDir); err != nil {
		t.Fatalf("iterate: %v", err)
	}

	if _, err := os.Stat(txt); err != nil {
		t.Errorf("txt file should not have been moved: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outputDir, "Test Artist", "Test Album", "Test Title.mp3")); err != nil {
		t.Errorf("mp3 file should have been moved: %v", err)
	}
}

func TestIterateMissingRoot(t *testing.T) {
	if err := iterate("/does/not/exist", t.TempDir()); err == nil {
		t.Error("expected error for missing root, got nil")
	}
}

func TestIterateSkipsOutputDirNestedInInput(t *testing.T) {
	inputDir := t.TempDir()
	outputDir := filepath.Join(inputDir, "library")

	// A file already in the output layout must not be reprocessed.
	alreadyMoved := filepath.Join(outputDir, "Test Artist", "Test Album", "Existing.mp3")
	if err := os.MkdirAll(filepath.Dir(alreadyMoved), 0777); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(alreadyMoved, makeMP3Bytes(), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	contentBefore, _ := os.ReadFile(alreadyMoved)

	// A raw input file that should be moved into the nested output dir.
	raw := filepath.Join(inputDir, "raw.mp3")
	if err := os.WriteFile(raw, makeID3v2Title("New Title"), 0644); err != nil {
		t.Fatalf("write raw: %v", err)
	}

	if err := iterate(inputDir, outputDir); err != nil {
		t.Fatalf("iterate: %v", err)
	}

	// Raw file should now sit in the nested output layout.
	want := filepath.Join(outputDir, "Test Artist", "Test Album", "New Title.mp3")
	if _, err := os.Stat(want); err != nil {
		t.Errorf("expected moved file at %q: %v", want, err)
	}
	if _, err := os.Stat(raw); !os.IsNotExist(err) {
		t.Errorf("raw file should have been moved: %v", err)
	}

	// The pre-existing organized file must be untouched (not rewritten/duplicated).
	contentAfter, _ := os.ReadFile(alreadyMoved)
	if string(contentAfter) != string(contentBefore) {
		t.Error("pre-existing organized file was modified")
	}
}
