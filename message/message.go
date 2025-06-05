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

package message

import (
	"bytes"
	"crypto/rand"
	"math/big"
)

const (
	TextMessageId     = uint8(0x01)
	FileMessageId     = uint8(0x17)
	LocationMessageId = uint8(0x10)
	DeliveryReceiptId = uint8(0x80)
	PollSetupId       = uint8(0x15)
	PollVoteId        = uint8(0x16)
)

type ThreemaMessage interface {
	MessageType() uint8
	HasDeliveryReceipt() bool
	HasPushNotification() bool
	HasGroupFlag() bool
	Data() []byte
}

func PaddedData(msg ThreemaMessage) ([]byte, error) {
	paddingBytesB, err := rand.Int(rand.Reader, big.NewInt(254))
	if err != nil {
		return nil, err
	}
	paddingBytes := int(paddingBytesB.Int64()) + 1

	data := msg.Data()
	if 1+len(data)+paddingBytes < 32 {
		paddingBytes += 32 - (1 + len(data) + paddingBytes)
	}

	paddedData := make([]byte, 1+len(data)+paddingBytes)
	paddedData[0] = msg.MessageType()
	copy(paddedData[1:], data)
	copy(paddedData[1+len(data):], bytes.Repeat([]byte{byte(paddingBytes)}, paddingBytes))

	return paddedData, nil
}
