const OAuth = require('oauth').OAuth;

var foo;

var oauth = new OAuth("a");
oauth.someMethod();

if (oauth) {
    foo = new OAuth("b");
} else {
    foo = require("auth0");
}

foo.someMethod();
