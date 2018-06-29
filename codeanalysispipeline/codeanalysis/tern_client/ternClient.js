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

    const tern = new ternModule.Server({
        plugins: {
            "modules": {},
            "node": {},
            "es_modules": {},
        },
    });

    // helper "class" to store call expressions
    function CallExpression(file, start, end, name, outerMethod, receiver) {
        this.file = file;
        this.start = start;
        this.end = end;
        this.name = name;
        this.outerMethod = outerMethod;
        this.receiver = receiver;
    }

    let callExpressions = [];

    let requiredModules = {};

    // register AST visitors that get called when tern parses the files
    const astVisitors = {
        VariableDeclaration: function (declNode, _) {
            for (let decl of declNode.declarations) {
                if (decl.init.type === "CallExpression") {
                    if (decl.init.callee.name === "require") {
                        const moduleName = decl.init.arguments[0].value;
                        requiredModules[decl.id.name] = moduleName;
                    }
                }
            }
        },
        FunctionDeclaration: function (fctNode, _) {
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
        CallExpression: function (callNode, ancestors) {
            if (debug) console.log({callNode, ancestors});
            const decl = ancestors.filter(node => node.type === 'FunctionDeclaration').pop();
            let outerMethod = callNode.sourceFile.name;
            if (decl) {
                outerMethod = decl.id.name;
            }

            let functionName = callNode.callee.name;
            let receiver = "this";
            let receiverStart = 0;

            if (!functionName) {
                // parse child expression here as it is not an identifier
                const callee = callNode.callee;
                if (callee.type === "MemberExpression") {
                    functionName = callee.property.name;
                    receiver = callee.object.name;
                }
            }

            callExpressions.push(
                new CallExpression(
                    callNode.sourceFile.name,
                    callNode.start,
                    callNode.end,
                    functionName,
                    outerMethod,
                    receiver
                )
            );

        }
    };
    tern.on("postParse", function (ast, text) {
        acornWalk.ancestor(ast, astVisitors);
    });

    calls = [];

    // parse and analyze files
    try {
        readdirp(
            {root: folderPath, fileFilter: "*.js"},
            (fileInfo) => {
                tern.addFile(fileInfo.name, fs.readFileSync(fileInfo.fullPath));
                if (debug) console.log(`added file ${fileInfo.name} to tern`);
            },
            () => {
                // for each call expression, find the function definition that the call resolves to
                for (let i = 0; i < callExpressions.length; i++) {
                    const callExpression = callExpressions[i];

                    const queryFuncDef = {
                        //query:{type:"definition", end:{line:0, ch:1}, file:"call.js"}
                        query: {
                            type: "definition",
                            end: callExpression.start,
                            file: callExpression.file
                        }
                    };
                    tern.request(queryFuncDef, function (err, data) {
                        if (debug) {
                            console.log("\nCall at:");
                            console.log(callExpression);
                            console.log(".. goes to function defined at:");
                            console.log(data);
                        }

                        calls.push({
                            fromFile: callExpression.file,
                            fromFunction: callExpression.outerMethod,
                            receiver: callExpression.receiver,
                            module: requiredModules[callExpression.receiver],
                            toFile: data.origin,
                            toFunction: callExpression.name,
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
