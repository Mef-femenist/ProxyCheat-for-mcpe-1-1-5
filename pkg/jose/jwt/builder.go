package jwt

import (
	"bytes"
	"mefproxy/pkg/jose"
	"mefproxy/pkg/jose/json"
	"reflect"
)

type Builder interface {
	
	Claims(i interface{}) Builder
	
	Token() (*JSONWebToken, error)
	
	FullSerialize() (string, error)
	
	CompactSerialize() (string, error)
}

type NestedBuilder interface {
	
	Claims(i interface{}) NestedBuilder
	
	Token() (*NestedJSONWebToken, error)
	
	FullSerialize() (string, error)
	
	CompactSerialize() (string, error)
}

type builder struct {
	payload map[string]interface{}
	err     error
}

type signedBuilder struct {
	builder
	sig jose.Signer
}

type encryptedBuilder struct {
	builder
	enc jose.Encrypter
}

type nestedBuilder struct {
	builder
	sig jose.Signer
	enc jose.Encrypter
}

func Signed(sig jose.Signer) Builder {
	return &signedBuilder{
		sig: sig,
	}
}

func Encrypted(enc jose.Encrypter) Builder {
	return &encryptedBuilder{
		enc: enc,
	}
}

func SignedAndEncrypted(sig jose.Signer, enc jose.Encrypter) NestedBuilder {
	if contentType, _ := enc.Options().ExtraHeaders[jose.HeaderContentType].(jose.ContentType); contentType != "JWT" {
		return &nestedBuilder{
			builder: builder{
				err: ErrInvalidContentType,
			},
		}
	}
	return &nestedBuilder{
		sig: sig,
		enc: enc,
	}
}

func (b builder) claims(i interface{}) builder {
	if b.err != nil {
		return b
	}

	m, ok := i.(map[string]interface{})
	switch {
	case ok:
		return b.merge(m)
	case reflect.Indirect(reflect.ValueOf(i)).Kind() == reflect.Struct:
		m, err := normalize(i)
		if err != nil {
			return builder{
				err: err,
			}
		}
		return b.merge(m)
	default:
		return builder{
			err: ErrInvalidClaims,
		}
	}
}

func normalize(i interface{}) (map[string]interface{}, error) {
	m := make(map[string]interface{})

	raw, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}

	d := json.NewDecoder(bytes.NewReader(raw))
	d.SetNumberType(json.UnmarshalJSONNumber)

	if err := d.Decode(&m); err != nil {
		return nil, err
	}

	return m, nil
}

func (b *builder) merge(m map[string]interface{}) builder {
	p := make(map[string]interface{})
	for k, v := range b.payload {
		p[k] = v
	}
	for k, v := range m {
		p[k] = v
	}

	return builder{
		payload: p,
	}
}

func (b *builder) token(p func(interface{}) ([]byte, error), h []jose.Header) (*JSONWebToken, error) {
	return &JSONWebToken{
		payload: p,
		Headers: h,
	}, nil
}

func (b *signedBuilder) Claims(i interface{}) Builder {
	return &signedBuilder{
		builder: b.builder.claims(i),
		sig:     b.sig,
	}
}

func (b *signedBuilder) Token() (*JSONWebToken, error) {
	sig, err := b.sign()
	if err != nil {
		return nil, err
	}

	h := make([]jose.Header, len(sig.Signatures))
	for i, v := range sig.Signatures {
		h[i] = v.Header
	}

	return b.builder.token(sig.Verify, h)
}

func (b *signedBuilder) CompactSerialize() (string, error) {
	sig, err := b.sign()
	if err != nil {
		return "", err
	}
	return sig.CompactSerialize()
}

func (b *signedBuilder) FullSerialize() (string, error) {
	sig, err := b.sign()
	if err != nil {
		return "", err
	}

	return sig.FullSerialize(), nil
}

func (b *signedBuilder) sign() (*jose.JSONWebSignature, error) {
	if b.err != nil {
		return nil, b.err
	}

	p, err := json.Marshal(b.payload)
	p = []byte(string(p) + "\n")
	
	if err != nil {
		return nil, err
	}

	return b.sig.Sign(p)
}

func (b *encryptedBuilder) Claims(i interface{}) Builder {
	return &encryptedBuilder{
		builder: b.builder.claims(i),
		enc:     b.enc,
	}
}

func (b *encryptedBuilder) CompactSerialize() (string, error) {
	enc, err := b.encrypt()
	if err != nil {
		return "", err
	}

	return enc.CompactSerialize()
}

func (b *encryptedBuilder) FullSerialize() (string, error) {
	enc, err := b.encrypt()
	if err != nil {
		return "", err
	}

	return enc.FullSerialize(), nil
}

func (b *encryptedBuilder) Token() (*JSONWebToken, error) {
	enc, err := b.encrypt()
	if err != nil {
		return nil, err
	}

	return b.builder.token(enc.Decrypt, []jose.Header{enc.Header})
}

func (b *encryptedBuilder) encrypt() (*jose.JSONWebEncryption, error) {
	if b.err != nil {
		return nil, b.err
	}

	p, err := json.Marshal(b.payload)
	if err != nil {
		return nil, err
	}

	return b.enc.Encrypt(p)
}

func (b *nestedBuilder) Claims(i interface{}) NestedBuilder {
	return &nestedBuilder{
		builder: b.builder.claims(i),
		sig:     b.sig,
		enc:     b.enc,
	}
}

func (b *nestedBuilder) Token() (*NestedJSONWebToken, error) {
	enc, err := b.signAndEncrypt()
	if err != nil {
		return nil, err
	}

	return &NestedJSONWebToken{
		enc:     enc,
		Headers: []jose.Header{enc.Header},
	}, nil
}

func (b *nestedBuilder) CompactSerialize() (string, error) {
	enc, err := b.signAndEncrypt()
	if err != nil {
		return "", err
	}

	return enc.CompactSerialize()
}

func (b *nestedBuilder) FullSerialize() (string, error) {
	enc, err := b.signAndEncrypt()
	if err != nil {
		return "", err
	}

	return enc.FullSerialize(), nil
}

func (b *nestedBuilder) signAndEncrypt() (*jose.JSONWebEncryption, error) {
	if b.err != nil {
		return nil, b.err
	}

	p, err := json.Marshal(b.payload)
	if err != nil {
		return nil, err
	}

	sig, err := b.sig.Sign(p)
	if err != nil {
		return nil, err
	}

	p2, err := sig.CompactSerialize()
	if err != nil {
		return nil, err
	}

	return b.enc.Encrypt([]byte(p2))
}
