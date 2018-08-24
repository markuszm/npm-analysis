// Author: Michael Pradel, Markus Zimmermann

import * as fs from "fs";

import * as ternModule from "tern";

import * as acornWalk from "acorn/dist/walk";

import { Call, CallExpression, Function } from "./model";
import * as path from "path";
import { Visitors } from "./traversal";

export class TernClient {
    private ternServer: any;

    constructor(
        callExpressions: CallExpression[],
        requiredModules: any,
        declaredFunctions: Function[],
        importedMethods: any,
        private debug: boolean
    ) {
        const tern = new ternModule.Server({
            plugins: {
                modules: {},
                node: {},
                es_modules: {}
            }
        });

        const visitors = Visitors(
            this,
            callExpressions,
            requiredModules,
            declaredFunctions,
            importedMethods,
            debug
        );

        tern.on("afterLoad", (file: any) => {
            if (file.ast) {
                acornWalk.ancestor(file.ast, visitors);
            }
        });

        this.ternServer = tern;
    }

    addFile(fileName: string, filePath: string): void {
        this.ternServer.addFile(fileName, fs.readFileSync(filePath));
        this.ternServer.flush(() => void 0)
    }

    requestCallExpression(
        callExpression: CallExpression,
        requiredModules: any,
        declaredFunctions: Function[],
        importedMethods: any,
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

                let toFunction = callExpression.name;
                let importedName = importedMethods[callExpression.name];
                if (importedName) {
                    toFunction = importedName;
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
                        callExpression.className,
                        Array.of(...modules.values()),
                        toFunction,
                        callExpression.args,
                        declaredFunctions.some(declFunc => declFunc.id === callExpression.name) &&
                            (callExpression.receiver === "" || callExpression.receiver === "this")
                    )
                );
            });
        });
    }

    requestReferences(start: number, file: string, cb: (err: string, data: any) => void) {
        const queryRefs = {
            query: {
                type: "refs",
                end: start,
                file: file
            }
        };

        this.ternServer.request(queryRefs, cb);
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
    return fileName.replace(path.extname(fileName), "");
}
