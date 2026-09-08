# tagger

A small Go tool that organizes audio files by their metadata. It walks an input
directory and moves every supported audio file into an `Artist/Album/` layout
under an output directory, using the file's embedded ID3 tags.

## Supported formats

- `.mp3`
- `.mp4`
- `.m4b`

Tags are read using [github.com/dhowden/tag](https://github.com/dhowden/tag).

## Usage

```sh
tagger -in <input directory> -out <output directory>
```

Files are moved from:

```
<input>/some/path/track.mp3
```

into:

```
<output>/<Artist>/<Album>/<Title>.mp3
```

Colons (`:`) are stripped from Artist, Album and Title so the resulting paths
are valid on Windows filesystems.

### Example

```sh
tagger -in ~/Music/incoming -out ~/Music/library
```

## Build

```sh
make build          # produces ./tagger in the project root
make install        # installs to $(PREFIX)/bin (PREFIX defaults to /usr/local)
make test           # run unit tests
make vet            # run go vet
make clean          # remove the built binary
```

## Options

| Flag    | Description                      |
|---------|----------------------------------|
| `-in`   | Input directory (required)       |
| `-out`  | Output directory (required)      |
