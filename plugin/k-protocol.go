// Code generated by thriftgo (0.4.1) (fastgo). DO NOT EDIT.
package plugin

import (
	"fmt"
	"unsafe"

	"github.com/cloudwego/gopkg/protocol/thrift"
	"github.com/cloudwego/thriftgo/parser"
)

var ThriftGoUnusedProtection = struct{}{}

var (
	_ = parser.ThriftGoUnusedProtection
)

func (p *Request) BLength() int {
	if p == nil {
		return 1
	}
	off := 0

	// p.Version ID:1 thrift.STRING
	off += 3
	off += 4 + len(p.Version)

	// p.GeneratorParameters ID:2 thrift.LIST
	off += 3
	off += 5
	for _, v := range p.GeneratorParameters {
		off += 4 + len(v)
	}

	// p.PluginParameters ID:3 thrift.LIST
	off += 3
	off += 5
	for _, v := range p.PluginParameters {
		off += 4 + len(v)
	}

	// p.Language ID:4 thrift.STRING
	off += 3
	off += 4 + len(p.Language)

	// p.OutputPath ID:5 thrift.STRING
	off += 3
	off += 4 + len(p.OutputPath)

	// p.Recursive ID:6 thrift.BOOL
	off += 3
	off += 1

	// p.AST ID:7 thrift.STRUCT
	off += 3
	off += p.AST.BLength()
	return off + 1
}

func (p *Request) FastWrite(b []byte) int { return p.FastWriteNocopy(b, nil) }

func (p *Request) FastWriteNocopy(b []byte, w thrift.NocopyWriter) (n int) {
	if n = len(p.FastAppend(b[:0])); n > len(b) {
		panic("buffer overflow. concurrency issue?")
	}
	return
}

func (p *Request) FastAppend(b []byte) []byte {
	if p == nil {
		return append(b, 0)
	}
	x := thrift.BinaryProtocol{}
	_ = x

	// p.Version
	b = append(b, 11, 0, 1)
	b = x.AppendI32(b, int32(len(p.Version)))
	b = append(b, p.Version...)

	// p.GeneratorParameters
	b = append(b, 15, 0, 2)
	b = x.AppendListBegin(b, thrift.STRING, len(p.GeneratorParameters))
	for _, v := range p.GeneratorParameters {
		b = x.AppendI32(b, int32(len(v)))
		b = append(b, v...)
	}

	// p.PluginParameters
	b = append(b, 15, 0, 3)
	b = x.AppendListBegin(b, thrift.STRING, len(p.PluginParameters))
	for _, v := range p.PluginParameters {
		b = x.AppendI32(b, int32(len(v)))
		b = append(b, v...)
	}

	// p.Language
	b = append(b, 11, 0, 4)
	b = x.AppendI32(b, int32(len(p.Language)))
	b = append(b, p.Language...)

	// p.OutputPath
	b = append(b, 11, 0, 5)
	b = x.AppendI32(b, int32(len(p.OutputPath)))
	b = append(b, p.OutputPath...)

	// p.Recursive
	b = append(b, 2, 0, 6)
	b = append(b, *(*byte)(unsafe.Pointer(&p.Recursive)))

	// p.AST
	b = append(b, 12, 0, 7)
	b = p.AST.FastAppend(b)

	return append(b, 0)
}

