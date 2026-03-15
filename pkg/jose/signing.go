package jose

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"fmt"
	"golang.org/x/crypto/ed25519"
	"mefproxy/pkg/jose/json"
)

type NonceSource interface {
	Nonce() (string, error)
}

type Signer interface {
	Sign(payload []byte) (*JSONWebSignature, error)
	Options() SignerOptions
}

type SigningKey struct {
	Algorithm SignatureAlgorithm
	Key       interface{}
}

type SignerOptions struct {
	NonceSource NonceSource
	EmbedJWK    bool

	ExtraHeaders map[HeaderKey]interface{}
}

func (so *SignerOptions) WithHeader(k HeaderKey, v interface{}) *SignerOptions {
	if so.ExtraHeaders == nil {
		so.ExtraHeaders = map[HeaderKey]interface{}{}
	}
	so.ExtraHeaders[k] = v
	return so
}

func (so *SignerOptions) WithContentType(contentType ContentType) *SignerOptions {
	return so.WithHeader(HeaderContentType, contentType)
}

func (so *SignerOptions) WithType(typ ContentType) *SignerOptions {
	return so.WithHeader(HeaderType, typ)
}

func (so *SignerOptions) WithCritical(names ...string) *SignerOptions {
	if so.ExtraHeaders[headerCritical] == nil {
		so.WithHeader(headerCritical, make([]string, 0, len(names)))
	}
	crit := so.ExtraHeaders[headerCritical].([]string)
	so.ExtraHeaders[headerCritical] = append(crit, names...)
	return so
}

func (so *SignerOptions) WithBase64(b64 bool) *SignerOptions {
	if !b64 {
		so.WithHeader(headerB64, b64)
		so.WithCritical(headerB64)
	}
	return so
}

type payloadSigner interface {
	signPayload(payload []byte, alg SignatureAlgorithm) (Signature, error)
}

type payloadVerifier interface {
	verifyPayload(payload []byte, signature []byte, alg SignatureAlgorithm) error
}

type genericSigner struct {
	recipients   []recipientSigInfo
	nonceSource  NonceSource
	embedJWK     bool
	extraHeaders map[HeaderKey]interface{}
}

type recipientSigInfo struct {
	sigAlg    SignatureAlgorithm
	publicKey func() *JSONWebKey
	signer    payloadSigner
}

func staticPublicKey(jwk *JSONWebKey) func() *JSONWebKey {
	return func() *JSONWebKey {
		return jwk
	}
}

func NewSigner(sig SigningKey, opts *SignerOptions) (Signer, error) {
	return NewMultiSigner([]SigningKey{sig}, opts)
}

func NewMultiSigner(sigs []SigningKey, opts *SignerOptions) (Signer, error) {
	signer := &genericSigner{recipients: []recipientSigInfo{}}

	if opts != nil {
		signer.nonceSource = opts.NonceSource
		signer.embedJWK = opts.EmbedJWK
		signer.extraHeaders = opts.ExtraHeaders
	}

	for _, sig := range sigs {
		err := signer.addRecipient(sig.Algorithm, sig.Key)
		if err != nil {
			return nil, err
		}
	}

	return signer, nil
}

func newVerifier(verificationKey interface{}) (payloadVerifier, error) {
	switch verificationKey := verificationKey.(type) {
	case ed25519.PublicKey:
		return &edEncrypterVerifier{
			publicKey: verificationKey,
		}, nil
	case *rsa.PublicKey:
		return &rsaEncrypterVerifier{
			publicKey: verificationKey,
		}, nil
	case *ecdsa.PublicKey:
		return &ecEncrypterVerifier{
			publicKey: verificationKey,
		}, nil
	case []byte:
		return &symmetricMac{
			key: verificationKey,
		}, nil
	case JSONWebKey:
		return newVerifier(verificationKey.Key)
	case *JSONWebKey:
		return newVerifier(verificationKey.Key)
	}
	if ov, ok := verificationKey.(OpaqueVerifier); ok {
		return &opaqueVerifier{verifier: ov}, nil
	}
	return nil, ErrUnsupportedKeyType
}

func (ctx *genericSigner) addRecipient(alg SignatureAlgorithm, signingKey interface{}) error {
	recipient, err := makeJWSRecipient(alg, signingKey)
	if err != nil {
		return err
	}

	ctx.recipients = append(ctx.recipients, recipient)
	return nil
}

func makeJWSRecipient(alg SignatureAlgorithm, signingKey interface{}) (recipientSigInfo, error) {
	switch signingKey := signingKey.(type) {
	case ed25519.PrivateKey:
		return newEd25519Signer(alg, signingKey)
	case *rsa.PrivateKey:
		return newRSASigner(alg, signingKey)
	case *ecdsa.PrivateKey:
		return newECDSASigner(alg, signingKey)
	case []byte:
		return newSymmetricSigner(alg, signingKey)
	case JSONWebKey:
		return newJWKSigner(alg, signingKey)
	case *JSONWebKey:
		return newJWKSigner(alg, *signingKey)
	}
	if signer, ok := signingKey.(OpaqueSigner); ok {
		return newOpaqueSigner(alg, signer)
	}
	return recipientSigInfo{}, ErrUnsupportedKeyType
}

