var b = require("microtime");

function func() {
  function nested() {
    var b = "";
    var e = b;
    e.normalize();
    var d;
    d = b;
    d.normalize();
  }
  var foo = b;
  foo.now();
}
