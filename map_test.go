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

		Context("with int8 keys", func() {
			It("should quote keys as strings", func() {
				om := mapx.New[int8, string]()
				om.Set(1, "one")
				om.Set(-2, "neg two")

				b, err := json.Marshal(om)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(b)).To(Equal(`{"1":"one","-2":"neg two"}`))
			})
		})

		Context("with int16 keys", func() {
			It("should quote keys as strings", func() {
				om := mapx.New[int16, string]()
				om.Set(300, "three hundred")

				b, err := json.Marshal(om)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(b)).To(Equal(`{"300":"three hundred"}`))
			})
		})

		Context("with int32 keys", func() {
			It("should quote keys as strings", func() {
				om := mapx.New[int32, string]()
				om.Set(100000, "big")

				b, err := json.Marshal(om)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(b)).To(Equal(`{"100000":"big"}`))
			})
		})

		Context("with int64 keys", func() {
			It("should quote keys as strings", func() {
				om := mapx.New[int64, string]()
				om.Set(9999999999, "huge")

				b, err := json.Marshal(om)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(b)).To(Equal(`{"9999999999":"huge"}`))
			})
		})

		Context("with uint keys", func() {
			It("should quote keys as strings", func() {
				om := mapx.New[uint, string]()
				om.Set(42, "answer")

				b, err := json.Marshal(om)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(b)).To(Equal(`{"42":"answer"}`))
			})
		})

		Context("with uint8 keys", func() {
			It("should quote keys as strings", func() {
				om := mapx.New[uint8, string]()
				om.Set(255, "max")

				b, err := json.Marshal(om)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(b)).To(Equal(`{"255":"max"}`))
			})
		})

		Context("with uint16 keys", func() {
			It("should quote keys as strings", func() {
				om := mapx.New[uint16, string]()
				om.Set(65535, "max")

				b, err := json.Marshal(om)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(b)).To(Equal(`{"65535":"max"}`))
			})
		})

		Context("with uint32 keys", func() {
			It("should quote keys as strings", func() {
				om := mapx.New[uint32, string]()
				om.Set(4294967295, "max")

				b, err := json.Marshal(om)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(b)).To(Equal(`{"4294967295":"max"}`))
			})
		})

		Context("with uint64 keys", func() {
			It("should quote keys as strings", func() {
				om := mapx.New[uint64, string]()
				om.Set(18446744073709551615, "max")

				b, err := json.Marshal(om)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(b)).To(Equal(`{"18446744073709551615":"max"}`))
			})
		})

		Context("with uintptr keys", func() {
			It("should quote keys as strings", func() {
				om := mapx.New[uintptr, string]()
				om.Set(12345, "ptr")

				b, err := json.Marshal(om)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(b)).To(Equal(`{"12345":"ptr"}`))
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

		Context("with unsupported key type", func() {
			It("should return an error", func() {
				om := mapx.New[float64, string]()
				om.Set(1.5, "x")

				_, err := json.Marshal(om)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("with boolean values", func() {
			It("should marshal true and false", func() {
				om := mapx.New[string, bool]()
				om.Set("yes", true)
				om.Set("no", false)

				b, err := json.Marshal(om)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(b)).To(Equal(`{"yes":true,"no":false}`))
			})
		})

		Context("with float values", func() {
			It("should marshal float64 numbers", func() {
				om := mapx.New[string, float64]()
				om.Set("pi", 3.14)
				om.Set("neg", -0.5)

				b, err := json.Marshal(om)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(b)).To(Equal(`{"pi":3.14,"neg":-0.5}`))
			})
		})

		Context("with nil/pointer values", func() {
			It("should marshal nil pointers as null", func() {
				om := mapx.New[string, *int]()
				v := 42
				om.Set("present", &v)
				om.Set("absent", nil)

				b, err := json.Marshal(om)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(b)).To(Equal(`{"present":42,"absent":null}`))
			})
		})

		Context("with slice/array values", func() {
			It("should marshal slices as JSON arrays", func() {
				om := mapx.New[string, []int]()
				om.Set("nums", []int{1, 2, 3})
				om.Set("empty", []int{})

				b, err := json.Marshal(om)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(b)).To(Equal(`{"nums":[1,2,3],"empty":[]}`))
			})
		})

		Context("with nested object values", func() {
			It("should marshal struct values as JSON objects", func() {
				type inner struct {
					X int    `json:"x"`
					Y string `json:"y"`
				}
				om := mapx.New[string, inner]()
				om.Set("point", inner{X: 10, Y: "hello"})

				b, err := json.Marshal(om)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(b)).To(Equal(`{"point":{"x":10,"y":"hello"}}`))
			})
		})

		Context("with map values", func() {
			It("should marshal map values as JSON objects", func() {
				om := mapx.New[string, map[string]int]()
				om.Set("scores", map[string]int{"a": 1})

				b, err := json.Marshal(om)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(b)).To(Equal(`{"scores":{"a":1}}`))
			})
		})

		Context("with interface{} values", func() {
			It("should marshal mixed JSON types", func() {
				om := mapx.New[string, any]()
				om.Set("str", "hello")
				om.Set("num", 42)
				om.Set("bool", true)
				om.Set("null", nil)
				om.Set("arr", []int{1, 2})

				b, err := json.Marshal(om)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(b)).To(Equal(`{"str":"hello","num":42,"bool":true,"null":null,"arr":[1,2]}`))
			})
		})

		Context("with nested OrderedMap values", func() {
			It("should recursively marshal nested ordered maps", func() {
				child := mapx.New[string, int]()
				child.Set("b", 2)
				child.Set("a", 1)

				parent := mapx.New[string, *mapx.OrderedMap[string, int]]()
				parent.Set("nested", child)

				b, err := json.Marshal(parent)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(b)).To(Equal(`{"nested":{"b":2,"a":1}}`))
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

		Context("with int8 keys", func() {
			It("should parse string-encoded int8 keys", func() {
				data := []byte(`{"1":"one","-2":"neg two"}`)
				om := mapx.New[int8, string]()

				Expect(json.Unmarshal(data, om)).To(Succeed())
				Expect(om.Keys()).To(Equal([]int8{1, -2}))
				Expect(om.Values()).To(Equal([]string{"one", "neg two"}))
			})
		})

		Context("with int16 keys", func() {
			It("should parse string-encoded int16 keys", func() {
				data := []byte(`{"300":"x"}`)
				om := mapx.New[int16, string]()

				Expect(json.Unmarshal(data, om)).To(Succeed())
				Expect(om.Keys()).To(Equal([]int16{300}))
			})
		})

		Context("with int32 keys", func() {
			It("should parse string-encoded int32 keys", func() {
				data := []byte(`{"100000":"x"}`)
				om := mapx.New[int32, string]()

				Expect(json.Unmarshal(data, om)).To(Succeed())
				Expect(om.Keys()).To(Equal([]int32{100000}))
			})
		})

		Context("with int64 keys", func() {
			It("should parse string-encoded int64 keys", func() {
				data := []byte(`{"9999999999":"x"}`)
				om := mapx.New[int64, string]()

				Expect(json.Unmarshal(data, om)).To(Succeed())
				Expect(om.Keys()).To(Equal([]int64{9999999999}))
			})
		})

		Context("with uint keys", func() {
			It("should parse string-encoded uint keys", func() {
				data := []byte(`{"42":"answer"}`)
				om := mapx.New[uint, string]()

				Expect(json.Unmarshal(data, om)).To(Succeed())
				Expect(om.Keys()).To(Equal([]uint{42}))
			})
		})

		Context("with uint8 keys", func() {
			It("should parse string-encoded uint8 keys", func() {
				data := []byte(`{"255":"max"}`)
				om := mapx.New[uint8, string]()

				Expect(json.Unmarshal(data, om)).To(Succeed())
				Expect(om.Keys()).To(Equal([]uint8{255}))
			})
		})

		Context("with uint16 keys", func() {
			It("should parse string-encoded uint16 keys", func() {
				data := []byte(`{"65535":"max"}`)
				om := mapx.New[uint16, string]()

				Expect(json.Unmarshal(data, om)).To(Succeed())
				Expect(om.Keys()).To(Equal([]uint16{65535}))
			})
		})

		Context("with uint32 keys", func() {
			It("should parse string-encoded uint32 keys", func() {
				data := []byte(`{"4294967295":"max"}`)
				om := mapx.New[uint32, string]()

				Expect(json.Unmarshal(data, om)).To(Succeed())
				Expect(om.Keys()).To(Equal([]uint32{4294967295}))
			})
		})

		Context("with uint64 keys", func() {
			It("should parse string-encoded uint64 keys", func() {
				data := []byte(`{"18446744073709551615":"max"}`)
				om := mapx.New[uint64, string]()

				Expect(json.Unmarshal(data, om)).To(Succeed())
				Expect(om.Keys()).To(Equal([]uint64{18446744073709551615}))
			})
		})

		Context("with uintptr keys", func() {
			It("should parse string-encoded uintptr keys", func() {
				data := []byte(`{"12345":"ptr"}`)
				om := mapx.New[uintptr, string]()

				Expect(json.Unmarshal(data, om)).To(Succeed())
				Expect(om.Keys()).To(Equal([]uintptr{12345}))
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

		Context("with unsupported key type", func() {
			It("should return an error", func() {
				data := []byte(`{"1.5":"x"}`)
				om := mapx.New[float64, string]()

				err := json.Unmarshal(data, om)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("with boolean values", func() {
			It("should decode true and false", func() {
				data := []byte(`{"yes":true,"no":false}`)
				om := mapx.New[string, bool]()

				Expect(json.Unmarshal(data, om)).To(Succeed())
				Expect(om.Keys()).To(Equal([]string{"yes", "no"}))

				v1, _ := om.Get("yes")
				v2, _ := om.Get("no")
				Expect(v1).To(BeTrue())
				Expect(v2).To(BeFalse())
			})
		})

		Context("with float values", func() {
			It("should decode float64 numbers", func() {
				data := []byte(`{"pi":3.14,"neg":-0.5}`)
				om := mapx.New[string, float64]()

				Expect(json.Unmarshal(data, om)).To(Succeed())
				Expect(om.Keys()).To(Equal([]string{"pi", "neg"}))

				v1, _ := om.Get("pi")
				v2, _ := om.Get("neg")
				Expect(v1).To(BeNumerically("~", 3.14))
				Expect(v2).To(BeNumerically("~", -0.5))
			})
		})

		Context("with null values", func() {
			It("should decode null as nil pointer", func() {
				data := []byte(`{"present":42,"absent":null}`)
				om := mapx.New[string, *int]()

				Expect(json.Unmarshal(data, om)).To(Succeed())
				Expect(om.Keys()).To(Equal([]string{"present", "absent"}))

				v1, _ := om.Get("present")
				v2, _ := om.Get("absent")
				Expect(*v1).To(Equal(42))
				Expect(v2).To(BeNil())
			})
		})

		Context("with array values", func() {
			It("should decode JSON arrays into slices", func() {
				data := []byte(`{"nums":[1,2,3],"empty":[]}`)
				om := mapx.New[string, []int]()

				Expect(json.Unmarshal(data, om)).To(Succeed())
				Expect(om.Keys()).To(Equal([]string{"nums", "empty"}))

				v1, _ := om.Get("nums")
				v2, _ := om.Get("empty")
				Expect(v1).To(Equal([]int{1, 2, 3}))
				Expect(v2).To(BeEmpty())
			})
		})

		Context("with nested object values", func() {
			It("should decode JSON objects into structs", func() {
				type inner struct {
					X int    `json:"x"`
					Y string `json:"y"`
				}
				data := []byte(`{"point":{"x":10,"y":"hello"}}`)
				om := mapx.New[string, inner]()

				Expect(json.Unmarshal(data, om)).To(Succeed())
				v, ok := om.Get("point")
				Expect(ok).To(BeTrue())
				Expect(v.X).To(Equal(10))
				Expect(v.Y).To(Equal("hello"))
			})
		})

		Context("with interface{} values", func() {
			It("should decode mixed JSON types", func() {
				data := []byte(`{"str":"hello","num":42,"bool":true,"null":null,"arr":[1,2]}`)
				om := mapx.New[string, any]()

				Expect(json.Unmarshal(data, om)).To(Succeed())
				Expect(om.Keys()).To(Equal([]string{"str", "num", "bool", "null", "arr"}))

				v1, _ := om.Get("str")
				Expect(v1).To(Equal("hello"))

				v2, _ := om.Get("num")
				Expect(v2).To(BeNumerically("==", 42))

				v3, _ := om.Get("bool")
				Expect(v3).To(Equal(true))

				v4, _ := om.Get("null")
				Expect(v4).To(BeNil())

				v5, _ := om.Get("arr")
				Expect(v5).To(HaveLen(2))
			})
		})

		Context("with nested OrderedMap values", func() {
			It("should recursively unmarshal nested ordered maps", func() {
				data := []byte(`{"nested":{"b":2,"a":1}}`)
				parent := mapx.New[string, *mapx.OrderedMap[string, int]]()

				Expect(json.Unmarshal(data, parent)).To(Succeed())

				child, ok := parent.Get("nested")
				Expect(ok).To(BeTrue())
				Expect(child.Keys()).To(Equal([]string{"b", "a"}))
				Expect(child.Values()).To(Equal([]int{2, 1}))
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

			It("should round-trip boolean values", func() {
				original := mapx.New[string, bool]()
				original.Set("t", true)
				original.Set("f", false)

				b, err := json.Marshal(original)
				Expect(err).NotTo(HaveOccurred())

				restored := mapx.New[string, bool]()
				Expect(json.Unmarshal(b, restored)).To(Succeed())
				Expect(restored.Keys()).To(Equal(original.Keys()))
				Expect(restored.Values()).To(Equal(original.Values()))
			})

			It("should round-trip float values", func() {
				original := mapx.New[string, float64]()
				original.Set("pi", 3.14)
				original.Set("e", 2.718)

				b, err := json.Marshal(original)
				Expect(err).NotTo(HaveOccurred())

				restored := mapx.New[string, float64]()
				Expect(json.Unmarshal(b, restored)).To(Succeed())
				Expect(restored.Keys()).To(Equal(original.Keys()))
				Expect(restored.Values()).To(Equal(original.Values()))
			})

			It("should round-trip slice values", func() {
				original := mapx.New[string, []string]()
				original.Set("tags", []string{"go", "json"})
				original.Set("empty", []string{})

				b, err := json.Marshal(original)
				Expect(err).NotTo(HaveOccurred())

				restored := mapx.New[string, []string]()
				Expect(json.Unmarshal(b, restored)).To(Succeed())
				Expect(restored.Keys()).To(Equal(original.Keys()))
				Expect(restored.Values()).To(Equal(original.Values()))
			})

			It("should round-trip nested OrderedMap values", func() {
				child := mapx.New[string, int]()
				child.Set("b", 2)
				child.Set("a", 1)

				original := mapx.New[string, *mapx.OrderedMap[string, int]]()
				original.Set("inner", child)

				b, err := json.Marshal(original)
				Expect(err).NotTo(HaveOccurred())

				restored := mapx.New[string, *mapx.OrderedMap[string, int]]()
				Expect(json.Unmarshal(b, restored)).To(Succeed())

				restoredChild, ok := restored.Get("inner")
				Expect(ok).To(BeTrue())
				Expect(restoredChild.Keys()).To(Equal(child.Keys()))
				Expect(restoredChild.Values()).To(Equal(child.Values()))
			})
		})
	})

	Describe("IsZero", func() {
		It("should return true for a new empty map", func() {
			om := mapx.New[string, int]()
			Expect(om.IsZero()).To(BeTrue())
		})

		It("should return true for a zero-value receiver", func() {
			var om mapx.OrderedMap[string, int]
			Expect(om.IsZero()).To(BeTrue())
		})

		It("should return false after adding an entry", func() {
			om := mapx.New[string, int]()
			om.Set("a", 1)
			Expect(om.IsZero()).To(BeFalse())
		})

		It("should return true after removing all entries", func() {
			om := mapx.New[string, int]()
			om.Set("a", 1)
			om.Delete("a")
			Expect(om.IsZero()).To(BeTrue())
		})

		It("should return true after Clear", func() {
			om := mapx.New[string, int]()
			om.Set("a", 1)
			om.Set("b", 2)
			om.Clear()
			Expect(om.IsZero()).To(BeTrue())
		})
	})

	Describe("omitempty", func() {
		type wrapper struct {
			Name  string                        `json:"name"`
			Items *mapx.OrderedMap[string, int] `json:"items,omitempty"`
		}

		It("should omit a nil OrderedMap field", func() {
			w := wrapper{Name: "test"}
			b, err := json.Marshal(w)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(b)).To(Equal(`{"name":"test"}`))
		})

		It("should include a non-empty OrderedMap field", func() {
			items := mapx.New[string, int]()
			items.Set("x", 42)
			w := wrapper{Name: "test", Items: items}
			b, err := json.Marshal(w)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(b)).To(Equal(`{"name":"test","items":{"x":42}}`))
		})
	})

	Describe("omitzero", func() {
		type wrapper struct {
			Name  string                        `json:"name"`
			Items *mapx.OrderedMap[string, int] `json:"items,omitzero"`
		}

		It("should omit a nil OrderedMap field", func() {
			w := wrapper{Name: "test"}
			b, err := json.Marshal(w)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(b)).To(Equal(`{"name":"test"}`))
		})

		It("should omit an empty non-nil OrderedMap field", func() {
			w := wrapper{Name: "test", Items: mapx.New[string, int]()}
			b, err := json.Marshal(w)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(b)).To(Equal(`{"name":"test"}`))
		})

		It("should include a non-empty OrderedMap field", func() {
			items := mapx.New[string, int]()
			items.Set("x", 42)
			w := wrapper{Name: "test", Items: items}
			b, err := json.Marshal(w)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(b)).To(Equal(`{"name":"test","items":{"x":42}}`))
		})
	})
})
