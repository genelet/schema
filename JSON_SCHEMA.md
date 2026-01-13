# JSON Schema Representation of Genelet Schema Types

The `schema` package allows you to define Genelet Schema types (`Struct`, `ListStruct`, `MapStruct`, `Map2Struct`) using standard JSON Schema (Draft-7) with minimal extensions. This document details how different JSON Schema structures map to their corresponding Genelet Schema types.

## Philosophy: Minimal Schema

**You do not need to provide a full JSON Schema.**

The purpose of this package is to assist in two specific scenarios during unmarshalling/marshaling:
1.  **Interface Types**: Go requires knowing the concrete implementation to unmarshal into an Interface field.
2.  **Service Injection**: Specifying which microservice allows access to a specific field.

**Standard primitive fields** (e.g., `string`, `int`, `bool`, `float`) and standard structs **do not need to be included** in the schema. You should omit them entirely.

**Rule of Thumb**: Only include a field in the JSON Schema if:
*   It is an **Interface** type (requires a concrete `className` definition).
*   It needs a `serviceName` annotation.
*   It is a nested structure (Array/Map/Object) that *contains* one of the above.

## Overview

The mapping logic keywords:

| JSON Schema Keyword | Genelet Schema Type | Description |
|---------------------|---------------------|-------------|
| `properties` | **SingleStruct** | A single object with known fields. |
| `items` | **ListStruct** | A list of objects of the same schema. |
| `additionalProperties` | **MapStruct** | A map with string keys and values of the same schema. |
| `x-map2` (extension) | **Map2Struct** | A map of maps (two-layer keys). |
| `className` (custom) | **SingleStruct** | A custom class (e.g., `MyClass`) used to implement an Interface. |

## 1. SingleStruct (Struct)

A `SingleStruct` represents an object with a class name, potentially with nested fields.

### Explicit Object with Properties
Use `properties` to define a struct that contains Interface fields.

**JSON Schema:**
```json
{
  "className": "Person",
  "properties": {
    "Avatar": { "className": "Image", "serviceName": "s1" }
  }
}
```
*Note: Fields like `Name` (string) or `Age` (int) are omitted.*

**Resulting Schema Type:** `SingleStruct`
- ClassName: "Person" (Implicitly an object because of `properties`)
- Fields: "Avatar" (SingleStruct "Image" with Service "s1")

### Leaf Node (Interface Implementation)
A schema with a **custom className** string is treated as a `SingleStruct` leaf node. This tells the parser what concrete class to use for an Interface field.

**JSON Schema:**
```json
{
  "className": "MyConcreteType"
}
```
**Resulting Schema Type:** `SingleStruct`
- ClassName: "MyConcreteType"

## 2. ListStruct

A `ListStruct` represents an array or slice of items. It corresponds to `[]T` in Go.

Use `items` to define the element schema. `items` implies it is an array.

**JSON Schema:**
```json
{
  "items": {
    "className": "Person"
  }
}
```

**Resulting Schema Type:** `ListStruct`
- ListFields: `[SingleStruct("Person")]` (describes the element schema)

## 3. MapStruct

A `MapStruct` represents a map with string keys. It corresponds to `map[string]T` in Go.

Use `additionalProperties` to define the value schema. `additionalProperties` implies it is a map.

**JSON Schema:**
```json
{
  "additionalProperties": {
    "className": "Person"
  }
}
```

**Resulting Schema Type:** `MapStruct`
- MapFields: `{"*": SingleStruct("Person")}` (wildcard key describing value schema)

## 4. Map2Struct

A `Map2Struct` represents a two-level nested map. It corresponds to `map[[2]string]T` in Go, where the key is effectively a composite `(key1, key2)`.

To represent this structure, use the custom extension `x-map2: true`. The structure requires two layers of `properties`.

**JSON Schema:**
```json
{
  "x-map2": true,
  "properties": {
    "region1": {
      "properties": {
        "key1": { "className": "ServiceA" },
        "key2": { "className": "ServiceB" }
      }
    },
    "region2": {
      "properties": {
        "key3": { "className": "ServiceC" }
      }
    }
  }
}
```

**Resulting Schema Type:** `Map2Struct`
- Map2Fields:
  - "region1" -> MapStruct
      - "key1" -> SingleStruct("ServiceA")
      - "key2" -> SingleStruct("ServiceB")
  - "region2" -> MapStruct
  - "key3" -> SingleStruct("ServiceC")

## Nested Collections

`items` and `additionalProperties` can be nested to represent multi-level collection types such as `[][]T`,
`[]map[string]T`, `map[string][]T`, and `map[string]map[string]T`.

**Examples:**

List of list (`[][]T`):
```json
{
  "items": {
    "items": { "className": "Person" }
  }
}
```

List of map (`[]map[string]T`):
```json
{
  "items": {
    "additionalProperties": { "className": "Person" }
  }
}
```

Map of list (`map[string][]T`):
```json
{
  "additionalProperties": {
    "items": { "className": "Person" }
  }
}
```

Map of map (`map[string]map[string]T`):
```json
{
  "additionalProperties": {
    "additionalProperties": { "className": "Person" }
  }
}
```

## Service Decoration

The `serviceName` keyword can be added to any schema node to specify the service responsible for that data.

