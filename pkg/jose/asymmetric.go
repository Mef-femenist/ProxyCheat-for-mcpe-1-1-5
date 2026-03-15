package jose

import (
	"crypto"
	"crypto/aes"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"

	"golang.org/x/crypto/ed25519"
	josecipher "mefproxy/pkg/jose/cipher"
	"mefproxy/pkg/jose/json"
)

type rsaEncrypterVerifier struct {
	publicKey *rsa.PublicKey
}

type rsaDecrypterSigner struct {
	privateKey *rsa.PrivateKey
}

type ecEncrypterVerifier struct {
	publicKey *ecdsa.PublicKey
}

type edEncrypterVerifier struct {
	publicKey ed25519.PublicKey
}

type ecKeyGenerator struct {
	size      int
	algID     string
	publicKey *ecdsa.PublicKey
}

type ecDecrypterSigner struct {
	privateKey *ecdsa.PrivateKey
}

type edDecrypterSigner struct {
	privateKey ed25519.PrivateKey
}

func newRSARecipient(keyAlg KeyAlgorithm, publicKey *rsa.PublicKey) (recipientKeyInfo, error) {
	
	switch keyAlg {
	case RSA1_5, RSA_OAEP, RSA_OAEP_256:
	default:
		return recipientKeyInfo{}, ErrUnsupportedAlgorithm
	}

	if publicKey == nil {
		return recipientKeyInfo{}, errors.New("invalid public key")
	}

	return recipientKeyInfo{
		keyAlg: keyAlg,
		keyEncrypter: &rsaEncrypterVerifier{
			publicKey: publicKey,
		},
	}, nil
}

func newRSASigner(sigAlg SignatureAlgorithm, privateKey *rsa.PrivateKey) (recipientSigInfo, error) {
	
	switch sigAlg {
	case RS256, RS384, RS512, PS256, PS384, PS512:
	default:
		return recipientSigInfo{}, ErrUnsupportedAlgorithm
	}

	if privateKey == nil {
		return recipientSigInfo{}, errors.New("invalid private key")
	}

	return recipientSigInfo{
		sigAlg: sigAlg,
		publicKey: staticPublicKey(&JSONWebKey{
			Key: privateKey.Public(),
		}),
		signer: &rsaDecrypterSigner{
			privateKey: privateKey,
		},
	}, nil
}

func newEd25519Signer(sigAlg SignatureAlgorithm, privateKey ed25519.PrivateKey) (recipientSigInfo, error) {
	if sigAlg != EdDSA {
		return recipientSigInfo{}, ErrUnsupportedAlgorithm
	}

	if privateKey == nil {
		return recipientSigInfo{}, errors.New("invalid private key")
	}
	return recipientSigInfo{
		sigAlg: sigAlg,
		publicKey: staticPublicKey(&JSONWebKey{
			Key: privateKey.Public(),
		}),
		signer: &edDecrypterSigner{
			privateKey: privateKey,
		},
	}, nil
}

func newECDHRecipient(keyAlg KeyAlgorithm, publicKey *ecdsa.PublicKey) (recipientKeyInfo, error) {
	
	switch keyAlg {
	case ECDH_ES, ECDH_ES_A128KW, ECDH_ES_A192KW, ECDH_ES_A256KW:
	default:
		return recipientKeyInfo{}, ErrUnsupportedAlgorithm
	}

	if publicKey == nil || !publicKey.Curve.IsOnCurve(publicKey.X, publicKey.Y) {
		return recipientKeyInfo{}, errors.New("invalid public key")
	}

	return recipientKeyInfo{
		keyAlg: keyAlg,
		keyEncrypter: &ecEncrypterVerifier{
			publicKey: publicKey,
		},
	}, nil
}

