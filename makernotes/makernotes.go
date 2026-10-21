// Package makernotes maps the tags exiv2 reports for camera makernotes to the
// names and values exiftool uses for them.
//
// exiv2 leaves many makernote tags unnamed (Exif.CanonFi.0x000a) and uses its
// own names for others (Exif.CanonCs.Macro where exiftool says MacroMode). The
// tables in this package are generated from exiftool's own tag definitions, so
// they track its naming. They are keyed on the exiv2 group name and numeric
// tag, which exiv2 reports for every makernote datum.
package makernotes

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Tag describes one makernote tag as exiftool knows it.
type Tag struct {
	// Name is the exiftool tag name, e.g. "MacroMode".
	Name string
	// Print maps raw integer values to the text exiftool prints for them. It
	// is nil for tags without such a conversion.
	Print map[int64]string
}

// Entry is one translated value.
type Entry struct {
	// Key is the exiftool style key, "<family 1 group>:<Name>", e.g.
	// "Canon:MacroMode".
	Key string
	// Value is the raw value, or its exiftool print conversion if requested
	// and known.
	Value string
}

// Lookup returns the exiftool tag for the given exiv2 group (the second part
// of a key like Exif.CanonCs.Macro) and numeric tag.
func Lookup(group string, tag int) (Tag, bool) {
	if t, ok := canonTables[group][tag]; ok {
		return t, true
	}
	return Tag{}, false
}

// Group returns the exiftool family 1 group name for an exiv2 group, or false
// if the group is not a known makernote group.
func Group(group string) (string, bool) {
	if _, ok := canonTables[group]; ok {
		return "Canon", true
	}
	return "", false
}

// Translate maps one exiv2 datum to its exiftool entries.
//
// group is the exiv2 group name, tag the numeric tag of the datum, count its
// number of values and value its space separated raw string value. In the
// array style makernote groups (CanonCs, CanonSi, ...) a datum with count n
// covers the n tags starting at tag, which exiftool lists separately, so it
// yields up to n entries. If print is true, values are converted to the text
// exiftool prints where a conversion is known.
//
// Translate returns nil if the group is unknown or the tag is not named.
func Translate(group string, tag, count int, value string, print bool) []Entry {
	family, ok := Group(group)
	if !ok {
		return nil
	}

	values := strings.Fields(value)
	if count > 1 && len(values) == count && isArrayGroup(group) {
		var entries []Entry
		for i, v := range values {
			if e, ok := entry(family, group, tag+i, v, print); ok {
				entries = append(entries, e)
			}
		}
		return entries
	}

	if e, ok := entry(family, group, tag, value, print); ok {
		return []Entry{e}
	}
	return nil
}

func entry(family, group string, tag int, value string, print bool) (Entry, bool) {
	t, ok := Lookup(group, tag)
	if !ok {
		return Entry{}, false
	}
	if print && t.Print != nil {
		if n, err := strconv.ParseInt(value, 10, 64); err == nil {
			s, ok := t.Print[n]
			if !ok && n > math.MaxInt16 && n <= math.MaxUint16 {
				// exiv2 reads these as unsigned shorts, exiftool as signed
				s, ok = t.Print[n-(math.MaxUint16+1)]
			}
			if ok {
				value = s
			}
		}
	}
	return Entry{Key: fmt.Sprintf("%s:%s", family, t.Name), Value: value}, true
}

// isArrayGroup reports whether the group is stored as an array of values
// indexed by tag, as opposed to a regular IFD of independent tags.
func isArrayGroup(group string) bool {
	return group != "Canon"
}
