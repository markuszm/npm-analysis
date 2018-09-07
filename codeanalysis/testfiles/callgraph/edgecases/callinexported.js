const mem = require("mem");

const foo = require("foo");

module.exports.sync = () => {
  foo.apiA();
};

module.exports.load = function() {
  foo.apiB();
};

module.exports.save = mem(() => {
  foo.apiC();
});

const push = mem(exec());

const save = option => {
    foo.apiB();
  },
  loadNested = option => {
    foo.apiC();
  };

const load = function(option) {
  foo.apiC();
};

module.exports.safe4 = () => save("v4");
module.exports.load4 = () => load("v4");
