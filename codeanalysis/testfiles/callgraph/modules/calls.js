const f = require("foo"),
  bar = require("bar");
const foobar = require("foobar");

const OAuth = require('oauth').OAuth;

const a = f.a();
let b = bar.b(a);

foobar.func(a, b);
OAuth.someMethod();