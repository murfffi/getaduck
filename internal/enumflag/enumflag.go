package enumflag

// Adapted from https://github.com/creachadair/goflags/blob/main/enumflag/flag.go
// under BSD 3-Clause license

import (
	"fmt"
	"strings"
)

// A Value represents an enumeration of string values.  A pointer to a Value
// satisfies the flag.Value interface. Use the Key method to recover the
// currently-selected value of the enumeration.
type Value struct {
	keys  []string
	index int // The selected index in the enumeration
}

// Help concatenates a human-readable string summarizing the legal values of v
// to h, for use in generating a documentation string.
func (v *Value) Help(h string) string {
	return fmt.Sprintf("%s (%s)", h, strings.Join(v.keys, "|"))
}

// New returns a *Value for the specified enumerators, where defaultKey is the
// default value and otherKeys are additional options. The index of a selected
// key reflects its position in the order given to this function, so that if:
//
//	v := enumflag.New("a", "b", "c", "d")
//
// then the index of "a" is 0, "b" is 1, "c" is 2, "d" is 3. The default key is
// always stored at index 0.
func New(defaultKey string, otherKeys ...string) *Value {
	return &Value{keys: append([]string{defaultKey}, otherKeys...)}
}

// Key returns the currently-selected key in the enumeration.  The original
// spelling of the selected value is returned, as given to the constructor, not
// the value as parsed.
func (v *Value) Key() string {
	if len(v.keys) == 0 {
		return "" // BUG: https://github.com/golang/go/issues/16694
	}
	return v.keys[v.index]
}

// Get satisfies the flag.Getter interface.
// The concrete value is the the string of the current key.
func (v *Value) Get() any { return v.Key() }

// Index returns the currently-selected index in the enumeration.
// The order of keys reflects the original order in which they were passed to
// the constructor, so index 0 is the default value.
func (v *Value) Index() int { return v.index }

// String satisfies part of the flag.Value interface.
func (v *Value) String() string { return fmt.Sprintf("%q", v.Key()) }

// Set satisfies part of the flag.Value interface.
func (v *Value) Set(s string) error {
	for i, key := range v.keys {
		if strings.EqualFold(s, key) {
			v.index = i
			return nil
		}
	}
	return fmt.Errorf("expected one of (%s)", strings.Join(v.keys, "|"))
}
