import { default as traverse } from "@babel/traverse";
// -- Need to use estree types because we use estree AST format instead of babel-parser format --
import { NodePath } from "babel-traverse";
import { Import } from "./model";
import {
    AssignmentExpression,
    CallExpression,
    Expression,
    ImportDeclaration,
    Node,
    Super,
    VariableDeclaration
} from "estree";
import * as util from "./util";
import { patternToString } from "./util";

export class Traversal {
    BUNDLE_TYPE_ES6 = "es6";
    BUNDLE_TYPE_COMMONJS = "commonjs";
    IMPORT_SIDE_EFFECT = "@side-effect";

    constructor(private debug: boolean) {}

    public traverseAst(ast: any, fileName: string): Array<Import> {
        const definedImports: Array<Import> = [];
        const self = this;

        traverse(ast, {
            /* detect imports */
            VariableDeclaration(path: NodePath) {
                const node = (path.node as Node) as VariableDeclaration;
                if (self.debug) {
                    console.log({ node });
                }
                for (let decl of node.declarations) {
                    if (self.debug) {
                        console.log(decl);
                    }
                    const declarator = decl.init;
                    if (declarator) {
                        const callExpr = Traversal.getRequireCallExpr(declarator);
                        if (!callExpr) continue;

                        const moduleArg = callExpr.arguments[0];
                        if (decl.id.type !== "Identifier" || moduleArg.type !== "Literal") {
                            continue;
                        }
                        const variableName = decl.id.name;
                        const moduleName = moduleArg.value || "unknown";
                        const imported =
                            declarator.type === "MemberExpression"
                                ? util
                                      .expressionToString(declarator)
                                      .replace(`require(${moduleName}).`, "")
                                : undefined;
                        definedImports.push(
                            new Import(
                                variableName,
                                fileName,
                                moduleName.toString(),
                                self.BUNDLE_TYPE_COMMONJS,
                                imported
                            )
                        );
                        if (self.debug) {
                            console.log("\nModule Declaration: \n", {
                                Variable: variableName,
                                ModuleName: moduleName
                            });
                        }
                    }
                }
            },

            AssignmentExpression: function(path: NodePath) {
                const assignmentExpr = (path.node as Node) as AssignmentExpression;

                if (self.debug) {
                    console.log(assignmentExpr);
                }

                const callExpr = Traversal.getRequireCallExpr(assignmentExpr.right);
                if (!callExpr) return;

                const moduleArg = callExpr.arguments[0];
                if (moduleArg.type !== "Literal") {
                    return;
                }
                const variableName = patternToString(assignmentExpr.left);
                const moduleName = moduleArg.value || "unknown";
                const imported =
                    assignmentExpr.right.type === "MemberExpression"
                        ? util
                              .expressionToString(assignmentExpr.right)
                              .replace(`require(${moduleName}).`, "")
                        : undefined;
                definedImports.push(
                    new Import(
                        variableName,
                        fileName,
                        moduleName.toString(),
                        self.BUNDLE_TYPE_COMMONJS,
                        imported
                    )
                );
                if (self.debug) {
                    console.log("\nModule Declaration: \n", {
                        Variable: variableName,
                        ModuleName: moduleName
                    });
                }
            },

            CallExpression(path: NodePath) {
                const node = (path.node as Node) as CallExpression;

                if (Traversal.isRequireCall(node)) {
                    let argument = node.arguments[0];
                    if (argument.type !== "Literal" || !argument.value) {
                        return;
                    }
                    const moduleName = argument.value.toString();
                    const alreadyExistingImport = definedImports.find(i => i.moduleName === moduleName);
                    if(alreadyExistingImport === undefined) {
                        definedImports.push(new Import(self.IMPORT_SIDE_EFFECT, fileName, moduleName, self.BUNDLE_TYPE_COMMONJS, ""))
                    }
                }
            },

            ImportDeclaration(path: NodePath) {
                const node = (path.node as Node) as ImportDeclaration;

                if (self.debug) {
                    console.log({ ImportDeclaration: node });
                }

                const moduleName = node.source.value || "unknown";

                if (node.specifiers.length == 0) {
                    definedImports.push(
                        new Import(
                            self.IMPORT_SIDE_EFFECT,
                            fileName,
                            moduleName.toString(),
                            self.BUNDLE_TYPE_ES6
                        )
                    );
                }

                for (let specifier of node.specifiers) {
                    if (specifier.type === "ImportSpecifier") {
                        definedImports.push(
                            new Import(
                                specifier.local.name,
                                fileName,
                                moduleName.toString(),
                                self.BUNDLE_TYPE_ES6,
                                specifier.imported.name
                            )
                        );
                    } else {
                        definedImports.push(
                            new Import(
                                specifier.local.name,
                                fileName,
                                moduleName.toString(),
                                self.BUNDLE_TYPE_ES6
                            )
                        );
                    }
                }
            }
        });

        return definedImports;
    }

    private static getRequireCallExpr(decl: Expression | Super): CallExpression | null {
        switch (decl.type) {
            case "CallExpression":
                if (Traversal.isRequireCall(decl)) {
                    return decl;
                } else {
                    if (decl.callee.type === "MemberExpression") {
                        return Traversal.getRequireCallExpr(decl.callee);
                    } else {
                        return null;
                    }
                }
            case "MemberExpression":
                return Traversal.getRequireCallExpr(decl.object);
            default:
                return null;
        }
    }

    private static isRequireCall(decl: CallExpression) {
        return (
            decl &&
            decl.type === "CallExpression" &&
            decl.callee.type === "Identifier" &&
            decl.callee.name === "require"
        );
    }
}
