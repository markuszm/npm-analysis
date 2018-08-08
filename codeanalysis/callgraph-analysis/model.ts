// Author: Michael Pradel, Markus Zimmermann

export class CallExpression {
    constructor(
        public file: string,
        public start: number,
        public end: number,
        public name: string,
        public outerMethod: string,
        public receiver: string,
        public args: string[]
    ) {}
}

export class Function {
    constructor(
        public id: string,
        public start: number,
        public params: string[]
    ) {}
}

export class Call {
    constructor(
        public fromModule: string,
        public fromFunction: string,
        public receiver: string,
        public modules: string[],
        public toFunction: string,
        public args: string[],
        public isLocal: boolean
    ) {}
}
