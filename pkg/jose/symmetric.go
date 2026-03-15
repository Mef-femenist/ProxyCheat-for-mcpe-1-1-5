package jose

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/subtle"
	"errors"
	"fmt"
	"hash"
	"io"

	"golang.org/x/crypto/pbkdf2"
	"mefproxy/pkg/jose/cipher"
)

var RandReader = rand.Reader

const (
	
	defaultP2C = 100000
	
	defaultP2SSize = 16
)

type symmetricKeyCipher struct {
	key []byte 
	p2c int    
	p2s []byte 
}

type symmetricMac struct {
	key []byte
}

type aeadParts struct {
	iv, ciphertext, tag []byte
}

type aeadContentCipher struct {
	keyBytes     int
	authtagBytes int
	getAead      func(key []byte) (cipher.AEAD, error)
}

type randomKeyGenerator struct {
	size int
}

type staticKeyGenerator struct {
	key []byte
}

func newAESGCM(keySize int) contentCipher {
	return &aeadContentCipher{
		keyBytes:     keySize,
		authtagBytes: 16,
		getAead: func(key []byte) (cipher.AEAD, error) {
			aes, err := aes.NewCipher(key)
			if err != nil {
				return nil, err
			}

			return cipher.NewGCM(aes)
		},
	}
}

func newAESCBC(keySize int) contentCipher {
	return &aeadContentCipher{
		keyBytes:     keySize * 2,
		authtagBytes: keySize,
		getAead: func(key []byte) (cipher.AEAD, error) {
			return josecipher.NewCBCHMAC(key, aes.NewCipher)
		},
	}
}

func getContentCipher(alg ContentEncryption) contentCipher {
	switch alg {
	case A128GCM:
		return newAESGCM(16)
	case A192GCM:
		return newAESGCM(24)
	case A256GCM:
		return newAESGCM(32)
	case A128CBC_HS256:
		return newAESCBC(16)
	case A192CBC_HS384:
		return newAESCBC(24)
	case A256CBC_HS512:
		return newAESCBC(32)
	default:
		return nil
	}
}

func getPbkdf2Params(alg KeyAlgorithm) (int, func() hash.Hash) {
	switch alg {
	case PBES2_HS256_A128KW:
		return 16, sha256.New
	case PBES2_HS384_A192KW:
		return 24, sha512.New384
	case PBES2_HS512_A256KW:
		return 32, sha512.New
	default:
		panic("invalid algorithm")
	}
}

func getRandomSalt(size int) ([]byte, error) {
	salt := make([]byte, size)
	_, err := io.ReadFull(RandReader, salt)
	if err != nil {
		return nil, err
	}

	return salt, nil
}

func newSymmetricRecipient(keyAlg KeyAlgorithm, key []byte) (recipientKeyInfo, error) {
	switch keyAlg {
	case DIRECT, A128GCMKW, A192GCMKW, A256GCMKW, A128KW, A192KW, A256KW:
	case PBES2_HS256_A128KW, PBES2_HS384_A192KW, PBES2_HS512_A256KW:
	default:
		return recipientKeyInfo{}, ErrUnsupportedAlgorithm
	}

	return recipientKeyInfo{
		keyAlg: keyAlg,
		keyEncrypter: &symmetricKeyCipher{
			key: key,
		},
	}, nil
}

func newSymmetricSigner(sigAlg SignatureAlgorithm, key []byte) (recipientSigInfo, error) {
	
	switch sigAlg {
	case HS256, HS384, HS512:
	default:
		return recipientSigInfo{}, ErrUnsupportedAlgorithm
	}

	return recipientSigInfo{
		sigAlg: sigAlg,
		signer: &symmetricMac{
			key: key,
		},
	}, nil
}

func (ctx randomKeyGenerator) genKey() ([]byte, rawHeader, error) {
	key := make([]byte, ctx.size)
	_, err := io.ReadFull(RandReader, key)
	if err != nil {
		return nil, rawHeader{}, err
	}

	return key, rawHeader{}, nil
}

