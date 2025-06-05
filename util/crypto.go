/*
 * MIT License
 *
 * Copyright (c) 2025 Marlon Pohl
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package util

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"golang.org/x/crypto/nacl/box"
)

const ThreemaPhoneHmacKey = "85adf8226953f3d96cfd5d09bf29555eb955fcd8aa5ec4f9fcd869e258370723"

type NaClKey *[32]byte

// EncryptBytes returns the encrypted message, the nonce used, or an error
func EncryptBytes(b *[]byte, pubKey NaClKey, privKey NaClKey) ([]byte, []byte, error) {
	nonce, err := generateNonce()
	if err != nil {
		return nil, nil, err
	}

	enc := make([]byte, 0)
	enc = box.Seal(enc[:], *b, nonce, pubKey, privKey)
	return enc, nonce[:], nil
}

func NaClKeyFromString(key string) (NaClKey, error) {
	keyBytes, err := hex.DecodeString(key)
	if err != nil {
		return nil, err
	}

	if len(keyBytes) != 32 {
		return nil, fmt.Errorf("invalid key")
	}

	naclKey := new([32]byte)
	copy(naclKey[:], keyBytes)
	return naclKey, nil
}

func HashPhoneNumber(phoneNumber string) (string, error) {
	mac := hmac.New(sha256.New, []byte(ThreemaPhoneHmacKey))
	_, err := mac.Write([]byte(phoneNumber))
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(mac.Sum(nil)), nil
}

func generateNonce() (*[24]byte, error) {
	nonce := new([24]byte)
	n, err := rand.Read(nonce[:])
	if err != nil {
		return nil, err
	}

	if n != len(nonce) {
		return nil, fmt.Errorf("could not read enough random bytes")
	}

	return nonce, nil
}
