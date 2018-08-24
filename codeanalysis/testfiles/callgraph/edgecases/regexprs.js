let regex3;

function f() {
  let regex = new RegExp("(ab)*");
  let regex2 = /(ab)*/;
  regex3 = /(ab)*/;
  regex.exec("abababab");
  regex2.exec("ababab");
  regex3.exec("ab");
  /(ab)*/.exec("abab");
}
