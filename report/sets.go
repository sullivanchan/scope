package report

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"

	"github.com/ugorji/go/codec"
	"github.com/weaveworks/ps"
)

// Sets is a string->set-of-strings map.
// It is immutable.
type Sets struct {
	psMap ps.Map
}

// EmptySets is an empty Sets.  Starts with this.
var EmptySets = Sets{ps.NewMap()}

// MakeSets returns EmptySets
func MakeSets() Sets {
	return EmptySets
}

// Keys returns the keys for this set
func (s Sets) Keys() []string {
	if s.psMap == nil {
		return nil
	}
	return s.psMap.Keys()
}

// Add the given value to the Sets.
func (s Sets) Add(key string, value StringSet) Sets {
	if s.psMap == nil {
		s = EmptySets
	}
	if existingValue, ok := s.psMap.Lookup(key); ok {
		value = value.Merge(existingValue.(StringSet))
	}
	return Sets{
		psMap: s.psMap.Set(key, value),
	}
}

// Delete the given set from the Sets.
func (s Sets) Delete(key string) Sets {
	if s.psMap == nil {
		return EmptySets
	}
	return Sets{
		psMap: s.psMap.Delete(key),
	}
}

// Lookup returns the sets stored under key.
func (s Sets) Lookup(key string) (StringSet, bool) {
	if s.psMap == nil {
		return EmptyStringSet, false
	}
	if value, ok := s.psMap.Lookup(key); ok {
		return value.(StringSet), true
	}
	return EmptyStringSet, false
}

// Size returns the number of elements
func (s Sets) Size() int {
	if s.psMap == nil {
		return 0
	}
	return s.psMap.Size()
}

// Merge merges two sets maps into a fresh set, performing set-union merges as
// appropriate.
func (s Sets) Merge(other Sets) Sets {
	var (
		sSize     = s.Size()
		otherSize = other.Size()
		result    = s.psMap
		iter      = other.psMap
	)
	switch {
	case sSize == 0:
		return other
	case otherSize == 0:
		return s
	case sSize < otherSize:
		result, iter = iter, result
	}

	iter.ForEach(func(key string, value interface{}) {
		set := value.(StringSet)
		if existingSet, ok := result.Lookup(key); ok {
			set = set.Merge(existingSet.(StringSet))
		}
		result = result.Set(key, set)
	})

	return Sets{result}
}

// Copy is a noop
func (s Sets) Copy() Sets {
	return s
}

func (s Sets) String() string {
	if s.psMap == nil {
		s = EmptySets
	}
	keys := []string{}
	for _, k := range s.psMap.Keys() {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	buf := bytes.NewBufferString("{")
	for _, key := range keys {
		val, _ := s.psMap.Lookup(key)
		fmt.Fprintf(buf, "%s: %v, ", key, val)
	}
	fmt.Fprintf(buf, "}")
	return buf.String()
}

// DeepEqual tests equality with other Sets
func (s Sets) DeepEqual(t Sets) bool {
	if s.Size() != t.Size() {
		return false
	}
	if s.Size() == 0 {
		return true
	}

	equal := true
	s.psMap.ForEach(func(k string, val interface{}) {
		if otherValue, ok := t.psMap.Lookup(k); !ok {
			equal = false
		} else {
			equal = equal && reflect.DeepEqual(val, otherValue)
		}
	})
	return equal
}

// CodecEncodeSelf implements codec.Selfer
func (s *Sets) CodecEncodeSelf(encoder *codec.Encoder) {
	mapWrite(s.psMap, encoder, func(encoder *codec.Encoder, val interface{}) {
		encoder.Encode(val.(StringSet))
	})
}

// CodecDecodeSelf implements codec.Selfer
func (s *Sets) CodecDecodeSelf(decoder *codec.Decoder) {
	out := mapRead(decoder, func(isNil bool) interface{} {
		var value StringSet
		if !isNil {
			decoder.Decode(&value)
		}
		return value
	})
	*s = Sets{out}
}

// MarshalJSON shouldn't be used, use CodecEncodeSelf instead
func (Sets) MarshalJSON() ([]byte, error) {
	panic("MarshalJSON shouldn't be used, use CodecEncodeSelf instead")
}

// UnmarshalJSON shouldn't be used, use CodecDecodeSelf instead
func (*Sets) UnmarshalJSON(b []byte) error {
	panic("UnmarshalJSON shouldn't be used, use CodecDecodeSelf instead")
}
