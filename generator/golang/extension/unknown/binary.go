// Copyright 2023 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package unknown .
package unknown

import (
	"encoding/binary"
	"errors"
	"math"
)

var InvalidDataLength = errors.New("invalid data length")

// Binary protocol for bthrift.
var Binary binaryProtocol

type binaryProtocol struct{}

func (binaryProtocol) WriteStructBegin(buf []byte, name string) int {
	return 0
}

func (binaryProtocol) WriteStructEnd(buf []byte) int {
	return 0
}

func (binaryProtocol) WriteFieldBegin(buf []byte, name string, typeID int, id int16) int {
	return Binary.WriteByte(buf, int8(typeID)) + Binary.WriteI16(buf[1:], id)
}

func (binaryProtocol) WriteFieldEnd(buf []byte) int {
	return 0
}

func (binaryProtocol) WriteFieldStop(buf []byte) int {
	return Binary.WriteByte(buf, TStop)
}

func (binaryProtocol) WriteMapBegin(buf []byte, keyType, valueType, size int) int {
	return Binary.WriteByte(buf, int8(keyType)) +
		Binary.WriteByte(buf[1:], int8(valueType)) +
		Binary.WriteI32(buf[2:], int32(size))
}

func (binaryProtocol) WriteMapEnd(buf []byte) int {
	return 0
}

func (binaryProtocol) WriteListBegin(buf []byte, elemType, size int) int {
	return Binary.WriteByte(buf, int8(elemType)) +
		Binary.WriteI32(buf[1:], int32(size))
}

func (binaryProtocol) WriteListEnd(buf []byte) int {
	return 0
}

func (binaryProtocol) WriteSetBegin(buf []byte, elemType, size int) int {
	return Binary.WriteByte(buf, int8(elemType)) +
		Binary.WriteI32(buf[1:], int32(size))
}

func (binaryProtocol) WriteSetEnd(buf []byte) int {
	return 0
}

func (binaryProtocol) WriteBool(buf []byte, value bool) int {
	if value {
		return Binary.WriteByte(buf, 1)
	}
	return Binary.WriteByte(buf, 0)
}

func (binaryProtocol) WriteByte(buf []byte, value int8) int {
	buf[0] = byte(value)
	return 1
}

func (binaryProtocol) WriteI16(buf []byte, value int16) int {
	binary.BigEndian.PutUint16(buf, uint16(value))
	return 2
}

func (binaryProtocol) WriteI32(buf []byte, value int32) int {
	binary.BigEndian.PutUint32(buf, uint32(value))
	return 4
}

func (binaryProtocol) WriteI64(buf []byte, value int64) int {
	binary.BigEndian.PutUint64(buf, uint64(value))
	return 8
}

func (binaryProtocol) WriteDouble(buf []byte, value float64) int {
	return Binary.WriteI64(buf, int64(math.Float64bits(value)))
}

func (binaryProtocol) WriteString(buf []byte, value string) int {
	l := Binary.WriteI32(buf, int32(len(value)))
	copy(buf[l:], value)
	return l + len(value)
}

func (binaryProtocol) WriteBinary(buf, value []byte) int {
	l := Binary.WriteI32(buf, int32(len(value)))
	copy(buf[l:], value)
	return l + len(value)
}

func (binaryProtocol) StructBeginLength(name string) int {
	return 0
}

func (binaryProtocol) StructEndLength() int {
	return 0
}

func (binaryProtocol) FieldBeginLength(name string, typeID int, id int16) int {
	return Binary.ByteLength(int8(typeID)) + Binary.I16Length(id)
}

func (binaryProtocol) FieldEndLength() int {
	return 0
}

func (binaryProtocol) FieldStopLength() int {
	return Binary.ByteLength(TStop)
}

func (binaryProtocol) MapBeginLength(keyType, valueType, size int) int {
	return Binary.ByteLength(int8(keyType)) +
		Binary.ByteLength(int8(valueType)) +
		Binary.I32Length(int32(size))
}

func (binaryProtocol) MapEndLength() int {
	return 0
}

func (binaryProtocol) ListBeginLength(elemType, size int) int {
	return Binary.ByteLength(int8(elemType)) +
		Binary.I32Length(int32(size))
}

func (binaryProtocol) ListEndLength() int {
	return 0
}

func (binaryProtocol) SetBeginLength(elemType, size int) int {
	return Binary.ByteLength(int8(elemType)) +
		Binary.I32Length(int32(size))
}

func (binaryProtocol) SetEndLength() int {
	return 0
}

func (binaryProtocol) BoolLength(value bool) int {
	if value {
		return Binary.ByteLength(1)
	}
	return Binary.ByteLength(0)
}

func (binaryProtocol) ByteLength(value int8) int {
	return 1
}

func (binaryProtocol) I16Length(value int16) int {
	return 2
}

func (binaryProtocol) I32Length(value int32) int {
	return 4
}

func (binaryProtocol) I64Length(value int64) int {
	return 8
}

func (binaryProtocol) DoubleLength(value float64) int {
	return Binary.I64Length(int64(math.Float64bits(value)))
}

