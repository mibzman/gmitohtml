# mibzman's gmitohtml
A fork of [tslocum's gmitohtml](https://gitlab.com/tslocum/gmitohtml) that targets Titan.

### Changes:
- Removed search box
- Removed bookmarks 
- Modify pathing so paths render as `hostname/pagename` instead of `hostname/gemini/hostname/pagename`
- Added [water css](https://watercss.kognise.dev/)

# Original README
[![GoDoc](https://gitlab.com/tslocum/godoc-static/-/raw/master/badge.svg)](https://docs.rocketnine.space/gitlab.com/tslocum/gmitohtml/pkg/gmitohtml)
[![CI status](https://gitlab.com/tslocum/gmitohtml/badges/master/pipeline.svg)](https://gitlab.com/tslocum/gmitohtml/commits/master)
[![Donate](https://img.shields.io/liberapay/receives/rocketnine.space.svg?logo=liberapay)](https://liberapay.com/rocketnine.space)

[Gemini](https://gemini.circumlunar.space) to [HTML](https://en.wikipedia.org/wiki/HTML)
conversion tool and daemon

## Download

[**Download gmitohtml**](https://gmitohtml.rocketnine.space/download/?sort=name&order=desc)

gmitohtml is available on Android as [Xenia](https://gitlab.com/tslocum/xenia).

## Compile

gmitohtml is written in [Go](https://golang.org). Run the following command to
download and build gmitohtml from source.

```bash
go get gitlab.com/tslocum/gmitohtml
```

The resulting binary is available as `~/go/bin/gmitohtml`.

## Configure

See [CONFIGURATION.md](https://gitlab.com/tslocum/gmitohtml/blob/master/CONFIGURATION.md)

## Usage

Run daemon at [http://localhost:1967](http://localhost:1967):

```bash
gmitohtml --daemon=localhost:1967
```

Convert a single document:

```bash
gmitohtml < document.gmi
```

## Support

Please share issues and suggestions [here](https://gitlab.com/tslocum/gmitohtml/issues).
