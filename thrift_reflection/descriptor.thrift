namespace go reflection

struct TypeDescriptor{
    1:required string filepath
    2:required string type_name
    3:optional TypeDescriptor key_type
    4:optional TypeDescriptor value_type
}

// todo
//struct ConstDescriptor{
//    1:required string filepath
//    2:required string type_name
//    3:required TypeDescriptor key_type
//    4:required TypeDescriptor value_type
//}

struct TypedefDescriptor{
    1:required string filepath
    2:required TypeDescriptor type
    3:required string alias
    6:required map<string,list<string>> annotations // annotation of this field
    7:required string comments // comments of this field
    8:optional map<string,string> extra // extra info?
}

struct EnumDescriptor{
    1:required string filepath
    2:required string name
    3:required list<EnumValueDescriptor> values
    6:required map<string,list<string>> annotations // annotation of this field
    7:required string comments // comments of this field
    8:optional map<string,string> extra // extra info?
}

struct EnumValueDescriptor{
    1:required string filepath
    2:required string name
    3:required i64 value
    6:required map<string,list<string>> annotations // annotation of this field
    7:required string comments // comments of this field
    8:optional map<string,string> extra // extra info?
}

struct FieldDescriptor{
    1:required string filepath // the name of idl file to which this field belongs
    2:required string name  // name of the field
    3:required TypeDescriptor type  // struct type name of the field, if it's container,

    24:required string requiredness  // required, optional, or default
    5:required i32 id // field id
    6:required map<string,list<string>> annotations // annotation of this field
    7:required string comments // comments of this field
    8:optional map<string,string> extra // extra info?
}

struct StructDescriptor{
    1:required string filepath // the name of idl file to which this struct belongs
    2:required string name  // name of this struct
    3:required list<FieldDescriptor> fields
    8:required map<string,list<string>> annotation // annotation of this field
    9:required string comments // comments of this field
    10:optional map<string,string> extra // extra info?
}

struct MethodDescriptor{
    1:required string filepath // the name of idl file which this method is belonged
    2:required string name // name of the method
    3:required TypeDescriptor response // response, if it's oneway method, this should be nil
    4:required list<FieldDescriptor> args   // requests
    5:required map<string,list<string>> annotations // annotation of this field
    6:required string comments // comments of this field
    7:optional map<string,string> extra // extra info?
    8:required list<FieldDescriptor> throw_exceptions
}

struct ServiceDescriptor{
    1:required string filepath // the name of idl file which this service is belonged
    2:required string name // name of the service
    3:required list<MethodDescriptor> methods  // requests
    4:required map<string,list<string>> annotations // annotation of this field
    5:required string comments // comments of this field
    6:optional map<string,string> extra // extra info?
}

struct FileDescriptor{
    1:required string filepath // the path of idl file
    2:required map<string,string> includes  // include IDL, key is alias and value is path
    3:required map<string,string> namespaces // namespace, key is language and value is namespace for this language
    4:required list<ServiceDescriptor> services // services
    5:required list<StructDescriptor> structs   // structs
    6:required list<StructDescriptor> exceptions   // structs
    7:required list<EnumDescriptor> enums   // structs
    8:required list<TypedefDescriptor> typedefs   // structs
    9:required list<StructDescriptor> unions   // structs
    10:optional map<string,string> extra // extra info?
}