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
    Node,
    MemberExpression,
    VariableDeclaration,
    VariableDeclarator,
    ObjectExpression
} from "./@types/estree";
import { NodePath } from "babel-traverse";
import { expressionToString } from "./util";
import { Class, Variable, Function, Export } from "./model";
import { DefinitionQuery, TernClient } from "./ternClient";
import { extractFunctionInfo } from "./util";
import { createMethodSignatureString } from "./util";

export class Traversal {
    BUNDLE_TYPE_ES6 = "es6";
    BUNDLE_TYPE_COMMONJS = "commonjs";
    EXPORT_TYPE_FUNCTION = "function";

    private declaredFunctions: Array<Function> = [];
    private declaredVariables: Array<Variable> = [];
    private declaredClasses: Array<Class> = [];

    constructor(private ternClient: TernClient, private fileName: string, private debug: boolean) {}

    collectAllMethodsFromClasses(className: string): Array<string> {
        const classDecl = this.declaredClasses.find(value => value.id === className);

        if (classDecl) {
            const methods: Array<string> = [];
            for (let method of classDecl.methods) {
                methods.push(`${classDecl.id}.${method}`);
            }
            if (classDecl.superClass) {
                methods.push(...this.collectAllMethodsFromClasses(classDecl.superClass));
                return methods;
            }
            return methods;
        }

        return [];
    }

