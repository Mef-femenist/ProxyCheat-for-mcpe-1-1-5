package josecipher

import (
	"bytes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/binary"
	"errors"
	"hash"
)

const (
	nonceBytes = 16
)

func NewCBCHMAC(key []byte, newBlockCipher func([]byte) (cipher.Block, error)) (cipher.AEAD, error) {
	keySize := len(key) / 2
	integrityKey := key[:keySize]
	encryptionKey := key[keySize:]

	blockCipher, err := newBlockCipher(encryptionKey)
	if err != nil {
		return nil, err
	}

	var hash func() hash.Hash
	switch keySize {
	case 16:
		hash = sha256.New
	case 24:
		hash = sha512.New384
	case 32:
		hash = sha512.New
	}

	return &cbcAEAD{
		hash:         hash,
		blockCipher:  blockCipher,
		authtagBytes: keySize,
		integrityKey: integrityKey,
	}, nil
}

type cbcAEAD struct {
	hash         func() hash.Hash
	authtagBytes int
	integrityKey []byte
	blockCipher  cipher.Block
}

func (ctx *cbcAEAD) NonceSize() int {
	return nonceBytes
}

func (ctx *cbcAEAD) Overhead() int {
	
	return ctx.blockCipher.BlockSize() + ctx.authtagBytes
}

func (ctx *cbcAEAD) Seal(dst, nonce, plaintext, data []byte) []byte {
	
	ciphertext := make([]byte, uint64(len(plaintext))+uint64(ctx.Overhead()))[:len(plaintext)]
	copy(ciphertext, plaintext)
	ciphertext = padBuffer(ciphertext, ctx.blockCipher.BlockSize())

	cbc := cipher.NewCBCEncrypter(ctx.blockCipher, nonce)

	cbc.CryptBlocks(ciphertext, ciphertext)
	authtag := ctx.computeAuthTag(data, nonce, ciphertext)

	ret, out := resize(dst, uint64(len(dst))+uint64(len(ciphertext))+uint64(len(authtag)))
	copy(out, ciphertext)
	copy(out[len(ciphertext):], authtag)

	return ret
}

func (ctx *cbcAEAD) Open(dst, nonce, ciphertext, data []byte) ([]byte, error) {
	if len(ciphertext) < ctx.authtagBytes {
		return nil, errors.New("square/go-jose: invalid ciphertext (too short)")
	}

	offset := len(ciphertext) - ctx.authtagBytes
	expectedTag := ctx.computeAuthTag(data, nonce, ciphertext[:offset])
	match := subtle.ConstantTimeCompare(expectedTag, ciphertext[offset:])
	if match != 1 {
		return nil, errors.New("square/go-jose: invalid ciphertext (auth tag mismatch)")
	}

	cbc := cipher.NewCBCDecrypter(ctx.blockCipher, nonce)

	buffer := append([]byte{}, []byte(ciphertext[:offset])...)

	if len(buffer)%ctx.blockCipher.BlockSize() > 0 {
		return nil, errors.New("square/go-jose: invalid ciphertext (invalid length)")
	}

	cbc.CryptBlocks(buffer, buffer)

	plaintext, err := unpadBuffer(buffer, ctx.blockCipher.BlockSize())
	if err != nil {
		return nil, err
	}

	ret, out := resize(dst, uint64(len(dst))+uint64(len(plaintext)))
	copy(out, plaintext)

	return ret, nil
}

func (ctx *cbcAEAD) computeAuthTag(aad, nonce, ciphertext []byte) []byte {
	buffer := make([]byte, uint64(len(aad))+uint64(len(nonce))+uint64(len(ciphertext))+8)
	n := 0
	n += copy(buffer, aad)
	n += copy(buffer[n:], nonce)
	n += copy(buffer[n:], ciphertext)
	binary.BigEndian.PutUint64(buffer[n:], uint64(len(aad))*8)

	hmac := hmac.New(ctx.hash, ctx.integrityKey)
	_, _ = hmac.Write(buffer)

	return hmac.Sum(nil)[:ctx.authtagBytes]
}

func resize(in []byte, n uint64) (head, tail []byte) {
	if uint64(cap(in)) >= n {
		head = in[:n]
	} else {
		head = make([]byte, n)
		copy(head, in)
	}

	tail = head[len(in):]
	return
}

func padBuffer(buffer []byte, blockSize int) []byte {
	missing := blockSize - (len(buffer) % blockSize)
	ret, out := resize(buffer, uint64(len(buffer))+uint64(missing))
	padding := bytes.Repeat([]byte{byte(missing)}, missing)
	copy(out, padding)
	return ret
}

func unpadBuffer(buffer []byte, blockSize int) ([]byte, error) {
	if len(buffer)%blockSize != 0 {
		return nil, errors.New("square/go-jose: invalid padding")
	}

	last := buffer[len(buffer)-1]
	count := int(last)

	if count == 0 || count > blockSize || count > len(buffer) {
		return nil, errors.New("square/go-jose: invalid padding")
	}

	padding := bytes.Repeat([]byte{last}, count)
	if !bytes.HasSuffix(buffer, padding) {
		return nil, errors.New("square/go-jose: invalid padding")
	}

	return buffer[:len(buffer)-count], nil
}
