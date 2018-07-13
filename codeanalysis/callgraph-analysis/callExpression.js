// Author: Michael Pradel, Markus Zimmermann

// helper "class" to store call expressions
function CallExpression(file, start, end, name, outerMethod, receiver) {
    this.file = file;
    this.start = start;
    this.end = end;
    this.name = name;
    this.outerMethod = outerMethod;
    this.receiver = receiver;
}

exports.CallExpression = CallExpression;
