// Author: Michael Pradel, Markus Zimmermann

import { Call, CallExpression, Function } from "./model";

import * as fs from "fs";

import * as process from "process";

import readdirp from "readdirp";

import { TernClient } from "./ternClient";
import * as path from "path";


/* argument parsing */
let debug = false;
const args = process.argv.slice(2);
if (args.length === 0) {
    console.error("missing folder path");
}
const entryPath = args[0];
const sizeLimit = Number(args[1]);
if (args.length > 2 && args[2] === "debug") {
    debug = true;
}

const calls: Call[] = [];

const stats = fs.statSync(entryPath);

function getFileNameInsidePackage(fileInfo: any) {
    const fullPath: string = fileInfo.fullPath;
    const regexFileName = /(?:\/?.+)(?:\/package\/)(.+)/;
    if (fullPath.indexOf("package") != -1) {
        let [, fileName]: RegExpMatchArray = fullPath.match(regexFileName) || [];
        return !fileName || fileName === "" ? fileInfo.name : fileName;
    }
    return fileInfo.name;
}

if (stats.isDirectory()) {
    /* process files from folder */
    try {
        readdirp(
            {
                root: entryPath,
                fileFilter: function(fileInfo: any) {
                    const fileName = fileInfo.name;
                    const ext = path.extname(fileName);
                    switch (ext) {
                        case ".js":
                            return fileInfo.stat.size <= sizeLimit;
                        case "":
                            const file = fs.readFileSync(fileInfo.fullPath, { encoding: "utf8" });
                            return file.startsWith("#!/usr/bin/env node");
                        default:
                            return false;
                    }
                },
                directoryFilter: ["!.git", "!node_modules", "!assets"]
            },
            (fileInfo: any) => {
                const callExpressions: CallExpression[] = [];
                const declaredFunctions: Function[] = [];
                const requiredModules = new Map<string|number, any>(); // map global variable name -> module name
                const importedMethods = new Map<string, string>(); // map local import name to imported name
                const ternClient = new TernClient(
                    callExpressions,
                    requiredModules,
                    declaredFunctions,
                    importedMethods,
                    debug
                );

                const fileName = getFileNameInsidePackage(fileInfo);
                ternClient.addFile(fileName, fileInfo.fullPath);

                // for each call expression, find the function definition that the call resolves to
                for (let i = 0; i < callExpressions.length; i++) {
                    const callExpression = callExpressions[i];
                    ternClient.requestCallExpression(
                        callExpression,
                        requiredModules,
                        declaredFunctions,
                        importedMethods,
                        calls
                    );
                }

                if (debug) console.log({ requiredModules, declaredFunctions });
            },
            () => {
                console.log(JSON.stringify(calls));
            }
        );
    } catch (e) {
        console.error(e);
        console.error("could not find any files");
    }
} else {
    const callExpressions: CallExpression[] = [];
    const declaredFunctions: Function[] = [];
    const requiredModules = new Map<string|number, any>(); // map global variable name -> module name
    const importedMethods = new Map<string, string>(); // map local import name to imported name

    const ternClient = new TernClient(
        callExpressions,
        requiredModules,
        declaredFunctions,
        importedMethods,
        debug
    );

    ternClient.addFile(entryPath, entryPath);
    if (debug) console.log(`Added file ${entryPath} to tern`);

    // for each call expression, find the function definition that the call resolves to
    for (let i = 0; i < callExpressions.length; i++) {
        const callExpression = callExpressions[i];
        ternClient.requestCallExpression(
            callExpression,
            requiredModules,
            declaredFunctions,
            importedMethods,
            calls
        );
    }

    if (debug) console.log({ requiredModules, importedMethods });

    console.log(JSON.stringify(calls));
}