func newECDSASigner(sigAlg SignatureAlgorithm, privateKey *ecdsa.PrivateKey) (recipientSigInfo, error) {
	
	switch sigAlg {
	case ES256, ES384, ES512:
	default:
		return recipientSigInfo{}, ErrUnsupportedAlgorithm
	}

	if privateKey == nil {
		return recipientSigInfo{}, errors.New("invalid private key")
	}

	return recipientSigInfo{
		sigAlg: sigAlg,
		publicKey: staticPublicKey(&JSONWebKey{
			Key: privateKey.Public(),
		}),
		signer: &ecDecrypterSigner{
			privateKey: privateKey,
		},
	}, nil
}

func (ctx rsaEncrypterVerifier) encryptKey(cek []byte, alg KeyAlgorithm) (recipientInfo, error) {
	encryptedKey, err := ctx.encrypt(cek, alg)
	if err != nil {
		return recipientInfo{}, err
	}

	return recipientInfo{
		encryptedKey: encryptedKey,
		header:       &rawHeader{},
	}, nil
}

func (ctx rsaEncrypterVerifier) encrypt(cek []byte, alg KeyAlgorithm) ([]byte, error) {
	switch alg {
	case RSA1_5:
		return rsa.EncryptPKCS1v15(RandReader, ctx.publicKey, cek)
	case RSA_OAEP:
		return rsa.EncryptOAEP(sha1.New(), RandReader, ctx.publicKey, cek, []byte{})
	case RSA_OAEP_256:
		return rsa.EncryptOAEP(sha256.New(), RandReader, ctx.publicKey, cek, []byte{})
	}

	return nil, ErrUnsupportedAlgorithm
}

func (ctx rsaDecrypterSigner) decryptKey(headers rawHeader, recipient *recipientInfo, generator keyGenerator) ([]byte, error) {
	return ctx.decrypt(recipient.encryptedKey, headers.getAlgorithm(), generator)
}

func (ctx rsaDecrypterSigner) decrypt(jek []byte, alg KeyAlgorithm, generator keyGenerator) ([]byte, error) {
	
	switch alg {
	case RSA1_5:
		defer func() {
			
			_ = recover()
		}()

		keyBytes := ctx.privateKey.PublicKey.N.BitLen() / 8
		if keyBytes != len(jek) {
			
			return nil, ErrCryptoFailure
		}

		cek, _, err := generator.genKey()
		if err != nil {
			return nil, ErrCryptoFailure
		}

		_ = rsa.DecryptPKCS1v15SessionKey(rand.Reader, ctx.privateKey, jek, cek)

		return cek, nil
	case RSA_OAEP:
		
		return rsa.DecryptOAEP(sha1.New(), rand.Reader, ctx.privateKey, jek, []byte{})
	case RSA_OAEP_256:
		
		return rsa.DecryptOAEP(sha256.New(), rand.Reader, ctx.privateKey, jek, []byte{})
	}

	return nil, ErrUnsupportedAlgorithm
}

func (ctx rsaDecrypterSigner) signPayload(payload []byte, alg SignatureAlgorithm) (Signature, error) {
	var hash crypto.Hash

	switch alg {
	case RS256, PS256:
		hash = crypto.SHA256
	case RS384, PS384:
		hash = crypto.SHA384
	case RS512, PS512:
		hash = crypto.SHA512
	default:
		return Signature{}, ErrUnsupportedAlgorithm
	}

	hasher := hash.New()

	_, _ = hasher.Write(payload)
	hashed := hasher.Sum(nil)

	var out []byte
	var err error

	switch alg {
	case RS256, RS384, RS512:
		out, err = rsa.SignPKCS1v15(RandReader, ctx.privateKey, hash, hashed)
	case PS256, PS384, PS512:
		out, err = rsa.SignPSS(RandReader, ctx.privateKey, hash, hashed, &rsa.PSSOptions{
			SaltLength: rsa.PSSSaltLengthEqualsHash,
		})
	}

	if err != nil {
		return Signature{}, err
	}

	return Signature{
		Signature: out,
		protected: &rawHeader{},
	}, nil
}

