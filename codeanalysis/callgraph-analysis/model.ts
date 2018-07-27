// Author: Michael Pradel, Markus Zimmermann

export class CallExpression {
    constructor(
        public file: string,
        public start: number,
        public end: number,
        public name: string,
        public outerMethod: string,
        public receiver: string,
        public args: Array<string>
    ) {}
}

export class Call {
    constructor(
        public fromFile: string,
        public fromFunction: string,
        public receiver: string,
        public module: string,
        public toFile: string,
        public toFunction: string,
        public args: Array<string>
    ) {}
}
