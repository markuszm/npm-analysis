// Author: Michael Pradel, Markus Zimmermann

import * as fs from "fs";

import * as ternModule from "tern";

import * as acornWalk from "acorn/dist/walk";

import { Call, CallExpression } from "./model";

export class TernClient {
    private ternServer: any;

    constructor(visitors: any, private debug: boolean) {
        const tern = new ternModule.Server({
            plugins: {
                modules: {},
                node: {},
                es_modules: {}
            }
        });

        tern.on("postParse", (ast: any, _: any) => {
            if (ast) {
                acornWalk.ancestor(ast, visitors);
            }
        });

        this.ternServer = tern;
    }

    addFile(fileName: string, filePath: string): void {
        this.ternServer.addFile(fileName, fs.readFileSync(filePath));
    }

    requestCallExpression(
        callExpression: CallExpression,
        requiredModules: any,
        calls: Array<Call>
    ): void {
        const queryFuncDef = {
            query: {
                type: "definition",
                end: callExpression.start,
                file: callExpression.file
            }
        };

        this.ternServer.request(queryFuncDef, (err: any, data: any) => {
            if (!data) return;

            if (this.debug) {
                console.log(
                    "\nCall at: %o \n .. goes to function defined at: \n %o \n Context: %s",
                    callExpression,
                    data,
                    data.context
                );
            }

            const moduleName: string =
                requiredModules[data.start] ||
                requiredModules[callExpression.receiver] ||
                requiredModules[callExpression.name];
            calls.push(
                new Call(
                    callExpression.file,
                    callExpression.outerMethod,
                    callExpression.receiver,
                    moduleName,
                    data.origin,
                    callExpression.name,
                    callExpression.args
                )
            );
        });
    }
}
