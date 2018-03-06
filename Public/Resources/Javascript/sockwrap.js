;var SockWrap = function(io, forcedLongPolling) {
	var cs = {}
	, ack = function(what, args) {
		var cb;
		if (args.length == 0) return;
		cb = args[args.length-1];
		if (typeof cb !== 'function') return;
		cb({"acked":what});
	}
	;
	cs.io = io;
	cs.forcedLongPolling = forcedLongPolling;
	cs.encode = function(str) {
		return encodeURIComponent(str);
	};
	cs.decode = function (str) {
		return decodeURIComponent(str);
	};
	cs.on = function(what, callback) {
		cs.io.on(what, function() {
			callback.apply(null, arguments);
			ack('plain', arguments)
		});
	};
	cs.onString = function(what, callback) {
		cs.io.on(what, function(data) {
			arguments[0] = cs.decode(data)
			callback.apply(null, arguments);
			ack('onString', arguments)
		});
	};
	cs.onJson = function(what, callback) {
		cs.io.on(what, function(data) {
			var decoded, json;
			try {
				decoded = cs.decode(data);
			} catch (e) {
				console.log("ERROR SockWrap.decode(): error decoding data: [" + data + "]: " + e);
			}
			try {
				json = JSON.parse(decoded);
			} catch (e) {
				console.log("ERROR SockWrap.decode(): error parsing json `" + decoded + " `: " + e);
			}
			arguments[0] = json;
			callback.apply(null, arguments);
			ack('onJson', arguments)
		});
	};
	cs.emit = function(what, data) {
		switch (typeof data) {
			case "object":
				arguments[1] = cs.encode(JSON.stringify(data));
				break;
			case "string":
				arguments[1] = cs.encode(data);
				break;
		}
		cs.io.emit.apply(cs.io, arguments);
	};
	cs.connect = function() {
		cs.io.connect();
	};
	cs.disconnect = function() {
		cs.io.disconnect();
	};
	return cs;
};
