const libVar = require("./anotherFile");

function aFunction() {
  async.series([_.curry(libVar.referencedFn)]);
  libVar.aFnInAnotherFile(n + 1);
}
libVar.aFnInAnotherFile(2);
aFunction();

// method assignment - not a call
var b = libVar.referencedFn;
