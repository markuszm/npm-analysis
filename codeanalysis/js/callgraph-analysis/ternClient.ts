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
        requiredModules: Map<string|number, any>,
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

    deleteFile(fileName: string): void {
        this.ternServer.delFile(fileName);
        this.ternServer.flush(() => void 0)
    }

    requestCallExpression(
        callExpression: CallExpression,
        requiredModules: Map<string|number, any>,
        declaredFunctions: Function[],
        importedMethods: Map<string,string>,
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
                end: callExpression.end,
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
                        safePush(modules, requiredModules.get(ref.end));
                    }
                }

                let toFunction = callExpression.name;
                let importedName = importedMethods.get(callExpression.name);
                if (importedName) {
                    toFunction = importedName;
                }

                safePush(
                    modules,
                    requiredModules.get(dataFunc.start),
                    requiredModules.get(callExpression.receiver),
                    requiredModules.get(callExpression.name)
                );
                const call = new Call(
                    trimExt(callExpression.file),
                    callExpression.outerMethod,
                    callExpression.receiver,
                    callExpression.className,
                    isLocalFunction(declaredFunctions, callExpression) ? [] : Array.of(...modules.values()),
                    toFunction,
                    callExpression.args,
                    isLocalFunction(declaredFunctions, callExpression)
                );

                // set function name to default if receiver is empty but module reference is found
                if (call.receiver === "" && call.modules.length > 0 && !importedName) {
                    call.toFunction = "default"
                }

                calls.push(
                    call
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

function isLocalFunction(declaredFunctions: Function[], callExpression: CallExpression) {
    return declaredFunctions.some(declFunc => declFunc.id === callExpression.name) &&
        (callExpression.receiver === "" || callExpression.receiver === "this");
}