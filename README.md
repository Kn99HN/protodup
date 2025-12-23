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

For serializing, the high level steps are:
1. Get the descriptors of the structured data generated from the schema, which 
would yield a map of an integer to a value with type information.
2. Start serialzing integer into varints and string into record

For deserialzing, the high level steps are:
1. Parse the bytes into the structured data. The first field is always
combination of the field number and the tag value.
2. For each field, parse the bytes and create the structured data.

