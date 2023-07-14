// Code generated by thriftgo (0.3.0). DO NOT EDIT.

package entity

import (
	"fmt"
	"github.com/apache/thrift/lib/go/thrift"
	"github.com/cloudwego/thriftgo/thrift_reflection"
)

type InnerStruct struct {
	Email string `thrift:"email,1,required" json:"email"`
}

func NewInnerStruct() *InnerStruct {
	return &InnerStruct{}
}

func (p *InnerStruct) GetEmail() (v string) {
	return p.Email
}

var fieldIDToName_InnerStruct = map[int16]string{
	1: "email",
}

func (p *InnerStruct) Read(iprot thrift.TProtocol) (err error) {

	var fieldTypeId thrift.TType
	var fieldId int16
	var issetEmail bool = false

	if _, err = iprot.ReadStructBegin(); err != nil {
		goto ReadStructBeginError
	}

	for {
		_, fieldTypeId, fieldId, err = iprot.ReadFieldBegin()
		if err != nil {
			goto ReadFieldBeginError
		}
		if fieldTypeId == thrift.STOP {
			break
		}

		switch fieldId {
		case 1:
			if fieldTypeId == thrift.STRING {
				if err = p.ReadField1(iprot); err != nil {
					goto ReadFieldError
				}
				issetEmail = true
			} else {
				if err = iprot.Skip(fieldTypeId); err != nil {
					goto SkipFieldError
				}
			}
		default:
			if err = iprot.Skip(fieldTypeId); err != nil {
				goto SkipFieldError
			}
		}

		if err = iprot.ReadFieldEnd(); err != nil {
			goto ReadFieldEndError
		}
	}
	if err = iprot.ReadStructEnd(); err != nil {
		goto ReadStructEndError
	}

	if !issetEmail {
		fieldId = 1
		goto RequiredFieldNotSetError
	}
	return nil
ReadStructBeginError:
	return thrift.PrependError(fmt.Sprintf("%T read struct begin error: ", p), err)
ReadFieldBeginError:
	return thrift.PrependError(fmt.Sprintf("%T read field %d begin error: ", p, fieldId), err)
ReadFieldError:
	return thrift.PrependError(fmt.Sprintf("%T read field %d '%s' error: ", p, fieldId, fieldIDToName_InnerStruct[fieldId]), err)
SkipFieldError:
	return thrift.PrependError(fmt.Sprintf("%T field %d skip type %d error: ", p, fieldId, fieldTypeId), err)

ReadFieldEndError:
	return thrift.PrependError(fmt.Sprintf("%T read field end error", p), err)
ReadStructEndError:
	return thrift.PrependError(fmt.Sprintf("%T read struct end error: ", p), err)
RequiredFieldNotSetError:
	return thrift.NewTProtocolExceptionWithType(thrift.INVALID_DATA, fmt.Errorf("required field %s is not set", fieldIDToName_InnerStruct[fieldId]))
}

func (p *InnerStruct) ReadField1(iprot thrift.TProtocol) error {
	if v, err := iprot.ReadString(); err != nil {
		return err
	} else {
		p.Email = v
	}
	return nil
}

func (p *InnerStruct) Write(oprot thrift.TProtocol) (err error) {
	var fieldId int16
	if err = oprot.WriteStructBegin("InnerStruct"); err != nil {
		goto WriteStructBeginError
	}
	if p != nil {
		if err = p.writeField1(oprot); err != nil {
			fieldId = 1
			goto WriteFieldError
		}

	}
	if err = oprot.WriteFieldStop(); err != nil {
		goto WriteFieldStopError
	}
	if err = oprot.WriteStructEnd(); err != nil {
		goto WriteStructEndError
	}
	return nil
WriteStructBeginError:
	return thrift.PrependError(fmt.Sprintf("%T write struct begin error: ", p), err)
WriteFieldError:
	return thrift.PrependError(fmt.Sprintf("%T write field %d error: ", p, fieldId), err)
WriteFieldStopError:
	return thrift.PrependError(fmt.Sprintf("%T write field stop error: ", p), err)
WriteStructEndError:
	return thrift.PrependError(fmt.Sprintf("%T write struct end error: ", p), err)
}

