// Copyright 2022 CloudWeGo Authors
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

package meta

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
)

// Protcol magic numbers.
const (
	VersionMask = 0xffff0000
	Version1    = 0x80010000

	DefaultMaxDepth = 64
	bufferSize      = 64
)

var (
	errInvalidSize    = errors.New("invalid data length")
	errMissingVersion = errors.New("ReadMessageBegin: missing version")
	errBadVersion     = errors.New("ReadMessageBegin: bad version")
	errTooDeep        = errors.New("depth limit exceeded")
)

// BinaryProtocol implements the binary protocol of thrift.
type BinaryProtocol struct {
	transport   RichTransport
	buffer      [bufferSize]byte
	strictWrite bool
	strictRead  bool
}

// NewBinaryProtocol return a BinaryProtocol over the given transport.
func NewBinaryProtocol(t Transport) *BinaryProtocol {
	return &BinaryProtocol{
		transport: MakeRichTransport(t),
	}
}

// WithStrictRead forces the BinaryProtocol to do strict read.
func (p *BinaryProtocol) WithStrictRead() *BinaryProtocol {
	p.strictRead = true
	return p
}

// WithStrictWrite forces the BinaryProtocol to do strict write.
func (p *BinaryProtocol) WithStrictWrite() *BinaryProtocol {
	p.strictWrite = true
	return p
}

func (p *BinaryProtocol) readTypeID(ctx context.Context) (TTypeID, error) {
	bite, err := p.transport.ReadByte()
	if err != nil {
		return 0, err
	}
	return TTypeID(bite), nil
}

func (p *BinaryProtocol) readLength(ctx context.Context) (int, error) {
	size, err := p.ReadI32(ctx)
	if err != nil {
		return 0, err
	}

	if size < 0 {
		return 0, errInvalidSize
	}
	return int(size), nil
}

func (p *BinaryProtocol) readStringWithSize(ctx context.Context, size int) (string, error) {
	if size <= 0 {
		return "", nil
	}
	var res bytes.Buffer
	buf := p.buffer[:]
	for size > 0 {
		if size <= bufferSize {
			buf = p.buffer[:size]
		}
		n, err := io.ReadFull(p.transport, buf)
		if err != nil {
			return "", err
		}
		res.Write(buf)
		size -= n
	}
	return res.String(), nil
}

// ReadMessageBegin .
func (p *BinaryProtocol) ReadMessageBegin(ctx context.Context) (name string, typeID TMessageType, seqID int32, err error) {
	var size int32
	size, err = p.ReadI32(ctx)
	if err != nil {
		return
	}

	if size >= 0 {
		if p.strictRead {
			err = errMissingVersion
			return
		}

		if name, err = p.readStringWithSize(ctx, int(size)); err != nil {
			return
		}
		if b, e := p.ReadByte(ctx); e == nil {
			typeID = TMessageType(b)
		} else {
			err = e
			return
		}
		seqID, err = p.ReadI32(ctx)
		return
	}

	typeID = TMessageType(size & 0x0ff)
	if version := int64(size) & VersionMask; version != Version1 {
		err = errBadVersion
		return
	}
	name, err = p.ReadString(ctx)
	if err != nil {
		return
	}

	seqID, err = p.ReadI32(ctx)
	return
}

// ReadMessageEnd .
func (p *BinaryProtocol) ReadMessageEnd(ctx context.Context) error {
	return nil
}

// ReadBool .
func (p *BinaryProtocol) ReadBool(ctx context.Context) (bool, error) {
	bite, err := p.transport.ReadByte()
	return bite == 1, err
}

// ReadByte .
func (p *BinaryProtocol) ReadByte(ctx context.Context) (int8, error) {
	bite, err := p.transport.ReadByte()
	return int8(bite), err
}

// ReadI16 .
func (p *BinaryProtocol) ReadI16(ctx context.Context) (value int16, err error) {
	buf := p.buffer[0:2]
	_, err = io.ReadFull(p.transport, buf)
	value = int16(binary.BigEndian.Uint16(buf))
	return
}

// ReadI32 .
func (p *BinaryProtocol) ReadI32(ctx context.Context) (value int32, err error) {
	buf := p.buffer[0:4]
	_, err = io.ReadFull(p.transport, buf)
	value = int32(binary.BigEndian.Uint32(buf))
	return
}

// ReadI64 .
func (p *BinaryProtocol) ReadI64(ctx context.Context) (value int64, err error) {
	buf := p.buffer[0:8]
	_, err = io.ReadFull(p.transport, buf)
	value = int64(binary.BigEndian.Uint64(buf))
	return
}

// ReadDouble .
func (p *BinaryProtocol) ReadDouble(ctx context.Context) (value float64, err error) {
	buf := p.buffer[0:8]
	_, err = io.ReadFull(p.transport, buf)
	value = math.Float64frombits(binary.BigEndian.Uint64(buf))
	return
}

// ReadString .
func (p *BinaryProtocol) ReadString(ctx context.Context) (string, error) {
	size, err := p.readLength(ctx)
	if err != nil {
		return "", err
	}
	return p.readStringWithSize(ctx, size)
}