func (p *Request) FastRead(b []byte) (off int, err error) {
	var ftyp thrift.TType
	var fid int16
	var l int
	var isset uint8
	x := thrift.BinaryProtocol{}
	for {
		ftyp, fid, l, err = x.ReadFieldBegin(b[off:])
		off += l
		if err != nil {
			goto ReadFieldBeginError
		}
		if ftyp == thrift.STOP {
			break
		}
		switch uint32(fid)<<8 | uint32(ftyp) {
		case 0x10b: // p.Version ID:1 thrift.STRING
			p.Version, l, err = x.ReadString(b[off:])
			off += l
			if err != nil {
				goto ReadFieldError
			}
			isset |= 0x1
		case 0x20f: // p.GeneratorParameters ID:2 thrift.LIST
			var sz int
			_, sz, l, err = x.ReadListBegin(b[off:])
			off += l
			if err != nil {
				goto ReadFieldError
			}
			p.GeneratorParameters = make([]string, sz)
			for i := 0; i < sz; i++ {
				p.GeneratorParameters[i], l, err = x.ReadString(b[off:])
				off += l
				if err != nil {
					goto ReadFieldError
				}
			}
			isset |= 0x2
		case 0x30f: // p.PluginParameters ID:3 thrift.LIST
			var sz int
			_, sz, l, err = x.ReadListBegin(b[off:])
			off += l
			if err != nil {
				goto ReadFieldError
			}
			p.PluginParameters = make([]string, sz)
			for i := 0; i < sz; i++ {
				p.PluginParameters[i], l, err = x.ReadString(b[off:])
				off += l
				if err != nil {
					goto ReadFieldError
				}
			}
			isset |= 0x4
		case 0x40b: // p.Language ID:4 thrift.STRING
			p.Language, l, err = x.ReadString(b[off:])
			off += l
			if err != nil {
				goto ReadFieldError
			}
			isset |= 0x8
		case 0x50b: // p.OutputPath ID:5 thrift.STRING
			p.OutputPath, l, err = x.ReadString(b[off:])
			off += l
			if err != nil {
				goto ReadFieldError
			}
			isset |= 0x10
		case 0x602: // p.Recursive ID:6 thrift.BOOL
			p.Recursive, l, err = x.ReadBool(b[off:])
			off += l
			if err != nil {
				goto ReadFieldError
			}
			isset |= 0x20
		case 0x70c: // p.AST ID:7 thrift.STRUCT
			p.AST = parser.NewThrift()
			l, err = p.AST.FastRead(b[off:])
			off += l
			if err != nil {
				goto ReadFieldError
			}
			isset |= 0x40
		default:
			l, err = x.Skip(b[off:], ftyp)
			off += l
			if err != nil {
				goto SkipFieldError
			}
		}
	}
	if isset != 0x7f {
		if isset&0x1 == 0 {
			fid = 1 // Version
			goto RequiredFieldNotSetError
		}
		if isset&0x2 == 0 {
			fid = 2 // GeneratorParameters
			goto RequiredFieldNotSetError
		}
		if isset&0x4 == 0 {
			fid = 3 // PluginParameters
			goto RequiredFieldNotSetError
		}
		if isset&0x8 == 0 {
			fid = 4 // Language
			goto RequiredFieldNotSetError
		}
		if isset&0x10 == 0 {
			fid = 5 // OutputPath
			goto RequiredFieldNotSetError
		}
		if isset&0x20 == 0 {
			fid = 6 // Recursive
			goto RequiredFieldNotSetError
		}
		if isset&0x40 == 0 {
			fid = 7 // AST
			goto RequiredFieldNotSetError
		}
	}
	return
ReadFieldBeginError:
	return off, thrift.PrependError(fmt.Sprintf("%T read field begin error: ", p), err)
ReadFieldError:
	return off, thrift.PrependError(
		fmt.Sprintf("%T read field %d '%s' error: ", p, fid, fieldIDToName_Request[fid]), err)
SkipFieldError:
	return off, thrift.PrependError(
		fmt.Sprintf("%T skip field %d type %d error: ", p, fid, ftyp), err)
RequiredFieldNotSetError:
	return off, thrift.NewProtocolException(thrift.INVALID_DATA,
		fmt.Sprintf("required field %s is not set", fieldIDToName_Request[fid]))
}

