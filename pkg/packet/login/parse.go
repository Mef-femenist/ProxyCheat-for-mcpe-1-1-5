package login

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"mefproxy/pkg/jose/jwt"
)

func init() {
	
	const mojangPublicKey = `MHYwEAYHKoZIzj0CAQYFK4EEACIDYgAE8ELkixyLcwlZryUQcu1TvPOmI2B7vX83ndnWRUaXm74wFfa5f/lwQNTfrLVHa2PmenpGI6JhIMUJaWZrjmMj90NoKNFSNBuKdm8rYiXsfaz3K36x/1U26HpG0ZxK/V1V`

	data, _ := base64.StdEncoding.DecodeString(mojangPublicKey)
	publicKey, _ := x509.ParsePKIXPublicKey(data)
	mojangKey = publicKey.(*ecdsa.PublicKey)
}

var mojangKey = new(ecdsa.PublicKey)

func Parse(request []byte) (IdentityData, ClientData, AuthResult, error) {
	var (
		iData IdentityData
		cData ClientData
		res   AuthResult
		key   = &ecdsa.PublicKey{}
	)
	req, err := parseLoginRequest(request)
	if err != nil {
		return iData, cData, res, fmt.Errorf("parse login request: %w", err)
	}
	tok, err := jwt.ParseSigned(req.Chain[0])
	if err != nil {
		return iData, cData, res, fmt.Errorf("parse token 0: %w", err)
	}

	raw, _ := tok.Headers[0].ExtraHeaders["x5u"]
	if err := parseAsKey(raw, key); err != nil {
		return iData, cData, res, fmt.Errorf("parse x5u: %w", err)
	}

	var identityClaims identityClaims
	var authenticated bool
	
	switch len(req.Chain) {
	case 1:
		
		if err := parseFullClaim(req.Chain[0], key, &identityClaims); err != nil {
			return iData, cData, res, err
		}
		
	case 3:
		
		var c jwt.Claims
		if err := parseFullClaim(req.Chain[0], key, &c); err != nil {
			return iData, cData, res, fmt.Errorf("parse token 0: %w", err)
		}
		
		authenticated = bytes.Equal(key.X.Bytes(), mojangKey.X.Bytes()) && bytes.Equal(key.Y.Bytes(), mojangKey.Y.Bytes())

		if err := parseFullClaim(req.Chain[1], key, &c); err != nil {
			return iData, cData, res, fmt.Errorf("parse token 1: %w", err)
		}
		
		if err := parseFullClaim(req.Chain[2], key, &identityClaims); err != nil {
			return iData, cData, res, fmt.Errorf("parse token 2: %w", err)
		}
		
	default:
		return iData, cData, res, fmt.Errorf("unexpected login chain length %v", len(req.Chain))
	}
	if err := parseFullClaim(req.RawToken, key, &cData); err != nil {
		return iData, cData, res, fmt.Errorf("parse client data: %w", err)
	}
	return identityClaims.ExtraData, cData, AuthResult{PublicKey: key, XBOXLiveAuthenticated: authenticated}, nil
}

func parseLoginRequest(requestData []byte) (*request, error) {
	buf := bytes.NewBuffer(requestData)
	chain, err := decodeChain(buf)
	if err != nil {
		return nil, err
	}
	if len(chain) < 1 {
		return nil, fmt.Errorf("JWT chain must be at least 1 token long")
	}
	var rawLength int32
	if err := binary.Read(buf, binary.LittleEndian, &rawLength); err != nil {
		return nil, fmt.Errorf("error reading raw token length: %v", err)
	}
	return &request{Chain: chain, RawToken: string(buf.Next(int(rawLength)))}, nil
}

func parseFullClaim(claim string, key *ecdsa.PublicKey, v interface{}) error {
	tok, err := jwt.ParseSigned(claim)
	if err != nil {
		return fmt.Errorf("error parsing signed token: %w", err)
	}
	var m map[string]interface{}
	if err := tok.Claims(key, v, &m); err != nil {
		return fmt.Errorf("error verifying claims of token: %w", err)
	}
	newKey, present := m["identityPublicKey"]
	if present {
		if err := parseAsKey(newKey, key); err != nil {
			return fmt.Errorf("error parsing identity public key: %w", err)
		}
	}
	return nil
}

func parseAsKey(k interface{}, pub *ecdsa.PublicKey) error {
	kStr, _ := k.(string)
	if err := ParsePublicKey(kStr, pub); err != nil {
		return fmt.Errorf("error parsing public key: %w", err)
	}
	return nil
}

func ParsePublicKey(b64Data string, key *ecdsa.PublicKey) error {
	data, err := base64.StdEncoding.DecodeString(b64Data)
	if err != nil {
		return fmt.Errorf("error base64 decoding public key data: %v", err)
	}
	publicKey, err := x509.ParsePKIXPublicKey(data)
	if err != nil {
		return fmt.Errorf("error parsing public key: %v", err)
	}
	ecdsaKey, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("expected ECDSA public key, but got %v", key)
	}
	*key = *ecdsaKey
	return nil
}
