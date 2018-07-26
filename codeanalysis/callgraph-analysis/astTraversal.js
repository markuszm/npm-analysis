// Author: Michael Pradel, Markus Zimmermann
const model = require("./model");
const util = require("./util");

// register AST visitors that get called when tern parses the files
function Visitors(callExpressions, requiredModules, debug) {
    function getRequireCallExpr(decl) {
        switch (decl.type) {
            case "CallExpression":
                return decl &&
                decl.type === "CallExpression" &&
                decl.callee.type === "Identifier" &&
                decl.callee.name === "require"
                    ? decl
                    : null;
            case "MemberExpression":
                return getRequireCallExpr(decl.object);
            default:
                return null;
        }
    }

    return {
        /* detect imports */
        VariableDeclaration: function(declNode, _) {
            for (let decl of declNode.declarations) {
                const declarator = decl.init;
                if (declarator) {
                    const callExpr = getRequireCallExpr(declarator)
                    if (callExpr) {
                        const variableName = decl.id.name;
                        const moduleName = callExpr.arguments[0].value;
                        requiredModules[decl.start] = moduleName;
                        if (debug) {
                            console.log("\nModule Declaration: \n", {
                                Variable: variableName,
                                ModuleName: moduleName
                            });
                        }
                    }
                }

            }
        },

        ImportDeclaration: function(importDecl, _) {
            if (debug) {
                console.log({ImportDeclaration: importDecl});
            }

            const moduleName = importDecl.source.value;

            for (let specifier of importDecl.specifiers) {
                requiredModules[specifier.local.name] = moduleName;
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

        /* track function calls */
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
