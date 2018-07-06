import * as util from "./util";
import { default as traverse } from "@babel/traverse";
// -- Need to use estree types because we use estree AST format instead of babel-parser format --
import {
    AssignmentExpression,
    BaseFunction,
    ClassBody,
    ClassDeclaration,
    FunctionDeclaration,
    Identifier,
    MemberExpression,
    Node,
    ObjectExpression,
    VariableDeclaration
} from "estree";
import { NodePath } from "babel-traverse";

export class Export {
    constructor(public type: string, public id: string, public bundleType) {}
}

export class Function {
    constructor(public id: string, public params: Array<string>) {}
}

export class Variable {
    constructor(public id: string, public kind: string, public value?: string) {}
}

export class Class {
    constructor(
        public id: string,
        public methods: Array<string>,
        public superClass: string | null
    ) {}
}

const declaredFunctions: Array<Function> = [];
const declaredVariables: Array<Variable> = [];
const declaredClasses: Array<Class> = [];

function extractFunctionInfo(id: Identifier | null | undefined, baseFunc: BaseFunction): Function {
    const functionName = id ? id.name : "default function";
    const params = baseFunc.params;

    const paramsToString: Array<string> = [];
    for (let param of params) {
        paramsToString.push(util.patternToString(param));
    }
    return new Function(functionName, paramsToString);
}

function extractMethodsFromClassBody(body: ClassBody): Array<string> {
    const methods: Array<string> = [];
    const bodyElements = body.body;
    for (let element of bodyElements) {
        if (element.type === "MethodDefinition") {
            const classMethod = extractFunctionInfo(element.value.id, element.value);
            if (element.key.type === "Identifier") {
                const methodSignature = util.createMethodSignatureString(
                    element.key.name,
                    classMethod.params
                );
                methods.push(methodSignature);
            }
        }
    }
    return methods;
}

function extractExportsFromObject(object: ObjectExpression): Array<Export> {
    const exports: Array<Export> = [];
    const properties = object.properties;
    for (let property of properties) {
        if (property.key.type === "Identifier") {
            if (
                property.value.type === "FunctionExpression" ||
                property.value.type === "ArrowFunctionExpression"
            ) {
                const func = extractFunctionInfo(null, property.value);
                exports.push(
                    new Export(
                        "function",
                        util.createMethodSignatureString(property.key.name, func.params),
                        "commonjs"
                    )
                );
                continue;
            }
            exports.push(new Export("member", property.key.name, "commonjs"));
        }
    }
    return exports;
}

function collectAllMethodsFromClasses(className: string): Array<string> {
    const classDecl = declaredClasses.find(value => value.id === className);

    if (classDecl) {
        const methods: Array<string> = [];
        for (let method of classDecl.methods) {
            methods.push(`${classDecl.id}.${method}`);
        }
        if (classDecl.superClass) {
            methods.push(...collectAllMethodsFromClasses(classDecl.superClass));
            return methods;
        }
        return methods;
    }

    return [];
}

export function traverseAst(ast: any, debug: boolean): Array<Export> {
    const definedExports: Array<Export> = [];

    traverse(ast, {
        FunctionDeclaration(path: NodePath) {
            const node = (path.node as Node) as FunctionDeclaration;
            const func = extractFunctionInfo(node.id, node);
            declaredFunctions.push(func);
        },
        VariableDeclaration(path: NodePath) {
            const node = (path.node as Node) as VariableDeclaration;

            for (let decl of node.declarations) {
                const variable = decl.init
                    ? new Variable(
                          util.patternToString(decl.id),
                          node.kind,
                          util.expressionToString(decl.init)
                      )
                    : new Variable(util.patternToString(decl.id), node.kind);
                declaredVariables.push(variable);
            }
        },
        ClassDeclaration(path: NodePath) {
            const node = (path.node as Node) as ClassDeclaration;

            const superClass =
                node.superClass && node.superClass.type === "Identifier"
                    ? node.superClass.name
                    : null;
            const className = node.id ? node.id.name : "default";
            declaredClasses.push(
                new Class(className, extractMethodsFromClassBody(node.body), superClass)
            );
        },

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
                        definedExports.push(...extractExportsFromObject(right));
                        break;
                    case "ClassExpression":
                        definedExports.push(new Export("class", "default", "commonjs"));
                        const methods = extractMethodsFromClassBody(right.body);
                        for (let method of methods) {
                            if (debug) {
                                console.log(`Found export with name: ${method}`);
                            }
                            definedExports.push(new Export("function", method, "commonjs"));
                        }
                        break;
                    case "NewExpression":
                        const callee = right.callee;
                        if (callee.type === "Identifier") {
                            definedExports.push(new Export("class", callee.name, "commonjs"));
                            const methods = collectAllMethodsFromClasses(callee.name);
                            for (let method of methods) {
                                definedExports.push(new Export("function", method, "commonjs"));
                            }
                        }
                        break;
                    default:
                        definedExports.push(new Export("unknown", "default", "commonjs"));
                        break;
                }
            }

            if (util.isPropertyAssignment(left)) {
                const memberExpr = left as MemberExpression;
                if (memberExpr.property.type === "Identifier") {
                    switch (right.type) {
                        case "Identifier":
                            const variable = declaredVariables.find(
                                value => value.id === right.name
                            );
                            if (variable) {
                                definedExports.push(
                                    new Export(variable.kind, memberExpr.property.name, "commonjs")
                                );
                                break;
                            }

                            const method = declaredFunctions.find(value => value.id === right.name);
                            if (method) {
                                definedExports.push(
                                    new Export(
                                        "function",
                                        util.createMethodSignatureString(
                                            memberExpr.property.name,
                                            method.params
                                        ),
                                        "commonjs"
                                    )
                                );
                                break;
                            }

                            definedExports.push(new Export("unknown", memberExpr.property.name, "commonjs"));

                            if (debug) {
                                console.log(`Found export with name: ${memberExpr.property.name}`);
                            }
                            break;
                        case "ArrowFunctionExpression":
                        case "FunctionExpression":
                            const func = extractFunctionInfo(null, right);
                            definedExports.push(
                                new Export(
                                    "function",
                                    util.createMethodSignatureString(
                                        memberExpr.property.name,
                                        func.params
                                    ),
                                    "commonjs"
                                )
                            );
                            break;

                        case "ObjectExpression":
                            definedExports.push(
                                new Export("object", `${memberExpr.property.name}`, "commonjs")
                            );
                            const exports = extractExportsFromObject(right);
                            for (let exp of exports) {
                                definedExports.push(
                                    new Export(exp.type, `${memberExpr.property.name}.${exp.id}`, "commonjs")
                                );
                            }
                            break;
                        case "ClassExpression":
                            definedExports.push(new Export("class", `${memberExpr.property.name}`, "commonjs"));
                            const methods = extractMethodsFromClassBody(right.body);
                            for (let method of methods) {
                                if (debug) {
                                    console.log(`Found export with name: ${method}`);
                                }
                                definedExports.push(
                                    new Export("function", `${memberExpr.property.name}.${method}`, "commonjs")
                                );
                            }
                            break;
                        case "MemberExpression":
                            definedExports.push(
                                new Export("member", `${memberExpr.property.name}`, "commonjs")
                            );
                            break;
                        default:
                            definedExports.push(new Export("unknown", memberExpr.property.name, "commonjs"));
                            if (debug) {
                                console.log(`Found export with name: ${memberExpr.property.name}`);
                            }
                    }
                }
            }
        }
    });

    if (debug) {
        console.log({ declaredFunctions, declaredVariables, declaredClasses });
    }

    return definedExports;
}
