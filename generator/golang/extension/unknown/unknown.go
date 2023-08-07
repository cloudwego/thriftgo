// Copyright 2021 CloudWeGo Authors
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

// Package unknown provides definitions that work with the thriftgo `keep_unknown_fields` option.
// When the option is turned on, thriftgo generates an extra field `_unknownFields` for each structure
// defined in the IDL to store fields that are not recognized by the current IDL when deserializing data.
// Those unknown fields will be written out at the end of the data stream when serializing the structure
// that carries them.
package unknown

import (
	"errors"
	"fmt"
)

// WithUnknownFields is the interface of all structures that supports keeping unknown fields.
type WithUnknownFields interface {
	// CarryingUnknownFields tells whether the structure contains data from fields not recognized by the current IDL.
	CarryingUnknownFields() bool
}

// errors .
var (
	ErrExceedDepthLimit = errors.New("depth limit exceeded")

	ErrUnknownType = func(t int) error {
		return fmt.Errorf("unknown data type %d", t)
	}

	maxNestingDepth = 64
)

// SetNestingDepthLimit sets the max number of nesting level.
func SetNestingDepthLimit(d int) {
	maxNestingDepth = d
}

// Fields stores all undeserialized unknown fields.
type Fields []byte

// Append reads an object of a generalized type from xprot and serializes the object into Fields for compatibility
// with the thrift interface, and the performance is greatly discounted for this reason.
//
// [Deprecated]: Use the FastCodec api provided by Kitex for serialization/deserialization to improve performance.
func (fs *Fields) Append(xprot TProtocol, name string, fieldType TType, id int16) error {
	iprot, err := convert(xprot)
	if err != nil {
		return err
	}
	ft := asInt(fieldType)
	buf := ([]byte)(*fs)[:cap(*fs)]
	offset := len(*fs)
	ensureBytesLen(&buf, offset, Binary.FieldBeginLength(name, ft, id))
	offset += Binary.WriteFieldBegin(buf[offset:], name, ft, id)
	offset, err = read(&buf, offset, iprot, name, ft, id, maxNestingDepth)
	*fs = buf[:offset]
	return err
}

// Write reads an object of a generalized type from Fields and srializes the object into xprot for compatibility
// with the thrift interface, and the performance is greatly discounted for this reason.
//
// [Deprecated]: Use the FastCodec api provided by Kitex for serialization/deserialization to improve performance.
func (fs *Fields) Write(xprot TProtocol) (err error) {
	oprot, err := convert(xprot)
	if err != nil {
		return err
	}
	rbuf := []byte(*fs)
	var offset int
	for offset < len(rbuf) {
		name, fieldType, fieldID, l, err := Binary.ReadFieldBegin(rbuf[offset:])
		offset += l
		if err != nil {
			return fmt.Errorf("read field begin error: %w", err)
		}
		if err = oprot.WriteFieldBegin(ctx, name, fieldType, fieldID); err != nil {
			return fmt.Errorf("write field begin error: %w", err)
		}

		l, err = write(oprot, name, fieldType, fieldID, rbuf[offset:])
		offset += l
		if err != nil {
			return fmt.Errorf("write struct field error: %w", err)
		}

		l, err = Binary.ReadFieldEnd(rbuf[offset:])
		offset += l
		if err != nil {
			return fmt.Errorf("read field end error: %w", err)
		}
		if err = oprot.WriteFieldEnd(ctx); err != nil {
			return fmt.Errorf("write field end error: %w", err)
		}
	}
	if err != nil {
		err = fmt.Errorf("write field error unknown: %w", err)
	}
	return err
}