func (p *InnerStruct) writeField1(oprot thrift.TProtocol) (err error) {
	if err = oprot.WriteFieldBegin("email", thrift.STRING, 1); err != nil {
		goto WriteFieldBeginError
	}
	if err := oprot.WriteString(p.Email); err != nil {
		return err
	}
	if err = oprot.WriteFieldEnd(); err != nil {
		goto WriteFieldEndError
	}
	return nil
WriteFieldBeginError:
	return thrift.PrependError(fmt.Sprintf("%T write field 1 begin error: ", p), err)
WriteFieldEndError:
	return thrift.PrependError(fmt.Sprintf("%T write field 1 end error: ", p), err)
}

func (p *InnerStruct) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("InnerStruct(%+v)", *p)
}

var file_entity_struct_thrift_go_types = []interface{}{
	(*InnerStruct)(nil), // Struct 0: entity.InnerStruct
}
var file_idl_entity_struct_thrift *thrift_reflection.FileDescriptor
var file_idl_entity_struct_rawDesc = []byte{
	0x1f, 0x8b, 0x8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xff, 0xac, 0x90, 0x4b, 0x4e, 0xc3, 0x30,
	0x10, 0x40, 0x5f, 0x93, 0xb8, 0x35, 0x4c, 0x23, 0x9f, 0x80, 0x2b, 0xa4, 0x12, 0xb7, 0x60, 0xb,
	0x7, 0xa8, 0x22, 0x6a, 0x8a, 0xa5, 0x62, 0x83, 0x3d, 0x5d, 0x70, 0x7b, 0xe4, 0xc4, 0x88, 0x3d,
	0xb0, 0x7a, 0xfe, 0x68, 0x9e, 0x46, 0x4f, 0xd8, 0x0, 0xf7, 0xe9, 0x5d, 0x43, 0x8a, 0xc7, 0x70,
	0xba, 0x1c, 0xe6, 0x18, 0x93, 0xce, 0xf5, 0x5a, 0xe, 0x3e, 0x6a, 0xd0, 0xcf, 0x86, 0x63, 0xd1,
	0x7c, 0x7d, 0xd6, 0x49, 0x5f, 0x73, 0x78, 0xd1, 0x91, 0x4e, 0x4, 0x60, 0xa4, 0x5f, 0xe, 0xd5,
	0xd3, 0x9d, 0x13, 0x70, 0xd7, 0x6c, 0xea, 0x8b, 0x4e, 0x3f, 0xba, 0x69, 0xd5, 0x38, 0x86, 0x7d,
	0x9d, 0x73, 0x98, 0xca, 0xcd, 0xaf, 0x37, 0x10, 0x3a, 0x40, 0x1e, 0x62, 0xf4, 0xf9, 0x69, 0xf9,
	0x70, 0xf4, 0xff, 0xa1, 0x34, 0xfe, 0x6d, 0xe, 0x97, 0x3d, 0xfd, 0x1f, 0x3d, 0xdb, 0xa2, 0x39,
	0xc4, 0x33, 0xc2, 0x0, 0xd8, 0x47, 0xff, 0x71, 0xd, 0xd9, 0x9f, 0x2c, 0xa6, 0x2e, 0x39, 0xb2,
	0x13, 0x57, 0x43, 0x8, 0x96, 0x35, 0xe4, 0xf0, 0xfd, 0x60, 0x58, 0xb, 0x6d, 0x5b, 0xa9, 0x5d,
	0xa3, 0x6d, 0xbc, 0x69, 0xbc, 0x5d, 0xc8, 0x57, 0x0, 0x0, 0x0, 0xff, 0xff, 0x24, 0xd5, 0xce,
	0x65, 0xc5, 0x1, 0x0, 0x0,
}

func init() {
	if file_idl_entity_struct_thrift != nil {
		return
	}
	file_idl_entity_struct_thrift = thrift_reflection.BuildFileDescriptor(file_idl_entity_struct_rawDesc, file_entity_struct_thrift_go_types)
}

func GetFileDescriptorForEntityStruct() *thrift_reflection.FileDescriptor {
	return file_idl_entity_struct_thrift
}
func (p *InnerStruct) GetDescriptor() *thrift_reflection.StructDescriptor {
	return file_idl_entity_struct_thrift.GetStructDescriptor("InnerStruct")
}
