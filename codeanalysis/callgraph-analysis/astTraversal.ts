// Author: Michael Pradel, Markus Zimmermann
import * as model from "./model";

import * as util from "./util";
import {
    CallExpression,
    Expression,
    FunctionDeclaration,
    ImportDeclaration,
    Node,
    Super,
    VariableDeclaration
} from "./@types/estree";

// register AST visitors that get called when tern parses the files
export function Visitors(
    callExpressions: Array<model.CallExpression>,
    requiredModules: any,
    debug: boolean
): any {
    function getRequireCallExpr(decl: Expression | Super): CallExpression | null {
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
        VariableDeclaration: function(declNode: VariableDeclaration, _: any) {
            for (let decl of declNode.declarations) {
                const declarator = decl.init;
                if (declarator) {
                    const callExpr = getRequireCallExpr(declarator);
                    if (callExpr) {
                        const variableName = util.patternToString(decl.id);
                        const firstArg = callExpr.arguments[0];
                        const moduleName = firstArg.type === "Literal" ? firstArg.value : "";
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

        ImportDeclaration: function(importDecl: ImportDeclaration, _: any) {
            if (debug) {
                console.log({ ImportDeclaration: importDecl });
            }

            const moduleName = importDecl.source.value;

            for (let specifier of importDecl.specifiers) {
                requiredModules[specifier.local.name] = moduleName;
            }
        },

        FunctionDeclaration: function(functionDecl: FunctionDeclaration, _: any) {
            if (debug && functionDecl.id) {
                console.log("\nFunction Declaration: \n", {
                    FileName: functionDecl.sourceFile.name,
                    Start: functionDecl.id.start,
                    End: functionDecl.id.end,
                    Name: functionDecl.id.name
                });
            }
        },

        /* track function calls */
        CallExpression: function(callNode: CallExpression, ancestors: Array<Node>) {
            if (debug) console.log("\nCallExpression: \n", { callNode, ancestors });
            const outerMethod: FunctionDeclaration | undefined = ancestors
                .filter(node => node.type === "FunctionDeclaration")
                .pop() as FunctionDeclaration;
            let outerMethodName = callNode.sourceFile.name;
            if (outerMethod) {
                outerMethodName = outerMethod.id ? outerMethod.id.name : "default";
            }

            let functionName: string;
            let receiver: string;

            const callee = callNode.callee;

            switch (callee.type) {
                case "Identifier":
                    receiver = "this";
                    functionName = callee.name;
                    break;
                case "MemberExpression":
                    functionName = util.expressionToString(callee.property);
                    receiver =
                        callee.object.type === "Super"
                            ? "super"
                            : util.expressionToString(callee.object);
                    break;
                case "Super":
                    receiver = "";
                    functionName = "super";
                    break;
                default:
                    receiver = "this";
                    functionName = util.expressionToString(callee);
            }

            const args = [];

            for (let argument of callNode.arguments) {
                const argumentAsString =
                    argument.type === "SpreadElement"
                        ? "..." + util.expressionToString(argument.argument)
                        : util.expressionToString(argument);
                args.push(argumentAsString);
            }

            callExpressions.push(
                new model.CallExpression(
                    callNode.sourceFile.name,
                    callNode.start,
                    callNode.end,
                    functionName,
                    outerMethodName,
                    receiver,
                    args
                )
            );
        }
    };
}