func newJWKSigner(alg SignatureAlgorithm, signingKey JSONWebKey) (recipientSigInfo, error) {
	recipient, err := makeJWSRecipient(alg, signingKey.Key)
	if err != nil {
		return recipientSigInfo{}, err
	}
	if recipient.publicKey != nil && recipient.publicKey() != nil {
		
		publicKey := signingKey
		publicKey.Key = recipient.publicKey().Key
		recipient.publicKey = staticPublicKey(&publicKey)

		if !recipient.publicKey().IsPublic() {
			return recipientSigInfo{}, errors.New("square/go-jose: public key was unexpectedly not public")
		}
	}
	return recipient, nil
}

func (ctx *genericSigner) Sign(payload []byte) (*JSONWebSignature, error) {
	obj := &JSONWebSignature{}
	obj.payload = payload
	obj.Signatures = make([]Signature, len(ctx.recipients))

	for i, recipient := range ctx.recipients {
		protected := map[HeaderKey]interface{}{
			headerAlgorithm: string(recipient.sigAlg),
		}

		if recipient.publicKey != nil && recipient.publicKey() != nil {
			
			if ctx.embedJWK {
				protected[headerJWK] = recipient.publicKey()
			} else {
				keyID := recipient.publicKey().KeyID
				if keyID != "" {
					protected[headerKeyID] = keyID
				}
			}
		}

		if ctx.nonceSource != nil {
			nonce, err := ctx.nonceSource.Nonce()
			if err != nil {
				return nil, fmt.Errorf("square/go-jose: Error generating nonce: %v", err)
			}
			protected[headerNonce] = nonce
		}

		for k, v := range ctx.extraHeaders {
			protected[k] = v
		}
		serializedProtected := mustSerializeJSON(protected)

		needsBase64 := true

		if b64, ok := protected[headerB64]; ok {
			if needsBase64, ok = b64.(bool); !ok {
				return nil, errors.New("square/go-jose: Invalid b64 header parameter")
			}
		}

		var input bytes.Buffer

		input.WriteString(base64.RawURLEncoding.EncodeToString(serializedProtected))
		
		input.WriteByte('.')

		if needsBase64 {
			input.WriteString(base64.RawURLEncoding.EncodeToString(payload))
		} else {
			input.Write(payload)
		}

		signatureInfo, err := recipient.signer.signPayload(input.Bytes(), recipient.sigAlg)
		if err != nil {
			return nil, err
		}

		signatureInfo.protected = &rawHeader{}
		for k, v := range protected {
			b, err := json.Marshal(v)
			
			if err != nil {
				return nil, fmt.Errorf("square/go-jose: Error marshalling item %#v: %v", k, err)
			}
			(*signatureInfo.protected)[k] = makeRawMessage(b)
		}
		obj.Signatures[i] = signatureInfo
	}

	return obj, nil
}

func (ctx *genericSigner) Options() SignerOptions {
	return SignerOptions{
		NonceSource:  ctx.nonceSource,
		EmbedJWK:     ctx.embedJWK,
		ExtraHeaders: ctx.extraHeaders,
	}
}

func (obj JSONWebSignature) Verify(verificationKey interface{}) ([]byte, error) {
	err := obj.DetachedVerify(obj.payload, verificationKey)
	if err != nil {
		return nil, err
	}
	return obj.payload, nil
}

func (obj JSONWebSignature) UnsafePayloadWithoutVerification() []byte {
	return obj.payload
}

func (obj JSONWebSignature) DetachedVerify(payload []byte, verificationKey interface{}) error {
	verifier, err := newVerifier(verificationKey)
	if err != nil {
		return err
	}

	if len(obj.Signatures) > 1 {
		return errors.New("square/go-jose: too many signatures in payload; expecting only one")
	}

	signature := obj.Signatures[0]
	headers := signature.mergedHeaders()
	critical, err := headers.getCritical()
	if err != nil {
		return err
	}

	for _, name := range critical {
		if !supportedCritical[name] {
			return ErrCryptoFailure
		}
	}

	input, err := obj.computeAuthData(payload, &signature)
	if err != nil {
		return ErrCryptoFailure
	}

	alg := headers.getSignatureAlgorithm()
	err = verifier.verifyPayload(input, signature.Signature, alg)
	if err == nil {
		return nil
	}

	return ErrCryptoFailure
}

func (obj JSONWebSignature) VerifyMulti(verificationKey interface{}) (int, Signature, []byte, error) {
	idx, sig, err := obj.DetachedVerifyMulti(obj.payload, verificationKey)
	if err != nil {
		return -1, Signature{}, nil, err
	}
	return idx, sig, obj.payload, nil
}

func (obj JSONWebSignature) DetachedVerifyMulti(payload []byte, verificationKey interface{}) (int, Signature, error) {
	verifier, err := newVerifier(verificationKey)
	if err != nil {
		return -1, Signature{}, err
	}

outer:
	for i, signature := range obj.Signatures {
		headers := signature.mergedHeaders()
		critical, err := headers.getCritical()
		if err != nil {
			continue
		}

		for _, name := range critical {
			if !supportedCritical[name] {
				continue outer
			}
		}

		input, err := obj.computeAuthData(payload, &signature)
		if err != nil {
			continue
		}

		alg := headers.getSignatureAlgorithm()
		err = verifier.verifyPayload(input, signature.Signature, alg)
		if err == nil {
			return i, signature, nil
		}
	}

	return -1, Signature{}, ErrCryptoFailure
}
