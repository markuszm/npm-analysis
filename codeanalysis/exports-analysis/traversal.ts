import * as util from "./util";
import { default as traverse } from "@babel/traverse";
// -- Need to use estree types because we use estree AST format instead of babel-parser format --
import {
    AssignmentExpression,
    ClassDeclaration,
    ExportAllDeclaration,
    ExportDefaultDeclaration,
    ExportNamedDeclaration,
    FunctionDeclaration,
    Identifier,
    MemberExpression,
    Node,
    ObjectExpression,
    VariableDeclaration
} from "estree";
import { NodePath } from "babel-traverse";

const BUNDLE_TYPE_ES6 = "es6";
const BUNDLE_TYPE_COMMONJS = "commonjs";

const EXPORT_TYPE_FUNCTION = "function";

export class Export {
    constructor(public type: string, public id: string, public bundleType: string) {}
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

function findExportsForIdentifer(
    identifier: Identifier,
    name: string,
    bundleType: string
): Array<Export> {
    const variable = declaredVariables.find(value => value.id === identifier.name);
    if (variable) {
        const exports: Array<Export> = [];
        for (let declaredVariable of declaredVariables) {
            if (declaredVariable.id.startsWith(`${variable.id}.`)) {
                exports.push(new Export(declaredVariable.kind, declaredVariable.id, bundleType));
            }
        }
        for (let declaredFunction of declaredFunctions) {
            if (declaredFunction.id.startsWith(`${variable.id}.`)) {
                exports.push(new Export(EXPORT_TYPE_FUNCTION, declaredFunction.id, bundleType));
            }
        }
        exports.push(new Export(variable.kind, name, bundleType));
        return exports;
    }

    const method = declaredFunctions.find(value => value.id === identifier.name);
    if (method) {
        return [
            new Export(
                EXPORT_TYPE_FUNCTION,
                util.createMethodSignatureString(name, method.params),
                bundleType
            )
        ];
    }

    return [new Export("unknown", name, bundleType)];
}

function extractMembersFromObjectExpression(decl) {
    const properties = decl.init.properties;
    for (let property of properties) {
        if (property.key.type === "Identifier") {
            if (
                property.value.type === "FunctionExpression" ||
                property.value.type === "ArrowFunctionExpression"
            ) {
                const func = util.extractFunctionInfo(null, property.value);
                declaredFunctions.push(
                    new Function(
                        `${util.patternToString(decl.id)}.${property.key.name}`,
                        func.params
                    )
                );
            } else {
                declaredVariables.push(
                    new Variable(
                        `${util.patternToString(decl.id)}.${property.key.name}`,
                        "var",
                        util.expressionToString(decl.init)
                    )
                );
            }
        }
    }
}

export function traverseAst(ast: any, debug: boolean): Array<Export> {
    const definedExports: Array<Export> = [];

    traverse(ast, {
        /* --- Collecting Declared Members ---*/
        FunctionDeclaration(path: NodePath) {
            const node = (path.node as Node) as FunctionDeclaration;
            const func = util.extractFunctionInfo(node.id, node);
            declaredFunctions.push(func);
        },
        VariableDeclaration(path: NodePath) {
            const node = (path.node as Node) as VariableDeclaration;

            for (let decl of node.declarations) {
                if (decl.init && decl.init.type === "ObjectExpression") {
                    extractMembersFromObjectExpression(decl);
                }

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
                new Class(className, util.extractMethodsFromClassBody(node.body), superClass)
            );
        },

        /* --- ES6 Handling --- */
        ExportAllDeclaration(path: NodePath) {
            const node = (path.node as Node) as ExportAllDeclaration;

            if (debug) console.log(node);

            const id = node.source.value ? node.source.value.toString() : "unknown";

            definedExports.push(new Export("all", id, BUNDLE_TYPE_ES6));
        },
        ExportNamedDeclaration(path: NodePath) {
            const node = (path.node as Node) as ExportNamedDeclaration;

            if (debug) console.log(node);

            /* Handle declarations */
            const declaration = node.declaration;
            if (declaration) {
                switch (declaration.type) {
                    case "ClassDeclaration":
                        const name = declaration.id ? declaration.id.name : "default";
                        definedExports.push(new Export("class", `${name}`, BUNDLE_TYPE_ES6));
                        const methods = util.extractMethodsFromClassBody(declaration.body);
                        for (let method of methods) {
                            if (debug) {
                                console.log(`Found export with name: ${method}`);
                            }
                            definedExports.push(
                                new Export(
                                    EXPORT_TYPE_FUNCTION,
                                    `${name}.${method}`,
                                    BUNDLE_TYPE_ES6
                                )
                            );
                        }
                        break;
                    case "FunctionDeclaration":
                        const func = util.extractFunctionInfo(declaration.id, declaration);
                        definedExports.push(
                            new Export(
                                EXPORT_TYPE_FUNCTION,
                                util.createMethodSignatureString(func.id, func.params),
                                BUNDLE_TYPE_ES6
                            )
                        );
                        break;
                    case "VariableDeclaration":
                        const kind = declaration.kind;
                        for (let varDecl of declaration.declarations) {
                            definedExports.push(
                                new Export(kind, util.patternToString(varDecl.id), BUNDLE_TYPE_ES6)
                            );
                        }
                        break;
                }
            }

            /* Handle specifiers */
            const specifiers = node.specifiers;
            for (let specifier of specifiers) {
                const specifierExports = findExportsForIdentifer(
                    specifier.exported,
                    specifier.exported.name,
                    BUNDLE_TYPE_ES6
                );
                definedExports.push(...specifierExports);
                if (debug) {
                    specifierExports.forEach(exp =>
                        console.log(`Found export with name: ${exp.id}`)
                    );
                }
            }
        },
        ExportDefaultDeclaration(path: NodePath) {
            const node = (path.node as Node) as ExportDefaultDeclaration;

            if (debug) console.log(node);

            const declaration = node.declaration;
            if (declaration) {
                switch (declaration.type) {
                    case "ClassDeclaration":
                        const name = declaration.id ? declaration.id.name : "default";
                        definedExports.push(new Export("class", `${name}`, BUNDLE_TYPE_ES6));
                        const methods = util.extractMethodsFromClassBody(declaration.body);
                        for (let method of methods) {
                            if (debug) {
                                console.log(`Found export with name: ${method}`);
                            }
                            definedExports.push(
                                new Export(
                                    EXPORT_TYPE_FUNCTION,
                                    `default.${name}.${method}`,
                                    BUNDLE_TYPE_ES6
                                )
                            );
                        }
                        break;
                    case "FunctionDeclaration":
                        const func = util.extractFunctionInfo(declaration.id, declaration);
                        definedExports.push(
                            new Export(
                                EXPORT_TYPE_FUNCTION,
                                `default.${util.createMethodSignatureString(func.id, func.params)}`,
                                BUNDLE_TYPE_ES6
                            )
                        );
                        break;
                    case "VariableDeclaration":
                        const kind = declaration.kind;
                        for (let varDecl of declaration.declarations) {
                            definedExports.push(
                                new Export(
                                    kind,
                                    `default.${util.patternToString(varDecl.id)}`,
                                    BUNDLE_TYPE_ES6
                                )
                            );
                        }
                        break;
                }
            }
        },

        /* --- commonJS Handling --- */
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
                        definedExports.push(...util.extractExportsFromObject(right));
                        break;
                    case "ClassExpression":
                        definedExports.push(new Export("class", "default", BUNDLE_TYPE_COMMONJS));
                        const methods = util.extractMethodsFromClassBody(right.body);
                        for (let method of methods) {
                            if (debug) {
                                console.log(`Found export with name: ${method}`);
                            }
                            definedExports.push(
                                new Export(EXPORT_TYPE_FUNCTION, method, BUNDLE_TYPE_COMMONJS)
                            );
                        }
                        break;
                    case "NewExpression":
                        const callee = right.callee;
                        if (callee.type === "Identifier") {
                            definedExports.push(
                                new Export("class", callee.name, BUNDLE_TYPE_COMMONJS)
                            );
                            const methods = collectAllMethodsFromClasses(callee.name);
                            for (let method of methods) {
                                definedExports.push(
                                    new Export(EXPORT_TYPE_FUNCTION, method, BUNDLE_TYPE_COMMONJS)
                                );
                            }
                        }
                        break;
                    default:
                        definedExports.push(new Export("unknown", "default", BUNDLE_TYPE_COMMONJS));
                        break;
                }
            }

            if (util.isPropertyAssignment(left)) {
                const memberExpr = left as MemberExpression;
                if (memberExpr.property.type === "Identifier") {
                    switch (right.type) {
                        case "Identifier":
                            const identifierExports = findExportsForIdentifer(
                                right,
                                memberExpr.property.name,
                                BUNDLE_TYPE_COMMONJS
                            );
                            definedExports.push(...identifierExports);
                            if (debug) {
                                identifierExports.forEach(exp =>
                                    console.log(`Found export with name: ${exp.id}`)
                                );
                            }
                            break;
                        case "ArrowFunctionExpression":
                        case "FunctionExpression":
                            const func = util.extractFunctionInfo(null, right);
                            definedExports.push(
                                new Export(
                                    EXPORT_TYPE_FUNCTION,
                                    util.createMethodSignatureString(
                                        memberExpr.property.name,
                                        func.params
                                    ),
                                    BUNDLE_TYPE_COMMONJS
                                )
                            );
                            break;

                        case "ObjectExpression":
                            definedExports.push(
                                new Export(
                                    "object",
                                    `${memberExpr.property.name}`,
                                    BUNDLE_TYPE_COMMONJS
                                )
                            );
                            const exports = util.extractExportsFromObject(right);
                            for (let exp of exports) {
                                definedExports.push(
                                    new Export(
                                        exp.type,
                                        `${memberExpr.property.name}.${exp.id}`,
                                        BUNDLE_TYPE_COMMONJS
                                    )
                                );
                            }
                            break;
                        case "ClassExpression":
                            definedExports.push(
                                new Export(
                                    "class",
                                    `${memberExpr.property.name}`,
                                    BUNDLE_TYPE_COMMONJS
                                )
                            );
                            const methods = util.extractMethodsFromClassBody(right.body);
                            for (let method of methods) {
                                if (debug) {
                                    console.log(`Found export with name: ${method}`);
                                }
                                definedExports.push(
                                    new Export(
                                        EXPORT_TYPE_FUNCTION,
                                        `${memberExpr.property.name}.${method}`,
                                        BUNDLE_TYPE_COMMONJS
                                    )
                                );
                            }
                            break;
                        case "MemberExpression":
                            definedExports.push(
                                new Export(
                                    "member",
                                    `${memberExpr.property.name}`,
                                    BUNDLE_TYPE_COMMONJS
                                )
                            );
                            break;
                        default:
                            definedExports.push(
                                new Export(
                                    "unknown",
                                    memberExpr.property.name,
                                    BUNDLE_TYPE_COMMONJS
                                )
                            );
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