// ReadBinary .
func (p *BinaryProtocol) ReadBinary(ctx context.Context) ([]byte, error) {
	size, err := p.readLength(ctx)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, size)
	_, err = io.ReadFull(p.transport, buf)
	return buf, err
}

// ReadMapBegin .
func (p *BinaryProtocol) ReadMapBegin(ctx context.Context) (keyType, valueType TTypeID, size int, err error) {
	if keyType, err = p.readTypeID(ctx); err != nil {
		return
	}
	if valueType, err = p.readTypeID(ctx); err != nil {
		return
	}
	size, err = p.readLength(ctx)
	return
}

// ReadMapEnd .
func (p *BinaryProtocol) ReadMapEnd(ctx context.Context) error {
	return nil
}

// ReadListBegin .
func (p *BinaryProtocol) ReadListBegin(ctx context.Context) (elemType TTypeID, size int, err error) {
	if elemType, err = p.readTypeID(ctx); err != nil {
		return
	}
	size, err = p.readLength(ctx)
	return
}

// ReadListEnd .
func (p *BinaryProtocol) ReadListEnd(ctx context.Context) error {
	return nil
}

// ReadSetBegin .
func (p *BinaryProtocol) ReadSetBegin(ctx context.Context) (elemType TTypeID, size int, err error) {
	return p.ReadListBegin(ctx)
}

// ReadSetEnd .
func (p *BinaryProtocol) ReadSetEnd(ctx context.Context) error {
	return nil
}

// ReadStructBegin .
func (p *BinaryProtocol) ReadStructBegin(ctx context.Context) (name string, err error) {
	return
}

// ReadStructEnd .
func (p *BinaryProtocol) ReadStructEnd(ctx context.Context) error {
	return nil
}

// ReadFieldBegin .
func (p *BinaryProtocol) ReadFieldBegin(ctx context.Context) (name string, typeID TTypeID, id int16, err error) {
	if typeID, err = p.readTypeID(ctx); err != nil {
		return
	}
	if typeID != TTypeID_STOP {
		id, err = p.ReadI16(ctx)
	}
	return
}

// ReadFieldEnd .
func (p *BinaryProtocol) ReadFieldEnd(ctx context.Context) error {
	return nil
}

// Skip .
func (p *BinaryProtocol) Skip(ctx context.Context, fieldType TTypeID) (err error) {
	return Skip(ctx, p, fieldType, DefaultMaxDepth)
}

// WriteMessageBegin .
func (p *BinaryProtocol) WriteMessageBegin(ctx context.Context, name string, typeID TMessageType, seqID int32) (err error) {
	if p.strictWrite {
		if err = p.WriteI32(ctx, int32(Version1|int64(typeID))); err == nil {
			if err = p.WriteString(ctx, name); err == nil {
				err = p.WriteI32(ctx, seqID)
			}
		}
	} else {
		if err = p.WriteString(ctx, name); err == nil {
			if err = p.WriteByte(ctx, int8(typeID)); err == nil {
				err = p.WriteI32(ctx, seqID)
			}
		}
	}
	return
}

// WriteMessageEnd .
func (p *BinaryProtocol) WriteMessageEnd(ctx context.Context) error {
	return nil
}

// WriteBool .
func (p *BinaryProtocol) WriteBool(ctx context.Context, value bool) error {
	if value {
		return p.WriteByte(ctx, 1)
	}
	return p.WriteByte(ctx, 0)
}

// WriteByte .
func (p *BinaryProtocol) WriteByte(ctx context.Context, value int8) error {
	return p.transport.WriteByte(byte(value))
}

// WriteI16 .
func (p *BinaryProtocol) WriteI16(ctx context.Context, value int16) error {
	buf := p.buffer[0:2]
	binary.BigEndian.PutUint16(buf, uint16(value))
	_, err := p.transport.Write(buf)
	return err
}

// WriteI32 .
func (p *BinaryProtocol) WriteI32(ctx context.Context, value int32) error {
	buf := p.buffer[0:4]
	binary.BigEndian.PutUint32(buf, uint32(value))
	_, err := p.transport.Write(buf)
	return err
}

// WriteI64 .
func (p *BinaryProtocol) WriteI64(ctx context.Context, value int64) error {
	buf := p.buffer[0:8]
	binary.BigEndian.PutUint64(buf, uint64(value))
	_, err := p.transport.Write(buf)
	return err
}

// WriteDouble .
func (p *BinaryProtocol) WriteDouble(ctx context.Context, value float64) error {
	return p.WriteI64(ctx, int64(math.Float64bits(value)))
}

// WriteString .
func (p *BinaryProtocol) WriteString(ctx context.Context, value string) (err error) {
	if err = p.WriteI32(ctx, int32(len(value))); err == nil {
		_, err = p.transport.WriteString(value)
	}
	return
}

// WriteBinary .
func (p *BinaryProtocol) WriteBinary(ctx context.Context, value []byte) (err error) {
	if err = p.WriteI32(ctx, int32(len(value))); err == nil {
		_, err = p.transport.Write(value)
	}
	return
}

