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
	1: map<string, string> F1
	2: map<i64, string> F2,
	3: list<string> F3
	4: set<string> F4,
	5: map<double, Val> F5
}

struct MetaInfo {
	1: map<Int, Val> IntMap,
	2: map<Str, Key> StrMap,
	3: list<Key> List,
	4: set<Val> Set,
	255: Base Base,
}

typedef Val Key 

struct Val {
	1: string id
}

typedef double Float

typedef i64 Int

typedef string Str

enum Ex {
	A = 1,
	B = 2,
	C = 3
}

struct BaseResp {
	1: string StatusMessage = "",
	2: i32 StatusCode = 0,
	3: optional map<string, string> Extra,

	4: map<Str, Str> F1
	5: map<Int, string> F2,
	6: list<string> F3
	7: set<string> F4,
	8: map<Float, Val> F5
	9: map<double, string> F6
	10: map<Ex, string> F7
}
