function f() {
  let regex = new RegExp("(ab)*");
  let regex2 = /(ab)*/;
  regex.exec("abababab");
  regex2.exec("ababab");
  /(ab)*/.exec("abab");
}
