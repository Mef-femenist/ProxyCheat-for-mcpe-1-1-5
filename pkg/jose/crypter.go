package jose

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"errors"
	"fmt"
	"reflect"

	"mefproxy/pkg/jose/json"
)

type Encrypter interface {
	Encrypt(plaintext []byte) (*JSONWebEncryption, error)
	EncryptWithAuthData(plaintext []byte, aad []byte) (*JSONWebEncryption, error)
	Options() EncrypterOptions
}

type contentCipher interface {
	keySize() int
	encrypt(cek []byte, aad, plaintext []byte) (*aeadParts, error)
	decrypt(cek []byte, aad []byte, parts *aeadParts) ([]byte, error)
}

type keyGenerator interface {
	keySize() int
	genKey() ([]byte, rawHeader, error)
}

type keyEncrypter interface {
	encryptKey(cek []byte, alg KeyAlgorithm) (recipientInfo, error) 
}

type keyDecrypter interface {
	decryptKey(headers rawHeader, recipient *recipientInfo, generator keyGenerator) ([]byte, error) 
}

type genericEncrypter struct {
	contentAlg     ContentEncryption
	compressionAlg CompressionAlgorithm
	cipher         contentCipher
	recipients     []recipientKeyInfo
	keyGenerator   keyGenerator
	extraHeaders   map[HeaderKey]interface{}
}

type recipientKeyInfo struct {
	keyID        string
	keyAlg       KeyAlgorithm
	keyEncrypter keyEncrypter
}

type EncrypterOptions struct {
	Compression CompressionAlgorithm

	ExtraHeaders map[HeaderKey]interface{}
}

func (eo *EncrypterOptions) WithHeader(k HeaderKey, v interface{}) *EncrypterOptions {
	if eo.ExtraHeaders == nil {
		eo.ExtraHeaders = map[HeaderKey]interface{}{}
	}
	eo.ExtraHeaders[k] = v
	return eo
}

func (eo *EncrypterOptions) WithContentType(contentType ContentType) *EncrypterOptions {
	return eo.WithHeader(HeaderContentType, contentType)
}

func (eo *EncrypterOptions) WithType(typ ContentType) *EncrypterOptions {
	return eo.WithHeader(HeaderType, typ)
}

type Recipient struct {
	Algorithm  KeyAlgorithm
	Key        interface{}
	KeyID      string
	PBES2Count int
	PBES2Salt  []byte
}

func NewEncrypter(enc ContentEncryption, rcpt Recipient, opts *EncrypterOptions) (Encrypter, error) {
	encrypter := &genericEncrypter{
		contentAlg: enc,
		recipients: []recipientKeyInfo{},
		cipher:     getContentCipher(enc),
	}
	if opts != nil {
		encrypter.compressionAlg = opts.Compression
		encrypter.extraHeaders = opts.ExtraHeaders
	}

	if encrypter.cipher == nil {
		return nil, ErrUnsupportedAlgorithm
	}

	var keyID string
	var rawKey interface{}
	switch encryptionKey := rcpt.Key.(type) {
	case JSONWebKey:
		keyID, rawKey = encryptionKey.KeyID, encryptionKey.Key
	case *JSONWebKey:
		keyID, rawKey = encryptionKey.KeyID, encryptionKey.Key
	case OpaqueKeyEncrypter:
		keyID, rawKey = encryptionKey.KeyID(), encryptionKey
	default:
		rawKey = encryptionKey
	}

	switch rcpt.Algorithm {
	case DIRECT:
		
		if reflect.TypeOf(rawKey) != reflect.TypeOf([]byte{}) {
			return nil, ErrUnsupportedKeyType
		}
		if encrypter.cipher.keySize() != len(rawKey.([]byte)) {
			return nil, ErrInvalidKeySize
		}
		encrypter.keyGenerator = staticKeyGenerator{
			key: rawKey.([]byte),
		}
		recipientInfo, _ := newSymmetricRecipient(rcpt.Algorithm, rawKey.([]byte))
		recipientInfo.keyID = keyID
		if rcpt.KeyID != "" {
			recipientInfo.keyID = rcpt.KeyID
		}
		encrypter.recipients = []recipientKeyInfo{recipientInfo}
		return encrypter, nil
	case ECDH_ES:
		
		typeOf := reflect.TypeOf(rawKey)
		if typeOf != reflect.TypeOf(&ecdsa.PublicKey{}) {
			return nil, ErrUnsupportedKeyType
		}
		encrypter.keyGenerator = ecKeyGenerator{
			size:      encrypter.cipher.keySize(),
			algID:     string(enc),
			publicKey: rawKey.(*ecdsa.PublicKey),
		}
		recipientInfo, _ := newECDHRecipient(rcpt.Algorithm, rawKey.(*ecdsa.PublicKey))
		recipientInfo.keyID = keyID
		if rcpt.KeyID != "" {
			recipientInfo.keyID = rcpt.KeyID
		}
		encrypter.recipients = []recipientKeyInfo{recipientInfo}
		return encrypter, nil
	default:
		
		encrypter.keyGenerator = randomKeyGenerator{
			size: encrypter.cipher.keySize(),
		}
		err := encrypter.addRecipient(rcpt)
		return encrypter, err
	}
}

