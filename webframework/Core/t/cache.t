#!/usr/bin/env ferite
uses "tap";

uses "stream";
uses "../Cache.feh";
uses "memcached";

function wf_profile(string what) {
	object block = recipient();
	block.invoke();
}

string t;
object ss = new Stream.StringStream("");
array lines;

Console.stderr = ss;
t = Cache.fetch("Foo", "not cached") using {
	raise new Error("error caching Foo");
};

is (t, "not cached", "Default value must be used on error");
lines = (ss.toString()).toArray("\n");
if (ok (lines.size() > 0, "Must receive error")) {
	string e = lines[0];
	is (e, "error caching Foo", "Cache set failure  must raise error");
}

done_testing();
