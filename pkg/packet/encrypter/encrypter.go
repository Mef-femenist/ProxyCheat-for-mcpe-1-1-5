package encrypter

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"mefproxy/pkg/packet"
	"mefproxy/pkg/packet/encrypter/cfb8"
	"sync"
)

type Instance struct {
	IsEnabled        bool
	ServerToken      string
	BotKeyPair       *ecdsa.PrivateKey
	_encryptCounter  int64
	_encrypter       cipher.Stream
	_decrypter       cipher.Stream
	_serverPublicKey *ecdsa.PublicKey
	key              []byte
	_mux             sync.Mutex
}

func (inst *Instance) CreateSession(pubkey, token string) *Instance {
	inst._mux = sync.Mutex{}
	data, _ := base64.StdEncoding.DecodeString(pubkey)
	publicKey, _ := x509.ParsePKIXPublicKey(data)
	inst._serverPublicKey = publicKey.(*ecdsa.PublicKey)

	sharedKey, _ := inst._serverPublicKey.Curve.ScalarMult(inst._serverPublicKey.X, inst._serverPublicKey.Y, inst.BotKeyPair.D.Bytes())
	inst.key = hash256prepend([]byte(token), sharedKey.Bytes())
	inst.newSymmetricEncryption(inst.key)
	return inst
}

func hash256prepend(prep, byted []byte) []byte {
	hasher := sha256.New()
	tmpbuf := bytes.NewBuffer([]byte{})
	tmpbuf.Write(prep)
	tmpbuf.Write(byted)
	_, _ = hasher.Write(tmpbuf.Bytes())
	hs := hasher.Sum(nil)
	return hs[:32]
}

func (inst *Instance) LockForHandshake() {
	inst._mux.Lock()
}

func (inst *Instance) UnlockAfterHandshake() {
	inst._mux.Unlock()
}

func (inst *Instance) Reset() {
	inst._mux.Lock()
	inst._mux.Unlock()
	inst._encryptCounter = 0
	inst.IsEnabled = false
}

func (inst *Instance) newSymmetricEncryption(key []byte) {
	b, _ := aes.NewCipher(key)
	inst._decrypter = cfb8.NewCFB8Decrypt(b, key[:16])
	inst._encrypter = cfb8.NewCFB8Encrypt(b, key[:16])
	return
}

func (inst *Instance) EncodePayload(payload []byte) []byte {
	inst._mux.Lock()
	defer inst._mux.Unlock()
	tmpbuf := bytes.NewBuffer([]byte{})
	tmpbuf.Write(payload)
	tmpbuf.Write(inst.calculateChecksum(inst._encryptCounter, payload))
	outd := tmpbuf.Bytes()
	out := make([]byte, len(outd))
	inst._encrypter.XORKeyStream(out, outd)
	inst._encryptCounter++
	return out
}

func (inst *Instance) EncodePayloadOnce(payload []byte) []byte {
	tmpbuf := bytes.NewBuffer([]byte{})
	tmpbuf.Write(payload)
	tmpbuf.Write(inst.calculateChecksum(inst._encryptCounter, payload))
	outd := tmpbuf.Bytes()
	out := make([]byte, len(outd))
	inst._encrypter.XORKeyStream(out, outd)
	inst._encryptCounter++
	return out
}

func (inst *Instance) DecodePayload(payload []byte) []byte {
	inst._mux.Lock()
	defer inst._mux.Unlock()
	if len(payload) < 9 {
		return []byte{}
	}
	out := make([]byte, len(payload))
	inst._decrypter.XORKeyStream(out, payload)
	decpay := out[:len(out)-8]
	return decpay
}

func (inst *Instance) calculateChecksum(counter int64, payload []byte) []byte {
	hasher := sha256.New()
	tmpbuf := bytes.NewBuffer([]byte{})
	pw := packet.NewWriter(tmpbuf, 29)

	pw.Int64(&counter)
	pw.Bytes(&payload)
	pw.Bytes(&inst.key)
	_, _ = hasher.Write(tmpbuf.Bytes())
	hash := hasher.Sum(nil)
	return hash[:8]
}