    extractExportsFromObject(object: ObjectExpression, isDefault: boolean): Array<Export> {
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
                            createMethodSignatureString(property.key.name, func.params),
                            "commonjs",
                            this.fileName,
                            isDefault
                        )
                    );
                    continue;
                }
                exports.push(
                    new Export("member", property.key.name, "commonjs", this.fileName, isDefault)
                );
            }
        }
        return exports;
    }

    findExportsForIdentifer(
        identifier: Identifier,
        name: string,
        bundleType: string,
        isDefault: boolean,
        exports: Array<Export>
    ) {
        const start = identifier.start;
        this.ternClient.requestDefinition(
            new DefinitionQuery(start, this.fileName),
            (err, data) => {
                if (this.debug && data) {
                    console.log(
                        "\nDefinition at: %o \n .. for identifier at: \n %o ",
                        data,
                        identifier
                    );
                }

                const variable = this.declaredVariables.find(
                    value =>
                        data
                            ? value.id === identifier.name && value.start === data.start
                            : value.id === identifier.name
                );
                if (variable) {
                    for (let declaredVariable of this.declaredVariables) {
                        if (
                            data
                                ? declaredVariable.id.startsWith(`${variable.id}.`) &&
                                  declaredVariable.start === data.start
                                : declaredVariable.id.startsWith(`${variable.id}.`)
                        ) {
                            exports.push(
                                new Export(
                                    declaredVariable.kind,
                                    declaredVariable.id,
                                    bundleType,
                                    this.fileName,
                                    isDefault
                                )
                            );
                        }
                    }
                    for (let declaredFunction of this.declaredFunctions) {
                        if (
                            data
                                ? declaredFunction.id.startsWith(`${variable.id}.`) &&
                                  declaredFunction.start === data.start
                                : declaredFunction.id.startsWith(`${variable.id}.`)
                        ) {
                            exports.push(
                                new Export(
                                    this.EXPORT_TYPE_FUNCTION,
                                    declaredFunction.id,
                                    bundleType,
                                    this.fileName,
                                    isDefault
                                )
                            );
                        }
                    }
                    exports.push(
                        new Export(variable.kind, name, bundleType, this.fileName, isDefault)
                    );
                    return;
                }

                const method = this.declaredFunctions.find(
                    value =>
                        data
                            ? value.id === identifier.name && value.start === data.start
                            : value.id === identifier.name
                );
                if (method) {
                    exports.push(
                        new Export(
                            this.EXPORT_TYPE_FUNCTION,
                            util.createMethodSignatureString(name, method.params),
                            bundleType,
                            this.fileName,
                            isDefault
                        )
                    );
                    return;
                }

                exports.push(new Export("unknown", name, bundleType, this.fileName, isDefault));
            }
        );
    }

    extractMembersFromObjectExpression(decl: VariableDeclarator) {
        if (!decl.init || decl.init.type !== "ObjectExpression") {
            return;
        }

        const properties = decl.init.properties;
        for (let property of properties) {
            if (property.key.type === "Identifier") {
                if (
                    property.value.type === "FunctionExpression" ||
                    property.value.type === "ArrowFunctionExpression"
                ) {
                    const func = util.extractFunctionInfo(null, property.value);
                    this.declaredFunctions.push(
                        new Function(
                            `${util.patternToString(decl.id)}.${property.key.name}`,
                            decl.id.start,
                            func.params
                        )
                    );
                } else {
                    this.declaredVariables.push(
                        new Variable(
                            `${util.patternToString(decl.id)}.${property.key.name}`,
                            decl.start,
                            "var",
                            util.expressionToString(decl.init)
                        )
                    );
                }
            }
        }
    }

    public traverseAst(ast: any): Array<Export> {
        const definedExports: Array<Export> = [];
        const self = this;

        traverse(ast, {
            /* --- Collecting Declared Members ---*/
            FunctionDeclaration(path: NodePath) {
                const node = (path.node as Node) as FunctionDeclaration;
                const func = util.extractFunctionInfo(node.id, node);
                self.declaredFunctions.push(func);
            },
            VariableDeclaration(path: NodePath) {
                const node = (path.node as Node) as VariableDeclaration;

                for (let decl of node.declarations) {
                    if (decl.init && decl.init.type === "ObjectExpression") {
                        self.extractMembersFromObjectExpression(decl);
                    }

                    const variable = decl.init
                        ? new Variable(
                              util.patternToString(decl.id),
                              decl.start,
                              node.kind,
                              util.expressionToString(decl.init)
                          )
                        : new Variable(util.patternToString(decl.id), decl.start, node.kind);
                    self.declaredVariables.push(variable);
                }
            },
            ClassDeclaration(path: NodePath) {
                const node = (path.node as Node) as ClassDeclaration;

                const superClass =
                    node.superClass && node.superClass.type === "Identifier"
                        ? node.superClass.name
                        : null;
                const className = node.id ? node.id.name : "default";
                self.declaredClasses.push(
                    new Class(className, util.extractMethodsFromClassBody(node.body), superClass)
                );
            },

            /* --- ES6 Handling --- */
            ExportAllDeclaration(path: NodePath) {
                const node = (path.node as Node) as ExportAllDeclaration;

                if (self.debug) console.log(node);

                const id = node.source.value ? node.source.value.toString() : "unknown";

                definedExports.push(
                    new Export("all", id, self.BUNDLE_TYPE_ES6, self.fileName, false)
                );
            },
            ExportNamedDeclaration(path: NodePath) {
                const node = (path.node as Node) as ExportNamedDeclaration;

                if (self.debug) console.log(node);

                /* Handle declarations */
                const declaration = node.declaration;
                if (declaration) {
                    switch (declaration.type) {
                        case "ClassDeclaration":
                            const name = declaration.id ? declaration.id.name : "default";
                            definedExports.push(
                                new Export(
                                    "class",
                                    `${name}`,
                                    self.BUNDLE_TYPE_ES6,
                                    self.fileName,
                                    name === "default"
                                )
                            );
                            const methods = util.extractMethodsFromClassBody(declaration.body);
                            for (let method of methods) {
                                if (self.debug) {
                                    console.log(`Found export with name: ${method}`);
                                }
                                definedExports.push(
                                    new Export(
                                        self.EXPORT_TYPE_FUNCTION,
                                        `${name}.${method}`,
                                        self.BUNDLE_TYPE_ES6,
                                        self.fileName,
                                        name === "default"
                                    )
                                );
                            }
                            break;
                        case "FunctionDeclaration":
                            const func = util.extractFunctionInfo(declaration.id, declaration);
                            definedExports.push(
                                new Export(
                                    self.EXPORT_TYPE_FUNCTION,
                                    util.createMethodSignatureString(func.id, func.params),
                                    self.BUNDLE_TYPE_ES6,
                                    self.fileName,
                                    false
                                )
                            );
                            break;
                        case "VariableDeclaration":
                            const kind = declaration.kind;
                            for (let varDecl of declaration.declarations) {
                                definedExports.push(
                                    new Export(
                                        kind,
                                        util.patternToString(varDecl.id),
                                        self.BUNDLE_TYPE_ES6,
                                        self.fileName,
                                        false
                                    )
                                );
                            }
                            break;
                    }
                }

                /* Handle specifiers */
                const specifiers = node.specifiers;
                for (let specifier of specifiers) {
                    self.findExportsForIdentifer(
                        specifier.exported,
                        specifier.exported.name,
                        self.BUNDLE_TYPE_ES6,
                        false,
                        definedExports
                    );
                }
            },
            ExportDefaultDeclaration(path: NodePath) {
                const node = (path.node as Node) as ExportDefaultDeclaration;

                if (self.debug) console.log(node);

                const declaration = node.declaration;
                if (declaration) {
                    switch (declaration.type) {
                        case "Identifier":
                            self.findExportsForIdentifer(
                                declaration,
                                declaration.name,
                                self.BUNDLE_TYPE_ES6,
                                true,
                                definedExports
                            );
                            break;
                        case "ClassDeclaration":
                            const name = declaration.id ? declaration.id.name : "default";
                            definedExports.push(
                                new Export(
                                    "class",
                                    `${name}`,
                                    self.BUNDLE_TYPE_ES6,
                                    self.fileName,
                                    true
                                )
                            );
                            const methods = util.extractMethodsFromClassBody(declaration.body);
                            for (let method of methods) {
                                if (self.debug) {
                                    console.log(`Found export with name: ${method}`);
                                }
                                definedExports.push(
                                    new Export(
                                        self.EXPORT_TYPE_FUNCTION,
                                        `${name}.${method}`,
                                        self.BUNDLE_TYPE_ES6,
                                        self.fileName,
                                        true
                                    )
                                );
                            }
                            break;
                        case "FunctionDeclaration":
                            const func = util.extractFunctionInfo(declaration.id, declaration);
                            definedExports.push(
                                new Export(
                                    self.EXPORT_TYPE_FUNCTION,
                                    `${util.createMethodSignatureString(
                                        func.id,
                                        func.params
                                    )}`,
                                    self.BUNDLE_TYPE_ES6,
                                    self.fileName,
                                    true
                                )
                            );
                            break;
                        case "VariableDeclaration":
                            const kind = declaration.kind;
                            for (let varDecl of declaration.declarations) {
                                definedExports.push(
                                    new Export(
                                        kind,
                                        `${util.patternToString(varDecl.id)}`,
                                        self.BUNDLE_TYPE_ES6,
                                        self.fileName,
                                        true
                                    )
                                );
                            }
                            break;
                        default:
                            definedExports.push(
                                new Export(
                                    "expression",
                                    `${util.expressionToString(declaration)}`,
                                    self.BUNDLE_TYPE_ES6,
                                    self.fileName,
                                    true
                                )
                            );
                    }
                }
            },

            /* --- commonJS Handling --- */
            AssignmentExpression(path: NodePath) {
                // double type assertion to get from babel-types Node to estree Node
                const node = (path.node as Node) as AssignmentExpression;

                const left = node.left;
                const right = node.right;
                if (self.debug) {
                    console.log({ node });
                }

                if (util.isDirectAssignment(left)) {
                    switch (right.type) {
                        case "ObjectExpression":
                            definedExports.push(...self.extractExportsFromObject(right, true));
                            break;
                        case "ClassExpression":
                            definedExports.push(
                                new Export(
                                    "class",
                                    right.id ? right.id.name : "default",
                                    self.BUNDLE_TYPE_COMMONJS,
                                    self.fileName,
                                    true
                                )
                            );
                            const methods = util.extractMethodsFromClassBody(right.body);
                            for (let method of methods) {
                                if (self.debug) {
                                    console.log(`Found export with name: ${method}`);
                                }
                                definedExports.push(
                                    new Export(
                                        self.EXPORT_TYPE_FUNCTION,
                                        method,
                                        self.BUNDLE_TYPE_COMMONJS,
                                        self.fileName,
                                        true
                                    )
                                );
                            }
                            break;
                        case "NewExpression":
                            const callee = right.callee;
                            if (callee.type === "Identifier") {
                                definedExports.push(
                                    new Export(
                                        "class",
                                        callee.name,
                                        self.BUNDLE_TYPE_COMMONJS,
                                        self.fileName,
                                        true
                                    )
                                );
                                const methods = self.collectAllMethodsFromClasses(callee.name);
                                for (let method of methods) {
                                    definedExports.push(
                                        new Export(
                                            self.EXPORT_TYPE_FUNCTION,
                                            method,
                                            self.BUNDLE_TYPE_COMMONJS,
                                            self.fileName,
                                            true
                                        )
                                    );
                                }
                            }
                            break;
                        case "Identifier":
                            self.findExportsForIdentifer(
                                right,
                                right.name,
                                self.BUNDLE_TYPE_COMMONJS,
                                true,
                                definedExports
                            );
                            break;
                        default:
                            const expression = expressionToString(right);
                            if (expression === "module.exports") {
                                break;
                            }
                            definedExports.push(
                                new Export(
                                    "unknown",
                                    expression,
                                    self.BUNDLE_TYPE_COMMONJS,
                                    self.fileName,
                                    true
                                )
                            );
                            break;
                    }
                }

                if (util.isPropertyAssignment(left)) {
                    const memberExpr = left as MemberExpression;
                    if (memberExpr.property.type === "Identifier") {
                        switch (right.type) {
                            case "Identifier":
                                self.findExportsForIdentifer(
                                    right,
                                    memberExpr.property.name,
                                    self.BUNDLE_TYPE_COMMONJS,
                                    false,
                                    definedExports
                                );
                                break;
                            case "ArrowFunctionExpression":
                            case "FunctionExpression":
                                const func = util.extractFunctionInfo(null, right);
                                definedExports.push(
                                    new Export(
                                        self.EXPORT_TYPE_FUNCTION,
                                        util.createMethodSignatureString(
                                            memberExpr.property.name,
                                            func.params
                                        ),
                                        self.BUNDLE_TYPE_COMMONJS,
                                        self.fileName,
                                        false
                                    )
                                );
                                break;

                            case "ObjectExpression":
                                definedExports.push(
                                    new Export(
                                        "object",
                                        `${memberExpr.property.name}`,
                                        self.BUNDLE_TYPE_COMMONJS,
                                        self.fileName,
                                        false
                                    )
                                );
                                const exports = self.extractExportsFromObject(right,false);
                                for (let exp of exports) {
                                    definedExports.push(
                                        new Export(
                                            exp.type,
                                            `${memberExpr.property.name}.${exp.id}`,
                                            self.BUNDLE_TYPE_COMMONJS,
                                            self.fileName,
                                            false
                                        )
                                    );
                                }
                                break;
                            case "ClassExpression":
                                definedExports.push(
                                    new Export(
                                        "class",
                                        `${memberExpr.property.name}`,
                                        self.BUNDLE_TYPE_COMMONJS,
                                        self.fileName,
                                        false
                                    )
                                );
                                const methods = util.extractMethodsFromClassBody(right.body);
                                for (let method of methods) {
                                    if (self.debug) {
                                        console.log(`Found export with name: ${method}`);
                                    }
                                    definedExports.push(
                                        new Export(
                                            self.EXPORT_TYPE_FUNCTION,
                                            `${memberExpr.property.name}.${method}`,
                                            self.BUNDLE_TYPE_COMMONJS,
                                            self.fileName,
                                            false
                                        )
                                    );
                                }
                                break;
                            case "MemberExpression":
                                definedExports.push(
                                    new Export(
                                        "member",
                                        `${memberExpr.property.name}`,
                                        self.BUNDLE_TYPE_COMMONJS,
                                        self.fileName,
                                        false
                                    )
                                );
                                break;
                            default:
                                definedExports.push(
                                    new Export(
                                        "unknown",
                                        memberExpr.property.name,
                                        self.BUNDLE_TYPE_COMMONJS,
                                        self.fileName,
                                        false
                                    )
                                );
                                if (self.debug) {
                                    console.log(
                                        `Found export with name: ${memberExpr.property.name}`
                                    );
                                }
                        }
                    }
                }
            }
        });

        if (self.debug) {
            console.log({
                declaredFunctions: self.declaredFunctions,
                declaredVariables: self.declaredVariables,
                declaredClasses: self.declaredClasses
            });
        }

        return definedExports;
    }
}
