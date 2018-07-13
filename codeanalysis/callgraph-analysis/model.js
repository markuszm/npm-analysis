// Author: Michael Pradel, Markus Zimmermann

class CallExpression {
    constructor(file, start, end, name, outerMethod, receiver, args) {
        this.file = file;
        this.start = start;
        this.end = end;
        this.name = name;
        this.outerMethod = outerMethod;
        this.receiver = receiver;
        this.arguments = args;
    }
}

class Call {
    constructor(fromFile, fromFunction, receiver, module, toFile, toFunction, args) {
        this.fromFile = fromFile;
        this.fromFunction = fromFunction;
        this.receiver = receiver;
        this.module = module;
        this.toFile = toFile;
        this.toFunction = toFunction;
        this.arguments = args;
    }
}

exports.CallExpression = CallExpression;
exports.Call = Call;
