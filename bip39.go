// Package bip39 is the Golang implementation of the BIP39 spec.
package bip39

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"errors"
	"math/big"
	"strings"

	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/text/unicode/norm"
)

// Bip39 is bip39 instance
type Bip39 struct {
	length    int
	language  Language
	password  string
	menemonic string
}

// Error list
var (
	ErrWordLen       = errors.New("Invalid word list length")
	ErrEntropyLen    = errors.New("Invalid entropy length")
	ErrUnIntMnemonic = errors.New("Mnemonic is empty")
)

// NewBip39 create new Bip39 instance
// mlen should be 12 | 15 | 18 | 21 | 24, and recomend 12 and 24
// password can be empty string
func NewBip39(mlen int, lang Language, password string) (*Bip39, error) {
	if mlen < 12 || mlen > 24 || mlen%3 != 0 {
		return nil, ErrWordLen
	}
	ins := &Bip39{length: mlen, language: lang, password: password}
	return ins, nil
}

// NewMnemonic gets new mnemonic
func (b *Bip39) NewMnemonic() (string, error) {
	/*
		CS = ENT / 32
		MS = (ENT + CS) / 11

		|  ENT  | CS | ENT+CS |  MS  |
		+-------+----+--------+------+
		|  128  |  4 |   132  |  12  |
		|  160  |  5 |   165  |  15  |
		|  192  |  6 |   198  |  18  |
		|  224  |  7 |   231  |  21  |
		|  256  |  8 |   264  |  24  |
	*/
	entBits := b.length * 11 / (1 + 1/32)
	entropy := make([]byte, entBits/8)
	if _, err := rand.Read(entropy); err != nil {
		return "", err
	}
	return b.NewMnemonicByEntroy(entropy)
}

// NewMnemonicByEntroy creates new menemonic by entroy provied
func (b *Bip39) NewMnemonicByEntroy(entropy []byte) (string, error) {
	entBitsLen := b.length * 11 / (1 + 1/32)
	if entBitsLen != len(entropy) {
		return "", ErrEntropyLen
	}

	csBits := entBitsLen / 32
	hash := sha256.New()
	hash.Write(entropy)
	firstCsByte := hash.Sum(nil)[0]

	dataBigInt := new(big.Int).SetBytes(entropy)

	for i := 0; i < csBits; i++ {
		dataBigInt.Mul(dataBigInt, big.NewInt(2))
		if uint8(firstCsByte&(1<<(7-uint(i)))) > 0 {
			dataBigInt.Or(dataBigInt, big.NewInt(1))
		}
	}

	var padByteSlice = func(slice []byte, length int) []byte {
		offset := length - len(slice)
		if offset <= 0 {
			return slice
		}
		newSlice := make([]byte, length)
		copy(newSlice[offset:], slice)
		return newSlice
	}

	words := make([]string, b.length)
	word := big.NewInt(0)

	for i := b.length - 1; i >= 0; i-- {
		word.And(dataBigInt, big.NewInt(2047))
		dataBigInt.Div(dataBigInt, big.NewInt(2048))
		wordBytes := padByteSlice(word.Bytes(), 2)
		words[i] = b.language.List()[binary.BigEndian.Uint16(wordBytes)]
	}
	var mnemonic string
	if b.language == Japanese {
		mnemonic = strings.Join(words, "\u3000")
	}
	mnemonic = strings.Join(words, "\x20")
	b.menemonic = mnemonic
	return mnemonic, nil
}

// Seed gets bip32 root seed
func (b *Bip39) Seed() ([]byte, error) {
	if b.menemonic == "" {
		return nil, ErrUnIntMnemonic
	}
	reschan := make(chan []byte)
	go func() {
		password := []byte(norm.NFKD.String(b.menemonic))
		salt := []byte(norm.NFKD.String(b.menemonic + b.password))
		defer close(reschan)
		reschan <- pbkdf2.Key(password, salt, 2018, 64, sha512.New)
	}()
	return <-reschan, nil
}

// ValidateMnemonic validate menemonic
func ValidateMnemonic(mnemonic string, lang Language) bool {
	mnemonic = norm.NFKD.String(mnemonic)
	words := make(map[string]struct{})
	for _, v := range lang.List() {
		words[v] = struct{}{}
	}
	for _, v := range strings.Split(mnemonic, "\x20") {
		if _, ok := words[v]; !ok {
			return false
		}
	}
	// TODO(islishude): validate checksum
	return true
}
