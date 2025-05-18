# 2025-02-21 Define config struct

How to define the config struct

## Background

- [ ] There are two parts, decoding and merging, some library support merging from multiple sources.
  e.g. config file, environment variable, command line argument, etc.

Most encoding library supports unmarshal into struct (or `map[string]any`) and customzie the key name using struct tags.
For example json:

```go
type Config struct {
    FooBar string `json:"foo_bar"`
}
```

This allow typesafe access to the config value but lost context such as line number.
When config has validation error, the application cannot print the exact error location.

To keep context information, we can create `ConfigNode` interface to keep line number etc.

```go
type ConfigNode interface {
    Line() int
    Column() int
    Type() ConfigValueType
}

type ConfigValueString struct {
   line int
   column int
   Value string
}

func (c ConfigValueString) Line() int {
    return c.line
}

func (c ConfigValueString) Column() int {
    return c.column
}

func (c ConfigValueString) Type() ConfigValueType {
    return ConfigValueTypeString
}


```

Returning `map[string]ConfigNode` would allow keeping context information in `ConfigNode`
but it is not type safe as user need to access/traverse config by string key and relies on
runtime behavior instead of compile time check.

NOTE: Java jackson `JsonNode` is similar to `map[string]any` and does not provide position information out of the box.

Another way is to create struct using wrapped value, which can save context such as line number.
However, when a field is also struct, it becames harder to composite.

```go
type Config struct {
  FooBar MyString 
}
```

hmm ... seems I just used this repo for testing vscode extension development.