func (ctx randomKeyGenerator) keySize() int {
	return ctx.size
}

func (ctx staticKeyGenerator) genKey() ([]byte, rawHeader, error) {
	cek := make([]byte, len(ctx.key))
	copy(cek, ctx.key)
	return cek, rawHeader{}, nil
}

func (ctx staticKeyGenerator) keySize() int {
	return len(ctx.key)
}

func (ctx aeadContentCipher) keySize() int {
	return ctx.keyBytes
}

func (ctx aeadContentCipher) encrypt(key, aad, pt []byte) (*aeadParts, error) {
	
	aead, err := ctx.getAead(key)
	if err != nil {
		return nil, err
	}

	iv := make([]byte, aead.NonceSize())
	_, err = io.ReadFull(RandReader, iv)
	if err != nil {
		return nil, err
	}

	ciphertextAndTag := aead.Seal(nil, iv, pt, aad)
	offset := len(ciphertextAndTag) - ctx.authtagBytes

	return &aeadParts{
		iv:         iv,
		ciphertext: ciphertextAndTag[:offset],
		tag:        ciphertextAndTag[offset:],
	}, nil
}

func (ctx aeadContentCipher) decrypt(key, aad []byte, parts *aeadParts) ([]byte, error) {
	aead, err := ctx.getAead(key)
	if err != nil {
		return nil, err
	}

	if len(parts.iv) != aead.NonceSize() || len(parts.tag) < ctx.authtagBytes {
		return nil, ErrCryptoFailure
	}

	return aead.Open(nil, parts.iv, append(parts.ciphertext, parts.tag...), aad)
}

func (ctx *symmetricKeyCipher) encryptKey(cek []byte, alg KeyAlgorithm) (recipientInfo, error) {
	switch alg {
	case DIRECT:
		return recipientInfo{
			header: &rawHeader{},
		}, nil
	case A128GCMKW, A192GCMKW, A256GCMKW:
		aead := newAESGCM(len(ctx.key))

		parts, err := aead.encrypt(ctx.key, []byte{}, cek)
		if err != nil {
			return recipientInfo{}, err
		}

		header := &rawHeader{}
		header.set(headerIV, newBuffer(parts.iv))
		header.set(headerTag, newBuffer(parts.tag))

		return recipientInfo{
			header:       header,
			encryptedKey: parts.ciphertext,
		}, nil
	case A128KW, A192KW, A256KW:
		block, err := aes.NewCipher(ctx.key)
		if err != nil {
			return recipientInfo{}, err
		}

		jek, err := josecipher.KeyWrap(block, cek)
		if err != nil {
			return recipientInfo{}, err
		}

		return recipientInfo{
			encryptedKey: jek,
			header:       &rawHeader{},
		}, nil
	case PBES2_HS256_A128KW, PBES2_HS384_A192KW, PBES2_HS512_A256KW:
		if len(ctx.p2s) == 0 {
			salt, err := getRandomSalt(defaultP2SSize)
			if err != nil {
				return recipientInfo{}, err
			}
			ctx.p2s = salt
		}

		if ctx.p2c <= 0 {
			ctx.p2c = defaultP2C
		}

		salt := bytes.Join([][]byte{[]byte(alg), ctx.p2s}, []byte{0x00})

		keyLen, h := getPbkdf2Params(alg)
		key := pbkdf2.Key(ctx.key, salt, ctx.p2c, keyLen, h)

		block, err := aes.NewCipher(key)
		if err != nil {
			return recipientInfo{}, err
		}

		jek, err := josecipher.KeyWrap(block, cek)
		if err != nil {
			return recipientInfo{}, err
		}

		header := &rawHeader{}
		header.set(headerP2C, ctx.p2c)
		header.set(headerP2S, newBuffer(ctx.p2s))

		return recipientInfo{
			encryptedKey: jek,
			header:       header,
		}, nil
	}

	return recipientInfo{}, ErrUnsupportedAlgorithm
}

