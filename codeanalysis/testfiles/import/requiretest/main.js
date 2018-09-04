const foo = require("foo");
const bar = require("bar");
const OAuth = require('oauth').OAuth;
const OAuthC = require('oauth').OAuth.a.b.c;
const OAuthM = require('oauth').OAuth().a().b().c();
const someNotInstalledModule = require("abc");
const ignoreLocalModules2 = require("/someotherjsfile.js");
const ignoreLocalModules1 = require("./someotherjsfile.js");
const ignoreLocalModules3 = require("../someotherjsfile.js");

// side effect import
require("b");

function main() {
    require("f").f();
    console.log("Hello World")
}
