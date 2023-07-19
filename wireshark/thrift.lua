-- Copyright 2023 CloudWeGo Authors
--
-- Licensed under the Apache License, Version 2.0 (the "License");
-- you may not use this file except in compliance with the License.
-- You may obtain a copy of the License at
--
--   http://www.apache.org/licenses/LICENSE-2.0
--
-- Unless required by applicable law or agreed to in writing, software
-- distributed under the License is distributed on an "AS IS" BASIS,
-- WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
-- See the License for the specific language governing permissions and
-- limitations under the License.

-------------------------------------------------------------------------------
--- protocols
local ttheader_protocol = Proto("ttheader", "Thrift TTHeader Protocol")
local tbinary_protocol = Proto("tbinary", "Thrift UnframedBinary Protocol")

-------------------------------------------------------------------------------
--- lookup tables
local msgtype_valstr = {}
msgtype_valstr[1] = "CALL"
msgtype_valstr[2] = "REPLY"
msgtype_valstr[3] = "EXCEPTION"
msgtype_valstr[4] = "ONEWAY"

local fieldtype_valstr = {}
fieldtype_valstr[0] = "STOP"
fieldtype_valstr[1] = "VOID"
fieldtype_valstr[2] = "BOOL"
fieldtype_valstr[3] = "BYTE"
fieldtype_valstr[4] = "DOUBLE"
fieldtype_valstr[6] = "I16"
fieldtype_valstr[8] = "I32"
fieldtype_valstr[10] = "I64"
fieldtype_valstr[11] = "STRING"
fieldtype_valstr[12] = "STRUCT"
fieldtype_valstr[13] = "MAP"
fieldtype_valstr[14] = "SET"
fieldtype_valstr[15] = "LIST"
fieldtype_valstr[16] = "UTF8"
fieldtype_valstr[17] = "UTF16"
fieldtype_valstr[18] = "BINARY"

-------------------------------------------------------------------------------
--- protocol constants
THRIFT_VERSION_MASK = -65536
THRIFT_VERSION_1 = -2147418112
THRIFT_HEADER_MAGIC = 0x1000
THRIFT_HEADER_TYPE_STR_KV = 0x01
THRIFT_HEADER_TYPE_INT_KV = 0x10
THRIFT_HEADER_TYPE_ACL_KV = 0x11
THRIFT_TYPE_MASK = 0x000000ff

-------------------------------------------------------------------------------
--- fields
local tbinary_fields = {
    msg_type = ProtoField.uint8("tbinary.msgtype", "Message Type", base.DEC, msgtype_valstr),
    msg_name = ProtoField.string("tbinary.msgname", "Message Name"),
    msg_seq = ProtoField.uint32("tbinary.msgseq", "Message Sequence", base.DEC),
}

tbinary_protocol.fields = tbinary_fields

-------------------------------------------------------------------------------
--- ThriftBuffer is a stateful buffer helper
ThriftBuffer = {}
function ThriftBuffer:new(buf)
    o = {
        pos = 0,
        buf = buf
    }
    setmetatable(o, self)
    self.__index = self
    return o
end

function ThriftBuffer:seek(pos)
    self.pos = pos
end

function ThriftBuffer:skip(num)
    self.pos = self.pos + num
end

function ThriftBuffer:__call(len)
    rv = self.buf(self.pos, len)
    self.pos = self.pos + len
    return rv
end

function ThriftBuffer:bool()
    local byte = self(1):int()
    return byte ~= 0
end

function ThriftBuffer:byte()
    return self(1):int()
end

function ThriftBuffer:double()
    return self(8):float()
end

function ThriftBuffer:i16()
    return self(2):int()
end

function ThriftBuffer:i32()
    return self(4):int()
end

function ThriftBuffer:i64()
    return self(8):int64()
end

function ThriftBuffer:varint()
    local res = 0
    local pos = 0
    while true do
        local b = self.buf(self.pos + pos, 1):int()
        res = bit32.bor(res, bit32.lshift(bit32.band(b, 0x7f), pos * 7))
        pos = pos + 1

        if bit32.rshift(b, 7) == 0 then
            self.pos = self.pos + pos
            return res
        end
    end
end

function ThriftBuffer:varstring()
    local size = self:varint()
    local rv = self.buf(self.pos, size):string()
    self.pos = self.pos + size
    return rv
end

function ThriftBuffer:string()
    local size = self(4):int()
    local val = self(size):string()
    return val
end

function ThriftBuffer:infostring()
    local size = self(2):uint()
    local val = self(size):string()
    return val
end

local fieldtype_readers = {
    BOOL = ThriftBuffer.bool,
    BYTE = ThriftBuffer.byte,
    DOUBLE = ThriftBuffer.double,
    I16 = ThriftBuffer.i16,
    I32 = ThriftBuffer.i32,
    I64 = ThriftBuffer.i64,
    STRING = ThriftBuffer.string,
    BINARY = ThriftBuffer.string,
}

