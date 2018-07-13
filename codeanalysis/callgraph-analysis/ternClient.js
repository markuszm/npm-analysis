// Author: Michael Pradel, Markus Zimmermann

const fs = require("fs");

const acornWalk = require("acorn/dist/walk");
const ternModule = require("tern");

const model = require("./model");

class TernClient {
    constructor(visitors, debug) {
        this.debug = debug;

        const tern = new ternModule.Server({
            plugins: {
                modules: {},
                node: {},
                es_modules: {}
            }
        });

        tern.on("postParse", (ast, _) => {
            if(ast) {
                acornWalk.ancestor(ast, visitors)
            }
        });

        this.ternServer = tern;
    }

    AddFile(fileName, filePath) {
        this.ternServer.addFile(fileName, fs.readFileSync(filePath));
    }

    RequestCallExpression(callExpression, requiredModules, calls) {
        const queryFuncDef = {
            query: {
                type: "definition",
                end: callExpression.start,
                file: callExpression.file
            }
        };

        this.ternServer.request(queryFuncDef, (err, data) => {
            if (!data) return;

            if (this.debug) {
                console.log(
                    "\nCall at: %o \n .. goes to function defined at: \n %o",
                    callExpression,
                    data
                );
            }

            calls.push(
                new model.Call(
                    callExpression.file,
                    callExpression.outerMethod,
                    callExpression.receiver,
                    requiredModules[callExpression.receiver],
                    data.origin,
                    callExpression.name,
                    callExpression.arguments
                )
            );
        });
    }
}

module.exports = TernClient;
