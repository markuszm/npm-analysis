function myfun(x) {
  return otherfun(x);
}

function otherfun(y) {
  return anotherfun(y);
}

function anotherfun(y) {
  return y + 2
}

(function () {
  console.log(myfun(2));
})();