func (ctx *symmetricKeyCipher) decryptKey(headers rawHeader, recipient *recipientInfo, generator keyGenerator) ([]byte, error) {
	switch headers.getAlgorithm() {
	case DIRECT:
		cek := make([]byte, len(ctx.key))
		copy(cek, ctx.key)
		return cek, nil
	case A128GCMKW, A192GCMKW, A256GCMKW:
		aead := newAESGCM(len(ctx.key))

		iv, err := headers.getIV()
		if err != nil {
			return nil, fmt.Errorf("square/go-jose: invalid IV: %v", err)
		}
		tag, err := headers.getTag()
		if err != nil {
			return nil, fmt.Errorf("square/go-jose: invalid tag: %v", err)
		}

		parts := &aeadParts{
			iv:         iv.bytes(),
			ciphertext: recipient.encryptedKey,
			tag:        tag.bytes(),
		}

		cek, err := aead.decrypt(ctx.key, []byte{}, parts)
		if err != nil {
			return nil, err
		}

		return cek, nil
	case A128KW, A192KW, A256KW:
		block, err := aes.NewCipher(ctx.key)
		if err != nil {
			return nil, err
		}

		cek, err := josecipher.KeyUnwrap(block, recipient.encryptedKey)
		if err != nil {
			return nil, err
		}
		return cek, nil
	case PBES2_HS256_A128KW, PBES2_HS384_A192KW, PBES2_HS512_A256KW:
		p2s, err := headers.getP2S()
		if err != nil {
			return nil, fmt.Errorf("square/go-jose: invalid P2S: %v", err)
		}
		if p2s == nil || len(p2s.data) == 0 {
			return nil, fmt.Errorf("square/go-jose: invalid P2S: must be present")
		}

		p2c, err := headers.getP2C()
		if err != nil {
			return nil, fmt.Errorf("square/go-jose: invalid P2C: %v", err)
		}
		if p2c <= 0 {
			return nil, fmt.Errorf("square/go-jose: invalid P2C: must be a positive integer")
		}

		alg := headers.getAlgorithm()
		salt := bytes.Join([][]byte{[]byte(alg), p2s.bytes()}, []byte{0x00})

		keyLen, h := getPbkdf2Params(alg)
		key := pbkdf2.Key(ctx.key, salt, p2c, keyLen, h)

		block, err := aes.NewCipher(key)
		if err != nil {
			return nil, err
		}

		cek, err := josecipher.KeyUnwrap(block, recipient.encryptedKey)
		if err != nil {
			return nil, err
		}
		return cek, nil
	}

	return nil, ErrUnsupportedAlgorithm
}

func (ctx symmetricMac) signPayload(payload []byte, alg SignatureAlgorithm) (Signature, error) {
	mac, err := ctx.hmac(payload, alg)
	if err != nil {
		return Signature{}, errors.New("square/go-jose: failed to compute hmac")
	}

	return Signature{
		Signature: mac,
		protected: &rawHeader{},
	}, nil
}

func (ctx symmetricMac) verifyPayload(payload []byte, mac []byte, alg SignatureAlgorithm) error {
	expected, err := ctx.hmac(payload, alg)
	if err != nil {
		return errors.New("square/go-jose: failed to compute hmac")
	}

	if len(mac) != len(expected) {
		return errors.New("square/go-jose: invalid hmac")
	}

	match := subtle.ConstantTimeCompare(mac, expected)
	if match != 1 {
		return errors.New("square/go-jose: invalid hmac")
	}

	return nil
}

func (ctx symmetricMac) hmac(payload []byte, alg SignatureAlgorithm) ([]byte, error) {
	var hash func() hash.Hash

	switch alg {
	case HS256:
		hash = sha256.New
	case HS384:
		hash = sha512.New384
	case HS512:
		hash = sha512.New
	default:
		return nil, ErrUnsupportedAlgorithm
	}

	hmac := hmac.New(hash, ctx.key)

	_, _ = hmac.Write(payload)
	return hmac.Sum(nil), nil
}
