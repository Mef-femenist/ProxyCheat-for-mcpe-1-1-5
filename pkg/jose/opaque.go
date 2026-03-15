package jose

type OpaqueSigner interface {
	
	Public() *JSONWebKey
	
	Algs() []SignatureAlgorithm
	
	SignPayload(payload []byte, alg SignatureAlgorithm) ([]byte, error)
}

type opaqueSigner struct {
	signer OpaqueSigner
}

func newOpaqueSigner(alg SignatureAlgorithm, signer OpaqueSigner) (recipientSigInfo, error) {
	var algSupported bool
	for _, salg := range signer.Algs() {
		if alg == salg {
			algSupported = true
			break
		}
	}
	if !algSupported {
		return recipientSigInfo{}, ErrUnsupportedAlgorithm
	}

	return recipientSigInfo{
		sigAlg:    alg,
		publicKey: signer.Public,
		signer: &opaqueSigner{
			signer: signer,
		},
	}, nil
}

func (o *opaqueSigner) signPayload(payload []byte, alg SignatureAlgorithm) (Signature, error) {
	out, err := o.signer.SignPayload(payload, alg)
	if err != nil {
		return Signature{}, err
	}

	return Signature{
		Signature: out,
		protected: &rawHeader{},
	}, nil
}

type OpaqueVerifier interface {
	VerifyPayload(payload []byte, signature []byte, alg SignatureAlgorithm) error
}

type opaqueVerifier struct {
	verifier OpaqueVerifier
}

func (o *opaqueVerifier) verifyPayload(payload []byte, signature []byte, alg SignatureAlgorithm) error {
	return o.verifier.VerifyPayload(payload, signature, alg)
}

type OpaqueKeyEncrypter interface {
	
	KeyID() string
	
	Algs() []KeyAlgorithm
	
	encryptKey(cek []byte, alg KeyAlgorithm) (recipientInfo, error)
}

type opaqueKeyEncrypter struct {
	encrypter OpaqueKeyEncrypter
}

func newOpaqueKeyEncrypter(alg KeyAlgorithm, encrypter OpaqueKeyEncrypter) (recipientKeyInfo, error) {
	var algSupported bool
	for _, salg := range encrypter.Algs() {
		if alg == salg {
			algSupported = true
			break
		}
	}
	if !algSupported {
		return recipientKeyInfo{}, ErrUnsupportedAlgorithm
	}

	return recipientKeyInfo{
		keyID:  encrypter.KeyID(),
		keyAlg: alg,
		keyEncrypter: &opaqueKeyEncrypter{
			encrypter: encrypter,
		},
	}, nil
}

func (oke *opaqueKeyEncrypter) encryptKey(cek []byte, alg KeyAlgorithm) (recipientInfo, error) {
	return oke.encrypter.encryptKey(cek, alg)
}

type OpaqueKeyDecrypter interface {
	DecryptKey(encryptedKey []byte, header Header) ([]byte, error)
}

type opaqueKeyDecrypter struct {
	decrypter OpaqueKeyDecrypter
}

func (okd *opaqueKeyDecrypter) decryptKey(headers rawHeader, recipient *recipientInfo, generator keyGenerator) ([]byte, error) {
	mergedHeaders := rawHeader{}
	mergedHeaders.merge(&headers)
	mergedHeaders.merge(recipient.header)

	header, err := mergedHeaders.sanitized()
	if err != nil {
		return nil, err
	}

	return okd.decrypter.DecryptKey(recipient.encryptedKey, header)
}