// write writes fields out the oprot.
func write(oprot *protocol, name string, fieldType int, id int16, fs []byte) (offset int, err error) {
	switch fieldType {
	case TBool:
		v, l, err := Binary.ReadBool(fs[offset:])
		offset += l
		if err != nil {
			return offset, err
		}
		if err = oprot.WriteBool(ctx, v); err != nil {
			return offset, err
		}
	case TByte:
		v, l, err := Binary.ReadByte(fs[offset:])
		offset += l
		if err != nil {
			return offset, err
		}
		if err = oprot.WriteByte(ctx, v); err != nil {
			return offset, err
		}
	case TI16:
		v, l, err := Binary.ReadI16(fs[offset:])
		offset += l
		if err != nil {
			return offset, err
		}
		if err = oprot.WriteI16(ctx, v); err != nil {
			return offset, err
		}
	case TI32:
		v, l, err := Binary.ReadI32(fs[offset:])
		offset += l
		if err != nil {
			return offset, err
		}
		if err = oprot.WriteI32(ctx, v); err != nil {
			return offset, err
		}
	case TI64:
		v, l, err := Binary.ReadI64(fs[offset:])
		offset += l
		if err != nil {
			return offset, err
		}
		if err = oprot.WriteI64(ctx, v); err != nil {
			return offset, err
		}
	case TDouble:
		v, l, err := Binary.ReadDouble(fs[offset:])
		offset += l
		if err != nil {
			return offset, err
		}
		if err = oprot.WriteDouble(ctx, v); err != nil {
			return offset, err
		}
	case TString:
		v, l, err := Binary.ReadString(fs[offset:])
		offset += l
		if err != nil {
			return offset, err
		}
		if err = oprot.WriteString(ctx, v); err != nil {
			return offset, err
		}
	case TSet:
		ttype, size, l, err := Binary.ReadSetBegin(fs[offset:])
		offset += l
		if err != nil {
			return offset, fmt.Errorf("read set begin error: %w", err)
		}
		if err = oprot.WriteSetBegin(ctx, ttype, size); err != nil {
			return offset, fmt.Errorf("write set begin error: %w", err)
		}
		for i := 0; i < size; i++ {
			l, err = write(oprot, "", ttype, int16(i), fs[offset:])
			offset += l
			if err != nil {
				return offset, fmt.Errorf("write set elem error: %w", err)
			}
		}
		l, err = Binary.ReadSetEnd(fs[offset:])
		offset += l
		if err != nil {
			return offset, fmt.Errorf("read set end error: %w", err)
		}
		if err = oprot.WriteSetEnd(ctx); err != nil {
			return offset, fmt.Errorf("write set end error: %w", err)
		}
	case TList:
		ttype, size, l, err := Binary.ReadListBegin(fs[offset:])
		offset += l
		if err != nil {
			return offset, fmt.Errorf("read list begin error: %w", err)
		}
		if err = oprot.WriteListBegin(ctx, ttype, size); err != nil {
			return offset, err
		}
		for i := 0; i < size; i++ {
			l, err = write(oprot, "", ttype, int16(i), fs[offset:])
			offset += l
			if err != nil {
				return offset, fmt.Errorf("write list elem error: %w", err)
			}
		}
		l, err = Binary.ReadListEnd(fs[offset:])
		offset += l
		if err != nil {
			return offset, fmt.Errorf("read list end error: %w", err)
		}
		if err = oprot.WriteListEnd(ctx); err != nil {
			return offset, fmt.Errorf("write list end error: %w", err)
		}
	case TMap:
		kttype, vttype, size, l, err := Binary.ReadMapBegin(fs[offset:])
		offset += l
		if err != nil {
			return offset, fmt.Errorf("read map begin error: %w", err)
		}
		if err = oprot.WriteMapBegin(ctx, kttype, vttype, size); err != nil {
			return offset, fmt.Errorf("write map begin error: %w", err)
		}
		for i := 0; i < size; i++ {
			l, err = write(oprot, "", kttype, int16(i), fs[offset:])
			offset += l
			if err != nil {
				return offset, fmt.Errorf("write map key error: %w", err)
			}
			l, err = write(oprot, "", vttype, int16(i), fs[offset:])
			offset += l
			if err != nil {
				return offset, fmt.Errorf("write map value error: %w", err)
			}
		}
		l, err = Binary.ReadMapEnd(fs[offset:])
		offset += l
		if err != nil {
			return offset, fmt.Errorf("read map end error: %w", err)
		}
		if err = oprot.WriteMapEnd(ctx); err != nil {
			return offset, fmt.Errorf("write map end error: %w", err)
		}
	case TStruct:
		_, l, err := Binary.ReadStructBegin(fs[offset:])
		offset += l
		if err != nil {
			return offset, fmt.Errorf("read struct begin error: %w", err)
		}
		for {
			name, fieldTypeID, fieldID, l, err := Binary.ReadFieldBegin(fs[offset:])
			offset += l
			if err != nil {
				return offset, fmt.Errorf("read field begin error: %w", err)
			}
			if fieldTypeID == TStop {
				if err = oprot.WriteFieldStop(ctx); err != nil {
					return offset, fmt.Errorf("write field stop error: %w", err)
				}
				break
			}
			if err = oprot.WriteFieldBegin(ctx, name, fieldTypeID, fieldID); err != nil {
				return offset, fmt.Errorf("write field begin error: %w", err)
			}
			l, err = write(oprot, name, fieldTypeID, fieldID, fs[offset:])
			offset += l
			if err != nil {
				return offset, fmt.Errorf("write struct field error: %w", err)
			}
			l, err = Binary.ReadFieldEnd(fs[offset:])
			offset += l
			if err != nil {
				return offset, fmt.Errorf("read field end error: %w", err)
			}
			if err = oprot.WriteFieldEnd(ctx); err != nil {
				return offset, fmt.Errorf("write field end error: %w", err)
			}
		}
		l, err = Binary.ReadStructEnd(fs[offset:])
		offset += l
		if err != nil {
			return offset, fmt.Errorf("read struct end error: %w", err)
		}
		if err = oprot.WriteStructEnd(ctx); err != nil {
			return offset, fmt.Errorf("write struct end error: %w", err)
		}
	default:
		return offset, ErrUnknownType(fieldType)
	}
	return offset, nil
}

