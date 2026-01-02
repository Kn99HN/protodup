# protodup

Protobuf is a great system for defining structured data for (de)serialization. 
`protodup` is a barebone implementation of protobuf without code generation for
schema. Serialzing and deserialzing are processes of translating a schema into bytes 
and vice versa.

In protobuf, users would define a schema for their data

```
message Test {
	int field = 1
}
```

Each field is tagged with an integer. It's very important that we don't reuse
integer for different fields. If I have to guess, internally, the code would use
the integer as unique identifier for the field. Hence, it does not allow reusing of
integer (without special keyword).

Protodup supports the following data types: 

- `uint32, uint64, int32, int64, sint32, sint64`
- `string`

To encode the above types, protobuf uses `varints` and `record`

| Encoded Type | protobuf type |
| -------------|---------------|
|    varint    | uint32, uint64, int32, int64, sint32, sint64 |
|    record    | string |


`varint` is variable integer, which uses multiple bytes to represent an integer. 
The first bit of each byte for `varint` type indicates continuation. If set, we
need to keep parsing the bytes.

For example, `23` in binary would be `00010111`. This would be varint of size 1. The
leading value is 0 because there is only a single byte.

`257` is `00000001 00000001`. This wouldn't fit in a single `varint` byte. Therefore, 
we would need 2 bytes. To encode the value, we have `10000001 00000001`.

The most significant bit is `1`, indicating there we need to parse the next
byte. Grabbing 7 bits of each byte, we have `00000001 00000001`.

Given there are unsigned integer and signed integer, we have different schemes
for encoding it as bits. Unsigned integer is relatively straightforward. We just
need to parse the integer and write the bits. Signed integer is a bit more
complicated. In protobuf, signed integer can be encoded in 2 different ways. If 
encoded using the normal bits, it would consume lots of storage. Take -1 for
example. In binary, it would be 0xffffffff. This is because this is twos
complement. To get the binary value of `-1`, we do:

1. Convert to absolute value, which is 1
2. Flip all the bits 0xfffffffe
3. Add 1

Therefore, we have a new scheme for encoding signed integer. It's called zig-zag
encoding. The user can specify in the schema if they prefer this over the twos
complement encoding. The zig-zag encoding works as follow:

1. Map the integer to its correspoding positive integer.
2. Encode the signed integer as the mapped integer.

This saves bytes because a negative integer, which might have lots of most
significant bits set to 1 will have them set to 0 if zig-zag encoded.

For serializing, the high level steps are:
1. Get the descriptors of the structured data generated from the schema, which 
would yield a map of an integer to a value with type information.
2. Start serialzing integer into varints and string into record

For deserialzing, the high level steps are:
1. Parse the bytes into the structured data. The first field is always
combination of the field number and the tag value.
2. For each field, parse the bytes and create the structured data.

Things to note is that, int32, int64, intN is various ways of telling the
compiler to interpret the bits. It doesn't change the binary representation of
the value itself. One bug that we ran into is zig-zag encoding on Min integer
value. Carrying out the bit shift with type int32, the value would be -1. Given
that all the bits are 1 and we only right shift to encode the value as varints,
it would cause infinite loop. Therefore, we would need to cast value to uint32
before encoding to varints. When right shifting the integer, int32 would cause
the msb to preserve the signed bit, which would make it 0xff forever. Casting to
uint32 would make it insert a 0, leading to eventual loop termination.

Serializing string is more straightforward. Each character is written as a utf-8
encoded byte. Before the string is encoded, we write out the length of the
record as varints. Deserializing ends up parsing the record length and read X
bytes for the string itself.