func (p *Generated) BLength() int {
	if p == nil {
		return 1
	}
	off := 0

	// p.Content ID:1 thrift.STRING
	off += 3
	off += 4 + len(p.Content)

	// p.Name ID:2 thrift.STRING
	if p.Name != nil {
		off += 3
		off += 4 + len(*p.Name)
	}

	// p.InsertionPoint ID:3 thrift.STRING
	if p.InsertionPoint != nil {
		off += 3
		off += 4 + len(*p.InsertionPoint)
	}
	return off + 1
}

func (p *Generated) FastWrite(b []byte) int { return p.FastWriteNocopy(b, nil) }

func (p *Generated) FastWriteNocopy(b []byte, w thrift.NocopyWriter) (n int) {
	if n = len(p.FastAppend(b[:0])); n > len(b) {
		panic("buffer overflow. concurrency issue?")
	}
	return
}

func (p *Generated) FastAppend(b []byte) []byte {
	if p == nil {
		return append(b, 0)
	}
	x := thrift.BinaryProtocol{}
	_ = x

	// p.Content
	b = append(b, 11, 0, 1)
	b = x.AppendI32(b, int32(len(p.Content)))
	b = append(b, p.Content...)

	// p.Name
	if p.Name != nil {
		b = append(b, 11, 0, 2)
		b = x.AppendI32(b, int32(len(*p.Name)))
		b = append(b, *p.Name...)
	}

	// p.InsertionPoint
	if p.InsertionPoint != nil {
		b = append(b, 11, 0, 3)
		b = x.AppendI32(b, int32(len(*p.InsertionPoint)))
		b = append(b, *p.InsertionPoint...)
	}

	return append(b, 0)
}

func (p *Generated) FastRead(b []byte) (off int, err error) {
	var ftyp thrift.TType
	var fid int16
	var l int
	var isset uint8
	x := thrift.BinaryProtocol{}
	for {
		ftyp, fid, l, err = x.ReadFieldBegin(b[off:])
		off += l
		if err != nil {
			goto ReadFieldBeginError
		}
		if ftyp == thrift.STOP {
			break
		}
		switch uint32(fid)<<8 | uint32(ftyp) {
		case 0x10b: // p.Content ID:1 thrift.STRING
			p.Content, l, err = x.ReadString(b[off:])
			off += l
			if err != nil {
				goto ReadFieldError
			}
			isset |= 0x1
		case 0x20b: // p.Name ID:2 thrift.STRING
			if p.Name == nil {
				p.Name = new(string)
			}
			*p.Name, l, err = x.ReadString(b[off:])
			off += l
			if err != nil {
				goto ReadFieldError
			}
		case 0x30b: // p.InsertionPoint ID:3 thrift.STRING
			if p.InsertionPoint == nil {
				p.InsertionPoint = new(string)
			}
			*p.InsertionPoint, l, err = x.ReadString(b[off:])
			off += l
			if err != nil {
				goto ReadFieldError
			}
		default:
			l, err = x.Skip(b[off:], ftyp)
			off += l
			if err != nil {
				goto SkipFieldError
			}
		}
	}
	if isset&0x1 == 0 {
		fid = 1 // Content
		goto RequiredFieldNotSetError
	}
	return
ReadFieldBeginError:
	return off, thrift.PrependError(fmt.Sprintf("%T read field begin error: ", p), err)
ReadFieldError:
	return off, thrift.PrependError(
		fmt.Sprintf("%T read field %d '%s' error: ", p, fid, fieldIDToName_Generated[fid]), err)
SkipFieldError:
	return off, thrift.PrependError(
		fmt.Sprintf("%T skip field %d type %d error: ", p, fid, ftyp), err)
RequiredFieldNotSetError:
	return off, thrift.NewProtocolException(thrift.INVALID_DATA,
		fmt.Sprintf("required field %s is not set", fieldIDToName_Generated[fid]))
}

