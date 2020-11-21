# gmitohtml
[![CI status](https://gitlab.com/tslocum/gmitohtml/badges/master/pipeline.svg)](https://gitlab.com/tslocum/gmitohtml/commits/master)
[![Donate](https://img.shields.io/liberapay/receives/rocketnine.space.svg?logo=liberapay)](https://liberapay.com/rocketnine.space)

[Gemini](https://gemini.circumlunar.space) to HTML conversion tool and daemon

## Download

gmitohtml is written in [Go](https://golang.org). Run the following command to
download and build gmitohtml from source.

```bash
go get gitlab.com/tslocum/gmitohtml
```

The resulting binary is available as `~/go/bin/gmitohtml`.

## Usage

Convert a single document:

```bash
gmitohtml < document.gmi
```

Run as daemon at `http://localhost:8080`:

```bash
# Start the daemon:
gmitohtml --daemon=localhost:8080
# Access via browser:
xdg-open http://localhost:8080/gemini/twins.rocketnine.space/
```

## Support

Please share issues and suggestions [here](https://gitlab.com/tslocum/gmitohtml/issues).
