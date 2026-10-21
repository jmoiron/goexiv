package goexiv

// #cgo pkg-config: exiv2
// #include "helper.h"
// #include <stdlib.h>
import "C"

import (
	"runtime"
	"strings"
	"unsafe"

	"github.com/kolesa-team/goexiv/makernotes"
)

type ExifData struct {
	img  *Image // We point to img to keep it alive
	data *C.Exiv2ExifData
}

type ExifDatum struct {
	data  *ExifData
	datum *C.Exiv2ExifDatum
}

// ExifDatumIterator wraps the respective C++ structure.
type ExifDatumIterator struct {
	data *ExifData
	iter *C.Exiv2ExifDatumIterator
}

func makeExifData(img *Image, cdata *C.Exiv2ExifData) *ExifData {
	data := &ExifData{
		img,
		cdata,
	}

	runtime.SetFinalizer(data, func(x *ExifData) {
		C.exiv2_exif_data_free(x.data)
	})

	return data
}

func makeExifDatum(data *ExifData, cdatum *C.Exiv2ExifDatum) *ExifDatum {
	if cdatum == nil {
		return nil
	}

	datum := &ExifDatum{
		data,
		cdatum,
	}

	runtime.SetFinalizer(datum, func(x *ExifDatum) {
		C.exiv2_exif_datum_free(x.datum)
	})

	return datum
}

func (i *Image) GetExifData() *ExifData {
	return makeExifData(i, C.exiv2_image_get_exif_data(i.img))
}

func (i *Image) SetExifString(key, value string) error {
	return i.SetMetadataString("exif", key, value)
}

func (d *ExifData) GetString(key string) (string, error) {
	datum, err := d.FindKey(key)
	if err != nil {
		return "", err
	}

	if datum == nil {
		return "", ErrMetadataKeyNotFound
	}

	return datum.String(), nil
}

func (d *ExifData) FindKey(key string) (*ExifDatum, error) {
	ckey := C.CString(key)
	defer C.free(unsafe.Pointer(ckey))

	var cerr *C.Exiv2Error

	cdatum := C.exiv2_exif_data_find_key(d.data, ckey, &cerr)

	if cerr != nil {
		err := makeError(cerr)
		C.exiv2_error_free(cerr)
		return nil, err
	}

	return makeExifDatum(d, cdatum), nil
}

// Key returns the Exif key of the datum.
func (d *ExifDatum) Key() string {
	return C.GoString(C.exiv2_exif_datum_key(d.datum))
}

// Tag returns the numeric tag of the datum within its group. For the
// array-style maker note groups (e.g. Exif.CanonCs.*) this is the index of the
// first value in the array.
func (d *ExifDatum) Tag() int {
	return int(C.exiv2_exif_datum_tag(d.datum))
}

// Count returns the number of values of the datum.
func (d *ExifDatum) Count() int {
	return int(C.exiv2_exif_datum_count(d.datum))
}

func (d *ExifDatum) String() string {
	cstr := C.exiv2_exif_datum_to_string(d.datum)
	defer C.free(unsafe.Pointer(cstr))

	return C.GoString(cstr)
}

// Returns all EXIF tags
func (d *ExifData) AllTags() map[string]string {
	keyValues := map[string]string{}
	for i := d.Iterator(); i.HasNext(); {
		d := i.Next()
		keyValues[d.Key()] = d.String()
	}

	return keyValues
}

// Iterator returns a new ExifDatumIterator to iterate over all Exif data.
func (d *ExifData) Iterator() *ExifDatumIterator {
	return makeExifDatumIterator(d, C.exiv2_exif_data_iterator(d.data))
}

// HasNext returns true as long as the iterator has another datum to deliver.
func (i *ExifDatumIterator) HasNext() bool {
	return C.exiv2_exif_data_iterator_has_next(i.iter) != 0
}

// Next returns the next ExifDatum of the iterator or nil if iterator has reached the end.
func (i *ExifDatumIterator) Next() *ExifDatum {
	return makeExifDatum(i.data, C.exiv2_exif_datum_iterator_next(i.iter))
}

func makeExifDatumIterator(data *ExifData, cIter *C.Exiv2ExifDatumIterator) *ExifDatumIterator {
	datum := &ExifDatumIterator{data, cIter}

	runtime.SetFinalizer(datum, func(i *ExifDatumIterator) {
		C.exiv2_exif_datum_iterator_free(i.iter)
	})

	return datum
}

func (i *Image) ExifStripKey(key string) error {
	return i.StripKey(EXIF, key)
}

// ExiftoolTags is like AllTags, but names makernote tags the way exiftool does.
//
// Makernote tags known to the makernotes package are returned under a key of
// the form "Canon:MacroMode", with array style values split into their
// individual tags as exiftool lists them. All other tags keep their usual
// exiv2 key. If print is true, values are converted to the text exiftool
// prints (e.g. "Single" rather than 0) where a conversion is known.
func (d *ExifData) ExiftoolTags(print bool) map[string]string {
	keyValues := map[string]string{}
	for i := d.Iterator(); i.HasNext(); {
		d := i.Next()
		key, value := d.Key(), d.String()

		var entries []makernotes.Entry
		if parts := strings.Split(key, "."); len(parts) == 3 {
			entries = makernotes.Translate(parts[1], d.Tag(), d.Count(), value, print)
		}

		if entries == nil {
			keyValues[key] = value
			continue
		}
		for _, e := range entries {
			keyValues[e.Key] = e.Value
		}
	}

	return keyValues
}
