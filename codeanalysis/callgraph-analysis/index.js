// Author: Michael Pradel, Markus Zimmermann

const process = require("process");
const readdirp = require("readdirp");

const ternClient = require("./ternClient");
const traversal = require("./astTraversal");

// argument parsing
let debug = false;
const args = process.argv.slice(2);
if (args.length === 0) {
    console.error("missing folder path");
    return;
}
const folderPath = args[0];
if (args.length > 1 && args[1] === "debug") {
    debug = true;
}

// create resources
const calls = [];
const callExpressions = [];
const requiredModules = {};
const visitors = traversal.Visitors(callExpressions, requiredModules, debug);
const tern = ternClient.NewTernClient(visitors);

// process files from folder
try {
    readdirp(
        { root: folderPath, fileFilter: "*.js" },
        fileInfo => {
            ternClient.AddFile(tern, fileInfo.name, fileInfo.fullPath);
            if (debug) console.log(`added file ${fileInfo.name} to tern`);
        },
        () => {
            // for each call expression, find the function definition that the call resolves to
            for (let i = 0; i < callExpressions.length; i++) {
                const callExpression = callExpressions[i];
                ternClient.RequestCallExpression(
                    tern,
                    callExpression,
                    requiredModules,
                    calls,
                    debug
                );
            }
            console.log(JSON.stringify(calls));
        }
    );
} catch (e) {
    console.error(e);
    console.error("could not find any files");
}
