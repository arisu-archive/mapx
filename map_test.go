package mapx_test

import (
	"encoding/json"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/arisu-archive/mapx"
)

// textKey is a custom key type implementing encoding.TextMarshaler/TextUnmarshaler.
type textKey struct {
	A string
	B string
}

func (t textKey) MarshalText() ([]byte, error) {
	return []byte(t.A + ":" + t.B), nil
}

func (t *textKey) UnmarshalText(b []byte) error {
	for i, c := range b {
		if c == ':' {
			t.A = string(b[:i])
			t.B = string(b[i+1:])
			return nil
		}
	}
	return fmt.Errorf("invalid textKey: %q", string(b))
}

var _ = Describe("OrderedMap", func() {
	Describe("New", func() {
		It("should return an empty map", func() {
			om := mapx.New[string, int]()
			Expect(om).NotTo(BeNil())
			Expect(om.Len()).To(Equal(0))
		})
	})

	Describe("Set and Get", func() {
		var om *mapx.OrderedMap[string, int]

		BeforeEach(func() {
			om = mapx.New[string, int]()
		})

		Context("on an empty map", func() {
			It("should insert a new key and retrieve it", func() {
				om.Set("a", 1)

				v, ok := om.Get("a")
				Expect(ok).To(BeTrue())
				Expect(v).To(Equal(1))
				Expect(om.Len()).To(Equal(1))
			})
		})

		Context("on a zero-value receiver", func() {
			It("should initialize internal state and work correctly", func() {
				var zv mapx.OrderedMap[string, int]
				zv.Set("x", 42)

				v, ok := zv.Get("x")
				Expect(ok).To(BeTrue())
				Expect(v).To(Equal(42))
				Expect(zv.Len()).To(Equal(1))
			})
		})

		Context("when the key already exists", func() {
			It("should update the value without changing insertion order", func() {
				om.Set("a", 1)
				om.Set("b", 2)
				om.Set("c", 3)

				om.Set("a", 10)

				v, ok := om.Get("a")
				Expect(ok).To(BeTrue())
				Expect(v).To(Equal(10))
				Expect(om.Keys()).To(Equal([]string{"a", "b", "c"}))
			})
		})

		Context("with multiple keys", func() {
			It("should maintain insertion order", func() {
				om.Set("c", 3)
				om.Set("a", 1)
				om.Set("b", 2)

				Expect(om.Keys()).To(Equal([]string{"c", "a", "b"}))
				Expect(om.Values()).To(Equal([]int{3, 1, 2}))
			})
		})
	})

	Describe("Has", func() {
		var om *mapx.OrderedMap[string, int]

		BeforeEach(func() {
			om = mapx.New[string, int]()
			om.Set("x", 1)
		})

		It("should return true for an existing key", func() {
			Expect(om.Has("x")).To(BeTrue())
		})

		It("should return false for a missing key", func() {
			Expect(om.Has("y")).To(BeFalse())
		})
	})

	Describe("Len", func() {
		It("should return 0 for an empty map", func() {
			om := mapx.New[string, int]()
			Expect(om.Len()).To(Equal(0))
		})

		It("should reflect inserts and deletes", func() {
			om := mapx.New[string, int]()
			om.Set("a", 1)
			om.Set("b", 2)
			Expect(om.Len()).To(Equal(2))

			om.Delete("a")
			Expect(om.Len()).To(Equal(1))
		})
	})

	Describe("Delete", func() {
		It("should return false for a missing key", func() {
			om := mapx.New[string, int]()
			Expect(om.Delete("nope")).To(BeFalse())
		})

		Context("when deleting the only element", func() {
			It("should leave the map empty", func() {
				om := mapx.New[string, int]()
				om.Set("a", 1)

				Expect(om.Delete("a")).To(BeTrue())
				Expect(om.Len()).To(Equal(0))
				Expect(om.Keys()).To(BeEmpty())
			})
		})

		Context("when deleting the head", func() {
			It("should promote the next node and preserve order of rest", func() {
				om := mapx.New[string, int]()
				om.Set("a", 1)
				om.Set("b", 2)
				om.Set("c", 3)

				om.Delete("a")
				Expect(om.Keys()).To(Equal([]string{"b", "c"}))
			})
		})

		Context("when deleting the tail", func() {
			It("should move the tail to the previous node", func() {
				om := mapx.New[string, int]()
				om.Set("a", 1)
				om.Set("b", 2)
				om.Set("c", 3)

				om.Delete("c")
				Expect(om.Keys()).To(Equal([]string{"a", "b"}))
			})
		})

		Context("when deleting a middle element", func() {
			It("should stitch prev and next correctly", func() {
				om := mapx.New[string, int]()
				om.Set("a", 1)
				om.Set("b", 2)
				om.Set("c", 3)

				om.Delete("b")
				Expect(om.Keys()).To(Equal([]string{"a", "c"}))
				Expect(om.Values()).To(Equal([]int{1, 3}))
			})
		})
	})

	Describe("Clear", func() {
		It("should remove all entries and reset Len to 0", func() {
			om := mapx.New[string, int]()
			om.Set("a", 1)
			om.Set("b", 2)

			om.Clear()
			Expect(om.Len()).To(Equal(0))
			Expect(om.Keys()).To(BeEmpty())
		})
	})

	Describe("Keys", func() {
		It("should return keys in insertion order", func() {
			om := mapx.New[string, int]()
			om.Set("z", 26)
			om.Set("a", 1)
			om.Set("m", 13)

			Expect(om.Keys()).To(Equal([]string{"z", "a", "m"}))
		})

		It("should return an empty slice for an empty map", func() {
			om := mapx.New[string, int]()
			Expect(om.Keys()).To(BeEmpty())
		})
	})

	Describe("Values", func() {
		It("should return values in insertion order", func() {
			om := mapx.New[string, int]()
			om.Set("a", 10)
			om.Set("b", 20)
			om.Set("c", 30)

			Expect(om.Values()).To(Equal([]int{10, 20, 30}))
		})

		It("should return an empty slice for an empty map", func() {
			om := mapx.New[string, int]()
			Expect(om.Values()).To(BeEmpty())
		})
	})

	Describe("Range", func() {
		It("should visit all entries in insertion order", func() {
			om := mapx.New[string, int]()
			om.Set("x", 1)
			om.Set("y", 2)
			om.Set("z", 3)

			var keys []string
			var vals []int
			om.Range(func(k string, v int) bool {
				keys = append(keys, k)
				vals = append(vals, v)
				return true
			})
			Expect(keys).To(Equal([]string{"x", "y", "z"}))
			Expect(vals).To(Equal([]int{1, 2, 3}))
		})

		It("should stop early when the callback returns false", func() {
			om := mapx.New[string, int]()
			om.Set("a", 1)
			om.Set("b", 2)
			om.Set("c", 3)

			var visited []string
			om.Range(func(k string, v int) bool {
				visited = append(visited, k)
				return k != "b" // stop after "b"
			})
			Expect(visited).To(Equal([]string{"a", "b"}))
		})
	})

	Describe("ToMap", func() {
		It("should return a plain map with all key-value pairs", func() {
			om := mapx.New[string, int]()
			om.Set("a", 1)
			om.Set("b", 2)

			Expect(om.ToMap()).To(Equal(map[string]int{"a": 1, "b": 2}))
		})

		It("should return an empty map for an empty OrderedMap", func() {
			om := mapx.New[string, int]()
			Expect(om.ToMap()).To(BeEmpty())
		})
	})

	Describe("MarshalJSON", func() {
		It("should produce keys in insertion order", func() {
			om := mapx.New[string, string]()
			om.Set("b", "beta")
			om.Set("a", "alpha")
			om.Set("c", "gamma")

			b, err := json.Marshal(om)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(b)).To(Equal(`{"b":"beta","a":"alpha","c":"gamma"}`))
		})

		It("should marshal an empty map as {}", func() {
			om := mapx.New[string, int]()

			b, err := json.Marshal(om)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(b)).To(Equal(`{}`))
		})

		Context("with integer keys", func() {
			It("should quote keys as strings per JSON spec", func() {
				om := mapx.New[int, string]()
				om.Set(1, "one")
				om.Set(2, "two")

				b, err := json.Marshal(om)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(b)).To(Equal(`{"1":"one","2":"two"}`))
			})
		})

		Context("with TextMarshaler keys", func() {
			It("should use MarshalText for key encoding", func() {
				om := mapx.New[textKey, int]()
				om.Set(textKey{A: "foo", B: "bar"}, 1)

				b, err := json.Marshal(om)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(b)).To(Equal(`{"foo:bar":1}`))
			})
		})
	})

	Describe("UnmarshalJSON", func() {
		It("should decode a JSON object preserving key order", func() {
			data := []byte(`{"c":3,"a":1,"b":2}`)
			om := mapx.New[string, int]()

			Expect(json.Unmarshal(data, om)).To(Succeed())
			Expect(om.Keys()).To(Equal([]string{"c", "a", "b"}))
			Expect(om.Values()).To(Equal([]int{3, 1, 2}))
		})

		It("should be a no-op for JSON null", func() {
			om := mapx.New[string, int]()
			om.Set("pre", 1)

			Expect(json.Unmarshal([]byte("null"), om)).To(Succeed())
			// Existing data remains untouched.
			Expect(om.Len()).To(Equal(1))
		})

		It("should return an error for non-object input", func() {
			om := mapx.New[string, int]()
			err := json.Unmarshal([]byte(`[1,2,3]`), om)
			Expect(err).To(HaveOccurred())
		})

		Context("with integer keys", func() {
			It("should parse string-encoded numeric keys", func() {
				data := []byte(`{"10":"ten","20":"twenty"}`)
				om := mapx.New[int, string]()

				Expect(json.Unmarshal(data, om)).To(Succeed())
				Expect(om.Keys()).To(Equal([]int{10, 20}))
				Expect(om.Values()).To(Equal([]string{"ten", "twenty"}))
			})
		})

		Context("with TextUnmarshaler keys", func() {
			It("should use UnmarshalText for key decoding", func() {
				data := []byte(`{"foo:bar":1}`)
				om := mapx.New[textKey, int]()

				Expect(json.Unmarshal(data, om)).To(Succeed())

				keys := om.Keys()
				Expect(keys).To(HaveLen(1))
				Expect(keys[0]).To(Equal(textKey{A: "foo", B: "bar"}))
			})
		})

		Context("round-trip", func() {
			It("should marshal then unmarshal back to the same ordered map", func() {
				original := mapx.New[string, int]()
				original.Set("z", 26)
				original.Set("a", 1)
				original.Set("m", 13)

				b, err := json.Marshal(original)
				Expect(err).NotTo(HaveOccurred())

				restored := mapx.New[string, int]()
				Expect(json.Unmarshal(b, restored)).To(Succeed())

				Expect(restored.Keys()).To(Equal(original.Keys()))
				Expect(restored.Values()).To(Equal(original.Values()))
			})
		})
	})
})
