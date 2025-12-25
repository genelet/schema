# JSON Schema Representation of Genelet Schema Types

The `schema` package allows you to define Genelet Schema types (`Struct`, `ListStruct`, `MapStruct`, `Map2Struct`) using standard JSON Schema (Draft-7) with minimal extensions. This document details how different JSON Schema structures map to their corresponding Genelet Schema types.

## Overview

The mapping logic is primarily driven by the structural keywords present in the JSON Schema:

| JSON Schema Keyword | Genelet Schema Type | Description |
|---------------------|---------------------|-------------|
| `properties` | **SingleStruct** | A single object with known fields. |
| `items` | **ListStruct** | A list of objects of the same type. |
| `additionalProperties` | **MapStruct** | A map with string keys and values of the same type. |
| `x-map2` (extension) | **Map2Struct** | A map of maps (two-layer keys). |
| `type` (primitive) | **SingleStruct** | A "primitive" or "leaf" type (e.g., `string`, `integer`, `MyClass`). |

## 1. SingleStruct (Struct)

A `SingleStruct` represents a typed object, potentially with nested fields.

### Explicit Object with Properties
Use `type: "object"` and `properties` to define a struct with fields. The `type` field in the JSON schema corresponds to `ClassName`. If `type` is "object", the resulting `ClassName` might be empty or inferred (depending on function used), or you can provide a custom string for `type` to set the `ClassName`.

**JSON Schema:**
```json
{
  "type": "Person",
  "properties": {
    "Name": { "type": "string" },
    "Age": { "type": "integer" }
  }
}
```

**Resulting Schema Type:** `SingleStruct`
- ClassName: "Person"
- Fields: "Name" (SingleStruct "string"), "Age" (SingleStruct "integer")

### Leaf Node / Primitive
A schema with just a `type` string (and optional `serviceName`) is treated as a `SingleStruct` leaf node.

**JSON Schema:**
```json
{
  "type": "string"
}
```
**Resulting Schema Type:** `SingleStruct`
- ClassName: "string"

## 2. ListStruct

A `ListStruct` represents an array or slice of items. It corresponds to `[]T` in Go.

Use `type: "array"` and `items` to define the element type.

**JSON Schema:**
```json
{
  "type": "array",
  "items": {
    "type": "Person"
  }
}
```

**Resulting Schema Type:** `ListStruct`
- ListFields: `[SingleStruct("Person")]` (describes the element type)

## 3. MapStruct

A `MapStruct` represents a map with string keys. It corresponds to `map[string]T` in Go.

Use `type: "object"` and `additionalProperties` to define the value type.

**JSON Schema:**
```json
{
  "type": "object",
  "additionalProperties": {
    "type": "Person"
  }
}
```

**Resulting Schema Type:** `MapStruct`
- MapFields: `{"*": SingleStruct("Person")}` (wildcard key describing value type)

## 4. Map2Struct

A `Map2Struct` represents a two-level nested map. It corresponds to `map[[2]string]T` in Go, where the key is effectively a composite `(key1, key2)`.

To represent this structure, use the custom extension `x-map2: true`. The structure requires two layers of `properties`.

**JSON Schema:**
```json
{
  "type": "object",
  "x-map2": true,
  "properties": {
    "region1": {
      "type": "object",
      "properties": {
        "key1": { "type": "ServiceA" },
        "key2": { "type": "ServiceB" }
      }
    },
    "region2": {
      "type": "object",
      "properties": {
        "key3": { "type": "ServiceC" }
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

## Service Decoration

The `serviceName` keyword can be added to any schema node to specify the service responsible for that data.

**JSON Schema:**
```json
{
  "type": "UserProfile",
  "serviceName": "userService",
  "properties": {
    "Avatar": {
      "type": "Image",
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

*   `className` (string): The name assigned to the root object (SingleStruct). Even if the JSON Schema defines a `type` name, this argument takes precedence for the root.
*   `jsonSchemaStr` (string): The standard JSON Schema (Draft-7) string to parse.

**Conversion Rules:**

The following table summarizes how different JSON Schema structures are converted into Genelet Schema types, and how `ClassName` and `ServiceName` are determined.

| JSON Schema Pattern | Genelet Type | ClassName | ServiceName |
| :--- | :--- | :--- | :--- |
| `{"type": "Circle", "serviceName": "s1"}` | **SingleStruct** | "Circle" | "s1" |
| `{"type": "array", "items": {"type": "Circle", "serviceName": "s2"}}` | **ListStruct** | n/a | n/a (field level) |
| `{"type": "object", "additionalProperties": {"type": "Circle", "serviceName": "s3"}}` | **MapStruct** | n/a | n/a (field level) |
| `{"type": "Class1", "properties": {"Field1": {"type": "Circle"}}}` | **SingleStruct** | "Class1" | "" |
| `{"type": "object", "x-map2": true, "properties": ...}` | **Map2Struct** | n/a | n/a |

**Examples:**

#### Case 1: Simple Object (SingleStruct)

Defining a `Geo` object with a string field and a primitive type field.

```go
jsonStr := `{
    "type": "Geo",
    "properties": {
        "Name": { "type": "string" },
        "Shape": { "type": "Circle", "serviceName": "s_shape" }
    }
}`
s, err := JSMServiceStruct("Geo", jsonStr)
// Result: SingleStruct "Geo". Field "Shape" is SingleStruct("Circle") served by "s_shape".
```

#### Case 2: List (ListStruct)

Defining a list of `Circle` objects.

```go
jsonStr := `{
    "type": "array",
    "items": {
        "type": "Circle",
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
    "type": "object",
    "additionalProperties": {
        "type": "Circle",
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
    "type": "object",
    "x-map2": true,
    "properties": {
        "region1": {
            "type": "object",
            "properties": {
                "key1": { "type": "Circle", "serviceName": "s_map2_val" }
            }
        }
    }
}`
s, err := JSMServiceStruct("GeoMap", jsonStr)
// Result: Map2Struct with region "region1" -> key "key1" -> SingleStruct("Circle") served by "s_map2_val".
```

### `JSMStruct`

`JSMStruct` is a wrapper around `JSMServiceStruct`. It parses the JSON Schema but **removes all `serviceName` fields** from the resulting `Struct` tree. This is useful when you need the data structure definition but want to decouple it from specific backend services (e.g., for client-side generation or pure data validation).

```go
func JSMStruct(className, jsonSchemaStr string) (*Struct, error)
```

**Example:**

```go
jsonStr := `{
    "type": "Circle",
    "serviceName": "geometryService",
    "properties": {
        "Radius": { "type": "integer" }
    }
}`

// Using JSMServiceStruct
s1, _ := JSMServiceStruct("MyCircle", jsonStr)
// s1.ServiceName is "geometryService"

// Using JSMStruct
s2, _ := JSMStruct("MyCircle", jsonStr)
// s2.ServiceName is "" (empty)
```