-------------------------------------------------------------------------------
--- decodes a series of thrift fields until the STOP sentinel is reached
function decode_tfields(buf, tree)
    if buf:len() == 0 then
        return 0
    end

    local tbuf = ThriftBuffer:new(buf)
    local type = fieldtype_valstr[tbuf(1):int()]

    while type ~= nil and type ~= "STOP" do
        id = tbuf(2):int()
        local pos = tbuf.pos
        if type == "VOID" then
            tree:add(id, "Type: VOID")
        elseif type == "BOOL" then
            local val = tbuf:bool()
            tree:add(buf(pos, 1), id, "Type: BOOL", string.format("%s", val))
        elseif type == "BYTE" then
            local val = tbuf:byte()
            tree:add(buf(pos, 1), id, "Type: BYTE", val)
        elseif type == "DOUBLE" then
            local val = tbuf:double()
            tree:add(buf(pos, 8), id, "Type: DOUBLE", val)
        elseif type == "I16" then
            local val = tbuf:i16()
            tree:add(buf(pos, 2), id, "Type: I16", val)
        elseif type == "I32" then
            local val = tbuf:i32()
            tree:add(buf(pos, 4), id, "Type: I32", val)
        elseif type == "I64" then
            local val = tbuf:i64()
            tree:add(buf(pos, 8), id, "Type: I64", string.format("%s", val))
        elseif type == "STRING" then
            local size = tbuf(4):int()
            local val = tbuf(size):string()
            tree:add(buf(pos, 4+size), id, "Type: STRING", val)
        elseif type == "BINARY" then
            local size = tbuf(4):int()
            local val = tbuf(size):string()
            tree:add(buf(pos, 4+size), id, "Type: BINARY", val)
        elseif type == "STRUCT" then
            local child_tree = tree:add(id, "Type: STRUCT")
            local len = decode_tfields(buf(pos, buf:len() - pos), child_tree)
            tbuf:skip(len)
        elseif type == "MAP" then
            local ktype = tbuf(1):int()
            local vtype = tbuf(1):int()
            local size = tbuf(4):int()
            local ktype_str = fieldtype_valstr[ktype]
            local vtype_str = fieldtype_valstr[vtype]

            if ktype_str ~= nil and vtype_str ~= nil then
                local child_tree = tree:add(id, "Type: MAP" .. string.format("<%s, %s>", ktype_str, vtype_str))
                local kreader = fieldtype_readers[ktype_str]
                for i = 1, size do
                    fieldpos = tbuf.pos
                    key = kreader(tbuf)
                    -- TODO(eac): make handling non-scalars more elegant
                    if vtype_str == "STRUCT" then
                        local elem_tree = child_tree:add(i, key)
                        child_buf = tbuf.buf(tbuf.pos)
                        local len = decode_tfields(child_buf, elem_tree)
                        tbuf:skip(len)
                    else
                        local vreader = fieldtype_readers[vtype_str]
                        val = vreader(tbuf)
                        child_tree:add(buf(fieldpos, tbuf.pos-fieldpos), key, val)
                    end
                end
            end
        elseif type == "SET" or type == "LIST" then
            local etype = tbuf(1):int()
            local size = tbuf(4):int()
            local etype_str = fieldtype_valstr[etype]
            local child_tree = tree:add(id, "Type: " .. string.format("%s<%s>", type, etype_str))

            if etype_str ~= nil then
                local ereader = fieldtype_readers[etype_str]
                for i = 1, size do
                    local fieldpos = tbuf.pos
                    -- TODO(eac): make handling non-scalars more elegant
                    if etype_str == "STRUCT" then
                        local elem_tree = child_tree:add(string.format("%s", i))
                        child_buf = tbuf.buf(tbuf.pos)
                        local len = decode_tfields(child_buf, elem_tree)
                        tbuf:skip(len)
                    else
                        elem = ereader(tbuf)
                        child_tree:add(buf(fieldpos, tbuf.pos-fieldpos), i, elem)
                    end
                end
            end
        else
            print(type .. " not implemented")
        end

        type = fieldtype_valstr[tbuf(1):int()]
    end

    if type == nil then
        return 0
    end

    return tbuf.pos
end

-------------------------------------------------------------------------------
--- root tbinary dissector. will dissect a unframed tbinary message
function tbinary_protocol.dissector(buffer, pinfo, tree)
    local tbuf = ThriftBuffer:new(buffer)
    local sz = tbuf(4):int()

    if sz < 0 then
        local version = bit32.band(sz, THRIFT_VERSION_MASK)
        if not bit32.btest(version, THRIFT_VERSION_1) then
            return 0
        end

        local type = bit32.band(sz, THRIFT_TYPE_MASK)
        tree:add(tbinary_fields.msg_type, type)

        local name_pos = tbuf.pos
        local name = tbuf:string()
        if name:len() > 0 then
            tree:add(tbinary_fields.msg_name, buffer(name_pos, tbuf.pos-name_pos), name)
        end


        pinfo.cols.info = string.format("%s %s %s", pinfo.cols.info, msgtype_valstr[type], name)

        local seq_pos = tbuf.pos
        local seqid = tbuf(4):int()
        tree:add(tbinary_fields.msg_seq, buffer(seq_pos, 4), seqid)
    else
        -- TODO(eac): implement me
        print("non-versioned tbinary protocol unimplemented")
    end

    decode_tfields(buffer(tbuf.pos, buffer:len()-tbuf.pos), tree)
