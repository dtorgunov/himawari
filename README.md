# Himawari

This package is intended to be a simple clone of `puu.sh`, that can be self-hosted.

The client is able to send various kinds of files to the server, which can then serve those files back via standard HTTP means. The authentication scheme utilises public key cryptography.

# Building

Himawari is written in go and can be fetched and compiled using the Go tool as follows:

```bash
$ go get github.com/dtorgunov/himawari
```

This will build both the server and the client binaries. See the `doc/` directory and the `godoc` comments/documentation for further information on using said binaries.

# Architecture

## Design

Part of the goal of this project is for me to practice creating "ad-hoc" HTTP-based APIs. The detailed description of the API can be found in the `doc/` directory.

## Implementation

Note that the current version of the repository does not implement all of the features outlined on `doc/api.md`. The package is considered unstable until all (or most) features are implemented. Implementation of the majority of the features specified in `doc/api.md` will promote this package to version 1.0.