func NewMultiEncrypter(enc ContentEncryption, rcpts []Recipient, opts *EncrypterOptions) (Encrypter, error) {
	cipher := getContentCipher(enc)

	if cipher == nil {
		return nil, ErrUnsupportedAlgorithm
	}
	if rcpts == nil || len(rcpts) == 0 {
		return nil, fmt.Errorf("square/go-jose: recipients is nil or empty")
	}

	encrypter := &genericEncrypter{
		contentAlg: enc,
		recipients: []recipientKeyInfo{},
		cipher:     cipher,
		keyGenerator: randomKeyGenerator{
			size: cipher.keySize(),
		},
	}

	if opts != nil {
		encrypter.compressionAlg = opts.Compression
		encrypter.extraHeaders = opts.ExtraHeaders
	}

	for _, recipient := range rcpts {
		err := encrypter.addRecipient(recipient)
		if err != nil {
			return nil, err
		}
	}

	return encrypter, nil
}

func (ctx *genericEncrypter) addRecipient(recipient Recipient) (err error) {
	var recipientInfo recipientKeyInfo

	switch recipient.Algorithm {
	case DIRECT, ECDH_ES:
		return fmt.Errorf("square/go-jose: key algorithm '%s' not supported in multi-recipient mode", recipient.Algorithm)
	}

	recipientInfo, err = makeJWERecipient(recipient.Algorithm, recipient.Key)
	if recipient.KeyID != "" {
		recipientInfo.keyID = recipient.KeyID
	}

	switch recipient.Algorithm {
	case PBES2_HS256_A128KW, PBES2_HS384_A192KW, PBES2_HS512_A256KW:
		if sr, ok := recipientInfo.keyEncrypter.(*symmetricKeyCipher); ok {
			sr.p2c = recipient.PBES2Count
			sr.p2s = recipient.PBES2Salt
		}
	}

	if err == nil {
		ctx.recipients = append(ctx.recipients, recipientInfo)
	}
	return err
}

func makeJWERecipient(alg KeyAlgorithm, encryptionKey interface{}) (recipientKeyInfo, error) {
	switch encryptionKey := encryptionKey.(type) {
	case *rsa.PublicKey:
		return newRSARecipient(alg, encryptionKey)
	case *ecdsa.PublicKey:
		return newECDHRecipient(alg, encryptionKey)
	case []byte:
		return newSymmetricRecipient(alg, encryptionKey)
	case string:
		return newSymmetricRecipient(alg, []byte(encryptionKey))
	case *JSONWebKey:
		recipient, err := makeJWERecipient(alg, encryptionKey.Key)
		recipient.keyID = encryptionKey.KeyID
		return recipient, err
	}
	if encrypter, ok := encryptionKey.(OpaqueKeyEncrypter); ok {
		return newOpaqueKeyEncrypter(alg, encrypter)
	}
	return recipientKeyInfo{}, ErrUnsupportedKeyType
}

func newDecrypter(decryptionKey interface{}) (keyDecrypter, error) {
	switch decryptionKey := decryptionKey.(type) {
	case *rsa.PrivateKey:
		return &rsaDecrypterSigner{
			privateKey: decryptionKey,
		}, nil
	case *ecdsa.PrivateKey:
		return &ecDecrypterSigner{
			privateKey: decryptionKey,
		}, nil
	case []byte:
		return &symmetricKeyCipher{
			key: decryptionKey,
		}, nil
	case string:
		return &symmetricKeyCipher{
			key: []byte(decryptionKey),
		}, nil
	case JSONWebKey:
		return newDecrypter(decryptionKey.Key)
	case *JSONWebKey:
		return newDecrypter(decryptionKey.Key)
	}
	if okd, ok := decryptionKey.(OpaqueKeyDecrypter); ok {
		return &opaqueKeyDecrypter{decrypter: okd}, nil
	}
	return nil, ErrUnsupportedKeyType
}

func (ctx *genericEncrypter) Encrypt(plaintext []byte) (*JSONWebEncryption, error) {
	return ctx.EncryptWithAuthData(plaintext, nil)
}

