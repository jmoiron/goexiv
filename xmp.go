package goexiv

// #cgo pkg-config: exiv2
// #include "helper.h"
// #include <stdlib.h>
import "C"

import (
	"runtime"
	"unsafe"
)

// XmpData contains all Xmp Data of an image.
type XmpData struct {
	img  *Image // We point to img to keep it alive
	data *C.Exiv2XmpData
}

// XmpDatum stores the info of one xmp datum.
type XmpDatum struct {
	data  *XmpData
	datum *C.Exiv2XmpDatum
}

// XmpDatumIterator wraps the respective C++ structure.
type XmpDatumIterator struct {
	data *XmpData
	iter *C.Exiv2XmpDatumIterator
}

func makeXmpData(img *Image, cdata *C.Exiv2XmpData) *XmpData {
	data := &XmpData{
		img,
		cdata,
	}

	runtime.SetFinalizer(data, func(x *XmpData) {
		C.exiv2_xmp_data_free(x.data)
	})

	return data
}

func makeXmpDatum(data *XmpData, cdatum *C.Exiv2XmpDatum) *XmpDatum {
	if cdatum == nil {
		return nil
	}

	datum := &XmpDatum{
		data,
		cdatum,
	}

	runtime.SetFinalizer(datum, func(x *XmpDatum) {
		C.exiv2_xmp_datum_free(x.datum)
	})

	return datum
}

// GetXmpData returns the XmpData of an Image.
func (i *Image) GetXmpData() *XmpData {
	return makeXmpData(i, C.exiv2_image_get_xmp_data(i.img))
}

// FindKey tries to find the specified key and returns its data.
// It returns an error if the key is invalid. If the key is not found, a
// nil pointer will be returned
func (d *XmpData) FindKey(key string) (*XmpDatum, error) {
	ckey := C.CString(key)
	defer C.free(unsafe.Pointer(ckey))

	var cerr *C.Exiv2Error

	cdatum := C.exiv2_xmp_data_find_key(d.data, ckey, &cerr)

	if cerr != nil {
		err := makeError(cerr)
		C.exiv2_error_free(cerr)
		return nil, err
	}

	runtime.KeepAlive(d)
	return makeXmpDatum(d, cdatum), nil
}

// Iterator returns a new XmpDatumIterator to iterate over all XMP data.
func (d *XmpData) Iterator() *XmpDatumIterator {
	iter := &XmpDatumIterator{d, C.exiv2_xmp_data_iterator(d.data)}

	runtime.SetFinalizer(iter, func(i *XmpDatumIterator) {
		C.exiv2_xmp_datum_iterator_free(i.iter)
	})

	return iter
}

// HasNext returns true as long as the iterator has another datum to deliver.
func (i *XmpDatumIterator) HasNext() bool {
	return C.exiv2_xmp_data_iterator_has_next(i.iter) != 0
}

// Next returns the next XmpDatum of the iterator or nil if iterator has reached the end.
func (i *XmpDatumIterator) Next() *XmpDatum {
	return makeXmpDatum(i.data, C.exiv2_xmp_datum_iterator_next(i.iter))
}

// AllTags returns all XMP keys with their values as strings. Arrays (bags and
// sequences) are joined into one string, see XmpDatum.String; use
// XmpDatum.Values to get the elements of an array separately.
func (d *XmpData) AllTags() map[string]string {
	keyValues := map[string]string{}
	for i := d.Iterator(); i.HasNext(); {
		d := i.Next()
		keyValues[d.Key()] = d.String()
	}

	return keyValues
}

// Key returns the XMP key of the datum.
func (d *XmpDatum) Key() string {
	ckey := C.exiv2_xmp_datum_key(d.datum)
	defer C.free(unsafe.Pointer(ckey))

	return C.GoString(ckey)
}

// Values returns the elements of the datum. Simple properties yield a single
// element, while arrays (bags, sequences and language alternatives) yield one
// element per entry, in order.
func (d *XmpDatum) Values() []string {
	n := int(C.exiv2_xmp_datum_count(d.datum))
	values := make([]string, 0, n)
	for j := 0; j < n; j++ {
		cstr := C.exiv2_xmp_datum_to_string_at(d.datum, C.long(j))
		values = append(values, C.GoString(cstr))
		C.free(unsafe.Pointer(cstr))
	}

	return values
}

func (d *XmpDatum) String() string {
	cstr := C.exiv2_xmp_datum_to_string(d.datum)
	defer C.free(unsafe.Pointer(cstr))

	return C.GoString(cstr)
}

func (i *Image) XmpStripKey(key string) error {
	return i.StripKey(XMP, key)
}
