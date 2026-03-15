package jose

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"mefproxy/pkg/jose/json"
)

type rawJSONWebSignature struct {
	Payload    *byteBuffer        `json:"payload,omitempty"`
	Signatures []rawSignatureInfo `json:"signatures,omitempty"`
	Protected  *byteBuffer        `json:"protected,omitempty"`
	Header     *rawHeader         `json:"header,omitempty"`
	Signature  *byteBuffer        `json:"signature,omitempty"`
}

type rawSignatureInfo struct {
	Protected *byteBuffer `json:"protected,omitempty"`
	Header    *rawHeader  `json:"header,omitempty"`
	Signature *byteBuffer `json:"signature,omitempty"`
}

type JSONWebSignature struct {
	payload []byte
	
	Signatures []Signature
}

type Signature struct {
	
	Header Header

	Protected Header

	Unprotected Header

	Signature []byte

	protected *rawHeader
	header    *rawHeader
	original  *rawSignatureInfo
}

func ParseSigned(signature string) (*JSONWebSignature, error) {
	signature = stripWhitespace(signature)
	if strings.HasPrefix(signature, "{") {
		return parseSignedFull(signature)
	}

	return parseSignedCompact(signature, nil)
}

func ParseDetached(signature string, payload []byte) (*JSONWebSignature, error) {
	if payload == nil {
		return nil, errors.New("square/go-jose: nil payload")
	}
	return parseSignedCompact(stripWhitespace(signature), payload)
}

func (sig Signature) mergedHeaders() rawHeader {
	out := rawHeader{}
	out.merge(sig.protected)
	out.merge(sig.header)
	return out
}

func (obj JSONWebSignature) computeAuthData(payload []byte, signature *Signature) ([]byte, error) {
	var authData bytes.Buffer

	protectedHeader := new(rawHeader)

	if signature.original != nil && signature.original.Protected != nil {
		if err := json.Unmarshal(signature.original.Protected.bytes(), protectedHeader); err != nil {
			return nil, err
		}
		authData.WriteString(signature.original.Protected.base64())
	} else if signature.protected != nil {
		protectedHeader = signature.protected
		authData.WriteString(base64.RawURLEncoding.EncodeToString(mustSerializeJSON(protectedHeader)))
	}

	needsBase64 := true

	if protectedHeader != nil {
		var err error
		if needsBase64, err = protectedHeader.getB64(); err != nil {
			needsBase64 = true
		}
	}

	authData.WriteByte('.')

	if needsBase64 {
		authData.WriteString(base64.RawURLEncoding.EncodeToString(payload))
	} else {
		authData.Write(payload)
	}

	return authData.Bytes(), nil
}

func parseSignedFull(input string) (*JSONWebSignature, error) {
	var parsed rawJSONWebSignature
	err := json.Unmarshal([]byte(input), &parsed)
	if err != nil {
		return nil, err
	}

	return parsed.sanitized()
}

