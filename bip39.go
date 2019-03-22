// Package bip39 is the Golang implementation of the BIP39 spec.
package bip39

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"fmt"
	"strconv"
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
	ErrWordLen        = errors.New("Invalid word list length")
	ErrEntropyLen     = errors.New("Invalid entropy length")
	ErrUnIntMnemonic  = errors.New("Mnemonic is empty")
	ErrInvlidMnemonic = errors.New("Invlid mnemonic")
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

// NewBip39ByMnemonic create new Bip39 instance by mnemonic
func NewBip39ByMnemonic(mnemonic string, lang Language, password string) (*Bip39, error) {
	if !ValidateMnemonic(mnemonic, lang) {
		return nil, ErrInvlidMnemonic
	}

	ins := &Bip39{
		length:    len(strings.Split(norm.NFKD.String(mnemonic), "\x20")),
		menemonic: mnemonic, password: password, language: lang,
	}
	return ins, nil
}

// NewBip39ByEntropy create new Bip39 instance by entropy
func NewBip39ByEntropy(entropy []byte, lang Language, password string) (*Bip39, error) {
	var length int
	switch len(entropy) {
	case 128:
		length = 12
	case 160:
		length = 15
	case 192:
		length = 18
	case 224:
		length = 21
	case 256:
		length = 24
	default:
		return nil, ErrEntropyLen
	}

	ins := &Bip39{
		length:   length,
		language: lang,
		password: password,
	}

	if _, err := ins.NewMnemonicByEntroy(entropy); err != nil {
		return nil, err
	}
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
	entBitsLen := b.length * 11 / (1 + 1/32)
	entropy := make([]byte, entBitsLen/8)
	if _, err := rand.Read(entropy); err != nil {
		return "", err
	}
	return b.NewMnemonicByEntroy(entropy)
}

// NewMnemonicByEntroy creates new menemonic by entroy provied
func (b *Bip39) NewMnemonicByEntroy(entropy []byte) (string, error) {
	entBitsLen := b.length * 11 / (1 + 1/32)
	if entBitsLen/8 != len(entropy) {
		return "", ErrEntropyLen
	}

	var binEnt strings.Builder
	for _, v := range entropy {
		tmp := fmt.Sprintf("%08b", v)
		binEnt.WriteString(tmp)
	}

	hash := sha256.New()
	hash.Write(entropy)
	binEnt.WriteString(fmt.Sprintf("%08b", hash.Sum(nil)[0])[:entBitsLen/32])

	strEnt := binEnt.String()

	words := make([]string, 0, b.length)
	wordList := b.language.List()
	for i := 0; i < len(strEnt); i += 11 {
		idx, err := strconv.ParseInt(strEnt[i:i+11], 2, 32)
		if err != nil {
			return "", err
		}
		words = append(words, wordList[idx])
	}

	if b.language == Japanese {
		b.menemonic = strings.Join(words, "\u3000")
	} else {
		b.menemonic = strings.Join(words, "\x20")
	}
	return b.menemonic, nil
}

// Seed gets bip32 root seed
func (b *Bip39) Seed() ([]byte, error) {
	if b.menemonic == "" {
		return nil, ErrUnIntMnemonic
	}
	reschan := make(chan []byte)
	go func() {
		password := []byte(norm.NFKD.String(b.menemonic))
		salt := []byte(norm.NFKD.String("mnemonic" + b.password))
		defer close(reschan)
		reschan <- pbkdf2.Key(password, salt, 2048, 64, sha512.New)
	}()
	return <-reschan, nil
}

// ValidateMnemonic validate menemonic
func ValidateMnemonic(mnemonic string, lang Language) bool {
	mnemonic = norm.NFKD.String(mnemonic)
	wordList := strings.Split(mnemonic, "\x20")

	wordCount := len(wordList)
	if wordCount%3 != 0 || wordCount < 12 || wordCount > 24 {
		return false
	}

	// record index of word
	words := make(map[string]int)
	for idx, v := range lang.List() {
		words[v] = idx
	}

	var tmp strings.Builder
	for _, v := range wordList {
		idx, has := words[v]
		if !has {
			return false
		}
		x := fmt.Sprintf("%08b", idx)
		if rpt := 11 - len(x); rpt != 0 {
			x = strings.Repeat("0", rpt) + x
		}
		tmp.WriteString(x)
	}

	res := tmp.String()
	binLen := len(res)

	entBitsLen := len(wordList) * 11 / (1 + 1/32)
	csBitsLen := entBitsLen / 32

	entBytes := make([]byte, 0, (entBitsLen-csBitsLen)/8)
	for i := 0; i < binLen-4; i += 8 {
		b, err := strconv.ParseInt(res[i:i+8], 2, 32)
		if err != nil {
			return false
		}
		entBytes = append(entBytes, byte(b))
	}

	hash := sha256.New()
	hash.Write(entBytes)
	return res[binLen-4:binLen] == fmt.Sprintf("%08b", hash.Sum(nil)[0])[:csBitsLen]
}
