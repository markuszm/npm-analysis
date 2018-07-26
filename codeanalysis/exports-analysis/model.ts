export class Export {
    constructor(
        public type: string,
        public id: string,
        public bundleType: string,
        public file: string,
        public isDefault: boolean
    ) {}
}

export class Function {
    constructor(public id: string, public start: number, public params: Array<string>) {}
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
        public methods: Array<string>,
        public superClass: string | null
    ) {}
}