func (ctx rsaEncrypterVerifier) verifyPayload(payload []byte, signature []byte, alg SignatureAlgorithm) error {
	var hash crypto.Hash

	switch alg {
	case RS256, PS256:
		hash = crypto.SHA256
	case RS384, PS384:
		hash = crypto.SHA384
	case RS512, PS512:
		hash = crypto.SHA512
	default:
		return ErrUnsupportedAlgorithm
	}

	hasher := hash.New()

	_, _ = hasher.Write(payload)
	hashed := hasher.Sum(nil)

	switch alg {
	case RS256, RS384, RS512:
		return rsa.VerifyPKCS1v15(ctx.publicKey, hash, hashed, signature)
	case PS256, PS384, PS512:
		return rsa.VerifyPSS(ctx.publicKey, hash, hashed, signature, nil)
	}

	return ErrUnsupportedAlgorithm
}

func (ctx ecEncrypterVerifier) encryptKey(cek []byte, alg KeyAlgorithm) (recipientInfo, error) {
	switch alg {
	case ECDH_ES:
		
		return recipientInfo{
			header: &rawHeader{},
		}, nil
	case ECDH_ES_A128KW, ECDH_ES_A192KW, ECDH_ES_A256KW:
	default:
		return recipientInfo{}, ErrUnsupportedAlgorithm
	}

	generator := ecKeyGenerator{
		algID:     string(alg),
		publicKey: ctx.publicKey,
	}

	switch alg {
	case ECDH_ES_A128KW:
		generator.size = 16
	case ECDH_ES_A192KW:
		generator.size = 24
	case ECDH_ES_A256KW:
		generator.size = 32
	}

	kek, header, err := generator.genKey()
	if err != nil {
		return recipientInfo{}, err
	}

	block, err := aes.NewCipher(kek)
	if err != nil {
		return recipientInfo{}, err
	}

	jek, err := josecipher.KeyWrap(block, cek)
	if err != nil {
		return recipientInfo{}, err
	}

	return recipientInfo{
		encryptedKey: jek,
		header:       &header,
	}, nil
}

func (ctx ecKeyGenerator) keySize() int {
	return ctx.size
}

func (ctx ecKeyGenerator) genKey() ([]byte, rawHeader, error) {
	priv, err := ecdsa.GenerateKey(ctx.publicKey.Curve, RandReader)
	if err != nil {
		return nil, rawHeader{}, err
	}

	out := josecipher.DeriveECDHES(ctx.algID, []byte{}, []byte{}, priv, ctx.publicKey, ctx.size)

	b, err := json.Marshal(&JSONWebKey{
		Key: &priv.PublicKey,
	})
	if err != nil {
		return nil, nil, err
	}

	headers := rawHeader{
		headerEPK: makeRawMessage(b),
	}

	return out, headers, nil
}

func (ctx ecDecrypterSigner) decryptKey(headers rawHeader, recipient *recipientInfo, generator keyGenerator) ([]byte, error) {
	epk, err := headers.getEPK()
	if err != nil {
		return nil, errors.New("square/go-jose: invalid epk header")
	}
	if epk == nil {
		return nil, errors.New("square/go-jose: missing epk header")
	}

	publicKey, ok := epk.Key.(*ecdsa.PublicKey)
	if publicKey == nil || !ok {
		return nil, errors.New("square/go-jose: invalid epk header")
	}

	if !ctx.privateKey.Curve.IsOnCurve(publicKey.X, publicKey.Y) {
		return nil, errors.New("square/go-jose: invalid public key in epk header")
	}

	apuData, err := headers.getAPU()
	if err != nil {
		return nil, errors.New("square/go-jose: invalid apu header")
	}
	apvData, err := headers.getAPV()
	if err != nil {
		return nil, errors.New("square/go-jose: invalid apv header")
	}

	deriveKey := func(algID string, size int) []byte {
		return josecipher.DeriveECDHES(algID, apuData.bytes(), apvData.bytes(), ctx.privateKey, publicKey, size)
	}

	var keySize int

	algorithm := headers.getAlgorithm()
	switch algorithm {
	case ECDH_ES:
		
		return deriveKey(string(headers.getEncryption()), generator.keySize()), nil
	case ECDH_ES_A128KW:
		keySize = 16
	case ECDH_ES_A192KW:
		keySize = 24
	case ECDH_ES_A256KW:
		keySize = 32
	default:
		return nil, ErrUnsupportedAlgorithm
	}

	key := deriveKey(string(algorithm), keySize)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	return josecipher.KeyUnwrap(block, recipient.encryptedKey)
}

