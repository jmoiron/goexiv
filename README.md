# Go bindings for exiv2 (http://www.exiv2.org)

Note: This is a fork of [github.com/kolesa-team/goexiv](https://github.com/kolesa-team/goexiv) that makes
it compatible with newer versions (0.28.x) of exiv2. See the `Requirements` section for changes.

The library allows reading and writing EXIF and IPTC metadata to/from JPG, WEBP, and PNG images.

It is based on https://github.com/abustany/goexiv and https://github.com/gitschneider/goexiv with support added for writing metadata and various bugfixes.

## Requirements

A [libexiv2](http://www.exiv2.org) library v0.28.x is required. Older versions (0.27 and earlier) are not supported.

Building also requires a C++17 compiler (e.g. GCC 8+ or Clang 5+), which is itself a requirement of exiv2 0.28. The cgo build passes `-std=c++17` automatically.

On Ubuntu and Debian, libexiv2 0.28 can be installed from the package manager (`sudo apt install libexiv2-dev`); check the installed version with `pkg-config --modversion exiv2`.
If your distribution ships an older version, build exiv2 manually:

* Download and unpack the library from `https://github.com/Exiv2/exiv2/releases` (any v0.28.x release)
* Install the library (the steps are taken from libexiv2 README):
    ```
    mkdir build && cd build
    cmake .. -DCMAKE_BUILD_TYPE=Release
    cmake --build .
    sudo make install
    sudo ldconfig
    ```
* It may be needed to set the following variable: `export PKG_CONFIG_PATH=/usr/local/lib64/pkgconfig`

Now the Go code in this project can interface with the libexiv2 library.

The installation process for other operating systems should be similar.
On Alpine, libexiv2 can be installed with `apk --update add exiv2-dev` (make sure the packaged version is 0.28.x).

## Usage

Basic usage:

```
import "github.com/kolesa-team/goexiv"

// Open an image from disk
goexivImg, err := goexiv.Open("/path/to/image.jpg")
if err != nil {
    return err
}

// Write an EXIF comment
err = goexivImg.SetMetadataString("exif", "Exif.Photo.UserComment", "A comment. Might be a JSON string. Можно писать и по-русски!")
if err != nil {
    return err
}

// Write an EXIF short integer
err = goexivImg.SetMetadataShort("exif", "Exif.Photo.ExposureProgram", "2")
if err != nil {
    return err
}

// Read metadata
err = goexivImg.ReadMetadata()
if err != nil {
    return err
}

// Read an EXIF comment
userComment, err := goexivImg.GetExifData().GetString("Exif.Photo.UserComment")
if err != nil {
    return err
}

fmt.Println(userComment)
// "A comment. Might be a JSON string. Можно писать и по-русски!"

// It is advisable to insert this line at the end of `goexivImg` lifecycle.  
// see exiv_test.go:Test_GetBytes_Goroutine
runtime.KeepAlive(goexivImg)
```

Changing the image metadata in memory and returning the updated image (an approach fit for a web service):

```
// Say we have an image in memory: var img []byte
goexivImg, err := goexiv.OpenBytes(img)
if err != nil {
    return err
}

// Write an IPTC comment
err = goexivImg.SetMetadataString("iptc", "Iptc.Application2.Caption", "A comment. Might be a JSON string.")
if err != nil {
    return err
}

// Get back the modified image, so it can now be further processed (e.g. sent over the network)
img = goexivImg.GetBytes()
```

Retrieving all metadata keys and values:

```
img.ReadMetadata()
// map[string]string
exif := img.GetExifData().AllTags()

// map[string]string
iptc := img.GetIptcData().AllTags()
```

A complete image processing workflow in Go can be organized with the following additional libraries:

* https://github.com/kolesa-team/go-webp - Go bindings for libwebp to process WEBP images
* https://github.com/lEx0/go-libjpeg-nrgba - Go bindings for libjpeg-turbo, a fast JPEG processing library.
* https://github.com/disintegration/imaging - a generic Go library for working with images (covers many formats, but is not as fast as the libraries above, so it can be used as a fallback for PNG and GIF)
