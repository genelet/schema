# Introducing `genelet/schema`: Dynamic, Schema-Driven Type Orchestration for Go

If you search for "JSON Schema" in the Go ecosystem, you'll likely land on libraries like `invopop/jsonschema`. These are excellent tools that do one thing very well: they use **Reflection** to generate a JSON Schema from your existing Go structs.

**But what if you need to do the exact reverse?**

What if you have a JSON Schema (or a similar description) and you need to construct a Go object that represents that structure *dynamically* at runtime, without generating code? What if that structure needs to carry not just data, but metadata about *where* that data lives or *who* should handle it?

Enter [**`genelet/schema`**](https://github.com/genelet/schema).

## Chapter 1. Overview

The schema package defines types used across multiple packages for:

- **Dynamic HCL/JSON unmarshaling**: Using `ClassName` for type lookup when unmarshaling into interface fields
- **Service orchestration**: Using `ClassName` and `ServiceName` for delegating read/write operations to microservices

### Installation

```bash
go get github.com/genelet/schema
```

## Chapter 2. Philosophy: Minimal Schema

**You do not need to provide a full JSON Schema.**

The purpose of this package is to assist in two specific scenarios during unmarshalling/marshaling:
1.  **Interface Types**: Go requires knowing the concrete implementation to unmarshal into an Interface field.
2.  **Service Injection**: Specifying which microservice allows access to a specific field.

**Standard primitive fields** (e.g., `string`, `int`, `bool`, `float`), `structs`, `slices` and `maps` **do not need to be included** in the schema. You should omit them entirely.

**Rule of Thumb**: Only include a field in the JSON Schema if:
*   It is an **Interface** type (requires a concrete `className` definition).
*   It needs a `serviceName` annotation.
*   It is a nested structure (Array/Map/Object) that *contains* one of the above.

---

## Chapter 3. Schema Types

All types are defined in Go code.

### Struct

```go
type Struct struct {
  ClassName   string            // Go struct type name / object identifier
  ServiceName string            // Service name for delegation
  Fields      map[string]*Value // Nested field specifications
}
```

Represents a type specification for dynamic unmarshaling and service orchestration.

| Field | Type | Description |
|-------|------|-------------|
| `ClassName` | `string` | Go struct type name / object identifier |
| `ServiceName` | `string` | Service name for delegation (read/write operations) |
| `Fields` | `map[string]*Value` | Nested field specifications |


---

### Value

```go
type Value struct {
  Kind isValue_Kind // One of SingleStruct, ListStruct, MapStruct, Map2Struct
}
```

Represents a typed field specification. It can be one of four kinds.

**Oneof Wrapper Types:**

| Type | Description |
|------|-------------|
| `Value_SingleStruct` | Wraps a single `*Struct` |
| `Value_ListStruct` | Wraps a `*ListStruct` |
| `Value_MapStruct` | Wraps a `*MapStruct` |
| `Value_Map2Struct` | Wraps a `*Map2Struct` |

---

### ListStruct

```go
type ListStruct struct {
  ListFields []*Struct
}
```

Represents a list/slice of Struct specifications. Corresponds to `[]T` in Go.

---

### MapStruct

```go
type MapStruct struct {
  MapFields map[string]*Struct
}
```

Represents a map with string keys to Struct specifications. Corresponds to `map[string]T` in Go.

---

### Map2Struct

```go
type Map2Struct struct {
  Map2Fields map[string]*MapStruct
}
```

Represents a two-level nested map structure for `map[[2]string]T` types. The outer map uses the first key, the inner MapStruct uses the second key.

---

## Chapter 4. JSON Schema Representation

The package uses JSON Schema (Draft-7) with minimal extensions for defining types.

### Overview Table

| JSON Schema Keyword | Genelet Schema Type | Description |
|---------------------|---------------------|-------------|
| `properties` | **SingleStruct** | A single object with known fields. |
| `items` | **ListStruct** | A list of objects of the same schema. |
| `additionalProperties` | **MapStruct** | A map with string keys and values of the same schema. |
| `x-map2` (extension) | **Map2Struct** | A map of maps (two-layer keys). |
| `className` (custom) | **SingleStruct** | A custom class used to implement an Interface. |

### SingleStruct Examples

**Explicit Object with Properties:**
```json
{
  "className": "Person",
  "properties": {
    "Avatar": { "className": "Image", "serviceName": "s1" }
  }
}
```

**Leaf Node (Interface Implementation):**
```json
{
  "className": "MyConcreteType"
}
```

### ListStruct Example

```json
{
  "items": {
    "className": "Person"
  }
}
```

### MapStruct Example

```json
{
  "additionalProperties": {
    "className": "Person"
  }
}
```

### Map2Struct Example

```json
{
  "x-map2": true,
  "properties": {
    "region1": {
      "properties": {
        "key1": { "className": "ServiceA" },
        "key2": { "className": "ServiceB" }
      }
    }
  }
}
``` 

### Example: Mixed Fields with Services
```go
jsonStr := `{
    "properties": {
        "SimpleField": { "className": "Metric", "serviceName": "metric_service" },
        "ListField":   { "items": { "className": "Log", "serviceName": "log_service" } },
        "MapField":    { "additionalProperties": { "className": "Config", "serviceName": "config_service" } }
    }
}`
```

### Example: Nested Structs with Services
```go
jsonStr := `{
    "properties": {
        "Group1": {
            "className": "SubClass1",
            "properties": {
                "Item": { "className": "Detail", "serviceName": "s1" }
            }
        }
    }
}`
```

### Example: List of Structs with Services
```go
jsonStr := `{
    "properties": {
        "Items": {
            "items": {
                "className": "ItemClass",
                "properties": { "Info": { "className": "Data", "serviceName": "data_service" } }
            }
        }
    }
}`
```

### Example: Map of Structs with Services
```go
jsonStr := `{
    "properties": {
        "Registry": {
            "additionalProperties": {
                "className": "EntryClass",
                "properties": { "Record": { "className": "Row", "serviceName": "db_service" } }
            }
        }
    }
}`
```

---

## Chapter 5. Exported Functions

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

**Use Cases:**

1. **Dynamic JSON Interface Unmarshaling** - Used in [github.com/genelet/determined](https://github.com/genelet/determined) for JSON unmarshaling. See the [Medium article](https://github.com/genelet/determined) for details.

2. **Dynamic HCL Interface Unmarshaling** - Used in [github.com/genelet/horizon](https://github.com/genelet/horizon) for HCL unmarshaling. See the [Medium article: Marshal and Unmarshal HCL Files](https://medium.com/@peterbi_91340/marshal-and-unmarshal-hcl-files-1-3-d7591259a8d6) for details.

---

### NewValue

```go
func NewValue(v any) (*Value, error)
```

Constructs a Value from a generic Go interface. Used for dynamic HCL/JSON unmarshaling.

> **Note:** Don't use this function directly. It is exported for readers to understand how the `v any` parameter is passed to `NewStruct`.

**Conversion Rules:**

| Go Type | Conversion |
|---------|------------|
| `string` | SingleStruct (class name only) |
| `[]string` | ListStruct |
| `map[string]string` | MapStruct |
| `map[[2]string]string` | Map2Struct |
| `*Struct` | SingleStruct |
| `[]*Struct` | ListStruct |
| `map[string]*Struct` | MapStruct |

---

### NewServiceStruct

```go
func NewServiceStruct(className string, v any) (*Struct, error)
```

Constructs a Struct for service orchestration with ClassName and optional ServiceName.

**Parameters:**
- `className` - The class/object type identifier
- `v` - Either:
  - `string` - service name for delegation
  - `map[string]any` - field specifications

**Examples:**
```go
// With service delegation
spec, err := NewServiceStruct("Provider", "providerService")

// With nested field specs
spec, err := NewServiceStruct("Config", map[string]any{
    "Database": []string{"PostgresDB", "dbService"},
})
```

**Use Case:**

This function is used to build the [Grand Unmarshaler](https://medium.com/@peterbi_91340/the-grand-unmarshaller-project-dc8aeda76f41) project for service orchestration and delegation.

---

### NewServiceValue

```go
func NewServiceValue(v any) (*Value, error)
```

Constructs a Value for service orchestration. Similar to NewValue but supports `[]string` format for end-node structs with service names.

> **Note:** Don't use this function directly. It is exported for readers to understand how the `v any` parameter is passed to `NewServiceStruct`.

**Conversion Rules:**

| Go Type | Conversion | Element 0 | Element 1 |
|---------|------------|-----------|-----------|
| `[]string` | end SingleStruct | class name | service name |
| `[][]string` | end ListStruct | class name | service name |
| `map[string][]string` | end MapStruct | class name | service name |
| `map[[2]string][]string` | end Map2Struct | class name | service name |
| `[2]any` | SingleStruct | class name | service/fields/*Struct |
| `[][2]any` | ListStruct | class name | service/fields/*Struct |
| `map[string][2]any` | MapStruct | class name | service/fields/*Struct |
| `map[[2]string][2]any` | Map2Struct | class name | service/fields/*Struct |
| `*Struct` | SingleStruct | - | - |
| `[]*Struct` | ListStruct | - | - |
| `map[string]*Struct` | MapStruct | - | - |
| `map[string]*MapStruct` | Map2Struct | - | - |

---

### JSMStruct

```go
func JSMStruct(className, jsonSchemaStr string) (*Struct, error)
```

Parses a JSON Schema string and converts it into a Genelet `Struct`. **Removes all `serviceName` fields** from the resulting tree. Useful when you need the data structure definition but want to decouple it from specific backend services.

---

### JSMServiceStruct

```go
func JSMServiceStruct(className, jsonSchemaStr string) (*Struct, error)
```

Parses a JSON Schema string and converts it into a Genelet `Struct`. **Retains all `serviceName` annotations** found in the JSON Schema.

**Arguments:**
*   `className` (string): The name assigned to the root object. Takes precedence over any `className` in the JSON.
*   `jsonSchemaStr` (string): The standard JSON Schema (Draft-7) string to parse.

---

## Chapter 6. The Missing Link

To understand where `genelet/schema` fits, let's look at the existing landscape:

1.  **`mitchellh/mapstructure`**: This is the gold standard for decoding generic `map[string]interface{}` data into Go structs. However, it relies on static Go types. You must define your `struct` beforehand.
2.  **`gburgyan/go-poly` / `dhoelle/oneof`**: These tackle polymorphic JSON (e.g., a list containing both `Circle` and `Square` objects). They are **data-driven**—they peek at specific fields (like `type`) in the incoming JSON to decide what to do.
3.  **`invopop/jsonschema`**: As mentioned, this goes from **Go Struct → JSON Schema**.
4.  **`santhosh-tekuri/jsonschema`**: A powerful **validator**. It tells you if data matches a schema, but it doesn't give you a manipulatable type definition of that data.

**`genelet/schema`** is unique because it is **Schema-Driven**. It takes a description (a Schema) and instantiates a robust `Struct` object that acts as a dynamic type system. It doesn't just validate; it *defines* and *orchestrates*.

---

## Chapter 7. Conclusion

`genelet/schema` provides the missing reverse-gear in the Go JSON ecosystem. By treating Schemas as first-class citizens that can be instantiated into rich, traversable Go objects, it opens the door for:

*   **Dynamic Configurations**: Define types in JSON, use them in Go.
*   **Service Mesh & Orchestration**: Route data based on schema definitions.
*   **Legacy Systems Integration**: Wrap old endpoints with modern schema definitions.
*   **Low-Code/No-Code Backends**: Handle arbitrary data structures safely.

It's time to stop fighting your JSON and start orchestrating it.

Check it out at [github.com/genelet/schema](https://github.com/genelet/schema).

