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

        const queryRefs = {
            query: {
                type: "refs",
                end: callExpression.start,
                file: callExpression.file
            }
        };

        this.ternServer.request(queryFuncDef, (_: any, dataFunc: any) => {
            if (!dataFunc) return;

            if (this.debug) {
                console.log(
                    "\nCall at: %o \n .. goes to function defined at: \n %o",
                    callExpression,
                    dataFunc
                );
            }

            this.ternServer.request(queryRefs, (_: any, dataRefs: any) => {
                if (this.debug) {
                    console.log("\n All Refs of %o \n %o", callExpression, dataRefs);
                }

                let moduleName = "";

                if(dataRefs) {
                    for (let ref of dataRefs.refs) {
                        moduleName = requiredModules[ref.start] || moduleName;
                    }
                }

                moduleName =
                    moduleName ||
                    requiredModules[dataFunc.start] ||
                    requiredModules[callExpression.receiver] ||
                    requiredModules[callExpression.name];
                calls.push(
                    new Call(
                        callExpression.file,
                        callExpression.outerMethod,
                        callExpression.receiver,
                        moduleName,
                        dataFunc.origin,
                        callExpression.name,
                        callExpression.args
                    )
                );
            });
        });
    }
}
