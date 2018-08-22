let bar = require("foo");

let bar2 = bar;

let bar3;
bar3 = bar;

function x() {
  bar2.api();
  bar3.api();
}
