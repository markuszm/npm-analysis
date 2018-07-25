function foo() {
    return "GLOBAL"
}

function f() {
    const foo = 20;
    module.exports.bar = foo;
}

module.exports.foo = foo;
f();