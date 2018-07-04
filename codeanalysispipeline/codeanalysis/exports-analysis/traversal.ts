import * as util from "./util";
import { default as traverse } from "@babel/traverse";
import { Node, MemberExpression } from "estree";
import { AssignmentExpression } from "estree";
import { NodePath } from "babel-traverse";

// Need to use estree types because we use estree AST format instead of babel-parser format

export class Export {
    constructor(public type: string, public identifier: string) {}
}

export function traverseAst(ast: any, debug: boolean): Array<Export> {
    const definedExports: Array<Export> = [];

    traverse(ast, {
        // TODO: retrieve full method signature by also visiting function declarations
        AssignmentExpression(path: NodePath) {
            // double type assertion to get from babel-types Node to estree Node
            const node = (path.node as Node) as AssignmentExpression;

            const left = node.left;
            const right = node.right;
            if (debug) {
                console.log({ node });
            }

            if (util.isDirectAssignment(left)) {
                switch (right.type) {
                    case "ObjectExpression":
                        const properties = right.properties;
                        for (let property of properties) {
                            if (property.key.type === "Identifier") {
                                if (debug) {
                                    console.log(
                                        `Found export with name: ${
                                            property.key.name
                                        }`
                                    );
                                }
                                definedExports.push(
                                    new Export("method", property.key.name)
                                );
                            }
                        }
                        break;
                    case "ClassExpression":
                        const bodyElements = right.body.body;
                        for (let element of bodyElements) {
                            if (element.type === "MethodDefinition") {
                                if (element.key.type === "Identifier") {
                                    if (debug) {
                                        console.log(
                                            `Found export with name: ${
                                                element.key.name
                                            }`
                                        );
                                    }
                                    definedExports.push(
                                        new Export("method", element.key.name)
                                    );
                                }
                            }
                        }
                    default:
                        break;
                }
            }

            if (util.isPropertyAssignment(left)) {
                const memberExpr = left as MemberExpression;
                if (memberExpr.property.type === "Identifier") {
                    definedExports.push(
                        new Export("method", memberExpr.property.name)
                    );
                    if (debug) {
                        console.log(
                            `Found export with name: ${
                                memberExpr.property.name
                            }`
                        );
                    }
                }
            }
        }
    });

    return definedExports;
}
