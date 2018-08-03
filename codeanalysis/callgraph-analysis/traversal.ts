// Author: Michael Pradel, Markus Zimmermann
import * as model from "./model";

import { patternToString, expressionToString } from "./util";
import {
    AssignmentExpression,
    CallExpression,
    Expression,
    FunctionDeclaration,
    ImportDeclaration,
    NewExpression,
    Node,
    Super,
    VariableDeclaration,
    VariableDeclarator
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

    var crossReferences: any = {};

    return {
        /* detect imports */
        VariableDeclaration: function(declNode: VariableDeclaration, _: Array<Node>) {
            for (let decl of declNode.declarations) {
                const declarator = decl.init;
                if (declarator) {
                    const callExpr = getRequireCallExpr(declarator);
                    if (callExpr) {
                        const variableName = patternToString(decl.id);
                        const firstArg = callExpr.arguments[0];
                        const moduleName = firstArg.type === "Literal" ? firstArg.value : "";
                        requiredModules[decl.start] = moduleName;
                        crossReferences[variableName] = decl.start;
                        if (debug) {
                            console.log("\nModule Declaration: \n", {
                                Variable: variableName,
                                ModuleName: moduleName
                            });
                        }
                    } else {
                        const rightSideExpr = expressionToString(declarator);
                        if (crossReferences[rightSideExpr]) {
                            requiredModules[decl.start] =
                                requiredModules[crossReferences[rightSideExpr]];
                        }
                    }
                }
            }
        },

        AssignmentExpression: function(assignmentExpr: AssignmentExpression, _: Array<Node>) {
            const callExpr = getRequireCallExpr(assignmentExpr.right);
            if (callExpr) {
                const variableName = patternToString(assignmentExpr.left);
                const firstArg = callExpr.arguments[0];
                const moduleName = firstArg.type === "Literal" ? firstArg.value : "";
                requiredModules[assignmentExpr.left.start] = moduleName;
                crossReferences[variableName] = assignmentExpr.left.start;
                if (debug) {
                    console.log("\nModule Declaration: \n", {
                        Variable: variableName,
                        ModuleName: moduleName
                    });
                }
            } else {
                const rightSideExpr = expressionToString(assignmentExpr.right);
                if (crossReferences[rightSideExpr]) {
                    requiredModules[assignmentExpr.left.start] =
                        requiredModules[crossReferences[rightSideExpr]];
                }
            }
        },

        ImportDeclaration: function(importDecl: ImportDeclaration, _: Array<Node>) {
            if (debug) {
                console.log({ ImportDeclaration: importDecl });
            }

            const moduleName = importDecl.source.value;

            for (let specifier of importDecl.specifiers) {
                requiredModules[specifier.local.name] = moduleName;
            }
        },

        FunctionDeclaration: function(functionDecl: FunctionDeclaration, _: Array<Node>) {
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
            let outerMethodName = ".root";
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
                    functionName = expressionToString(callee.property);
                    receiver =
                        callee.object.type === "Super"
                            ? "super"
                            : expressionToString(callee.object);
                    break;
                case "Super":
                    receiver = "";
                    functionName = "super";
                    break;
                default:
                    receiver = "";
                    functionName = expressionToString(callee);
            }

            const args = [];

            for (let argument of callNode.arguments) {
                const argumentAsString =
                    argument.type === "SpreadElement"
                        ? "..." + expressionToString(argument.argument)
                        : expressionToString(argument);
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
        },

        NewExpression: function(newExpr: NewExpression, ancestors: Array<Node>) {
            if (debug) console.log("\nNewExpression: \n", { newExpr, ancestors });
            const outerMethod: FunctionDeclaration | undefined = ancestors
                .filter(node => node.type === "FunctionDeclaration")
                .pop() as FunctionDeclaration;
            let outerMethodName = ".root";
            if (outerMethod) {
                outerMethodName = outerMethod.id ? outerMethod.id.name : "default";
            }

            let functionName: string;
            let receiver: string = "";

            const callee = newExpr.callee;

            switch (callee.type) {
                case "Identifier":
                    functionName = callee.name;
                    break;
                case "MemberExpression":
                    functionName = expressionToString(callee.property);
                    break;
                case "Super":
                    functionName = "super";
                    break;
                default:
                    functionName = expressionToString(callee);
            }

            const args = [];

            for (let argument of newExpr.arguments) {
                const argumentAsString =
                    argument.type === "SpreadElement"
                        ? "..." + expressionToString(argument.argument)
                        : expressionToString(argument);
                args.push(argumentAsString);
            }

            // when functionName is module then add left side of new expression to requiredModules
            if (crossReferences[functionName]) {
                const outerDeclarator: VariableDeclarator = ancestors
                    .filter(node => node.type === "VariableDeclarator")
                    .pop() as VariableDeclarator;

                const outerAssignment: AssignmentExpression = ancestors
                    .filter(node => node.type === "AssignmentExpression")
                    .pop() as AssignmentExpression;

                if (outerDeclarator) {
                    receiver = patternToString(outerDeclarator.id);
                    requiredModules[outerDeclarator.start] =
                        requiredModules[crossReferences[functionName]];
                }

                if (outerAssignment) {
                    receiver = patternToString(outerAssignment.left);
                    requiredModules[outerAssignment.start] =
                        requiredModules[crossReferences[functionName]];
                }
            }

            callExpressions.push(
                new model.CallExpression(
                    newExpr.sourceFile.name,
                    newExpr.callee.start,
                    newExpr.callee.end,
                    "new " + functionName,
                    outerMethodName,
                    receiver,
                    args
                )
            );
        }
    };
}
