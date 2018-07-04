function abs(x) {
    return Math.abs(x);
}
function sqrt(x) {
    return Math.sqrt(x);
}

module.exports.abs = abs;
module.exports.sqrt = sqrt;
module.exports.pow = function(x, exp) {
    return Math.pow(x, exp);
};
module.exports.floor = x => Math.floor(x);
module.exports.parseInt = (x, r) => Number.parseInt(x, r);
