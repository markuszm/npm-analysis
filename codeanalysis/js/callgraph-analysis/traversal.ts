// Author: Michael Pradel, Markus Zimmermann
import * as model from "./model";
import { Function } from "./model";

import { expressionToString, extractFunctionInfo, patternToString } from "./util";
import {
    AssignmentExpression, BaseCallExpression,
    CallExpression, ClassBody,
    Expression,
    FunctionDeclaration,
    ImportDeclaration, MethodDefinition,
    NewExpression,
    Node,
    RegExpLiteral,
    Super,
    VariableDeclaration,
    VariableDeclarator
} from "./@types/estree";
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
            decl.callee &&
            decl.callee.type === "Identifier" &&
            decl.callee.name === "require"
        );
    }

    function replaceCommonjsExportPrefix(functionName: string): string {
        if (functionName.startsWith("module.exports.")) {
            functionName = functionName.replace("module.exports.", "");
        }
        if (functionName.startsWith("exports.")) {
            functionName = functionName.replace("exports.", "");
        }
        return functionName;
    }

    function findOuterFunction(ancestors: Node[]) {
        const outerDeclaration: FunctionDeclaration | undefined = ancestors
            .filter((node: Node) => node.type === "FunctionDeclaration")
            .pop() as FunctionDeclaration;
        const outerClassMethod: MethodDefinition | undefined = ancestors
            .filter((node: Node) => node.type === "MethodDefinition")
            .pop() as MethodDefinition;
        const outerExpression: Node | undefined = ancestors
            .filter(
                (node: Node) =>
                    node.type === "FunctionExpression" ||
                    node.type === "ArrowFunctionExpression" ||
                    node.type === "CallExpression"
            )
            .shift();
        let outerMethodName = ".root";
        if (outerDeclaration) {
            outerMethodName = outerDeclaration.id ? outerDeclaration.id.name : "default";
            return outerMethodName;
        }
        if (outerClassMethod && outerClassMethod.key.type === "Identifier") {
            outerMethodName = outerClassMethod.key.name;
        }
        if (outerExpression) {
            for (let ancestor of ancestors) {
                if (ancestor.type === "VariableDeclarator") {
                    const leftSideName = patternToString(ancestor.id);
                    if (
                        outerExpression.type === "CallExpression" &&
                        !leftSideName.includes("exports.")
                    ) {
                        break;
                    }

                    if (ancestor.init === outerExpression) {
                        outerMethodName = leftSideName;
                        outerMethodName = replaceCommonjsExportPrefix(outerMethodName);
                        break;
                    }
                }
                if (ancestor.type === "AssignmentExpression") {
                    const leftSideName = patternToString(ancestor.left);
                    if (
                        outerExpression.type === "CallExpression" &&
                        !leftSideName.includes("exports.")
                    ) {
                        break;
                    }
                    if(ancestor.right === outerExpression) {
                        outerMethodName = leftSideName;
                        outerMethodName = replaceCommonjsExportPrefix(outerMethodName);
                        break;
                    }
                }
            }
        }

        if (outerMethodName === "module.exports" || outerMethodName === "exports") {
            outerMethodName = "default";
        }
        return outerMethodName;
    }

    var crossReferences: any = {};

    var classReceivers: any = {};

    function getEndForCallExpression(callNode: BaseCallExpression) {
        if (callNode.callee.type === "MemberExpression") {
            const calleeObject = callNode.callee.object;
            if (
                calleeObject.type === "Identifier" &&
                (calleeObject.name === "this" || calleeObject.name === "self")
            ) {
                return callNode.callee.end;
            } else {
                return callNode.callee.object.end;
            }
        }
        return callNode.callee.end;
    }

    return {
        /* detect imports */
        VariableDeclaration: function(declNode: VariableDeclaration, _: Node[]) {
            for (let decl of declNode.declarations) {
                const declarator = decl.init;
                if (declarator) {
                    // handle function expressions
                    if (
                        declarator.type === "FunctionExpression" ||
                        declarator.type === "ArrowFunctionExpression"
                    ) {
                        // replace cjs export prefix
                        let functionName = patternToString(decl.id);
                        functionName = replaceCommonjsExportPrefix(functionName);

                        const func = new Function(
                            functionName,
                            decl.id.start,
                            declarator.params.map(param => patternToString(param))
                        );
                        definedFunctions.push(func);
                    }

                    // handle require calls
                    const callExpr = getRequireCallExpr(declarator);
                    if (callExpr) {
                        const variableName = patternToString(decl.id);
                        const firstArg = callExpr.arguments[0];
                        const moduleName = firstArg.type === "Literal" ? firstArg.value : "";
                        requiredModules.set(decl.id.end, moduleName);
                        crossReferences[variableName] = decl.id.end;
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
                    }

                    // special handling of RegExp literals
                    if (declarator.type === "Literal") {
                        const regexp = declarator as RegExpLiteral;
                        if (regexp.regex) {
                            classReceivers[patternToString(decl.id)] = {
                                start: decl.id.end,
                                className: "RegExp"
                            };
                        }
                    }

                    // check for references of the expression in same scope to add module reference for left side
                    let rightSideExpr = expressionToString(declarator);
                    if (declarator.type === "CallExpression") {
                        rightSideExpr = expressionToString(declarator.callee);
                    }
                    const crPosition = crossReferences[rightSideExpr];
                    ternClient.requestReferences(
                        declarator.start,
                        declNode.sourceFile.name,
                        function(err, data) {
                            if (debug) {
                                console.log("Found following references: ", data);
                            }
                            if (err) return;
                            const crossRef = data.refs.find((ref: any) => ref.end === crPosition);
                            if (crossRef) {
                                requiredModules.set(decl.id.end, requiredModules.get(crPosition));
                            }
                        }
                    );
                }
            }
        },
        AssignmentExpression: function(assignmentExpr: AssignmentExpression, _: Node[]) {
            const left = assignmentExpr.left;
            const right = assignmentExpr.right;

            // handle function expressions
            if (right.type === "FunctionExpression" || right.type === "ArrowFunctionExpression") {
                const func = new Function(
                    patternToString(left),
                    left.start,
                    right.params.map(param => patternToString(param))
                );
                definedFunctions.push(func);
            }

            // handle require calls
            const callExpr = getRequireCallExpr(right);
            if (callExpr) {
                const variableName = patternToString(left);
                const firstArg = callExpr.arguments[0];
                const moduleName = firstArg.type === "Literal" ? firstArg.value : "";
                requiredModules.set(left.end, moduleName);
                crossReferences[variableName] = left.end;
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
            }

            // special handling of RegExp literals
            if (right.type === "Literal") {
                const regexp = right as RegExpLiteral;
                if (regexp.regex) {
                    classReceivers[patternToString(left)] = {
                        start: left.end,
                        className: "RegExp"
                    };
                }
            }

            // check for references of the expression in same scope to add module reference for left side
            let rightSideExpr = expressionToString(right);
            if (right.type === "CallExpression") {
                rightSideExpr = expressionToString(right.callee);
            }
            const crPosition = crossReferences[rightSideExpr];
            ternClient.requestReferences(right.start, assignmentExpr.sourceFile.name, function(
                err,
                data
            ) {
                if (err) return;
                const crossRef = data.refs.find((ref: any) => ref.end === crPosition);
                if (crossRef) {
                    requiredModules.set(assignmentExpr.left.end, requiredModules.get(crPosition));
                }
            });
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
            let outerMethodName = findOuterFunction(ancestors);

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
                    receiver = expressionToString(callee.object);
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
                    ternClient.requestReferences(getEndForCallExpression(callNode), callNode.sourceFile.name, function(
                        err,
                        data
                    ) {
                        if (err) return;
                        const ref = data.refs.find(
                            (ref: any) => ref.end === classForReceiver.start
                        );
                        if (ref) {
                            className = classForReceiver.className;
                        }
                    });
                }
            }


            const start = callNode.start;
            const end = getEndForCallExpression(callNode);
            let loc = {start: callNode.sourceFile.asLineChar(start), end: callNode.sourceFile.asLineChar(end)};

            callExpressions.push(
                new model.CallExpression(
                    callNode.sourceFile.name,
                    start,
                    end,
                    loc,
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
                receiverStart = outerDeclarator.id.end;
                if (crossReferences[functionName]) {
                    requiredModules.set(
                        receiverStart,
                        requiredModules.get(crossReferences[functionName])
                    );
                }
            }

            if (outerAssignment) {
                receiver = patternToString(outerAssignment.left);
                receiverStart = outerAssignment.left.end;
                if (crossReferences[functionName]) {
                    requiredModules.set(
                        receiverStart,
                        requiredModules.get(crossReferences[functionName])
                    );
                }
            }

            if (receiverStart != -1) {
                classReceivers[receiver] = { start: receiverStart, className: functionName };
            }

            const start = newExpr.callee.start;
            const end = getEndForCallExpression(newExpr);
            let loc = {start: newExpr.sourceFile.asLineChar(start), end: newExpr.sourceFile.asLineChar(end)};
            callExpressions.push(
                new model.CallExpression(
                    newExpr.sourceFile.name,
                    start,
                    end,
                    loc,
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