// WriteMapBegin .
func (p *BinaryProtocol) WriteMapBegin(ctx context.Context, keyType, valueType TTypeID, size int) (err error) {
	if err = p.WriteByte(ctx, int8(keyType)); err == nil {
		if err = p.WriteByte(ctx, int8(valueType)); err == nil {
			err = p.WriteI32(ctx, int32(size))
		}
	}
	return
}

// WriteMapEnd .
func (p *BinaryProtocol) WriteMapEnd(ctx context.Context) error {
	return nil
}

// WriteListBegin .
func (p *BinaryProtocol) WriteListBegin(ctx context.Context, elemType TTypeID, size int) (err error) {
	if err = p.WriteByte(ctx, int8(elemType)); err == nil {
		err = p.WriteI32(ctx, int32(size))
	}
	return
}

// WriteListEnd .
func (p *BinaryProtocol) WriteListEnd(ctx context.Context) error {
	return nil
}

// WriteSetBegin .
func (p *BinaryProtocol) WriteSetBegin(ctx context.Context, elemType TTypeID, size int) (err error) {
	if err = p.WriteByte(ctx, int8(elemType)); err == nil {
		err = p.WriteI32(ctx, int32(size))
	}
	return
}

// WriteSetEnd .
func (p *BinaryProtocol) WriteSetEnd(ctx context.Context) error {
	return nil
}

// WriteStructBegin .
func (p *BinaryProtocol) WriteStructBegin(ctx context.Context, name string) error {
	return nil
}

// WriteStructEnd .
func (p *BinaryProtocol) WriteStructEnd(ctx context.Context) error {
	return nil
}

// WriteFieldBegin .
func (p *BinaryProtocol) WriteFieldBegin(ctx context.Context, name string, typeID TTypeID, id int16) (err error) {
	if err = p.WriteByte(ctx, int8(typeID)); err == nil {
		err = p.WriteI16(ctx, id)
	}
	return
}

// WriteFieldEnd .
func (p *BinaryProtocol) WriteFieldEnd(ctx context.Context) error {
	return nil
}

// WriteFieldStop .
func (p *BinaryProtocol) WriteFieldStop(ctx context.Context) error {
	return p.WriteByte(ctx, int8(TTypeID_STOP))
}

// Flush .
func (p *BinaryProtocol) Flush(ctx context.Context) (err error) {
	return p.transport.Flush(ctx)
}

// Skip skips a thrift type with the given protocol.
func Skip(ctx context.Context, iprot Protocol, typeID TTypeID, maxDepth int) (err error) {
	if maxDepth <= 0 {
		return errTooDeep
	}

	switch typeID {
	case TTypeID_BOOL:
		_, err = iprot.ReadBool(ctx)
	case TTypeID_BYTE:
		_, err = iprot.ReadByte(ctx)
	case TTypeID_I16:
		_, err = iprot.ReadI16(ctx)
	case TTypeID_I32:
		_, err = iprot.ReadI32(ctx)
	case TTypeID_I64:
		_, err = iprot.ReadI64(ctx)
	case TTypeID_DOUBLE:
		_, err = iprot.ReadDouble(ctx)
	case TTypeID_STRING:
		_, err = iprot.ReadString(ctx)
	case TTypeID_STRUCT:
		if _, err = iprot.ReadStructBegin(ctx); err != nil {
			return
		}
		for {
			if _, typeID, _, err = iprot.ReadFieldBegin(ctx); err != nil {
				return
			}
			if typeID == TTypeID_STOP {
				break
			}
			if err = Skip(ctx, iprot, typeID, maxDepth-1); err != nil {
				return
			}
			if err = iprot.ReadFieldEnd(ctx); err != nil {
				return
			}
		}
		return iprot.ReadStructEnd(ctx)
	case TTypeID_MAP:
		keyType, valueType, size, err := iprot.ReadMapBegin(ctx)
		if err != nil {
			return err
		}
		for i := 0; i < size; i++ {
			if err = Skip(ctx, iprot, keyType, maxDepth-1); err != nil {
				return err
			}

			if err = Skip(ctx, iprot, valueType, maxDepth-1); err != nil {
				return err
			}
		}
		return iprot.ReadMapEnd(ctx)
	case TTypeID_SET:
		elemType, size, err := iprot.ReadSetBegin(ctx)
		if err != nil {
			return err
		}
		for i := 0; i < size; i++ {
			if err := Skip(ctx, iprot, elemType, maxDepth-1); err != nil {
				return err
			}
		}
		return iprot.ReadSetEnd(ctx)
	case TTypeID_LIST:
		elemType, size, err := iprot.ReadListBegin(ctx)
		if err != nil {
			return err
		}
		for i := 0; i < size; i++ {
			if err := Skip(ctx, iprot, elemType, maxDepth-1); err != nil {
				return err
			}
		}
		return iprot.ReadListEnd(ctx)
	default:
		return fmt.Errorf("Unknown data type %d", typeID)
	}
	return
}
