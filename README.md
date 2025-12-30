# protodup

Protobuf is a great system for serialzing structured data. Here, in protodup,
I attempt to implement a simple dup of protobuf. It's very simple. Serialzing
and deserialzing is basically translate a schema into bytes and from bytes into
the structured data.

In protobuf, users would define a schema for their data

```
message Test {
	int field = 1
}
```

Each field is tagged with an integer. It's very important that we don't reuse
integer for different fields. Interanlly, the way schema works is generating
a map between an index to a a field with type information. To represent
different type, protobuf uses the following encoding

varints -> int8, int16, ....
record -> string


varints is variable ints, which use multiple bytes to represent an integer. It
uses the first bit of each byte as an indicator if there are more to parse.
Imagine `23` in binary would be 0001 0111. This would be varint of size 1. The
leading value is 0 because there is only a single byte.

257 is 100000001. This wouldn't fit in a single varint byte. Therefore, we would 
need 2 bytes. To represent the value, we will need:

`1000 0001 0000 0001`
first bit tells us that we need to parse the next byte to get the full integer.
Once we parse everything, we would concatenate the bits together to form the 
field value.

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

