// Author: Michael Pradel, Markus Zimmermann

import * as fs from "fs";

import * as ternModule from "tern";

export class TernClient {
    private ternServer: any;

    constructor(private debug: boolean) {
        this.ternServer = new ternModule.Server({
            plugins: {
                modules: {},
                node: {},
                es_modules: {}
            }
        });
    }

    public addFile(fileName: string, filePath: string) {
        try {
            this.ternServer.addFile(fileName, fs.readFileSync(filePath));
        }
        catch (e) {
            if (this.debug) {
                console.error(e)
            }
        }
    }

    public requestDefinition(definitionQuery: DefinitionQuery, cb: (err:string, data: any) => void) {
        const queryFuncDef = {
            query: {
                type: "definition",
                end: definitionQuery.end,
                file: definitionQuery.file
            }
        };

        this.ternServer.request(queryFuncDef, cb);
    }
}

export class DefinitionQuery {
    constructor(public end: number, public file: string) {}
}