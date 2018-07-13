// Author: Michael Pradel, Markus Zimmermann
const model = require("./model");
const util = require("./util");

// register AST visitors that get called when tern parses the files
function Visitors(callExpressions, requiredModules, debug) {
    return {
        VariableDeclaration: function(declNode, _) {
            for (let decl of declNode.declarations) {
                const isRequireInit =
                    decl.init && decl.init.type === "CallExpression" && decl.init.callee.name === "require";
                if (isRequireInit) {
                    const variableName = decl.id.name;
                    const moduleName = decl.init.arguments[0].value;
                    requiredModules[variableName] = moduleName;
                    if (debug) {
                        console.log("\nModule Declaration: \n", {
                            Variable: variableName,
                            ModuleName: moduleName
                        });
                    }
                }
            }
        },

        FunctionDeclaration: function(functionDecl, _) {
            if (debug) {
                console.log("\nFunction Declaration: \n", {
                    FileName: functionDecl.sourceFile.name,
                    Start: functionDecl.id.start,
                    End: functionDecl.id.end,
                    Name: functionDecl.id.name
                });
            }
        },

        CallExpression: function(callNode, ancestors) {
            if (debug) console.log("\nCallExpression: \n", { callNode, ancestors });
            const outerMethod = ancestors.filter(node => node.type === "FunctionDeclaration").pop();
            let outerMethodName = callNode.sourceFile.name;
            if (outerMethod) {
                outerMethodName = outerMethod.id.name;
            }

            let functionName = callNode.callee.name;
            let receiver = "this";
            if (!functionName) {
                // parse child expression here as it is not an identifier
                const callee = callNode.callee;
                if (callee.type === "MemberExpression") {
                    functionName = callee.property.name;
                    receiver = callee.object.name;
                }
            }

            const arguments = [];

            for (let argument of callNode.arguments) {
                const argumentAsString = util.expressionToString(argument);
                arguments.push(argumentAsString);
            }

            callExpressions.push(
                new model.CallExpression(
                    callNode.sourceFile.name,
                    callNode.start,
                    callNode.end,
                    functionName,
                    outerMethodName,
                    receiver,
                    arguments
                )
            );
        }
    };
}

exports.Visitors = Visitors;
