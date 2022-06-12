# naml

This is a Go implementation of the [naml][spec] format.

## Usage

This implementation is used simiarily to the `encoding/json` package:

```go
package main

import (
  "fmt"
  "strings"

  "gihtub.com/qeaml/naml"
)

type Person struct {
  Name string
  Age int
}

func main() {
  r := strings.NewReader(`name = "John"; age = 24`)
  p := Person{}
  d := naml.NewDecoder(r)
  if err := d.Decode(&p); err != nil {
    fmt.Println(err)
    os.Exit(1)
  }
  fmt.Println(p.Name)
  fmt.Println(p.Age)
}
```

## Implementation-specific notes

This implementation is whitespace-agnostic. That means that no actual spaces are
required for the source to be parsed correctly. Another note is that a semicolon
(`;`) is also considered whitespace.

[spec]: https://github.com/naml-conf/naml
