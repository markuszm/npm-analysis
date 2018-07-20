var foo = require("foobar");

foo = require("foo");

var bar = foo;
bar = require("bar");

var fooVar = foo;

var foobar = require("foobar");

function f() {
    var foobar = require("foo");
    foo = require("bar");
}

function g() {
  foo.someMethod();
  foobar.otherMethod();
}

function h() {
  bar.someMethod();
  fooVar.someMethod();
}
