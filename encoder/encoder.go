// Package encoder handles source encoding formats
package encoder

// Encoder handles encoding and decoding of a variety of config formats
type Encoder interface {
	Encode(interface{}) ([]byte, error)
	Decode([]byte, interface{}) error
	String() string
}
