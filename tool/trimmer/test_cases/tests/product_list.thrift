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