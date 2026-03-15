package jwt

import (
	"fmt"
	"strings"

	jose "mefproxy/pkg/jose"
	"mefproxy/pkg/jose/json"
)

type JSONWebToken struct {
	payload           func(k interface{}) ([]byte, error)
	unverifiedPayload func() []byte
	Headers           []jose.Header
}

type NestedJSONWebToken struct {
	enc     *jose.JSONWebEncryption
	Headers []jose.Header
}

func (t *JSONWebToken) Claims(key interface{}, dest ...interface{}) error {
	payloadKey := tryJWKS(t.Headers, key)

	b, err := t.payload(payloadKey)
	if err != nil {
		return err
	}

	for _, d := range dest {
		if err := json.Unmarshal(b, d); err != nil {
			return err
		}
	}

	return nil
}

func (t *JSONWebToken) UnsafeClaimsWithoutVerification(dest ...interface{}) error {
	if t.unverifiedPayload == nil {
		return fmt.Errorf("square/go-jose: Cannot get unverified claims")
	}
	claims := t.unverifiedPayload()
	for _, d := range dest {
		if err := json.Unmarshal(claims, d); err != nil {
			return err
		}
	}
	return nil
}

func (t *NestedJSONWebToken) Decrypt(decryptionKey interface{}) (*JSONWebToken, error) {
	key := tryJWKS(t.Headers, decryptionKey)

	b, err := t.enc.Decrypt(key)
	if err != nil {
		return nil, err
	}

	sig, err := ParseSigned(string(b))
	if err != nil {
		return nil, err
	}

	return sig, nil
}

func ParseSigned(s string) (*JSONWebToken, error) {
	sig, err := jose.ParseSigned(s)
	if err != nil {
		return nil, err
	}
	headers := make([]jose.Header, len(sig.Signatures))
	for i, signature := range sig.Signatures {
		headers[i] = signature.Header
	}

	return &JSONWebToken{
		payload:           sig.Verify,
		unverifiedPayload: sig.UnsafePayloadWithoutVerification,
		Headers:           headers,
	}, nil
}

func ParseEncrypted(s string) (*JSONWebToken, error) {
	enc, err := jose.ParseEncrypted(s)
	if err != nil {
		return nil, err
	}

	return &JSONWebToken{
		payload: enc.Decrypt,
		Headers: []jose.Header{enc.Header},
	}, nil
}

func ParseSignedAndEncrypted(s string) (*NestedJSONWebToken, error) {
	enc, err := jose.ParseEncrypted(s)
	if err != nil {
		return nil, err
	}

	contentType, _ := enc.Header.ExtraHeaders[jose.HeaderContentType].(string)
	if strings.ToUpper(contentType) != "JWT" {
		return nil, ErrInvalidContentType
	}

	return &NestedJSONWebToken{
		enc:     enc,
		Headers: []jose.Header{enc.Header},
	}, nil
}

func tryJWKS(headers []jose.Header, key interface{}) interface{} {
	var jwks jose.JSONWebKeySet

	switch jwksType := key.(type) {
	case *jose.JSONWebKeySet:
		jwks = *jwksType
	case jose.JSONWebKeySet:
		jwks = jwksType
	default:
		return key
	}

	var kid string
	for _, header := range headers {
		if header.KeyID != "" {
			kid = header.KeyID
			break
		}
	}

	if kid == "" {
		return key
	}

	keys := jwks.Key(kid)
	if len(keys) == 0 {
		return key
	}

	return keys[0].Key
}
