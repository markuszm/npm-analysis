module.exports = {
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
    },
    addAll: function(a,...b){
        console.log('From graph draw function');
    },
    theSolution: 42
};

var bar = {
    sub2(a) {
        return a - 2;
    },
    sub4(a) {
        return a - 4;
    },
    sub8(a) {
        return a - 8;
    },
    sub16(a) {
        return a - 16;
    },
    theSolution: 42
};

module.exports.foo = bar;