// Author: Michael Pradel, Markus Zimmermann

import { Call, CallExpression } from "./model";

import * as fs from "fs";

import * as process from "process";

import readdirp from "readdirp";

import { TernClient } from "./ternClient";
import { Visitors } from "./astTraversal";

/* argument parsing */
let debug = false;
const args = process.argv.slice(2);
if (args.length === 0) {
    console.error("missing folder path");
}
const path = args[0];
if (args.length > 1 && args[1] === "debug") {
    debug = true;
}

/* create resources */
const calls: Array<Call> = [];
const callExpressions: Array<CallExpression> = [];
const requiredModules = {}; // map global variable name -> module name
const visitors = Visitors(callExpressions, requiredModules, debug);
const ternClient = new TernClient(visitors, debug);

const stats = fs.statSync(path);

if (stats.isDirectory()) {
    /* process files from folder */
    try {
        readdirp(
            {
                root: path,
                fileFilter: "*.js",
                directoryFilter: ["!.git", "!node_modules", "!assets"]
            },
            (fileInfo: any) => {
                ternClient.addFile(fileInfo.name, fileInfo.fullPath);
                if (debug) console.log(`Added file ${fileInfo.name} to tern`);
            },
            () => {
                if (debug) console.log(`Finished AST walking`);
                // for each call expression, find the function definition that the call resolves to
                for (let i = 0; i < callExpressions.length; i++) {
                    const callExpression = callExpressions[i];
                    ternClient.requestCallExpression(callExpression, requiredModules, calls);
                }

                if (debug) console.log({requiredModules});

                console.log(JSON.stringify(calls));
            }
        );
    } catch (e) {
        console.error(e);
        console.error("could not find any files");
    }
} else {
    ternClient.addFile(path, path);
    if (debug) console.log(`Added file ${path} to tern`);

    // for each call expression, find the function definition that the call resolves to
    for (let i = 0; i < callExpressions.length; i++) {
        const callExpression = callExpressions[i];
        ternClient.requestCallExpression(callExpression, requiredModules, calls);
    }

    if (debug) console.log({requiredModules});

    console.log(JSON.stringify(calls));
}