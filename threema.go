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

package threemagatewaygo

import (
	"fmt"
	"github.com/Keksnet/threema-gateway-go/message"
	"github.com/Keksnet/threema-gateway-go/util"
	"io"
	"log"
	"net/http"
	"strconv"
)

const ThreemaGatewayApi = "https://msgapi.threema.ch"
const ThreemaIdLength = 8
const ThreemaMaxMessageLength = 3500
const ThreemaMaxEncryptedMessageLength = 7812

type ApiCredentials struct {
	ThreemaId string
	ApiKey    string
}

type ThreemaClient struct {
	ApiUrl      string
	Credentials *ApiCredentials
	sender      ThreemaMessageSender
	keyCache    map[string]util.NaClKey
}

/* ------------ Constructors ------------ */

func WithoutEncryption(credentials *ApiCredentials) *ThreemaClient {
	client := ThreemaClient{
		ApiUrl:      ThreemaGatewayApi,
		Credentials: credentials,
		keyCache:    make(map[string]util.NaClKey),
	}

	client.sender = &ThreemaBasicMessageSender{
		client: client,
	}

	return &client
}

func WithEncryption(credentials *ApiCredentials, encryption *EncryptionSettings) *ThreemaClient {
	client := ThreemaClient{
		ApiUrl:      ThreemaGatewayApi,
		Credentials: credentials,
		keyCache:    make(map[string]util.NaClKey),
	}

	client.sender = &ThreemaEndToEndMessageSender{
		Encryption: encryption,
		client:     client,
	}

	return &client
}

/* ------------ ID Lookup ------------ */

func (t *ThreemaClient) LookupIdByPhoneNumber(phoneNumber string) (string, error) {
	phoneNumber = util.NormalizePhoneNumber(phoneNumber)
	err := util.ValidatePhoneNumber(phoneNumber)
	if err != nil {
		return "", err
	}

	hash, err := util.HashPhoneNumber(phoneNumber)
	if err != nil {
		return "", err
	}

	c := t.Credentials
	reqUrl := fmt.Sprintf("%s/lookup/phone_hash/%s?from=%s&secret=%s", t.ApiUrl, hash, c.ThreemaId, c.ApiKey)
	resp, err := http.Get(reqUrl)
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	err = resp.Body.Close()
	if err != nil {
		return "", err
	}

	return string(body), nil
}

/* ------------ Key Lookup ------------ */

func (t *ThreemaClient) LookupKey(id string) (util.NaClKey, error) {
	// Get cached key
	if t.keyCache[id] != nil {
		return t.keyCache[id], nil
	}

	// if no key is cached send a request to the api
	c := t.Credentials
	reqUrl := fmt.Sprintf("%s/pubkeys/%s?from=%s&secret=%s", t.ApiUrl, id, c.ThreemaId, c.ApiKey)
	res, err := http.Get(reqUrl)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	err = res.Body.Close()
	if err != nil {
		return nil, err
	}

	key, err := util.NaClKeyFromString(string(body))
	if err != nil {
		return nil, err
	}

	// save the key to the cache
	t.keyCache[id] = key

	return key, nil
}

/* ------------ Messaging ------------ */

func (t *ThreemaClient) SendMessage(rcv string, msg message.ThreemaMessage, pubKey util.NaClKey) (string, error) {
	if len(rcv) != ThreemaIdLength {
		return "", fmt.Errorf("rcv length invalid")
	}

	key := pubKey
	if pubKey == nil {
		remoteKey, err := t.LookupKey(rcv)
		if err != nil {
			return "", err
		}
		key = remoteKey
	}

	return t.sender.SendMessage(rcv, key, msg)
}

/* ------------ Account Information ------------ */

func (t *ThreemaClient) GetCredits() (int64, error) {
	creditsUrl := fmt.Sprintf("%s/credits?from=%s&secret=%s", ThreemaGatewayApi, t.Credentials.ThreemaId, t.Credentials.ApiKey)

	resp, err := http.Get(creditsUrl)
	if err != nil {
		return -1, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Threema Gateway API failed to close response body")
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return -1, err
	}

	return strconv.ParseInt(string(body), 10, 64)
}

func (t *ThreemaClient) ValidateConnection() (bool, error) {
	ok, err := t.GetCredits()
	if err != nil {
		return false, err
	}

	return ok != -1, nil
}
