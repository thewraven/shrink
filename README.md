# shrink
Image compressor for jpg/png/gif/webp formats.

## Installation
This is a CLI frontend for the effort of [bradberger](github.com/bradberger) API. Please check it out his [optimizer repo.](github.com/bradberger/optimizer).

Note: webp is only supported on Linux.

### Development version

If you have a Go installation set up, you can compile shrink yourself:

> go get -u github.com/wraven/shrink

### Releases

Grab the binaries for your platform [here](https://github.com/thewraven/shrink/releases),
Windows, Linux and macOS are supported.
Just put the binary in your executable path.

## Examples

Tip: check the available options yourself

> shrink --help

Say, you want to compress all your photos and creating
the compressed images on the compressed folder:

> shrink -dir ~/Pictures  -output ~/compressed

If you want to define the quality of the compressed
images, you may use the `-quality` flag.

> shrink -dir ~/Pictures  -output ~/compressed -quality 30

If you have a large set of images and want to limit the
concurrency of the images processed, you can use the
`-workers` flag. By default, there are 50 workers acting concurrently.

> shrink -dir ~/Pictures  -output ~/compressed -workers 75
