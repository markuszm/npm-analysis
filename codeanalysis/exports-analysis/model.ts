export class Export {
    constructor(
        public type: string,
        public id: string,
        public bundleType: string,
        public file: string,
        public isDefault: boolean,
        public local?: string,
    ) {}
}

export class Function {
    constructor(public id: string, public start: number, public params: string[]) {}
}

export class Variable {
    constructor(
        public id: string,
        public start: number,
        public kind: string,
        public value?: string
    ) {}
}

export class Class {
    constructor(
        public id: string,
        public methods: string[],
        public superClass: string | null
    ) {}
}

export class ObjectExpr{
    constructor(
        public id: string,
        public methods: Function[],
        public variables: Variable[]
    ){}
}