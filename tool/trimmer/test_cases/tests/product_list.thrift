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

include "./base.thrift"

namespace go oec.product.test_sentry_list

struct SellReq {
    1: required i64 seller_id,

    255: optional base.Base Base, 
}

struct SellResp {
    1: required bool hit,

    255: optional base.BaseResp BaseResp, 
}

struct BReq {
    1: optional list<i64>  product_ids,

    255: optional base.Base Base, 
}

struct BResp {
    1: optional string result, // deprecated
    2: optional map<i64,string> id2result,

    255: optional base.BaseResp BaseResp, 
}


struct InfoReq {
    1: optional list<i64>  ids,

    255: optional base.Base Base, 
}

struct InfoResp {
    1: optional map<i64,string> id2result,

    255: optional base.BaseResp BaseResp, 
}




service TestSentryListService {
    SellResp GetSeller (1: SellReq req),
    BResp GetAD (1: BReq req),
    InfoResp GetInfo(1: InfoReq req),
}