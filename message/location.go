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
	"fmt"
	"strings"
)

type LocationMessage struct {
	Latitude  float64
	Longitude float64
	Accuracy  float64
	Address   string
	Name      string
}

func (l *LocationMessage) MessageType() uint8 {
	return LocationMessageId
}

func (l *LocationMessage) HasDeliveryReceipt() bool {
	return true
}

func (l *LocationMessage) HasPushNotification() bool {
	return true
}

func (l *LocationMessage) HasGroupFlag() bool {
	return false
}

func (l *LocationMessage) Data() []byte {
	dataString := fmt.Sprintf("%f,%f", l.Latitude, l.Longitude)
	if l.Accuracy > 0 {
		dataString = fmt.Sprintf("%s,%f", dataString, l.Accuracy)
	}

	if l.Name != "" {
		dataString = fmt.Sprintf("%s\n%s", dataString, l.Name)
	}

	if l.Address != "" {
		l.Address = strings.ReplaceAll(l.Address, "\n", "\\n")
		dataString = fmt.Sprintf("%s\n%s", dataString, l.Address)
	}

	return []byte(dataString)
}

func NewLocationMessage(Latitude float64, Longitude float64) *LocationMessage {
	return &LocationMessage{
		Latitude:  Latitude,
		Longitude: Longitude,
		Accuracy:  -1,
		Address:   "",
		Name:      "",
	}
}
