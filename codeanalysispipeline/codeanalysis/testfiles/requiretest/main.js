const foo = require("foo");
const bar = require("bar");
const someNotInstalledModule = require("abc");
const ignoreLocalModules2 = require("/someotherjsfile.js");
const ignoreLocalModules1 = require("./someotherjsfile.js");
const ignoreLocalModules3 = require("../someotherjsfile.js");

function main() {
    console.log("Hello World")
}
