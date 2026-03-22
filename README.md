# mapx

> A generic, insertion-ordered map for Go with first-class JSON support.

`mapx` provides `OrderedMap[K, V]` — a hash map backed by a doubly linked list that preserves the order in which keys were inserted. It marshals to and from JSON with key order intact, making it ideal for APIs, config files, and anywhere deterministic output matters.

## Features

- **Generics** — works with any `comparable` key and any value type.
- **Insertion-order iteration** — `Keys()`, `Values()`, `Range()` all follow insertion order.
- **O(1) operations** — `Get`, `Set`, `Has`, and `Delete` are all constant-time.
- **JSON round-trip** — implements `json.Marshaler` and `json.Unmarshaler`; key order is preserved in both directions.
- **Custom key types** — supports `encoding.TextMarshaler` / `TextUnmarshaler` for JSON map keys, plus all integer and string kinds.
- **Zero dependencies** — only the Go standard library.

## Installation

```sh
go get github.com/arisu-archive/mapx
```

Requires **Go 1.25+**.

## Usage

```go
package main

import (
	"encoding/json"
	"fmt"

	"github.com/arisu-archive/mapx"
)

func main() {
	m := mapx.New[string, int]()
	m.Set("cherry", 3)
	m.Set("apple", 1)
	m.Set("banana", 2)

	// Iteration follows insertion order.
	m.Range(func(k string, v int) bool {
		fmt.Printf("%s = %d\n", k, v)
		return true
	})
	// cherry = 3
	// apple = 1
	// banana = 2

	// JSON output preserves order.
	b, _ := json.Marshal(m)
	fmt.Println(string(b))
	// {"cherry":3,"apple":1,"banana":2}

	// Round-trip back.
	m2 := mapx.New[string, int]()
	_ = json.Unmarshal(b, m2)
	fmt.Println(m2.Keys()) // [cherry apple banana]
}
```

## License

See [LICENSE](LICENSE) for details.
