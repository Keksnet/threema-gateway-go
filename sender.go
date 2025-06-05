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
	"encoding/hex"
	"fmt"
	"github.com/Keksnet/threema-gateway-go/message"
	"github.com/Keksnet/threema-gateway-go/util"
	"io"
	"net/http"
	"net/url"
)

type ThreemaMessageSender interface {
	SupportsPhoneOrEmail() bool
	SendMessage(rcv string, pubKey util.NaClKey, msg message.ThreemaMessage) (string, error)
}

type ThreemaBasicMessageSender struct {
	client ThreemaClient
	ThreemaMessageSender
}

func (s *ThreemaBasicMessageSender) SupportsPhoneOrEmail() bool {
	return true
}

func (s *ThreemaBasicMessageSender) SendMessage(rcv string, pubKey util.NaClKey, msg message.ThreemaMessage) (string, error) {
	if msg.MessageType() != message.TextMessageId {
		return "", fmt.Errorf("message type not supported")
	}

	txtMsg, ok := msg.(*message.TextMessage)
	if !ok {
		return "", fmt.Errorf("message type not supported")
	}

	if len(txtMsg.Text) > ThreemaMaxMessageLength {
		return "", fmt.Errorf("message too long")
	}

	reqUrl := fmt.Sprintf("%s/send_simple", s.client.ApiUrl)
	form := url.Values{
		"from":   []string{s.client.Credentials.ThreemaId},
		"to":     []string{rcv},
		"text":   []string{txtMsg.Text},
		"secret": []string{s.client.Credentials.ApiKey},
	}

	res, err := http.PostForm(reqUrl, form)
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	err = res.Body.Close()
	if err != nil {
		return "", err
	}

	return string(body), nil
}

type EncryptionSettings struct {
	PrivateKey util.NaClKey
	PublicKey  util.NaClKey
}

type ThreemaEndToEndMessageSender struct {
	Encryption *EncryptionSettings
	client     ThreemaClient
	ThreemaMessageSender
}

func (s *ThreemaEndToEndMessageSender) SupportsPhoneOrEmail() bool {
	return false
}

func (s *ThreemaEndToEndMessageSender) SendMessage(rcv string, pubKey util.NaClKey, msg message.ThreemaMessage) (string, error) {
	b, err := message.PaddedData(msg)
	if err != nil {
		return "", err
	}

	enc, nonce, err := util.EncryptBytes(&b, pubKey, s.Encryption.PrivateKey)
	if err != nil {
		return "", err
	}

	if len(enc) > ThreemaMaxEncryptedMessageLength {
		err = fmt.Errorf("message is too large")
		return "", err
	}

	noDeliveryReceipts := ternary(msg.HasDeliveryReceipt(), "0", "1")
	noPush := ternary(msg.HasPushNotification(), "0", "1")
	group := ternary(msg.HasGroupFlag(), "1", "0")

	reqUrl := fmt.Sprintf("%s/send_e2e", s.client.ApiUrl)
	form := url.Values{
		"from":               []string{s.client.Credentials.ThreemaId},
		"to":                 []string{rcv},
		"nonce":              []string{hex.EncodeToString(nonce)},
		"box":                []string{hex.EncodeToString(enc)},
		"secret":             []string{s.client.Credentials.ApiKey},
		"noDeliveryReceipts": []string{noDeliveryReceipts},
		"noPush":             []string{noPush},
		"group":              []string{group},
	}

	res, err := http.PostForm(reqUrl, form)
	if err != nil {
		return "", err
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return "", err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	err = res.Body.Close()
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func ternary[T interface{}](c bool, t T, f T) T {
	if c {
		return t
	}

	return f
}
