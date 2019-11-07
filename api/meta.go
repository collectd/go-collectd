// Package api defines data types representing core collectd data types.
package api // import "collectd.org/api"

// Author: Remi Ferrand <remi.ferrand_at_cc.in2p3.fr>

import (
	"sort"
	"strconv"
)

// Metadata represents a ValueList metadata. It's Go's equivalent to the
// linked list `meta_data_s`
// Currently only int64, uint64, float64, string and bool metadata
// types are supported
type Metadata map[string]interface{}

// Get fetches metadata associated with `key`
func (m Metadata) Get(key string) interface{} {
	return m[key]
}

// GetAsString returns the value as a string, regardless of the type
// It's Go's equivalent to the `meta_data_as_string` function
func (m Metadata) GetAsString(key string) string {
	v := m.Get(key)

	switch v.(type) {
	case int64:
		return strconv.FormatInt(v.(int64), 10)
	case uint64:
		return strconv.FormatUint(v.(uint64), 10)
	case float64:
		return strconv.FormatFloat(v.(float64), 'e', 5, 64)
	case bool:
		return strconv.FormatBool(v.(bool))
	case string:
		return v.(string)
	default:
		panic("unsupported type")
	}
}

// Set adds a new metadata
// If `key` is already present, old value will
// be overwritten
func (m *Metadata) Set(key string, value interface{}) {
	(*m)[key] = value
}

// Delete removes metadata associated with `key`
func (m *Metadata) Delete(key string) {
	delete(*m, key)
}

// Exists checks if a given metadata key exists
func (m Metadata) Exists(key string) bool {
	for k := range m {
		if k == key {
			return true
		}
	}
	return false
}

// Toc returns all the metadata keys
// keys will always be sorted using Go
// sort.Strings() function
func (m Metadata) Toc() []string {
	keys := make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return keys
}

// Clone returns a cloned copy of the Metadata
func (m Metadata) Clone() Metadata {
	nMeta := make(Metadata)
	for k := range m {
		nMeta.Set(k, m.Get(k))
	}
	return nMeta
}

// CloneMerge merges data metadata from `orig`
// into a cloned copy of the current Metadata object
// If a key exists both in the current Metadata object
// and in `orig`, the key from `orig` will be kept
func (m Metadata) CloneMerge(orig Metadata) Metadata {
	nMeta := m.Clone()
	for k := range orig {
		nMeta[k] = orig.Get(k)
	}
	return nMeta
}
