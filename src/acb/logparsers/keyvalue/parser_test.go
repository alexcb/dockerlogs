package keyvalue_test

import (
	"acb/logparsers/keyvalue"
	"reflect"
	"strings"
	"testing"
)

// Ensure the parser can parse strings into Statement ASTs.
func TestParser_ParseStatement(t *testing.T) {
	var tests = []struct {
		s         string
		keyvalues []keyvalue.KeyValuePair
		err       string
	}{
		{
			s: `key=value`,
			keyvalues: []keyvalue.KeyValuePair{
				{Key: "key", Value: "value"},
			},
		},
		{
			s: `key=value key2=val2`,
			keyvalues: []keyvalue.KeyValuePair{
				{Key: "key", Value: "value"},
				{Key: "key2", Value: "val2"},
			},
		},
		{
			s: `key=value key2="val2=val"`,
			keyvalues: []keyvalue.KeyValuePair{
				{Key: "key", Value: "value"},
				{Key: "key2", Value: "val2=val"},
			},
		},
	}

	for i, tt := range tests {
		keyvalues, err := keyvalue.NewParser(strings.NewReader(tt.s)).Parse()
		if !reflect.DeepEqual(tt.err, errstring(err)) {
			t.Errorf("%d. %q: error mismatch:\n  exp=%s\n  got=%s\n\n", i, tt.s, tt.err, err)
		} else if tt.err == "" && !reflect.DeepEqual(tt.keyvalues, keyvalues) {
			t.Errorf("%d. %q\n\nkeyvalue mismatch:\n\nexp=%#v\n\ngot=%#v\n\n", i, tt.s, tt.keyvalues, keyvalues)
		}
	}
}

// errstring returns the string representation of an error.
func errstring(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}
