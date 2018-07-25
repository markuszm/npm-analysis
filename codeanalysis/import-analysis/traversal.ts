import { default as traverse } from "@babel/traverse";
// -- Need to use estree types because we use estree AST format instead of babel-parser format --
import { NodePath } from "babel-traverse";
import { Import } from "./model";
import { ImportDeclaration, VariableDeclaration, Node } from "estree";

export class Traversal {
    BUNDLE_TYPE_ES6 = "es6";
    BUNDLE_TYPE_COMMONJS = "commonjs";
    IMPORT_SIDE_EFFECT = "@side-effect";

    constructor(private debug: boolean) {}

    public traverseAst(ast: any): Array<Import> {
        const definedImports: Array<Import> = [];
        const self = this;

        traverse(ast, {
            /* detect imports */
            VariableDeclaration(path: NodePath) {
                const node = (path.node as Node) as VariableDeclaration;
                for (let decl of node.declarations) {
                    if (decl.init &&
                        decl.init.type === "CallExpression" &&
                        decl.init.callee.type === "Identifier" &&
                        decl.init.callee.name === "require") {
                        const moduleArg = decl.init.arguments[0];
                        if(decl.id.type !== "Identifier" || moduleArg.type !== "Literal") {
                            continue
                        }
                        const variableName = decl.id.name;
                        const moduleName = moduleArg.value || "unknown";
                        definedImports.push(new Import(variableName, moduleName.toString(), self.BUNDLE_TYPE_COMMONJS)) ;
                        if (self.debug) {
                            console.log("\nModule Declaration: \n", {
                                Variable: variableName,
                                ModuleName: moduleName
                            });
                        }
                    }
                }
            },
            ImportDeclaration(path: NodePath) {
                const node = (path.node as Node) as ImportDeclaration;

                if (self.debug) {
                    console.log({ImportDeclaration: node});
                }

                const moduleName = node.source.value || "unknown";

                if(node.specifiers.length == 0) {
                    definedImports.push(new Import(self.IMPORT_SIDE_EFFECT, moduleName.toString(), self.BUNDLE_TYPE_ES6))
                }

                for (let specifier of node.specifiers) {
                    if (specifier.type === "ImportSpecifier") {
                        definedImports.push(new Import(specifier.local.name, moduleName.toString(), self.BUNDLE_TYPE_ES6, specifier.imported.name)) ;
                    } else {
                        definedImports.push(new Import(specifier.local.name, moduleName.toString(), self.BUNDLE_TYPE_ES6)) ;
                    }
                }
            },
        });

        return definedImports;
    }
}