func (ctx *genericEncrypter) EncryptWithAuthData(plaintext, aad []byte) (*JSONWebEncryption, error) {
	obj := &JSONWebEncryption{}
	obj.aad = aad

	obj.protected = &rawHeader{}
	err := obj.protected.set(headerEncryption, ctx.contentAlg)
	if err != nil {
		return nil, err
	}

	obj.recipients = make([]recipientInfo, len(ctx.recipients))

	if len(ctx.recipients) == 0 {
		return nil, fmt.Errorf("square/go-jose: no recipients to encrypt to")
	}

	cek, headers, err := ctx.keyGenerator.genKey()
	if err != nil {
		return nil, err
	}

	obj.protected.merge(&headers)

	for i, info := range ctx.recipients {
		recipient, err := info.keyEncrypter.encryptKey(cek, info.keyAlg)
		if err != nil {
			return nil, err
		}

		err = recipient.header.set(headerAlgorithm, info.keyAlg)
		if err != nil {
			return nil, err
		}

		if info.keyID != "" {
			err = recipient.header.set(headerKeyID, info.keyID)
			if err != nil {
				return nil, err
			}
		}
		obj.recipients[i] = recipient
	}

	if len(ctx.recipients) == 1 {
		
		obj.protected.merge(obj.recipients[0].header)
		obj.recipients[0].header = nil
	}

	if ctx.compressionAlg != NONE {
		plaintext, err = compress(ctx.compressionAlg, plaintext)
		if err != nil {
			return nil, err
		}

		err = obj.protected.set(headerCompression, ctx.compressionAlg)
		if err != nil {
			return nil, err
		}
	}

	for k, v := range ctx.extraHeaders {
		b, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		(*obj.protected)[k] = makeRawMessage(b)
	}

	authData := obj.computeAuthData()
	parts, err := ctx.cipher.encrypt(cek, authData, plaintext)
	if err != nil {
		return nil, err
	}

	obj.iv = parts.iv
	obj.ciphertext = parts.ciphertext
	obj.tag = parts.tag

	return obj, nil
}

func (ctx *genericEncrypter) Options() EncrypterOptions {
	return EncrypterOptions{
		Compression:  ctx.compressionAlg,
		ExtraHeaders: ctx.extraHeaders,
	}
}

func (obj JSONWebEncryption) Decrypt(decryptionKey interface{}) ([]byte, error) {
	headers := obj.mergedHeaders(nil)

	if len(obj.recipients) > 1 {
		return nil, errors.New("square/go-jose: too many recipients in payload; expecting only one")
	}

	critical, err := headers.getCritical()
	if err != nil {
		return nil, fmt.Errorf("square/go-jose: invalid crit header")
	}

	if len(critical) > 0 {
		return nil, fmt.Errorf("square/go-jose: unsupported crit header")
	}

	decrypter, err := newDecrypter(decryptionKey)
	if err != nil {
		return nil, err
	}

	cipher := getContentCipher(headers.getEncryption())
	if cipher == nil {
		return nil, fmt.Errorf("square/go-jose: unsupported enc value '%s'", string(headers.getEncryption()))
	}

	generator := randomKeyGenerator{
		size: cipher.keySize(),
	}

	parts := &aeadParts{
		iv:         obj.iv,
		ciphertext: obj.ciphertext,
		tag:        obj.tag,
	}

	authData := obj.computeAuthData()

	var plaintext []byte
	recipient := obj.recipients[0]
	recipientHeaders := obj.mergedHeaders(&recipient)

	cek, err := decrypter.decryptKey(recipientHeaders, &recipient, generator)
	if err == nil {
		
		plaintext, err = cipher.decrypt(cek, authData, parts)
	}

	if plaintext == nil {
		return nil, ErrCryptoFailure
	}

	if comp := obj.protected.getCompression(); comp != "" {
		plaintext, err = decompress(comp, plaintext)
	}

	return plaintext, err
}

func (obj JSONWebEncryption) DecryptMulti(decryptionKey interface{}) (int, Header, []byte, error) {
	globalHeaders := obj.mergedHeaders(nil)

	critical, err := globalHeaders.getCritical()
	if err != nil {
		return -1, Header{}, nil, fmt.Errorf("square/go-jose: invalid crit header")
	}

	if len(critical) > 0 {
		return -1, Header{}, nil, fmt.Errorf("square/go-jose: unsupported crit header")
	}

	decrypter, err := newDecrypter(decryptionKey)
	if err != nil {
		return -1, Header{}, nil, err
	}

	encryption := globalHeaders.getEncryption()
	cipher := getContentCipher(encryption)
	if cipher == nil {
		return -1, Header{}, nil, fmt.Errorf("square/go-jose: unsupported enc value '%s'", string(encryption))
	}

	generator := randomKeyGenerator{
		size: cipher.keySize(),
	}

	parts := &aeadParts{
		iv:         obj.iv,
		ciphertext: obj.ciphertext,
		tag:        obj.tag,
	}

	authData := obj.computeAuthData()

	index := -1
	var plaintext []byte
	var headers rawHeader

	for i, recipient := range obj.recipients {
		recipientHeaders := obj.mergedHeaders(&recipient)

		cek, err := decrypter.decryptKey(recipientHeaders, &recipient, generator)
		if err == nil {
			
			plaintext, err = cipher.decrypt(cek, authData, parts)
			if err == nil {
				index = i
				headers = recipientHeaders
				break
			}
		}
	}

	if plaintext == nil || err != nil {
		return -1, Header{}, nil, ErrCryptoFailure
	}

	if comp := obj.protected.getCompression(); comp != "" {
		plaintext, err = decompress(comp, plaintext)
	}

	sanitized, err := headers.sanitized()
	if err != nil {
		return -1, Header{}, nil, fmt.Errorf("square/go-jose: failed to sanitize header: %v", err)
	}

	return index, sanitized, plaintext, err
}
