package json

import (
	"encoding/json"
	"fmt"
	"io"

	"go.bytecodealliance.org/internal/codec"
)

type Decoder struct {
	dec *json.Decoder
	r   codec.Resolvers
}

func NewDecoder(r io.Reader, resolvers ...codec.Resolver) *Decoder {
	dec := json.NewDecoder(r)
	dec.UseNumber()
	return &Decoder{
		dec: dec,
		r:   codec.Resolvers(resolvers),
	}
}

func (dec *Decoder) Decode(v any) error {
	if c := dec.r.ResolveCodec(v); c != nil {
		v = c
	}

	err := dec.decodeToken(v)
	if err != nil && err != io.EOF {
		return err
	}

	return nil
}

func (dec *Decoder) decodeToken(v any) error {
	tok, err := dec.dec.Token()
	if err != nil {
		return err
	}
	if tok == nil {
		return codec.DecodeNil(v)
	}

	switch tok := tok.(type) {
	case bool:
		return codec.DecodeBool(v, tok)
	case json.Number:
		return codec.DecodeNumber(v, string(tok))
	case string:
		return codec.DecodeString(v, tok)
	case json.Delim:
		switch tok {
		case '{':
			return dec.decodeObject(v)
		case '[':
			return dec.decodeArray(v)
		default:
			return fmt.Errorf("unexpected JSON token %v at offset %d", tok, dec.dec.InputOffset())
		}
	}

	return nil
}

// decodeObject decodes a JSON object into v.
// It expects that the initial { token has already been decoded.
func (dec *Decoder) decodeObject(o any) error {
	d, ok := o.(codec.FieldDecoder)
	if !ok {
		d = &ignore{}
	}

	for dec.dec.More() {
		name, err := dec.stringToken()
		if err != nil {
			return err
		}
		fdec := &onceDecoder{Decoder: dec}
		err = d.DecodeField(fdec, name)
		if err != nil {
			return err
		}
		if fdec.calls == 0 {
			err = dec.Decode(nil)
			if err != nil {
				return err
			}
		}
	}

	tok, err := dec.dec.Token()
	if err != nil {
		return err
	}
	if tok != json.Delim('}') {
		return fmt.Errorf("unexpected JSON token %v at offset %d", tok, dec.dec.InputOffset())
	}

	return nil
}

// decodeArray decodes a JSON array into v.
// It expects that the initial [ token has already been decoded.
func (dec *Decoder) decodeArray(v any) error {
	d, ok := v.(codec.ElementDecoder)
	if !ok {
		d = &ignore{}
	}

	for i := 0; dec.dec.More(); i++ {
		edec := &onceDecoder{Decoder: dec}
		err := d.DecodeElement(edec, i)
		if err != nil {
			return err
		}
		if edec.calls == 0 {
			err = dec.Decode(nil)
			if err != nil {
				return err
			}
		}
	}

	tok, err := dec.dec.Token()
	if err != nil {
		return err
	}
	if tok != json.Delim(']') {
		return fmt.Errorf("unexpected JSON token %v at offset %d", tok, dec.dec.InputOffset())
	}

	return nil
}

func (dec *Decoder) stringToken() (string, error) {
	tok, err := dec.dec.Token()
	if err != nil {
		return "", err
	}
	s, ok := tok.(string)
	if !ok {
		return "", fmt.Errorf("unexpected JSON token %v at offset %d", tok, dec.dec.InputOffset())
	}
	return s, nil
}

type onceDecoder struct {
	*Decoder
	calls int
}

func (dec *onceDecoder) Decode(v any) error {
	dec.calls++
	if dec.calls > 1 {
		return fmt.Errorf("unexpected call to Decode (%d > 1)", dec.calls)
	}
	return dec.Decoder.Decode(v)
}

type ignore struct{}

func (ig *ignore) DecodeField(dec codec.Decoder, name string) error {
	return nil
}

func (ig *ignore) DecodeElement(dec codec.Decoder, i int) error {
	return nil
}

func (i *ignore) UnmarshalJSON([]byte) error {
	return nil
}