// read reads an unknown field from the given TProtocol.
func read(buf *[]byte, offset int, iprot *protocol, name string, fieldType int, id int16, maxDepth int) (noffset int, err error) {
	if maxDepth <= 0 {
		return offset, ErrExceedDepthLimit
	}
	switch fieldType {
	case TBool:
		var v bool
		v, err = iprot.ReadBool(ctx)
		ensureBytesLen(buf, offset, Binary.BoolLength(v))
		offset += Binary.WriteBool((*buf)[offset:], v)
	case TByte:
		var v int8
		v, err = iprot.ReadByte(ctx)
		ensureBytesLen(buf, offset, Binary.ByteLength(v))
		offset += Binary.WriteByte((*buf)[offset:], v)
	case TI16:
		var v int16
		v, err = iprot.ReadI16(ctx)
		ensureBytesLen(buf, offset, Binary.I16Length(v))
		offset += Binary.WriteI16((*buf)[offset:], v)
	case TI32:
		var v int32
		v, err = iprot.ReadI32(ctx)
		ensureBytesLen(buf, offset, Binary.I32Length(v))
		offset += Binary.WriteI32((*buf)[offset:], v)
	case TI64:
		var v int64
		v, err = iprot.ReadI64(ctx)
		ensureBytesLen(buf, offset, Binary.I64Length(v))
		offset += Binary.WriteI64((*buf)[offset:], v)
	case TDouble:
		var v float64
		v, err = iprot.ReadDouble(ctx)
		ensureBytesLen(buf, offset, Binary.DoubleLength(v))
		offset += Binary.WriteDouble((*buf)[offset:], v)
	case TString:
		var v string
		v, err = iprot.ReadString(ctx)
		ensureBytesLen(buf, offset, Binary.StringLength(v))
		offset += Binary.WriteString((*buf)[offset:], v)
	case TSet:
		var valType int
		var size int
		valType, size, err = iprot.ReadSetBegin(ctx)
		if err != nil {
			return offset, fmt.Errorf("read set begin error: %w", err)
		}
		ensureBytesLen(buf, offset, Binary.SetBeginLength(valType, size))
		offset += Binary.WriteSetBegin((*buf)[offset:], valType, size)
		for i := 0; i < size; i++ {
			offset, err = read(buf, offset, iprot, "", valType, int16(i), maxDepth-1)
			if err != nil {
				return offset, fmt.Errorf("read set elem error: %w", err)
			}
		}
		if err = iprot.ReadSetEnd(ctx); err != nil {
			return offset, fmt.Errorf("read set end error: %w", err)
		}
		ensureBytesLen(buf, offset, Binary.SetEndLength())
		offset += Binary.WriteSetEnd((*buf)[offset:])
	case TList:
		var valType int
		var size int
		valType, size, err = iprot.ReadListBegin(ctx)
		if err != nil {
			return offset, fmt.Errorf("read list begin error: %w", err)
		}
		ensureBytesLen(buf, offset, Binary.ListBeginLength(valType, size))
		offset += Binary.WriteListBegin((*buf)[offset:], valType, size)
		for i := 0; i < size; i++ {
			offset, err = read(buf, offset, iprot, "", valType, int16(i), maxDepth-1)
			if err != nil {
				return offset, fmt.Errorf("read list elem error: %w", err)
			}
		}
		if err = iprot.ReadListEnd(ctx); err != nil {
			return offset, fmt.Errorf("read list end error: %w", err)
		}
		ensureBytesLen(buf, offset, Binary.ListEndLength())
		offset += Binary.WriteListEnd((*buf)[offset:])
	case TMap:
		var keyType, valType int
		var size int
		keyType, valType, size, err = iprot.ReadMapBegin(ctx)
		if err != nil {
			return offset, fmt.Errorf("read map begin error: %w", err)
		}
		ensureBytesLen(buf, offset, Binary.MapBeginLength(keyType, valType, size))
		offset += Binary.WriteMapBegin((*buf)[offset:], keyType, valType, size)
		for i := 0; i < size; i++ {
			offset, err = read(buf, offset, iprot, "", keyType, int16(i), maxDepth-1)
			if err != nil {
				return offset, fmt.Errorf("read map key error: %w", err)
			}
			offset, err = read(buf, offset, iprot, "", valType, int16(i), maxDepth-1)
			if err != nil {
				return offset, fmt.Errorf("read map value error: %w", err)
			}
		}
		if err = iprot.ReadMapEnd(ctx); err != nil {
			return offset, fmt.Errorf("read map end error: %w", err)
		}
		ensureBytesLen(buf, offset, Binary.MapEndLength())
		offset += Binary.WriteMapEnd((*buf)[offset:])
	case TStruct:
		name, err := iprot.ReadStructBegin(ctx)
		if err != nil {
			return offset, fmt.Errorf("read struct begin error: %w", err)
		}
		ensureBytesLen(buf, offset, Binary.StructBeginLength(name))
		offset += Binary.WriteStructBegin((*buf)[offset:], name)
		for {
			name, fieldTypeID, fieldID, err := iprot.ReadFieldBegin(ctx)
			if err != nil {
				return offset, fmt.Errorf("read field begin error: %w", err)
			}
			if fieldTypeID == TStop {
				ensureBytesLen(buf, offset, Binary.FieldStopLength())
				offset += Binary.WriteFieldStop((*buf)[offset:])
				break
			}
			ensureBytesLen(buf, offset, Binary.FieldBeginLength(name, fieldTypeID, fieldID))
			offset += Binary.WriteFieldBegin((*buf)[offset:], name, fieldTypeID, fieldID)
			offset, err = read(buf, offset, iprot, name, fieldTypeID, fieldID, maxDepth-1)
			if err != nil {
				return offset, fmt.Errorf("read struct field error: %w", err)
			}
			if err = iprot.ReadFieldEnd(ctx); err != nil {
				return offset, fmt.Errorf("read field end error: %w", err)
			}
			ensureBytesLen(buf, offset, Binary.FieldEndLength())
			offset += Binary.WriteFieldEnd((*buf)[offset:])
		}
		if err = iprot.ReadStructEnd(ctx); err != nil {
			return offset, fmt.Errorf("read struct end error: %w", err)
		}
		ensureBytesLen(buf, offset, Binary.StructEndLength())
		offset += Binary.WriteStructEnd((*buf)[offset:])
	default:
		return offset, ErrUnknownType(fieldType)
	}
	return offset, nil
}

func ensureBytesLen(buf *[]byte, offset, l int) {
	if len(*buf)-offset < l {
		nb := make([]byte, (offset+l)*2)
		copy(nb, (*buf)[:offset])
		*buf = nb
	}
}
