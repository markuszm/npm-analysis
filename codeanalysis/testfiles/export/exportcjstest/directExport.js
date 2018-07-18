function addThree(a, b, c) {
    return a + b + c;
}

const foo = {
    subThree(a,b,c) {
        return a-b-c;
    }
};

module.exports = addThree;
module.exports = foo.subThree();

