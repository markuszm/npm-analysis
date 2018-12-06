// Author: Michael Pradel, Markus Zimmermann

export class CallExpression {
    constructor(
        public file: string,
        public start: number,
        public end: number,
        public loc: SourceLocation,
        public name: string,
        public outerMethod: string,
        public receiver: string,
        public className: string,
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
        public path: string,
        public loc: SourceLocation,
        public fromModule: string,
        public fromFunction: string,
        public receiver: string,
        public className: string,
        public modules: string[],
        public toFunction: string,
        public args: string[],
        public isLocal: boolean
    ) {}
}

export interface SourceLocation {
    start: Location,
    end: Location
}

export interface Location {
    line: number,
    ch: number
}