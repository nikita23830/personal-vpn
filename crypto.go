package main

import (
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"log"

	"golang.org/x/crypto/chacha20poly1305"
)

var aead cipher.AEAD

func initCrypto(passphrase string) {
	key := sha256.Sum256([]byte(passphrase))
	a, err := chacha20poly1305.NewX(key[:])
	if err != nil {
		log.Fatalf("encryption initialization: %v", err)
	}
	aead = a
}

func seal(plaintext []byte) []byte {
	nonce := make([]byte, chacha20poly1305.NonceSizeX)
	_, _ = rand.Read(nonce)

	var b [1]byte
	_, _ = rand.Read(b[:])
	padLen := int(b[0] & 0x3f)

	padded := make([]byte, 0, len(plaintext)+padLen+1)
	padded = append(padded, plaintext...)
	if padLen > 0 {
		pad := make([]byte, padLen)
		_, _ = rand.Read(pad)
		padded = append(padded, pad...)
	}
	padded = append(padded, byte(padLen))

	ct := aead.Seal(nil, nonce, padded, nil)
	out := make([]byte, len(nonce)+len(ct))
	copy(out, nonce)
	copy(out[len(nonce):], ct)
	return out
}

func open(packet []byte) (plaintext []byte, ok bool) {
	ns := chacha20poly1305.NonceSizeX
	if len(packet) < ns+aead.Overhead()+1 {
		return nil, false
	}
	pt, err := aead.Open(nil, packet[:ns], packet[ns:], nil)
	if err != nil || len(pt) < 1 {
		return nil, false
	}
	padLen := int(pt[len(pt)-1])
	if len(pt) < 1+padLen {
		return nil, false
	}
	return pt[:len(pt)-1-padLen], true
}
