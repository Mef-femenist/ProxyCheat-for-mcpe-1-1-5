package jose

import (
	"crypto/elliptic"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"

	"mefproxy/pkg/jose/json"
)

type KeyAlgorithm string

type SignatureAlgorithm string

type ContentEncryption string

type CompressionAlgorithm string

type ContentType string

var (
	
	ErrCryptoFailure = errors.New("square/go-jose: error in cryptographic primitive")

	ErrUnsupportedAlgorithm = errors.New("square/go-jose: unknown/unsupported algorithm")

	ErrUnsupportedKeyType = errors.New("square/go-jose: unsupported key type/format")

	ErrInvalidKeySize = errors.New("square/go-jose: invalid key size for algorithm")

	ErrNotSupported = errors.New("square/go-jose: compact serialization not supported for object")

	ErrUnprotectedNonce = errors.New("square/go-jose: Nonce parameter included in unprotected header")
)

const (
	ED25519            = KeyAlgorithm("ED25519")
	RSA1_5             = KeyAlgorithm("RSA1_5")             
	RSA_OAEP           = KeyAlgorithm("RSA-OAEP")           
	RSA_OAEP_256       = KeyAlgorithm("RSA-OAEP-256")       
	A128KW             = KeyAlgorithm("A128KW")             
	A192KW             = KeyAlgorithm("A192KW")             
	A256KW             = KeyAlgorithm("A256KW")             
	DIRECT             = KeyAlgorithm("dir")                
	ECDH_ES            = KeyAlgorithm("ECDH-ES")            
	ECDH_ES_A128KW     = KeyAlgorithm("ECDH-ES+A128KW")     
	ECDH_ES_A192KW     = KeyAlgorithm("ECDH-ES+A192KW")     
	ECDH_ES_A256KW     = KeyAlgorithm("ECDH-ES+A256KW")     
	A128GCMKW          = KeyAlgorithm("A128GCMKW")          
	A192GCMKW          = KeyAlgorithm("A192GCMKW")          
	A256GCMKW          = KeyAlgorithm("A256GCMKW")          
	PBES2_HS256_A128KW = KeyAlgorithm("PBES2-HS256+A128KW") 
	PBES2_HS384_A192KW = KeyAlgorithm("PBES2-HS384+A192KW") 
	PBES2_HS512_A256KW = KeyAlgorithm("PBES2-HS512+A256KW") 
)

const (
	EdDSA = SignatureAlgorithm("EdDSA")
	HS256 = SignatureAlgorithm("HS256") 
	HS384 = SignatureAlgorithm("HS384") 
	HS512 = SignatureAlgorithm("HS512") 
	RS256 = SignatureAlgorithm("RS256") 
	RS384 = SignatureAlgorithm("RS384") 
	RS512 = SignatureAlgorithm("RS512") 
	ES256 = SignatureAlgorithm("ES256") 
	ES384 = SignatureAlgorithm("ES384") 
	ES512 = SignatureAlgorithm("ES512") 
	PS256 = SignatureAlgorithm("PS256") 
	PS384 = SignatureAlgorithm("PS384") 
	PS512 = SignatureAlgorithm("PS512") 
)

const (
	A128CBC_HS256 = ContentEncryption("A128CBC-HS256") 
	A192CBC_HS384 = ContentEncryption("A192CBC-HS384") 
	A256CBC_HS512 = ContentEncryption("A256CBC-HS512") 
	A128GCM       = ContentEncryption("A128GCM")       
	A192GCM       = ContentEncryption("A192GCM")       
	A256GCM       = ContentEncryption("A256GCM")       
)

const (
	NONE    = CompressionAlgorithm("")    
	DEFLATE = CompressionAlgorithm("DEF") 
)

type HeaderKey string

const (
	HeaderType        HeaderKey = "typ" 
	HeaderContentType           = "cty" 

	headerAlgorithm   = "alg"  
	headerEncryption  = "enc"  
	headerCompression = "zip"  
	headerCritical    = "crit" 

	headerAPU = "apu" 
	headerAPV = "apv" 
	headerEPK = "epk" 
	headerIV  = "iv"  
	headerTag = "tag" 
	headerX5c = "x5c" 

	headerJWK   = "jwk"   
	headerKeyID = "kid"   
	headerNonce = "nonce" 
	headerB64   = "b64"   

	headerP2C = "p2c" 
	headerP2S = "p2s" 

)

