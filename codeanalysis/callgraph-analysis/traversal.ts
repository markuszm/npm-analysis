// Author: Michael Pradel, Markus Zimmermann
import * as model from "./model";

import { patternToString, expressionToString, extractFunctionInfo } from "./util";
import {
    AssignmentExpression,
    CallExpression,
    Expression,
    FunctionDeclaration,
    ImportDeclaration,
    NewExpression,
    Node,
    RegExpLiteral,
    Super,
    VariableDeclaration,
    VariableDeclarator
} from "./@types/estree";
import { Function } from "./model";
import { TernClient } from "./ternClient";

// register AST visitors that get called when tern parses the files
export function Visitors(
    ternClient: TernClient,
    callExpressions: model.CallExpression[],
    requiredModules: Map<string | number, any>,
    definedFunctions: Function[],
    importedMethods: Map<string, string>,
    debug: boolean
): any {
    function getRequireCallExpr(decl: Expression | Super): CallExpression | null {
        switch (decl.type) {
            case "CallExpression":
                if (isRequireCall(decl)) {
                    return decl;
                } else {
                    if (decl.callee.type === "MemberExpression") {
                        return getRequireCallExpr(decl.callee);
                    } else {
                        return null;
                    }
                }
            case "MemberExpression":
                return getRequireCallExpr(decl.object);
            default:
                return null;
        }
    }

    function isRequireCall(decl: CallExpression) {
        return (
            decl &&
            decl.type === "CallExpression" &&
            decl.callee.type === "Identifier" &&
            decl.callee.name === "require"
        );
    }

    var crossReferences: any = {};

    var classReceivers: any = {};

    return {
        /* detect imports */
        VariableDeclaration: function(declNode: VariableDeclaration, _: Node[]) {
            for (let decl of declNode.declarations) {
                const declarator = decl.init;
                if (declarator) {
                    const callExpr = getRequireCallExpr(declarator);
                    if (callExpr) {
                        const variableName = patternToString(decl.id);
                        const firstArg = callExpr.arguments[0];
                        const moduleName = firstArg.type === "Literal" ? firstArg.value : "";
                        requiredModules.set(decl.start, moduleName);
                        crossReferences[variableName] = decl.start;
                        if (debug) {
                            console.log("\nModule Declaration: \n", {
                                Variable: variableName,
                                ModuleName: moduleName
                            });
                        }

                        const imported =
                            declarator.type === "MemberExpression"
                                ? expressionToString(declarator).replace(
                                      `require(${moduleName}).`,
                                      ""
                                  )
                                : undefined;

                        if (imported) {
                            importedMethods.set(variableName, imported);
                        }
                    } else {
                        // special handling of RegExp literals
                        if (declarator.type === "Literal") {
                            const regexp = declarator as RegExpLiteral;
                            if (regexp.regex) {
                                classReceivers[patternToString(decl.id)] = {
                                    start: decl.id.start,
                                    className: "RegExp"
                                };
                            }
                        }

                        const rightSideExpr = expressionToString(declarator);
                        const crPosition = crossReferences[rightSideExpr];
                        ternClient.requestReferences(
                            declarator.start,
                            declNode.sourceFile.name,
                            function(err, data) {
                                if (err) return;
                                const crossRef = data.refs.find(
                                    (ref: any) => ref.start === crPosition
                                );
                                if (crossRef) {
                                    requiredModules.set(
                                        decl.start,
                                        requiredModules.get(crPosition)
                                    );
                                }
                            }
                        );
                    }
                }
            }
        },
        AssignmentExpression: function(assignmentExpr: AssignmentExpression, _: Node[]) {
            const right = assignmentExpr.right;
            const callExpr = getRequireCallExpr(right);
            if (callExpr) {
                const variableName = patternToString(assignmentExpr.left);
                const firstArg = callExpr.arguments[0];
                const moduleName = firstArg.type === "Literal" ? firstArg.value : "";
                requiredModules.set(assignmentExpr.left.start, moduleName);
                crossReferences[variableName] = assignmentExpr.left.start;
                if (debug) {
                    console.log("\nModule Declaration: \n", {
                        Variable: variableName,
                        ModuleName: moduleName
                    });
                }
                const imported =
                    right.type === "MemberExpression"
                        ? expressionToString(right).replace(`require(${moduleName}).`, "")
                        : undefined;

                if (imported) {
                    importedMethods.set(variableName, imported);
                }
            } else {
                // special handling of RegExp literals
                if (right.type === "Literal") {
                    const regexp = right as RegExpLiteral;
                    if (regexp.regex) {
                        classReceivers[patternToString(assignmentExpr.left)] = {
                            start: assignmentExpr.left.start,
                            className: "RegExp"
                        };
                    }
                }

                const rightSideExpr = expressionToString(right);
                const crPosition = crossReferences[rightSideExpr];
                ternClient.requestReferences(right.start, assignmentExpr.sourceFile.name, function(
                    err,
                    data
                ) {
                    if (err) return;
                    const crossRef = data.refs.find((ref: any) => ref.start === crPosition);
                    if (crossRef) {
                        requiredModules.set(
                            assignmentExpr.left.start,
                            requiredModules.get(crPosition)
                        );
                    }
                });
            }
        },
        ImportDeclaration: function(importDecl: ImportDeclaration, _: Node[]) {
            if (debug) {
                console.log({ ImportDeclaration: importDecl });
            }

            const moduleName = importDecl.source.value;

            for (let specifier of importDecl.specifiers) {
                requiredModules.set(specifier.local.name, moduleName);

                if (
                    specifier.type === "ImportSpecifier" &&
                    specifier.imported.name &&
                    specifier.imported.name != specifier.local.name
                ) {
                    importedMethods.set(specifier.local.name, specifier.imported.name);
                }
            }
        },

        /* collecting declared functions */
        FunctionDeclaration: function(declNode: FunctionDeclaration, _: Node[]) {
            if (debug && declNode.id) {
                console.log("\nFunction Declaration: \n", {
                    FileName: declNode.sourceFile.name,
                    Start: declNode.id.start,
                    End: declNode.id.end,
                    Name: declNode.id.name
                });
            }
            const func = extractFunctionInfo(declNode.id, declNode);
            definedFunctions.push(func);
        },

        /* track function calls */
        CallExpression: function(callNode: CallExpression, ancestors: Node[]) {
            if (debug) console.log("\nCallExpression: \n", { callNode, ancestors });
            const outerMethod: FunctionDeclaration | undefined = ancestors
                .filter((node: Node) => node.type === "FunctionDeclaration")
                .pop() as FunctionDeclaration;
            let outerMethodName = ".root";
            if (outerMethod) {
                outerMethodName = outerMethod.id ? outerMethod.id.name : "default";
            }

            let functionName: string;
            let receiver: string = "";
            let className = "";

            const callee = callNode.callee;

            switch (callee.type) {
                case "Identifier":
                    functionName = callee.name;
                    break;
                case "MemberExpression":
                    functionName = expressionToString(callee.property);

                    // special case handling calls on Regular Expressions Literals
                    if (callee.object.type === "Literal") {
                        const regexp = callee.object as RegExpLiteral;
                        if (regexp.regex) {
                            className = "RegExp";
                            receiver = "";
                        }
                        break;
                    }

                    if (
                        callee.object.type !== "Identifier" &&
                        callee.object.type !== "MemberExpression" &&
                        callee.object.type !== "Super"
                    )
                        // ignore other expressions for now
                        return;
                    receiver =
                        callee.object.type === "Super"
                            ? "super"
                            : expressionToString(callee.object);
                    break;
                case "Super":
                    functionName = "super";
                    break;
                default:
                    // ignore other expressions for now
                    // functionName = expressionToString(callee);
                    return;
            }

            const args = [];

            for (let argument of callNode.arguments) {
                const argumentAsString =
                    argument.type === "SpreadElement"
                        ? "..." + expressionToString(argument.argument)
                        : expressionToString(argument);
                args.push(argumentAsString);
            }

            if (receiver != "") {
                const classForReceiver = classReceivers[receiver];
                if (classForReceiver) {
                    ternClient.requestReferences(callNode.start, callNode.sourceFile.name, function(
                        err,
                        data
                    ) {
                        if (err) return;
                        const ref = data.refs.find(
                            (ref: any) => ref.start === classForReceiver.start
                        );
                        if (ref) {
                            className = classForReceiver.className;
                        }
                    });
                }
            }

            callExpressions.push(
                new model.CallExpression(
                    callNode.sourceFile.name,
                    callNode.start,
                    callNode.end,
                    functionName,
                    outerMethodName,
                    receiver,
                    className,
                    args
                )
            );
        },

        NewExpression: function(newExpr: NewExpression, ancestors: Node[]) {
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
                    // calls on other other expressions - ignore for now
                    // functionName = expressionToString(callee);
                    return;
            }

            const args = [];

            for (let argument of newExpr.arguments) {
                const argumentAsString =
                    argument.type === "SpreadElement"
                        ? "..." + expressionToString(argument.argument)
                        : expressionToString(argument);
                args.push(argumentAsString);
            }

            let receiverStart = -1;

            // when functionName is module then add left side of new expression to requiredModules
            const outerDeclarator: VariableDeclarator = ancestors
                .filter(node => node.type === "VariableDeclarator")
                .pop() as VariableDeclarator;
            const outerAssignment: AssignmentExpression = ancestors
                .filter(node => node.type === "AssignmentExpression")
                .pop() as AssignmentExpression;

            if (outerDeclarator) {
                receiver = patternToString(outerDeclarator.id);
                receiverStart = outerDeclarator.id.start;
                if (crossReferences[functionName]) {
                    requiredModules.set(
                        outerDeclarator.start,
                        requiredModules.get(crossReferences[functionName])
                    );
                }
            }

            if (outerAssignment) {
                receiver = patternToString(outerAssignment.left);
                receiverStart = outerAssignment.left.start;
                if (crossReferences[functionName]) {
                    requiredModules.set(
                        outerAssignment.start,
                        requiredModules.get(crossReferences[functionName])
                    );
                }
            }

            if (receiverStart != -1) {
                classReceivers[receiver] = { start: receiverStart, className: functionName };
            }

            callExpressions.push(
                new model.CallExpression(
                    newExpr.sourceFile.name,
                    newExpr.callee.start,
                    newExpr.callee.end,
                    "new " + functionName,
                    outerMethodName,
                    receiver,
                    functionName,
                    args
                )
            );
        }
    };
}