end

-------------------------------------------------------------------------------
--- root theader dissector. will dissect a framed theader message, chaining
--- the payload into the tbinary dissector
function ttheader_protocol.dissector(buffer, pinfo, tree)
    if buffer:len() == 0 then return end

    pinfo.cols.protocol = ttheader_protocol.name

    local subtree = tree:add(ttheader_protocol, buffer(), "Thrift Protocol Data")

    local frame_size = buffer(0, 4):int()

    if (buffer:len() - 4) < frame_size then
        pinfo.desegment_len = frame_size - (buffer:len() - 4)
        pinfo.desegment_offset = 0
        return
    end

    local framebuf = buffer(4, frame_size):tvb()

    local tb = ThriftBuffer:new(framebuf)
    local version = framebuf(0, 4):int()
    if bit32.rshift(version, 16) == THRIFT_HEADER_MAGIC then
        local flags, seq_id, header_length, end_of_headers, protocol_id, transform_count = nil
        tb:seek(2)

        flags = tb(2):uint()
        seq_id = tb(4):int()
        header_length = tb(2):uint() * 4
        end_of_headers = tb.pos + header_length
        local headers_tree = subtree:add(framebuf(tb.pos, header_length), "Headers")

        protocol_id = tb(1):uint()
        transform_count = tb(1):uint()

        -- TODO: now don't support transform
        --[[
        local transforms = {}
        for i = 1, transform_count do
            local transform_id = tb(1):uint()
            table.insert(transforms, transform_id)
        end
        --]]

        while tb.pos < end_of_headers do
            local header_type = tb(1):uint()
            if header_type == THRIFT_HEADER_TYPE_INT_KV then
                local count = tb(2):uint()
                for i = 1, count do
                    local header_start = tb.pos
                    local key = tb(2):uint()
                    local keystr = string.format("INT(%d)", key)
                    local val_len = tb(2):uint()

                    if val_len > 0 then
                        local value_tree = headers_tree:add(framebuf(header_start, (tb.pos-header_start) + val_len), keystr)
                        local val_range = framebuf(tb.pos, val_len)
                        local value = val_range:string()
                        value_tree:add(framebuf(tb.pos, val_len), value)
                        tb:seek(tb.pos + val_len)
                    else
                        print("empty")
                    end
                end
            elseif header_type == THRIFT_HEADER_TYPE_STR_KV then
                local count = tb(2):uint()
                for i = 1, count do
                    local header_start = tb.pos
                    local key = tb:infostring()
                    local val_len = tb(2):uint()

                    if val_len > 0 then
                        local value_tree = headers_tree:add(framebuf(header_start, (tb.pos-header_start) + val_len), key)
                        local val_range = framebuf(tb.pos, val_len)
                        local value = val_range:string()
                        value_tree:add(framebuf(tb.pos, val_len), value)
                        tb:seek(tb.pos + val_len)
                    else
                        print("empty")
                    end
                end
            elseif header_type == THRIFT_HEADER_TYPE_ACL_KV then
                local key = "acl_token"
                local header_start = tb.pos
                local val_len = tb(2):uint()
                local value_tree = headers_tree:add(framebuf(header_start, (tb.pos-header_start) + val_len), key)
                local val_range = framebuf(tb.pos, val_len)
                local value = val_range:string()
                value_tree:add(framebuf(tb.pos, val_len), value)
                tb:seek(tb.pos + val_len)
            end
        end

        --- TODO: unstrict
        local remaining_size = framebuf:len() - end_of_headers
        local first_word = tb(4):int()
        if first_word > 0
        then
            remaining_size = remaining_size - 4
            end_of_headers = end_of_headers + 4
            pinfo.cols.info = "TTHeader(Framed)"
        else
            pinfo.cols.info = "TTHeader(Buffered)"
        end

        remaining_buf = framebuf(end_of_headers, remaining_size)
        local payload_tree = subtree:add(tbinary_protocol, remaining_buf, "Payload")
        Dissector.get("tbinary"):call(remaining_buf:tvb(), pinfo, payload_tree)
    end
end

local function heuristic_checker(buffer, pinfo, tree)
    --- guard for length
    length = buffer:len()
    if length < 8 then return false end

    local frame_size = buffer(0, 4):int()
    if frame_size < 20 then return false end

    local version = buffer(4, 4):int()

    if bit32.rshift(version, 16) == THRIFT_HEADER_MAGIC
    then
        ttheader_protocol.dissector(buffer, pinfo, tree)
        return true
    else return false end
end

-------------------------------------------------------------------------------

ttheader_protocol:register_heuristic("tcp", heuristic_checker)
