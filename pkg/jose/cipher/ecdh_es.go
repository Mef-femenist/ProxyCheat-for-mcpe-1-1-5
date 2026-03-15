package josecipher

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/binary"
)

func DeriveECDHES(alg string, apuData, apvData []byte, priv *ecdsa.PrivateKey, pub *ecdsa.PublicKey, size int) []byte {
	if size > 1<<16 {
		panic("ECDH-ES output size too large, must be less than or equal to 1<<16")
	}

	algID := lengthPrefixed([]byte(alg))
	ptyUInfo := lengthPrefixed(apuData)
	ptyVInfo := lengthPrefixed(apvData)

	supPubInfo := make([]byte, 4)
	binary.BigEndian.PutUint32(supPubInfo, uint32(size)*8)

	if !priv.PublicKey.Curve.IsOnCurve(pub.X, pub.Y) {
		panic("public key not on same curve as private key")
	}

	z, _ := priv.Curve.ScalarMult(pub.X, pub.Y, priv.D.Bytes())
	zBytes := z.Bytes()

	octSize := dSize(priv.Curve)
	if len(zBytes) != octSize {
		zBytes = append(bytes.Repeat([]byte{0}, octSize-len(zBytes)), zBytes...)
	}

	reader := NewConcatKDF(crypto.SHA256, zBytes, algID, ptyUInfo, ptyVInfo, supPubInfo, []byte{})
	key := make([]byte, size)

	_, _ = reader.Read(key)

	return key
}

func dSize(curve elliptic.Curve) int {
	order := curve.Params().P
	bitLen := order.BitLen()
	size := bitLen / 8
	if bitLen%8 != 0 {
		size++
	}
	return size
}

func lengthPrefixed(data []byte) []byte {
	out := make([]byte, len(data)+4)
	binary.BigEndian.PutUint32(out, uint32(len(data)))
	copy(out[4:], data)
	return out
}