var supportedCritical = map[string]bool{
	headerB64: true,
}

type rawHeader map[HeaderKey]*json.RawMessage

type Header struct {
	KeyID      string
	JSONWebKey *JSONWebKey
	Algorithm  string
	Nonce      string

	certificates []*x509.Certificate

	ExtraHeaders map[HeaderKey]interface{}
}

func (h Header) Certificates(opts x509.VerifyOptions) ([][]*x509.Certificate, error) {
	if len(h.certificates) == 0 {
		return nil, errors.New("square/go-jose: no x5c header present in message")
	}

	leaf := h.certificates[0]
	if opts.Intermediates == nil {
		opts.Intermediates = x509.NewCertPool()
		for _, intermediate := range h.certificates[1:] {
			opts.Intermediates.AddCert(intermediate)
		}
	}

	return leaf.Verify(opts)
}

func (parsed rawHeader) set(k HeaderKey, v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	parsed[k] = makeRawMessage(b)
	return nil
}

func (parsed rawHeader) getString(k HeaderKey) string {
	v, ok := parsed[k]
	if !ok || v == nil {
		return ""
	}
	var s string
	err := json.Unmarshal(*v, &s)
	if err != nil {
		return ""
	}
	return s
}

func (parsed rawHeader) getByteBuffer(k HeaderKey) (*byteBuffer, error) {
	v := parsed[k]
	if v == nil {
		return nil, nil
	}
	var bb *byteBuffer
	err := json.Unmarshal(*v, &bb)
	if err != nil {
		return nil, err
	}
	return bb, nil
}

func (parsed rawHeader) getAlgorithm() KeyAlgorithm {
	return KeyAlgorithm(parsed.getString(headerAlgorithm))
}

func (parsed rawHeader) getSignatureAlgorithm() SignatureAlgorithm {
	return SignatureAlgorithm(parsed.getString(headerAlgorithm))
}

func (parsed rawHeader) getEncryption() ContentEncryption {
	return ContentEncryption(parsed.getString(headerEncryption))
}

func (parsed rawHeader) getCompression() CompressionAlgorithm {
	return CompressionAlgorithm(parsed.getString(headerCompression))
}

func (parsed rawHeader) getNonce() string {
	return parsed.getString(headerNonce)
}

func (parsed rawHeader) getEPK() (*JSONWebKey, error) {
	v := parsed[headerEPK]
	if v == nil {
		return nil, nil
	}
	var epk *JSONWebKey
	err := json.Unmarshal(*v, &epk)
	if err != nil {
		return nil, err
	}
	return epk, nil
}

func (parsed rawHeader) getAPU() (*byteBuffer, error) {
	return parsed.getByteBuffer(headerAPU)
}

func (parsed rawHeader) getAPV() (*byteBuffer, error) {
	return parsed.getByteBuffer(headerAPV)
}

func (parsed rawHeader) getIV() (*byteBuffer, error) {
	return parsed.getByteBuffer(headerIV)
}

func (parsed rawHeader) getTag() (*byteBuffer, error) {
	return parsed.getByteBuffer(headerTag)
}

func (parsed rawHeader) getJWK() (*JSONWebKey, error) {
	v := parsed[headerJWK]
	if v == nil {
		return nil, nil
	}
	var jwk *JSONWebKey
	err := json.Unmarshal(*v, &jwk)
	if err != nil {
		return nil, err
	}
	return jwk, nil
}

func (parsed rawHeader) getCritical() ([]string, error) {
	v := parsed[headerCritical]
	if v == nil {
		return nil, nil
	}

	var q []string
	err := json.Unmarshal(*v, &q)
	if err != nil {
		return nil, err
	}
	return q, nil
}

