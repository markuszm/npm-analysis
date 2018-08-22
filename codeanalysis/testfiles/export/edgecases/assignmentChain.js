var a;
var b;
const c = 2;

module.exports = a = b = c;
module.exports.a = module.exports.b = module.exports.c = c;
module.exports.foo = module.exports.bar = null;
