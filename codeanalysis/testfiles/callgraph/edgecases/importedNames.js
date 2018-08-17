import {foo as someFunction} from 'foo'

const foobar = require('bar').bar;

let foo;

foo = require('foobar').bar;

function f() {
    someFunction();
    foobar();
    foo();
}