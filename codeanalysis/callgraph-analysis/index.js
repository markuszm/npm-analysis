// Author: Michael Pradel, Markus Zimmermann

const process = require("process");
const readdirp = require("readdirp");

const TernClient = require("./ternClient");
const traversal = require("./astTraversal");

/* argument parsing */
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

/* create resources */
const calls = []; // type model.Call
const callExpressions = []; // type model.CallExpression
const requiredModules = {}; // map global variable name -> module name
const visitors = traversal.Visitors(callExpressions, requiredModules, debug);
const ternClient = new TernClient(visitors, debug);

/* process files from folder */
try {
    readdirp(
        { root: folderPath, fileFilter: "*.js", directoryFilter: [ '!.git', '!node_modules', '!assets' ]},
        fileInfo => {
            ternClient.AddFile(fileInfo.name, fileInfo.fullPath);
            if (debug) console.log(`Added file ${fileInfo.name} to tern`);
        },
        () => {
            if (debug) console.log(`Finished collecting AST walking`);
            // for each call expression, find the function definition that the call resolves to
            for (let i = 0; i < callExpressions.length; i++) {
                const callExpression = callExpressions[i];
                ternClient.RequestCallExpression(callExpression, requiredModules, calls);
            }
            console.log(JSON.stringify(calls));
        }
    );
} catch (e) {
    console.error(e);
    console.error("could not find any files");
}
