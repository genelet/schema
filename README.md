# Package schema

[![GoDoc](https://godoc.org/github.com/genelet/schema?status.svg)](https://godoc.org/github.com/genelet/schema)

Package `schema` provides unified protobuf-based data structures for describing hierarchical configuration schemas. It consolidates types from `horizon/utils`, `determined/det`, and `grand/spec` into a single shared definition.

## Overview

The schema package defines types used across multiple packages for:

- **Dynamic HCL/JSON unmarshaling**: Using `ClassName` for type lookup when unmarshaling into interface fields
- **Service orchestration**: Using `ClassName` and `ServiceName` for delegating read/write operations to microservices

## Installation

```bash
go get github.com/genelet/schema
```

## Dependencies

- `google.golang.org/protobuf` - Protocol Buffers runtime

---

## Protobuf Types

All types are defined in `proto/schema.proto` and generated into Go code.

### Struct

```protobuf
message Struct {
  string ClassName = 1;
  string ServiceName = 2;
  map<string, Value> fields = 3;
}
```

Represents a type specification for dynamic unmarshaling and service orchestration.

| Field | Type | Description |
|-------|------|-------------|
| `ClassName` | `string` | Go struct type name / object identifier |
| `ServiceName` | `string` | Service name for delegation (read/write operations) |
| `Fields` | `map[string]*Value` | Nested field specifications |

**Generated Methods:**

| Method | Description |
|--------|-------------|
| `GetClassName() string` | Returns the ClassName field |
| `GetServiceName() string` | Returns the ServiceName field |
| `GetFields() map[string]*Value` | Returns the Fields map |
| `GetObjectName() string` | Alias for GetClassName (backwards compatibility) |
| `Reset()` | Resets the struct to zero value |
| `String() string` | Returns string representation |
| `ProtoMessage()` | Marker method for protobuf |
| `ProtoReflect() protoreflect.Message` | Returns protobuf reflection interface |

---

### Value

```protobuf
message Value {
  oneof kind {
    Struct single_struct = 1;
    ListStruct list_struct = 2;
    MapStruct map_struct = 3;
    Map2Struct map2_struct = 4;
  }
}
```

Represents a typed field specification. It can be one of four kinds.

**Generated Methods:**

| Method | Description |
|--------|-------------|
| `GetKind() isValue_Kind` | Returns the oneof kind interface |
| `GetSingleStruct() *Struct` | Returns SingleStruct if set, nil otherwise |
| `GetListStruct() *ListStruct` | Returns ListStruct if set, nil otherwise |
| `GetMapStruct() *MapStruct` | Returns MapStruct if set, nil otherwise |
| `GetMap2Struct() *Map2Struct` | Returns Map2Struct if set, nil otherwise |

**Oneof Wrapper Types:**

| Type | Description |
|------|-------------|
| `Value_SingleStruct` | Wraps a single `*Struct` |
| `Value_ListStruct` | Wraps a `*ListStruct` |
| `Value_MapStruct` | Wraps a `*MapStruct` |
| `Value_Map2Struct` | Wraps a `*Map2Struct` |

---

### ListStruct

```protobuf
message ListStruct {
  repeated Struct list_fields = 1;
}
```

Represents a list/slice of Struct specifications.

| Field | Type | Description |
|-------|------|-------------|
| `ListFields` | `[]*Struct` | Slice of Struct specifications |

**Generated Methods:**

| Method | Description |
|--------|-------------|
| `GetListFields() []*Struct` | Returns the ListFields slice |

---

### MapStruct

```protobuf
message MapStruct {
  map<string, Struct> map_fields = 1;
}
```

Represents a map with string keys to Struct specifications.

| Field | Type | Description |
|-------|------|-------------|
| `MapFields` | `map[string]*Struct` | Map of key to Struct |

**Generated Methods:**

| Method | Description |
|--------|-------------|
| `GetMapFields() map[string]*Struct` | Returns the MapFields map |

---

### Map2Struct

```protobuf
message Map2Struct {
  map<string, MapStruct> map2_fields = 1;
}
```

Represents a two-level nested map structure for `map[[2]string]T` types. The outer map uses the first key, the inner MapStruct uses the second key.

| Field | Type | Description |
|-------|------|-------------|
| `Map2Fields` | `map[string]*MapStruct` | Two-level nested map |

**Generated Methods:**

| Method | Description |
|--------|-------------|
| `GetMap2Fields() map[string]*MapStruct` | Returns the Map2Fields map |

---

## Exported Functions

### NewValue

```go
func NewValue(v any) (*Value, error)
```

Constructs a Value from a generic Go interface. Used for dynamic HCL/JSON unmarshaling.

**Conversion Rules:**

| Go Type | Conversion |
|---------|------------|
| `string` | SingleStruct (class name only) |
| `[]string` | ListStruct |
| `map[string]string` | MapStruct |
| `map[[2]string]string` | Map2Struct |
| `[2]any` | SingleStruct with fields |
| `[][2]any` | ListStruct with fields |
| `map[string][2]any` | MapStruct with fields |
| `map[[2]string][2]any` | Map2Struct with fields |
| `*Struct` | SingleStruct |
| `[]*Struct` | ListStruct |
| `map[string]*Struct` | MapStruct |
| `map[string]*MapStruct` | Map2Struct |

**Returns:** `*Value` and error if type is unsupported.

---

### NewStruct

```go
func NewStruct(className string, fieldSpecs ...map[string]any) (*Struct, error)
```

Constructs a Struct specification for dynamic type unmarshaling.

**Parameters:**
- `className` - The Go struct type name (must be non-empty)
- `fieldSpecs` - Optional map specifying field types (field name → type specification)

**Example:**
```go
spec, err := NewStruct("Geo", map[string]any{
    "Shape": "Circle",  // Shape field should be a Circle
})
```

---

### NewServiceValue

```go
func NewServiceValue(v any) (*Value, error)
```

Constructs a Value for service orchestration. Similar to NewValue but supports `[]string` format for end-node structs with service names.

**Conversion Rules:**

| Go Type | Conversion | Element 0 | Element 1 |
|---------|------------|-----------|-----------|
| `[]string` | end SingleStruct | class name | service name |
| `[][]string` | end ListStruct | class name | service name |
| `map[string][]string` | end MapStruct | class name | service name |
| `map[[2]string][]string` | end Map2Struct | class name | service name |
| `[2]any` | SingleStruct | class name | field specs or service |
| `[][2]any` | ListStruct | class name | field specs or service |
| `map[string][2]any` | MapStruct | class name | field specs or service |
| `map[[2]string][2]any` | Map2Struct | class name | field specs or service |
| `*Struct` | SingleStruct | - | - |
| `[]*Struct` | ListStruct | - | - |
| `map[string]*Struct` | MapStruct | - | - |
| `map[string]*MapStruct` | Map2Struct | - | - |

---

### NewServiceStruct

```go
func NewServiceStruct(className string, v any) (*Struct, error)
```

Constructs a Struct for service orchestration with ClassName and optional ServiceName.

**Parameters:**
- `className` - The class/object type identifier
- `v` - Either a service name (`string`) or field specifications (`map[string]any`)

**Examples:**
```go
// With service delegation
spec, err := NewServiceStruct("Provider", "providerService")

// With nested field specs
spec, err := NewServiceStruct("Config", map[string]any{
    "Database": []string{"PostgresDB", "dbService"},
})
```

---

## Unexported Functions

### Type Specification Helpers

| Function | Description |
|----------|-------------|
| `createTypeSpec(className string, fields ...map[string]any) typeSpec` | Creates a typeSpec from class name and optional fields |
| `newValueFromString(className string) (*Value, error)` | Creates Value from simple class name |
| `newValueFromTypeSpec(spec typeSpec) (*Value, error)` | Creates Value from type specification |
| `newValueFromStringSlice(classNames []string) (*Value, error)` | Creates Value from slice of class names |
| `newValueFromTypeSpecSlice(specs [][2]any) (*Value, error)` | Creates Value from slice of type specs |
| `newValueFromStringMap(classNames map[string]string) (*Value, error)` | Creates Value from map of class names |
| `newValueFromTypeSpecMap(specs map[string][2]any) (*Value, error)` | Creates Value from map of type specs |
| `newValueFromString2DMap(classNames map[[2]string]string) (*Value, error)` | Creates Value from 2D map of class names |
| `newValueFromTypeSpec2DMap(specs map[[2]string][2]any) (*Value, error)` | Creates Value from 2D map of type specs |

### Struct Builders

| Function | Description |
|----------|-------------|
| `newSingleStruct(spec typeSpec) (*Struct, error)` | Creates Struct from type specification |
| `newListStruct(specs [][2]any) (*ListStruct, error)` | Creates ListStruct from slice of type specs |
| `newMapStruct(specs map[string][2]any) (*MapStruct, error)` | Creates MapStruct from map of type specs |
| `newMap2Struct(specs map[[2]string][2]any) (*Map2Struct, error)` | Creates Map2Struct from 2D map of type specs |

### Service Orchestration Helpers

| Function | Description |
|----------|-------------|
| `newServiceSingleStruct(spec typeSpec) (*Struct, error)` | Creates Struct for service orchestration |
| `newServiceListStruct(specs [][2]any) (*ListStruct, error)` | Creates ListStruct for service orchestration |
| `newServiceMapStruct(specs map[string][2]any) (*MapStruct, error)` | Creates MapStruct for service orchestration |
| `newServiceMap2Struct(specs map[[2]string][2]any) (*Map2Struct, error)` | Creates Map2Struct for service orchestration |

### End-Node Struct Helpers

| Function | Description |
|----------|-------------|
| `newEndStruct(v []string) (*Struct, error)` | Creates end-node Struct from `[className, serviceName]` |
| `newEndListStruct(v [][]string) (*ListStruct, error)` | Creates ListStruct of end-node Structs |
| `newEndMapStruct(v map[string][]string) (*MapStruct, error)` | Creates MapStruct of end-node Structs |
| `newEndMap2Struct(v map[[2]string][]string) (*Map2Struct, error)` | Creates Map2Struct of end-node Structs |

---

## Unexported Types

### typeSpec

```go
type typeSpec [2]any
```

Represents a type specification for a single struct field. A 2-element array where:
- `[0]`: class name (string)
- `[1]`: optional field specifications (`map[string]any`) or service name (`string`)

---

## Usage Examples

### Dynamic Unmarshaling Specification

```go
// Simple: just specify the type name
spec, _ := NewStruct("Config", map[string]any{
    "Database": "PostgresDB",
})

// Nested: specify type with its own field types
spec, _ := NewStruct("Config", map[string]any{
    "Database": [2]any{"PostgresDB", map[string]any{
        "Connection": "TCPConnection",
    }},
})

// List of types
spec, _ := NewStruct("Config", map[string]any{
    "Servers": []string{"HTTPServer", "GRPCServer"},
})

// Map of types
spec, _ := NewStruct("Config", map[string]any{
    "Handlers": map[string]string{
        "api": "APIHandler",
        "web": "WebHandler",
    },
})
```

### Service Orchestration Specification

```go
// End-node with service delegation
spec, _ := NewServiceStruct("Config", map[string]any{
    "Provider": []string{"CloudProvider", "providerService"},
})

// Nested structure with service delegation
spec, _ := NewServiceStruct("Root", map[string]any{
    "Child": [2]any{"ChildType", map[string]any{
        "Grandchild": []string{"GrandchildType", "grandchildService"},
    }},
})
```

### Using Pre-built Structs

```go
// Create structs directly
childSpec := &Struct{ClassName: "Child", ServiceName: "childService"}
parentSpec, _ := NewStruct("Parent", map[string]any{
    "Child": childSpec,
})

// List of pre-built structs
specs := []*Struct{
    {ClassName: "TypeA"},
    {ClassName: "TypeB"},
}
value, _ := NewValue(specs)
```

---

## Package Aliases

This package is used as the canonical source by:

- `github.com/genelet/horizon/utils` - Type aliases for HCL unmarshaling
- `github.com/genelet/determined/det` - Type aliases for JSON unmarshaling
- `github.com/genelet/grand/spec` - Type aliases for service orchestration

Example alias in dependent packages:
```go
import "github.com/genelet/schema"

type Struct = schema.Struct
type Value = schema.Value
var NewStruct = schema.NewStruct
```

---

## Regenerating Protobuf Code

```bash
cd ~/schema
protoc --go_out=. --go_opt=paths=source_relative proto/schema.proto
mv proto/schema.pb.go schema.pb.go
```
