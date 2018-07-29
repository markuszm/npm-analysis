function foo() {
    return "GLOBAL"
}

var bar = 20;

if (foo()) {
    bar = "a string"
} else {
    bar = function() {
        return 42
    }
}

function f() {
    const foo = 20;
    module.exports.bar = foo;
}

module.exports.foo = foo;
f();

module.exports.foobar = bar;