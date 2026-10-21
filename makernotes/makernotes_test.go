package makernotes_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kolesa-team/goexiv/makernotes"
)

func TestLookup(t *testing.T) {
	tag, ok := makernotes.Lookup("CanonCs", 1)
	assert.True(t, ok)
	assert.Equal(t, "MacroMode", tag.Name)

	_, ok = makernotes.Lookup("CanonCs", 9999)
	assert.False(t, ok)

	_, ok = makernotes.Lookup("Nikon", 1)
	assert.False(t, ok)
}

func TestTranslate(t *testing.T) {
	tests := []struct {
		name  string
		group string
		tag   int
		count int
		value string
		print bool
		want  []makernotes.Entry
	}{
		{"raw", "CanonCs", 1, 1, "2", false, []makernotes.Entry{{"Canon:MacroMode", "2"}}},
		{"print", "CanonCs", 1, 1, "2", true, []makernotes.Entry{{"Canon:MacroMode", "Normal"}}},
		{"print without conversion", "CanonPr", 2, 1, "4", true, []makernotes.Entry{{"Canon:Sharpness", "4"}}},
		{"print unknown value", "CanonCs", 1, 1, "12345", true, []makernotes.Entry{{"Canon:MacroMode", "12345"}}},
		{"print signed short", "CanonCs", 3, 1, "65535", true, []makernotes.Entry{{"Canon:Quality", "n/a"}}},
		{"array is split", "CanonCs", 23, 3, "105 24 1", false, []makernotes.Entry{
			{"Canon:MaxFocalLength", "105"},
			{"Canon:MinFocalLength", "24"},
			{"Canon:FocalUnits", "1"},
		}},
		{"unnamed tag", "CanonFi", 10, 1, "0", false, nil},
		{"unknown group", "Nikon3", 1, 1, "0", false, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, makernotes.Translate(tt.group, tt.tag, tt.count, tt.value, tt.print))
		})
	}
}