**JSON Schema:**
```json
{
  "className": "UserProfile",
  "serviceName": "userService",
  "properties": {
    "Avatar": {
      "className": "Image",
      "serviceName": "mediaService"
    }
  }
}
```

**Resulting Schema Type:**
- Root `SingleStruct` ("UserProfile") has `ServiceName` = "userService".
- Field "Avatar" (`SingleStruct` "Image") has `ServiceName` = "mediaService".

## Functions

### `JSMServiceStruct`

`JSMServiceStruct` parses a JSON Schema string and converts it into a Genelet `Struct`. It is the core function for schema conversion and **retains all `serviceName` annotations** found in the JSON Schema.

```go
func JSMServiceStruct(className, jsonSchemaStr string) (*Struct, error)
```

**Arguments:**

*   `className` (string): The name assigned to the root object (SingleStruct). Even if the JSON Schema defines a `className` field, this argument takes precedence for the root.
*   `jsonSchemaStr` (string): The standard JSON Schema (Draft-7) string to parse.

**Conversion Rules:**

The following table summarizes how different JSON Schema structures are converted into Genelet Schema types, and how `ClassName` and `ServiceName` are determined.

| JSON Schema Pattern | Genelet Type | ClassName | ServiceName |
| :--- | :--- | :--- | :--- |
| `{"className": "Circle", "serviceName": "s1"}` | **SingleStruct** | "Circle" | "s1" |
| `{"items": {"className": "Circle", "serviceName": "s2"}}` | **ListStruct** | n/a | n/a (field level) |
| `{"additionalProperties": {"className": "Circle", "serviceName": "s3"}}` | **MapStruct** | n/a | n/a (field level) |
| `{"className": "Class1", "properties": {"Field1": {"className": "Circle"}}}` | **SingleStruct** | "Class1" | "" |
| `{"x-map2": true, "properties": ...}` | **Map2Struct** | n/a | n/a |

**Examples:**

#### Case 1: Simple Object (SingleStruct)

Defining a `Geo` object with a string field and a primitive type field.

```go
jsonStr := `{
    "className": "Geo",
    "properties": {
        "Shape": { "className": "Circle", "serviceName": "s_shape" }
    }
}`
s, err := JSMServiceStruct("Geo", jsonStr)
// Result: SingleStruct "Geo". Field "Shape" is SingleStruct("Circle") served by "s_shape".
// Note: "Name" field (string) is omitted from this schema.
```

#### Case 2: List (ListStruct)

Defining a list of `Circle` objects.

```go
jsonStr := `{
    "items": {
        "className": "Circle",
        "serviceName": "s_list_item"
    }
}`
s, err := JSMServiceStruct("CircleList", jsonStr)
// Result: ListStruct where items are SingleStruct("Circle") served by "s_list_item".
```

#### Case 3: Map (MapStruct)

Defining a map where keys are strings and values are `Circle` objects.

```go
jsonStr := `{
    "additionalProperties": {
        "className": "Circle",
        "serviceName": "s_map_val"
    }
}`
s, err := JSMServiceStruct("CircleMap", jsonStr)
// Result: MapStruct where values are SingleStruct("Circle") served by "s_map_val".
```

#### Case 4: Map2 (Map2Struct)

Defining a two-level map using the `x-map2` extension.

```go
jsonStr := `{
    "x-map2": true,
    "properties": {
        "region1": {
            "properties": {
                "key1": { "className": "Circle", "serviceName": "s_map2_val" }
            }
        }
    }
}`
s, err := JSMServiceStruct("GeoMap", jsonStr)
// Result: Map2Struct with region "region1" -> key "key1" -> SingleStruct("Circle") served by "s_map2_val".
```

#### Case 5: Mixed Fields with Services

A struct containing various field types (simple, list, map), all backed by different services.

```go
jsonStr := `{
    "properties": {
        "SimpleField": { "className": "Metric", "serviceName": "metric_service" },
        "ListField":   { "items": { "className": "Log", "serviceName": "log_service" } },
        "MapField":    { "additionalProperties": { "className": "Config", "serviceName": "config_service" } }
    }
}`
```

#### Case 6: Nested Structs with Services

A struct where fields are themselves defined as structs (inline types), eventually leading to service-backed fields.

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

#### Case 7: List of Structs with Services

A list where the item type is a complex object (defined by properties), which in turn contains service-backed fields.

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

#### Case 8: Map of Structs with Services

A map where the value type is a complex object, which contains service-backed fields.

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

### `JSMStruct`

`JSMStruct` is a wrapper around `JSMServiceStruct`. It parses the JSON Schema but **removes all `serviceName` fields** from the resulting `Struct` tree. This is useful when you need the data structure definition but want to decouple it from specific backend services (e.g., for client-side generation or pure data validation).

```go
func JSMStruct(className, jsonSchemaStr string) (*Struct, error)
```

**Example:**

```go
jsonStr := `{
    "className": "Circle",
    "serviceName": "geometryService"
}`

// Using JSMServiceStruct
s1, _ := JSMServiceStruct("MyCircle", jsonStr)
// s1.ServiceName is "geometryService"

// Using JSMStruct
s2, _ := JSMStruct("MyCircle", jsonStr)
// s2.ServiceName is "" (empty)
```
