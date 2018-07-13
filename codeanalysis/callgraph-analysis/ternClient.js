// Author: Michael Pradel, Markus Zimmermann

const acornWalk = require("acorn/dist/walk");
const ternModule = require("tern");
const fs = require("fs");

function NewTernClient(visitors) {
    const tern = new ternModule.Server({
        plugins: {
            modules: {},
            node: {},
            es_modules: {}
        }
    });

    tern.on("postParse", function(ast, _) {
        acornWalk.ancestor(ast, visitors);
    });

    return tern;
}

function AddFile(tern, fileName, filePath) {
    const body = fs.readFileSync(filePath);
    tern.addFile(fileName, body);
}

function RequestCallExpression(tern, callExpression, requiredModules, calls, debug) {
    const queryFuncDef = {
        query: {
            type: "definition",
            end: callExpression.start,
            file: callExpression.file
        }
    };

    tern.request(queryFuncDef, function(err, data) {
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
            toFunction: callExpression.name
        });
    });
}

exports.NewTernClient = NewTernClient;
exports.RequestCallExpression = RequestCallExpression;
exports.AddFile = AddFile;
