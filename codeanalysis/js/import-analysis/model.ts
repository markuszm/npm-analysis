export class Import {
    constructor(public id: string, public fromModule: string, public moduleName: string, public bundleType: string, public imported?: string) {}
}
