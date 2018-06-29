// Author: Michael Pradel

const fs = require("fs");
const path = require("path");
const process = require("process");

const acornWalk = require("acorn/dist/walk");
const ternModule = require("tern");
const readdirp = require("readdirp");

let debug = false;

(function () {
    const args = process.argv.slice(2);

    if (args.length === 0) {
        console.error("missing folder path");
        return;
    }
    const folderPath = args[0];

    if (args.length > 1 && args[1] === "debug") {
        debug = true
    }

    const tern = new ternModule.Server({});

    // helper "class" to store call expressions
    function CallExpression(file, start, end, name) {
        this.file = file;
        this.start = start;
        this.end = end;
        this.name = name;
    }

    let callExpressions = [];

    // register AST visitors that get called when tern parses the files
    const astVisitors = {
        FunctionDeclaration: function (fctNode) {
            if (debug) console.log(
                fctNode.sourceFile.name +
                ", " +
                fctNode.id.start +
                ", " +
                fctNode.id.end +
                ", " +
                fctNode.id.name
            );
        },
        CallExpression: function (callNode) {
            if (debug) console.log(callNode);
            callExpressions.push(
                new CallExpression(
                    callNode.sourceFile.name,
                    callNode.start,
                    callNode.end,
                    callNode.callee.name
                )
            );
        }
    };
    tern.on("postParse", function (ast, text) {
        acornWalk.simple(ast, astVisitors);
    });

    calls = []

    // parse and analyze files
    try {
        readdirp(
            {root: folderPath, fileFilter: "*.js"},
            (fileInfo) => {
                tern.addFile(fileInfo.name, fs.readFileSync(fileInfo.fullPath));
                if (debug) console.log(`added file ${fileInfo.name} to tern`);
            },
            ()  => {
                // for each call expression, find the function definition that the call resolves to
                for (let i = 0; i < callExpressions.length; i++) {
                    const callExpression = callExpressions[i];
                    const query = {
                        //query:{type:"definition", end:{line:0, ch:1}, file:"call.js"}
                        query: {
                            type: "definition",
                            end: callExpression.start,
                            file: callExpression.file
                        }
                    };
                    tern.request(query, function (err, data) {
                        if (debug) {
                            console.log("\nCall at:");
                            console.log(callExpression);
                            console.log(".. goes to function defined at:");
                            console.log(data);
                        }

                        calls.push({
                            from: callExpression.file,
                            functionName: callExpression.name,
                            to: data.origin
                        })
                    });
                }
                console.log(JSON.stringify(calls))
            }
        );
    } catch (e) {
        console.error(e);
        console.error("could not find any files")
    }
})();
