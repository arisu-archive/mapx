package mapx

import (
	"bytes"
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
)

type node[K comparable, V any] struct {
	key  K
	val  V
	prev *node[K, V]
	next *node[K, V]
}

// OrderedMap keeps insertion order for iteration and JSON marshalling.
// Existing keys updated via Set keep their original position.
type OrderedMap[K comparable, V any] struct {
	m    map[K]*node[K, V]
	head *node[K, V]
	tail *node[K, V]
}

// New returns an empty ordered map.
func New[K comparable, V any]() *OrderedMap[K, V] {
	return &OrderedMap[K, V]{m: make(map[K]*node[K, V])}
}

// Len returns the number of items.
func (om *OrderedMap[K, V]) Len() int {
	return len(om.m)
}

// IsZero reports whether the map contains no entries.
// This enables encoding/json to respect omitempty and omitzero struct tags.
func (om *OrderedMap[K, V]) IsZero() bool {
	return len(om.m) == 0
}

// Has reports whether k exists.
func (om *OrderedMap[K, V]) Has(k K) bool {
	_, ok := om.m[k]
	return ok
}

// Get returns the value for k.
func (om *OrderedMap[K, V]) Get(k K) (V, bool) {
	if n, ok := om.m[k]; ok {
		return n.val, true
	}
	var zero V
	return zero, false
}

// Set inserts k with value v preserving insertion order.
// If k exists, only its value is updated.
func (om *OrderedMap[K, V]) Set(k K, v V) {
	if om.m == nil {
		om.m = make(map[K]*node[K, V])
		om.head = nil
		om.tail = nil
	}
	if n, ok := om.m[k]; ok {
		n.val = v
		return
	}
	n := &node[K, V]{key: k, val: v}
	if om.tail == nil {
		om.head = n
		om.tail = n
	} else {
		n.prev = om.tail
		om.tail.next = n
		om.tail = n
	}
	om.m[k] = n
}

// Delete removes k if present. Returns true if removed.
func (om *OrderedMap[K, V]) Delete(k K) bool {
	n, ok := om.m[k]
	if !ok {
		return false
	}
	if n.prev != nil {
		n.prev.next = n.next
	} else {
		om.head = n.next
	}
	if n.next != nil {
		n.next.prev = n.prev
	} else {
		om.tail = n.prev
	}
	delete(om.m, k)
	return true
}

// Clear removes all entries.
func (om *OrderedMap[K, V]) Clear() {
	om.m = make(map[K]*node[K, V])
	om.head = nil
	om.tail = nil
}

// Keys returns keys in insertion order.
func (om *OrderedMap[K, V]) Keys() []K {
	out := make([]K, 0, len(om.m))
	for n := om.head; n != nil; n = n.next {
		out = append(out, n.key)
	}
	return out
}

// Values returns values in insertion order.
func (om *OrderedMap[K, V]) Values() []V {
	out := make([]V, 0, len(om.m))
	for n := om.head; n != nil; n = n.next {
		out = append(out, n.val)
	}
	return out
}

// Range iterates in insertion order; stop when f returns false.
func (om *OrderedMap[K, V]) Range(f func(k K, v V) bool) {
	for n := om.head; n != nil; n = n.next {
		if !f(n.key, n.val) {
			return
		}
	}
}

// ToMap copies to a regular map (order lost).
func (om *OrderedMap[K, V]) ToMap() map[K]V {
	res := make(map[K]V, len(om.m))
	for k, n := range om.m {
		res[k] = n.val
	}
	return res
}

func keyToString[K comparable](k K) (string, error) {
	if tm, ok := any(k).(encoding.TextMarshaler); ok {
		b, err := tm.MarshalText()
		return string(b), err
	}
	rv := reflect.ValueOf(k)
	switch rv.Kind() {
	case reflect.String:
		return rv.String(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(rv.Int(), 10), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(rv.Uint(), 10), nil
	default:
		return "", fmt.Errorf("json: unsupported map key type: %T", k)
	}
}

func (om *OrderedMap[K, V]) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')
	first := true
	for n := om.head; n != nil; n = n.next {
		if !first {
			buf.WriteByte(',')
		}
		first = false
		ks, err := keyToString(n.key)
		if err != nil {
			return nil, err
		}
		kb, _ := json.Marshal(ks)
		vb, err := json.Marshal(n.val)
		if err != nil {
			return nil, err
		}
		buf.Write(kb)
		buf.WriteByte(':')
		buf.Write(vb)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

func (om *OrderedMap[K, V]) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if string(trimmed) == "null" {
		return nil
	}
	if om.m == nil { // zero-value receiver
		om.m = make(map[K]*node[K, V])
	}
	dec := json.NewDecoder(bytes.NewReader(trimmed))
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	if d, ok := tok.(json.Delim); !ok || d != '{' {
		return fmt.Errorf("json: cannot unmarshal %T into OrderedMap (expected object)", tok)
	}

	parseKey := func(s string) (K, error) {
		var zero K
		if u, ok := any(&zero).(encoding.TextUnmarshaler); ok {
			if err := u.UnmarshalText([]byte(s)); err != nil {
				return zero, err
			}
			return zero, nil
		}
		t := reflect.TypeOf(zero)
		switch t.Kind() {
		case reflect.String:
			rv := reflect.New(t).Elem()
			rv.SetString(s)
			return rv.Interface().(K), nil
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			i, err := strconv.ParseInt(s, 10, int(t.Bits()))
			if err != nil {
				return zero, err
			}
			rv := reflect.New(t).Elem()
			rv.SetInt(i)
			return rv.Interface().(K), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			u, err := strconv.ParseUint(s, 10, int(t.Bits()))
			if err != nil {
				return zero, err
			}
			rv := reflect.New(t).Elem()
			rv.SetUint(u)
			return rv.Interface().(K), nil
		default:
			return zero, fmt.Errorf("json: unsupported map key type: %T", zero)
		}
	}

	for dec.More() {
		tk, err := dec.Token()
		if err != nil {
			return err
		}
		keyStr, ok := tk.(string)
		if !ok {
			return fmt.Errorf("json: expected object key string, got %T", tk)
		}
		var v V
		if err := dec.Decode(&v); err != nil {
			return err
		}
		k, err := parseKey(keyStr)
		if err != nil {
			return err
		}
		om.Set(k, v)
	}
	_, err = dec.Token()
	return err
}