func (p *Response) BLength() int {
	if p == nil {
		return 1
	}
	off := 0

	// p.Error ID:1 thrift.STRING
	if p.Error != nil {
		off += 3
		off += 4 + len(*p.Error)
	}

	// p.Contents ID:2 thrift.LIST
	if p.Contents != nil {
		off += 3
		off += 5
		for _, v := range p.Contents {
			off += v.BLength()
		}
	}

	// p.Warnings ID:3 thrift.LIST
	if p.Warnings != nil {
		off += 3
		off += 5
		for _, v := range p.Warnings {
			off += 4 + len(v)
		}
	}
	return off + 1
}

func (p *Response) FastWrite(b []byte) int { return p.FastWriteNocopy(b, nil) }

func (p *Response) FastWriteNocopy(b []byte, w thrift.NocopyWriter) (n int) {
	if n = len(p.FastAppend(b[:0])); n > len(b) {
		panic("buffer overflow. concurrency issue?")
	}
	return
}

func (p *Response) FastAppend(b []byte) []byte {
	if p == nil {
		return append(b, 0)
	}
	x := thrift.BinaryProtocol{}
	_ = x

	// p.Error
	if p.Error != nil {
		b = append(b, 11, 0, 1)
		b = x.AppendI32(b, int32(len(*p.Error)))
		b = append(b, *p.Error...)
	}

	// p.Contents
	if p.Contents != nil {
		b = append(b, 15, 0, 2)
		b = x.AppendListBegin(b, thrift.STRUCT, len(p.Contents))
		for _, v := range p.Contents {
			b = v.FastAppend(b)
		}
	}

	// p.Warnings
	if p.Warnings != nil {
		b = append(b, 15, 0, 3)
		b = x.AppendListBegin(b, thrift.STRING, len(p.Warnings))
		for _, v := range p.Warnings {
			b = x.AppendI32(b, int32(len(v)))
			b = append(b, v...)
		}
	}

	return append(b, 0)
}

func (p *Response) FastRead(b []byte) (off int, err error) {
	var ftyp thrift.TType
	var fid int16
	var l int
	x := thrift.BinaryProtocol{}
	for {
		ftyp, fid, l, err = x.ReadFieldBegin(b[off:])
		off += l
		if err != nil {
			goto ReadFieldBeginError
		}
		if ftyp == thrift.STOP {
			break
		}
		switch uint32(fid)<<8 | uint32(ftyp) {
		case 0x10b: // p.Error ID:1 thrift.STRING
			if p.Error == nil {
				p.Error = new(string)
			}
			*p.Error, l, err = x.ReadString(b[off:])
			off += l
			if err != nil {
				goto ReadFieldError
			}
		case 0x20f: // p.Contents ID:2 thrift.LIST
			var sz int
			_, sz, l, err = x.ReadListBegin(b[off:])
			off += l
			if err != nil {
				goto ReadFieldError
			}
			p.Contents = make([]*Generated, sz)
			for i := 0; i < sz; i++ {
				p.Contents[i] = NewGenerated()
				l, err = p.Contents[i].FastRead(b[off:])
				off += l
				if err != nil {
					goto ReadFieldError
				}
			}
		case 0x30f: // p.Warnings ID:3 thrift.LIST
			var sz int
			_, sz, l, err = x.ReadListBegin(b[off:])
			off += l
			if err != nil {
				goto ReadFieldError
			}
			p.Warnings = make([]string, sz)
			for i := 0; i < sz; i++ {
				p.Warnings[i], l, err = x.ReadString(b[off:])
				off += l
				if err != nil {
					goto ReadFieldError
				}
			}
		default:
			l, err = x.Skip(b[off:], ftyp)
			off += l
			if err != nil {
				goto SkipFieldError
			}
		}
	}
	return
ReadFieldBeginError:
	return off, thrift.PrependError(fmt.Sprintf("%T read field begin error: ", p), err)
ReadFieldError:
	return off, thrift.PrependError(
		fmt.Sprintf("%T read field %d '%s' error: ", p, fid, fieldIDToName_Response[fid]), err)
SkipFieldError:
	return off, thrift.PrependError(
		fmt.Sprintf("%T skip field %d type %d error: ", p, fid, ftyp), err)
}
