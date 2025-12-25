# Introducing `genelet/schema`: Dynamic, Schema-Driven Type Orchestration for Go

If you search for "JSON Schema" in the Go ecosystem, you'll likely land on libraries like `invopop/jsonschema`. These are excellent tools that do one thing very well: they use **Reflection** to generate a JSON Schema from your existing Go structs.

**But what if you need to do the exact reverse?**

What if you have a JSON Schema (or a similar description) and you need to construct a Go object that represents that structure *dynamically* at runtime, without generating code? What if that structure needs to carry not just data, but metadata about *where* that data lives or *who* should handle it?

Enter [**`genelet/schema`**](https://github.com/genelet/schema).

## The Missing Link

To understand where `genelet/schema` fits, let's look at the existing landscape:

1.  **`mitchellh/mapstructure`**: This is the gold standard for decoding generic `map[string]interface{}` data into Go structs. However, it relies on static Go types. You must define your `struct` beforehand.
2.  **`gburgyan/go-poly` / `dhoelle/oneof`**: These tackle polymorphic JSON (e.g., a list containing both `Circle` and `Square` objects). They are **data-driven**—they peek at specific fields (like `type`) in the incoming JSON to decide what to do.
3.  **`invopop/jsonschema`**: As mentioned, this goes from **Go Struct → JSON Schema**.
4.  **`santhosh-tekuri/jsonschema`**: A powerful **validator**. It tells you if data matches a schema, but it doesn't give you a manipulatable type definition of that data.

**`genelet/schema`** is unique because it is **Schema-Driven**. It takes a description (a Schema) and instantiates a robust `Struct` object that acts as a dynamic type system. It doesn't just validate; it *defines* and *orchestrates*.

## The Core Concept: `Struct` vs. `jsonSchemaStr`

At the heart of the package are two representations of the same thing:

1.  **The `Struct`**: A Go type that represents your data structure. It holds a `ClassName`, specific `Fields` (which can be primitives, Lists, or Maps), and optional metadata.
2.  **The `jsonSchemaStr`**: A simplified JSON Schema string that defines the `Struct`.

You can convert between them seamlessly.

### Chapter 1: Dynamic Interface Types

Imagine you are building a low-code platform or a Terraform provider. You receive a JSON payload describing a resource, but you don't know at compile time if it's a `Database` or a `LoadBalancer`.

Instead of writing a massive switch statement or generating code on the fly, you can use `JSMStruct` (JSON Schema to Struct).

#### The Schema
Lests define a schema for a generic "Person" object:

```json
{
  "className": "Person",
  "properties": {
    "Name": { "className": "String" },
    "Address": {
      "className": "Address",
      "properties": {
        "City": { "className": "String" }
      }
    }
  }
}
```

#### The Code

```go
package main

import (
	"encoding/json"
	"fmt"
	"github.com/genelet/schema"
)

func main() {
	jsonStr := `{...}` // The JSON above

	// Parse the schema into a dynamic Struct
	personStruct, err := schema.JSMStruct("Person", jsonStr)
	if err != nil {
		panic(err)
	}

	// You now have a typed definition!
	fmt.Printf("Class: %s\n", personStruct.ClassName)
	// Output: Class: Person

	// You can traverse it dynamically
	addressField := personStruct.Fields["Address"].GetSingleStruct()
	fmt.Printf("Nested Class: %s\n", addressField.ClassName)
	// Output: Nested Class: Address
    
    // You can even marshal it back to JSON Schema
    out, _ := json.MarshalIndent(personStruct, "", "  ")
    fmt.Println(string(out))
}
```

This allows your application to handle entirely new types defined by configuration, not code.

### Chapter 2: Service Orchestration

This is where `genelet/schema` truly shines. In distributed systems, a schema often implies *ownership*. "This part of the configuration belongs to the **User Service**, but this nested part belongs to the **Billing Service**."

The package has built-in support for `ServiceName`.

#### The Orchestration Schema

We use the same structure, but notice the `serviceName` field:

```json
{
  "className": "UserProfile",
  "serviceName": "userService",
  "properties": {
    "PaymentMethod": {
      "className": "CreditCard",
      "serviceName": "billingService",
      "properties": { ... }
    }
  }
}
```

#### The Orchestration Code

We use `JSMServiceStruct` to parse this.

```go
func main() {
	orchestrationStr := `{...}` // The JSON above with serviceNames

	// Parse with Service awareness
	root, err := schema.JSMServiceStruct("UserProfile", orchestrationStr)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Root handles by: %s\n", root.ServiceName)
	// Output: Root handles by: userService

	payment := root.Fields["PaymentMethod"].GetSingleStruct()
	fmt.Printf("Payment handled by: %s\n", payment.ServiceName)
	// Output: Payment handled by: billingService
}
```

#### Why is this useful?

This structure allows you to build **Schema-Driven Gateways**. Your gateway can parse an incoming request, inspect the `Struct`, and automatically route parts of the payload to different downstream microservices based on the `ServiceName` property—all without hardcoding routing rules.




### Chapter 3: JSON Marshaling

The `Struct` type implements the standard `json.Marshaler` and `json.Unmarshaler` interfaces. This means you can easily serialize your dynamic type definitions to and from JSON.

This is particularly useful for tasks like:
*   Saving a user's configuration to a database.
*   Sending a schema definition over the wire to another service.
*   Debugging your dynamic structures.

```go
// 1. Create a Struct definition programmatically
spec, _ := schema.NewStruct("Person", map[string]any{
    "Name": "String",
    "Age":  "Integer",
})

// 2. Marshal it to JSON Schema
data, _ := json.MarshalIndent(spec, "", "  ")
fmt.Println(string(data))
/* Output:
{
  "className": "Person",
  "properties": {
    "Name": { "className": "String" },
    "Age": { "className": "Integer" }
  }
}
*/

// 3. Unmarshal it back into a Struct
var newSpec schema.Struct
checkErr(json.Unmarshal(data, &newSpec))
```

### Chapter 4: Typical Usage Patterns

Here are some common patterns you'll encounter when defining schemas.

#### 1. Primitives and Interfaces
You don't need to define every single field. Standard primitives like strings and integers involve no "orchestration" or "dynamic type lookup", so they are often omitted. You only explicitly define fields that are **Interfaces** (requiring a concrete type) or need **Service Injection**.

```json
// "Simple" fields like 'Name' are omitted. 
// "PaymentMethod" is an interface, so we define it.
{
  "className": "UserProfile",
  "properties": {
    "PaymentMethod": { "className": "CreditCard" }
  }
}
```

#### 2. Lists (Arrays)
To define a list of items, use the `items` keyword. This corresponds to `[]T` in Go.

```json
{
  "items": {
    "className": "Person",
    "serviceName": "userService"
  }
}
// Result: A list where every item is a "Person" handled by "userService".
```

#### 3. Maps
To define a map with string keys, use `additionalProperties`. This corresponds to `map[string]T`.

```json
{
  "additionalProperties": {
    "className": "ConfigItem"
  }
}
// Result: A map where every value is a "ConfigItem".
```

#### 4. Two-Level Maps (Map2)
Sometimes you need a map of maps (e.g., `Region -> Zone -> Service`). We use a custom `x-map2` extension for this.

```json
{
  "x-map2": true,
  "properties": {
    "us-east": {
      "properties": {
        "zone-1": { "className": "Node", "serviceName": "clusterA" }
      }
    }
  }
}
```


### Chapter 5: Nested Usage Patterns

For complex applications, you often need to mix different structures.

#### 5. Mixed Fields with Services
A struct can contain various field types (simple, list, map), all backed by different services.

```json
{
  "properties": {
    "SimpleField": { "className": "Metric", "serviceName": "metric_service" },
    "ListField":   { "items": { "className": "Log", "serviceName": "log_service" } },
    "MapField":    { "additionalProperties": { "className": "Config", "serviceName": "config_service" } }
  }
}
```

#### 6. Nested Structs with Services
Fields can be defined as inline structs, eventually leading to service-backed fields deep in the hierarchy.

```json
{
  "properties": {
    "Group1": {
      "className": "SubClass1",
      "properties": {
        "Item": { "className": "Detail", "serviceName": "s1" }
      }
    }
  }
}
```

#### 7. List of Structs with Services
A list where the item type is a complex object (defined by properties), which in turn contains service-backed fields.

```json
{
  "items": {
    "className": "ItemClass",
    "properties": {
      "Info": { "className": "Data", "serviceName": "data_service" }
    }
  }
}
```

#### 8. Map of Structs with Services
A map where the value type is a complex object, which contains service-backed fields.

```json
{
  "additionalProperties": {
    "className": "EntryClass",
    "properties": {
      "Record": { "className": "Row", "serviceName": "db_service" }
    }
  }
}
```

## Conclusion

`genelet/schema` provides the missing reverse-gear in the Go JSON ecosystem. By treating Schemas as first-class citizens that can be instantiated into rich, traversable Go objects, it opens the door for:

*   **Dynamic Configurations**: Define types in JSON, use them in Go.
*   **Service Mesh & Orchestration**: Route data based on schema definitions.
*   **Legacy Systems Integration**: Wrap old endpoints with modern schema definitions.
*   **Low-Code/No-Code Backends**: Handle arbitrary data structures safely.

It’s time to stop fighting your JSON and start orchestrating it.

Check it out at [github.com/genelet/schema](https://github.com/genelet/schema).


