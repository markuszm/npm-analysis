exports = module.exports;

function add(a, b) {
    return a + b;
}

exports.add = add;

exports.divide = (a, b) => a / b;

exports.multiply = function(a, b) {
    return a * b;
};

exports.foo = {
    add2(a) {
        return a + 2;
    },
    add4(a) {
        return a + 4;
    },
    add8(a) {
        return a + 8;
    },
    add16(a) {
        return a + 16;
    },
    add32(a) {
        return a + 32;
    }
};

exports.e = Math.E;

module.exports.calculator = class Calculator {
    add(a, b) {
        return a + b;
    }
    subtract(a, b) {
        return a - b;
    }
};
