thriftgo -g go:with_reflection,check_option_grammar -o . -r option_idl/test.thrift
thriftgo -g go:with_reflection -o . -r option_idl/test_grammar_error.thrift