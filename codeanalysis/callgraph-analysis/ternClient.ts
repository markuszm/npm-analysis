// Author: Michael Pradel, Markus Zimmermann

import * as fs from "fs";

import * as ternModule from "tern";

import * as acornWalk from "acorn/dist/walk";

import { Call, CallExpression, Function } from "./model";

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
        declaredFunctions: Function[],
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

                const modules: Set<string> = new Set<string>();

                if (dataRefs) {
                    for (let ref of dataRefs.refs) {
                        safePush(modules, requiredModules[ref.start]);
                    }
                }

                safePush(
                    modules,
                    requiredModules[dataFunc.start],
                    requiredModules[callExpression.receiver],
                    requiredModules[callExpression.name]
                );
                calls.push(
                    new Call(
                        trimExt(callExpression.file),
                        callExpression.outerMethod,
                        callExpression.receiver,
                        Array.of(...modules.values()),
                        callExpression.name,
                        callExpression.args,
                        declaredFunctions.some(declFunc => declFunc.id === callExpression.name) &&
                            callExpression.receiver === "this"
                    )
                );
            });
        });
    }
}

function safePush<T>(array: Set<T>, ...items: Array<T>) {
    for (let item of items) {
        if (item) {
            array.add(item);
        }
    }
}

function trimExt(fileName: string): string {
    return fileName.replace(".js", "")
}