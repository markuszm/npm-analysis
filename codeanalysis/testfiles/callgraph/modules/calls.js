const f = require("foo"),
  bar = require("bar");
const foobar = require("foobar");

const a = f.a();
let b = bar.b(a);

foobar.func(a, b);
