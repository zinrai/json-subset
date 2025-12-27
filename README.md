# json-subset

A command-line tool for checking if one JSON is a subset of another. Arrays are compared as sets, ignoring element order.

## Why json-subset?

You often need to verify that an HTTP API response contains expected fields:

```bash
# Does this response contain our required fields?
curl -s https://api.example.com/config > response.json
```

### Why Not jq?

jq's `contains` function can check subsets:

```bash
$ jq -s '.[0] as $a | .[1] | contains($a)' required.json response.json
```

However, it only returns `true` or `false`. When validation fails, you don't know *what* is missing:

```
false
```

You could write complex jq expressions to show differences, but they become hard to maintain:

```bash
$ jq -s '
  .[0] as $sub | .[1] as $sup |
  [
    $sub | paths(scalars) | . as $p |
    {
      path: ($p | map(tostring) | join(".")),
      subset_value: ($sub | getpath($p)),
      superset_value: ($sup | getpath($p))
    } |
    select(.subset_value != .superset_value)
  ]
' required.json response.json
```

This is jq wizardry that's difficult to understand, modify, or hand off to teammates.

### Why Not jd?

[jd](https://github.com/josephburnett/jd) is excellent for **equality** checks and showing diffs:

```bash
$ jd -set a.json b.json
```

But jd doesn't have a **subset** mode. It tells you *everything* that's different, not whether one JSON is contained within another.

### json-subsets Role

json-subset does one thing: subset checking with clear difference output.

```bash
$ json-subset required.json response.json
FAIL: First JSON is not a subset of second JSON.

 {
   "user": {
-    "email": "alice@example.com",
     "name": "alice"
   }
 }
```

When validation fails, you immediately see what's wrong. The `-` marker shows which fields are missing or have mismatched values.

### Tool Responsibilities

| Task                          | Tool        |
|-------------------------------|-------------|
| JSON equality check           | jd          |
| JSON transformation/filtering | jq          |
| Subset validation with diff   | json-subset |

## Installation

```bash
$ go install github.com/zinrai/json-subset@latest
```

## Usage

```bash
$ json-subset <subset.json> <superset.json>
```

### Examples

Check if required fields exist in API response:

```bash
$ json-subset required.json response.json
```

With curl and process substitution:

```bash
$ json-subset expected.json <(curl -s https://api.example.com/config)
```

With jq preprocessing to remove dynamic fields:

```bash
$ json-subset expected.json <(curl -s https://api.example.com/config | jq 'del(.timestamp)')
```

### Exit Codes

- `0`: Success (first JSON is a subset of second)
- `1`: Failure (first JSON is not a subset of second)
- `2`: Error (invalid input, file not found, etc.)

## Behavior

### Object Comparison

An object A is a subset of object B if:

- Every key in A exists in B
- For each key, the value in A is a subset of the value in B

```bash
# subset.json
{"name": "alice"}

# superset.json
{"name": "alice", "age": 30}

# Result: OK (subset)
```

### Array Comparison (Set Mode)

Arrays are compared as sets. Element order is ignored.

```bash
# subset.json
[2, 1]

# superset.json
[1, 2, 3]

# Result: OK (subset, order ignored)
```

### Nested Structures

Subset checking works recursively for nested objects and arrays.

```bash
# subset.json
{"user": {"name": "alice"}}

# superset.json
{"user": {"name": "alice", "age": 30}, "metadata": {}}

# Result: OK (subset)
```

## Difference Output

When the subset check fails, json-subset displays the subset JSON with diff markers. Lines prefixed with `-` indicate missing keys or mismatched values:

```
FAIL: First JSON is not a subset of second JSON.

 {
-  "license": "MIT",
   "name": "myapp",
   "version": "1.0.0"
 }
```

For nested structures:

```
FAIL: First JSON is not a subset of second JSON.

 {
   "user": {
-    "email": "alice@example.com",
     "name": "alice"
   }
 }
```

For arrays:

```
FAIL: First JSON is not a subset of second JSON.

 {
   "tags": [
     "production",
     "stable",
-    "admin"
   ]
 }
```

## Examples

The `examples/` directory contains sample JSON files for testing:

Success case: required fields exist in response:

```bash
json-subset examples/required.json examples/response.json
```

Failure case: required field missing:

```bash
$ json-subset examples/required_with_missing.json examples/response.json
```

Array subset (order ignored):

```bash
$ json-subset examples/required_tags.json examples/actual_tags.json
```

Nested object subset:

```bash
$ json-subset examples/required_nested.json examples/response_nested.json
```

Complex API response with multiple missing fields:

```bash
$ json-subset examples/required_api_response.json examples/actual_api_response.json
FAIL: First JSON is not a subset of second JSON.

 {
   "data": {
-    "permissions": [
-      "read",
-      "write"
-    ],
     "user": {
       "id": 123,
       "profile": {
-        "email": "alice@example.com",
         "name": "alice"
       }
     }
   },
   "status": "success"
 }
```

## License

This project is licensed under the [MIT License](./LICENSE).