func (parsed *rawJSONWebSignature) sanitized() (*JSONWebSignature, error) {
	if parsed.Payload == nil {
		return nil, fmt.Errorf("square/go-jose: missing payload in JWS message")
	}

	obj := &JSONWebSignature{
		payload:    parsed.Payload.bytes(),
		Signatures: make([]Signature, len(parsed.Signatures)),
	}

	if len(parsed.Signatures) == 0 {
		
		signature := Signature{}
		if parsed.Protected != nil && len(parsed.Protected.bytes()) > 0 {
			signature.protected = &rawHeader{}
			err := json.Unmarshal(parsed.Protected.bytes(), signature.protected)
			if err != nil {
				return nil, err
			}
		}

		if parsed.Header != nil && parsed.Header.getNonce() != "" {
			return nil, ErrUnprotectedNonce
		}

		signature.header = parsed.Header
		signature.Signature = parsed.Signature.bytes()
		
		signature.original = &rawSignatureInfo{
			Protected: parsed.Protected,
			Header:    parsed.Header,
			Signature: parsed.Signature,
		}

		var err error
		signature.Header, err = signature.mergedHeaders().sanitized()
		if err != nil {
			return nil, err
		}

		if signature.header != nil {
			signature.Unprotected, err = signature.header.sanitized()
			if err != nil {
				return nil, err
			}
		}

		if signature.protected != nil {
			signature.Protected, err = signature.protected.sanitized()
			if err != nil {
				return nil, err
			}
		}

		jwk := signature.Header.JSONWebKey
		if jwk != nil && (!jwk.Valid() || !jwk.IsPublic()) {
			return nil, errors.New("square/go-jose: invalid embedded jwk, must be public key")
		}

		obj.Signatures = append(obj.Signatures, signature)
	}

	for i, sig := range parsed.Signatures {
		if sig.Protected != nil && len(sig.Protected.bytes()) > 0 {
			obj.Signatures[i].protected = &rawHeader{}
			err := json.Unmarshal(sig.Protected.bytes(), obj.Signatures[i].protected)
			if err != nil {
				return nil, err
			}
		}

		if sig.Header != nil && sig.Header.getNonce() != "" {
			return nil, ErrUnprotectedNonce
		}

		var err error
		obj.Signatures[i].Header, err = obj.Signatures[i].mergedHeaders().sanitized()
		if err != nil {
			return nil, err
		}

		if obj.Signatures[i].header != nil {
			obj.Signatures[i].Unprotected, err = obj.Signatures[i].header.sanitized()
			if err != nil {
				return nil, err
			}
		}

		if obj.Signatures[i].protected != nil {
			obj.Signatures[i].Protected, err = obj.Signatures[i].protected.sanitized()
			if err != nil {
				return nil, err
			}
		}

		obj.Signatures[i].Signature = sig.Signature.bytes()

		jwk := obj.Signatures[i].Header.JSONWebKey
		if jwk != nil && (!jwk.Valid() || !jwk.IsPublic()) {
			return nil, errors.New("square/go-jose: invalid embedded jwk, must be public key")
		}

		original := sig

		obj.Signatures[i].header = sig.Header
		obj.Signatures[i].original = &original
	}

	return obj, nil
}

func parseSignedCompact(input string, payload []byte) (*JSONWebSignature, error) {
	parts := strings.Split(input, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("square/go-jose: compact JWS format must have three parts")
	}

	if parts[1] != "" && payload != nil {
		return nil, fmt.Errorf("square/go-jose: payload is not detached")
	}

	rawProtected, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, err
	}

	if payload == nil {
		payload, err = base64.RawURLEncoding.DecodeString(parts[1])
		if err != nil {
			return nil, err
		}
	}

	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, err
	}

	raw := &rawJSONWebSignature{
		Payload:   newBuffer(payload),
		Protected: newBuffer(rawProtected),
		Signature: newBuffer(signature),
	}
	return raw.sanitized()
}

func (obj JSONWebSignature) compactSerialize(detached bool) (string, error) {
	if len(obj.Signatures) != 1 || obj.Signatures[0].header != nil || obj.Signatures[0].protected == nil {
		return "", ErrNotSupported
	}
	serializedProtected := base64.RawURLEncoding.EncodeToString(mustSerializeJSON(obj.Signatures[0].protected))
	payload := ""
	signature := base64.RawURLEncoding.EncodeToString(obj.Signatures[0].Signature)

	if !detached {
		payload = base64.RawURLEncoding.EncodeToString(obj.payload)
	}
	return fmt.Sprintf("%s.%s.%s", serializedProtected, payload, signature), nil
}

func (obj JSONWebSignature) CompactSerialize() (string, error) {
	return obj.compactSerialize(false)
}

func (obj JSONWebSignature) DetachedCompactSerialize() (string, error) {
	return obj.compactSerialize(true)
}

func (obj JSONWebSignature) FullSerialize() string {
	raw := rawJSONWebSignature{
		Payload: newBuffer(obj.payload),
	}

	if len(obj.Signatures) == 1 {
		if obj.Signatures[0].protected != nil {
			serializedProtected := mustSerializeJSON(obj.Signatures[0].protected)
			raw.Protected = newBuffer(serializedProtected)
		}
		raw.Header = obj.Signatures[0].header
		raw.Signature = newBuffer(obj.Signatures[0].Signature)
	} else {
		raw.Signatures = make([]rawSignatureInfo, len(obj.Signatures))
		for i, signature := range obj.Signatures {
			raw.Signatures[i] = rawSignatureInfo{
				Header:    signature.header,
				Signature: newBuffer(signature.Signature),
			}

			if signature.protected != nil {
				raw.Signatures[i].Protected = newBuffer(mustSerializeJSON(signature.protected))
			}
		}
	}

	return string(mustSerializeJSON(raw))
}
