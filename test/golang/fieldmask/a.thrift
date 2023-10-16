namespace go base

struct TrafficEnv {
	0: string Name = "",
	1: bool Open = false,
	2: string Env = "",
	256: i64 Code,
}

struct Base {
	0: string Addr = "",
	1: string LogID = "",
	2: string Caller = "",
	5: optional TrafficEnv TrafficEnv,
	255: optional ExtraInfo Extra,
	256: MetaInfo Meta,
}

struct ExtraInfo {
	1: map<string, string> KVS
}

struct MetaInfo {
	1: map<string, string> PersistentKVS,
	2: map<Key, Val> TransientKVS,
	3: Base Base,
}

struct Key {
	1: string id
}

struct Val {
	1: string id
}

struct BaseResp {
	1: string StatusMessage = "",
	2: i32 StatusCode = 0,
	3: optional map<string, string> Extra,
}
