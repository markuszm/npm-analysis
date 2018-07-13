// Author: Michael Pradel, Markus Zimmermann

const CallExpression = require("./callExpression").CallExpression;

// register AST visitors that get called when tern parses the files
function Visitors(callExpressions, requiredModules, debug) {
    return (astVisitors = {
        VariableDeclaration: function(declNode, _) {
            for (let decl of declNode.declarations) {
                if (decl.init.type === "CallExpression") {
                    if (decl.init.callee.name === "require") {
                        requiredModules[decl.id.name] = decl.init.arguments[0].value;
                    }
                }
            }
        },
        FunctionDeclaration: function(fctNode, _) {
            if (debug)
                console.log(
                    fctNode.sourceFile.name +
                        ", " +
                        fctNode.id.start +
                        ", " +
                        fctNode.id.end +
                        ", " +
                        fctNode.id.name
                );
        },
        CallExpression: function(callNode, ancestors) {
            if (debug) console.log({ callNode, ancestors });
            const decl = ancestors.filter(node => node.type === "FunctionDeclaration").pop();
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
    });
}

exports.Visitors = Visitors;