func (parsed rawHeader) getP2C() (int, error) {
	v := parsed[headerP2C]
	if v == nil {
		return 0, nil
	}

	var p2c int
	err := json.Unmarshal(*v, &p2c)
	if err != nil {
		return 0, err
	}
	return p2c, nil
}

func (parsed rawHeader) getP2S() (*byteBuffer, error) {
	return parsed.getByteBuffer(headerP2S)
}

func (parsed rawHeader) getB64() (bool, error) {
	v := parsed[headerB64]
	if v == nil {
		return true, nil
	}

	var b64 bool
	err := json.Unmarshal(*v, &b64)
	if err != nil {
		return true, err
	}
	return b64, nil
}

func (parsed rawHeader) sanitized() (h Header, err error) {
	for k, v := range parsed {
		if v == nil {
			continue
		}
		switch k {
		case headerJWK:
			var jwk *JSONWebKey
			err = json.Unmarshal(*v, &jwk)
			if err != nil {
				err = fmt.Errorf("failed to unmarshal JWK: %v: %#v", err, string(*v))
				return
			}
			h.JSONWebKey = jwk
		case headerKeyID:
			var s string
			err = json.Unmarshal(*v, &s)
			if err != nil {
				err = fmt.Errorf("failed to unmarshal key ID: %v: %#v", err, string(*v))
				return
			}
			h.KeyID = s
		case headerAlgorithm:
			var s string
			err = json.Unmarshal(*v, &s)
			if err != nil {
				err = fmt.Errorf("failed to unmarshal algorithm: %v: %#v", err, string(*v))
				return
			}
			h.Algorithm = s
		case headerNonce:
			var s string
			err = json.Unmarshal(*v, &s)
			if err != nil {
				err = fmt.Errorf("failed to unmarshal nonce: %v: %#v", err, string(*v))
				return
			}
			h.Nonce = s
		case headerX5c:
			c := []string{}
			err = json.Unmarshal(*v, &c)
			if err != nil {
				err = fmt.Errorf("failed to unmarshal x5c header: %v: %#v", err, string(*v))
				return
			}
			h.certificates, err = parseCertificateChain(c)
			if err != nil {
				err = fmt.Errorf("failed to unmarshal x5c header: %v: %#v", err, string(*v))
				return
			}
		default:
			if h.ExtraHeaders == nil {
				h.ExtraHeaders = map[HeaderKey]interface{}{}
			}
			var v2 interface{}
			err = json.Unmarshal(*v, &v2)
			if err != nil {
				err = fmt.Errorf("failed to unmarshal value: %v: %#v", err, string(*v))
				return
			}
			h.ExtraHeaders[k] = v2
		}
	}
	return
}

func parseCertificateChain(chain []string) ([]*x509.Certificate, error) {
	out := make([]*x509.Certificate, len(chain))
	for i, cert := range chain {
		raw, err := base64.StdEncoding.DecodeString(cert)
		if err != nil {
			return nil, err
		}
		out[i], err = x509.ParseCertificate(raw)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (dst rawHeader) isSet(k HeaderKey) bool {
	dvr := dst[k]
	if dvr == nil {
		return false
	}

	var dv interface{}
	err := json.Unmarshal(*dvr, &dv)
	if err != nil {
		return true
	}

	if dvStr, ok := dv.(string); ok {
		return dvStr != ""
	}

	return true
}

func (dst rawHeader) merge(src *rawHeader) {
	if src == nil {
		return
	}

	for k, v := range *src {
		if dst.isSet(k) {
			continue
		}

		dst[k] = v
	}
}

func curveName(crv elliptic.Curve) (string, error) {
	switch crv {
	case elliptic.P256():
		return "P-256", nil
	case elliptic.P384():
		return "P-384", nil
	case elliptic.P521():
		return "P-521", nil
	default:
		return "", fmt.Errorf("square/go-jose: unsupported/unknown elliptic curve")
	}
}

func curveSize(crv elliptic.Curve) int {
	bits := crv.Params().BitSize

	div := bits / 8
	mod := bits % 8

	if mod == 0 {
		return div
	}

	return div + 1
}

func makeRawMessage(b []byte) *json.RawMessage {
	rm := json.RawMessage(b)
	return &rm
}