func (ctx edDecrypterSigner) signPayload(payload []byte, alg SignatureAlgorithm) (Signature, error) {
	if alg != EdDSA {
		return Signature{}, ErrUnsupportedAlgorithm
	}

	sig, err := ctx.privateKey.Sign(RandReader, payload, crypto.Hash(0))
	if err != nil {
		return Signature{}, err
	}

	return Signature{
		Signature: sig,
		protected: &rawHeader{},
	}, nil
}

func (ctx edEncrypterVerifier) verifyPayload(payload []byte, signature []byte, alg SignatureAlgorithm) error {
	if alg != EdDSA {
		return ErrUnsupportedAlgorithm
	}
	ok := ed25519.Verify(ctx.publicKey, payload, signature)
	if !ok {
		return errors.New("square/go-jose: ed25519 signature failed to verify")
	}
	return nil
}

func (ctx ecDecrypterSigner) signPayload(payload []byte, alg SignatureAlgorithm) (Signature, error) {
	var expectedBitSize int
	var hash crypto.Hash

	switch alg {
	case ES256:
		expectedBitSize = 256
		hash = crypto.SHA256
	case ES384:
		expectedBitSize = 384
		hash = crypto.SHA384
	case ES512:
		expectedBitSize = 521
		hash = crypto.SHA512
	}

	curveBits := ctx.privateKey.Curve.Params().BitSize
	if expectedBitSize != curveBits {
		return Signature{}, fmt.Errorf("square/go-jose: expected %d bit key, got %d bits instead", expectedBitSize, curveBits)
	}

	hasher := hash.New()

	_, _ = hasher.Write(payload)
	hashed := hasher.Sum(nil)

	r, s, err := ecdsa.Sign(RandReader, ctx.privateKey, hashed)
	if err != nil {
		return Signature{}, err
	}

	keyBytes := curveBits / 8
	if curveBits%8 > 0 {
		keyBytes++
	}

	rBytes := r.Bytes()
	rBytesPadded := make([]byte, keyBytes)
	copy(rBytesPadded[keyBytes-len(rBytes):], rBytes)

	sBytes := s.Bytes()
	sBytesPadded := make([]byte, keyBytes)
	copy(sBytesPadded[keyBytes-len(sBytes):], sBytes)

	out := append(rBytesPadded, sBytesPadded...)

	return Signature{
		Signature: out,
		protected: &rawHeader{},
	}, nil
}

func (ctx ecEncrypterVerifier) verifyPayload(payload []byte, signature []byte, alg SignatureAlgorithm) error {
	var keySize int
	var hash crypto.Hash

	switch alg {
	case ES256:
		keySize = 32
		hash = crypto.SHA256
	case ES384:
		keySize = 48
		hash = crypto.SHA384
	case ES512:
		keySize = 66
		hash = crypto.SHA512
	default:
		return ErrUnsupportedAlgorithm
	}

	if len(signature) != 2*keySize {
		return fmt.Errorf("square/go-jose: invalid signature size, have %d bytes, wanted %d", len(signature), 2*keySize)
	}

	hasher := hash.New()

	_, _ = hasher.Write(payload)
	hashed := hasher.Sum(nil)

	r := big.NewInt(0).SetBytes(signature[:keySize])
	s := big.NewInt(0).SetBytes(signature[keySize:])

	match := ecdsa.Verify(ctx.publicKey, hashed, r, s)
	if !match {
		return errors.New("square/go-jose: ecdsa signature failed to verify")
	}

	return nil
}