func (binaryProtocol) StringLength(value string) int {
	return Binary.I32Length(int32(len(value))) + len(value)
}

func (binaryProtocol) BinaryLength(value []byte) int {
	return Binary.I32Length(int32(len(value))) + len(value)
}

func (binaryProtocol) ReadMessageEnd(buf []byte) (int, error) {
	return 0, nil
}

func (binaryProtocol) ReadStructBegin(buf []byte) (name string, length int, err error) {
	return
}

func (binaryProtocol) ReadStructEnd(buf []byte) (int, error) {
	return 0, nil
}

func (binaryProtocol) ReadFieldBegin(buf []byte) (name string, typeID int, id int16, length int, err error) {
	t, l, e := Binary.ReadByte(buf)
	length += l
	typeID = int(t)
	if e != nil {
		err = e
		return
	}
	if typeID != TStop {
		id, l, err = Binary.ReadI16(buf[length:])
		length += l
	}
	return
}

func (binaryProtocol) ReadFieldEnd(buf []byte) (int, error) {
	return 0, nil
}

func (binaryProtocol) ReadMapBegin(buf []byte) (keyType, valueType, size, length int, err error) {
	k, l, e := Binary.ReadByte(buf)
	length += l
	if e != nil {
		err = e
		return
	}
	keyType = int(k)
	v, l, e := Binary.ReadByte(buf[length:])
	length += l
	if e != nil {
		err = e
		return
	}
	valueType = int(v)
	size32, l, e := Binary.ReadI32(buf[length:])
	length += l
	if e != nil {
		err = e
		return
	}
	if size32 < 0 {
		err = InvalidDataLength
		return
	}
	size = int(size32)
	return
}

func (binaryProtocol) ReadMapEnd(buf []byte) (int, error) {
	return 0, nil
}

func (binaryProtocol) ReadListBegin(buf []byte) (elemType, size, length int, err error) {
	b, l, e := Binary.ReadByte(buf)
	length += l
	if e != nil {
		err = e
		return
	}
	elemType = int(b)
	size32, l, e := Binary.ReadI32(buf[length:])
	length += l
	if e != nil {
		err = e
		return
	}
	if size32 < 0 {
		err = InvalidDataLength
		return
	}
	size = int(size32)

	return
}

func (binaryProtocol) ReadListEnd(buf []byte) (int, error) {
	return 0, nil
}

func (binaryProtocol) ReadSetBegin(buf []byte) (elemType, size, length int, err error) {
	b, l, e := Binary.ReadByte(buf)
	length += l
	if e != nil {
		err = e
		return
	}
	elemType = int(b)
	size32, l, e := Binary.ReadI32(buf[length:])
	length += l
	if e != nil {
		err = e
		return
	}
	if size32 < 0 {
		err = InvalidDataLength
		return
	}
	size = int(size32)
	return
}

func (binaryProtocol) ReadSetEnd(buf []byte) (int, error) {
	return 0, nil
}

func (binaryProtocol) ReadBool(buf []byte) (value bool, length int, err error) {
	b, l, e := Binary.ReadByte(buf)
	v := true
	if b != 1 {
		v = false
	}
	return v, l, e
}

func (binaryProtocol) ReadByte(buf []byte) (value int8, length int, err error) {
	if len(buf) < 1 {
		return value, length, InvalidDataLength
	}
	return int8(buf[0]), 1, err
}

func (binaryProtocol) ReadI16(buf []byte) (value int16, length int, err error) {
	if len(buf) < 2 {
		return value, length, InvalidDataLength
	}
	value = int16(binary.BigEndian.Uint16(buf))
	return value, 2, err
}

func (binaryProtocol) ReadI32(buf []byte) (value int32, length int, err error) {
	if len(buf) < 4 {
		return value, length, InvalidDataLength
	}
	value = int32(binary.BigEndian.Uint32(buf))
	return value, 4, err
}

func (binaryProtocol) ReadI64(buf []byte) (value int64, length int, err error) {
	if len(buf) < 8 {
		return value, length, InvalidDataLength
	}
	value = int64(binary.BigEndian.Uint64(buf))
	return value, 8, err
}

func (binaryProtocol) ReadDouble(buf []byte) (value float64, length int, err error) {
	if len(buf) < 8 {
		return value, length, InvalidDataLength
	}
	value = math.Float64frombits(binary.BigEndian.Uint64(buf))
	return value, 8, err
}

func (binaryProtocol) ReadString(buf []byte) (value string, length int, err error) {
	size, l, e := Binary.ReadI32(buf)
	length += l
	if e != nil {
		err = e
		return
	}
	if size < 0 || int(size) > len(buf) {
		return value, length, InvalidDataLength
	}
	value = string(buf[length : length+int(size)])
	length += int(size)
	return
}

func (binaryProtocol) ReadBinary(buf []byte) (value []byte, length int, err error) {
	size, l, e := Binary.ReadI32(buf)
	length += l
	if e != nil {
		err = e
		return
	}
	if size < 0 || int(size) > len(buf) {
		return value, length, InvalidDataLength
	}
	value = make([]byte, size)
	copy(value, buf[length:length+int(size)])
	length += int(size)
	return
}
