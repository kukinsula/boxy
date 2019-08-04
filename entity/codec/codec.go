package codec

type Codec interface {
	Encode(data interface{}) ([]byte, error)
	Decode(data []byte, v interface{}) error
}
