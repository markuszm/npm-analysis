const parser = require("@babel/parser");
const readdirp = require("readdirp");
const fs = require("fs");
const process = require("process");

const traverse = require("@babel/traverse").default;

let debug = false;

(function() {
    // parse command arguments
    const args = process.argv.slice(2);

    if (args.length === 0) {
        console.error("missing folder path");
    }
    const folderPath = args[0];

    if (args.length > 1 && args[1] === "debug") {
        debug = true;
    }

    const definedExports = [];

    const content = fs.readFileSync(folderPath, "utf-8");
    let ast = parser.parse(content, {
        sourceType: "module",
        allowImportExportEverywhere: true,

        plugins: ["jsx", "typescript", "estree"]
    });

    traverse(ast, {
        // TODO: retrieve full method signature by also visiting function declarations
        MemberExpression(path) {
            const node = path.node;
            const object = node.object;
            const property = node.property;
            if (debug) {
                console.log({ node, object, property });
            }
            if (object.type === "MemberExpression") {
                if (
                    object.object.name === "module" &&
                    object.property.name === "exports"
                ) {
                    // found type of export module.exports.foo = foo
                    definedExports.push(property.name);
                    if (debug) {
                        console.log(`Found export with name: ${property.name}`);
                    }
                }
            }

            if (object.type === "Identifier" && object.name === "exports") {
                // found type of export exports.foo = foo
                definedExports.push(property.name);
                if (debug) {
                    console.log(`Found export with name: ${property.name}`);
                }
            }
        },
        AssignmentExpression(path) {
            const node = path.node;
            const left = node.left;
            const right = node.right;
            if (debug) {
                console.log({ node, left, right });
            }

            if (
                (left.type === "MemberExpression" &&
                    left.object.name === "module" &&
                    left.property.name === "exports") ||
                (left.type === "Identifier" && left.name === "exports")
            ) {
                if (right.type === "ObjectExpression") {
                    const properties = right.properties;
                    for (let property of properties) {
                        if (debug) {
                            console.log(
                                `Found export with name: ${property.key.name}`
                            );
                        }
                        definedExports.push(property.key.name);
                    }
                }
                if (right.type === "ClassExpression") {
                    const bodyElements = right.body.body;
                    for (let element of bodyElements) {
                        if (element.type === "MethodDefinition") {
                            if (debug) {
                                console.log(
                                    `Found export with name: ${
                                        element.key.name
                                    }`
                                );
                            }
                            definedExports.push(element.key.name);
                        }
                    }
                }
            }
        }
    });

    console.log(JSON.stringify(definedExports));

    // // parse and analyze files
    // try {
    //     readdirp(
    //         { root: folderPath, fileFilter: ["*.ts", "*.js", "*.jsx"] },
    //         fileInfo => {
    //             const content = fs.readFileSync(fileInfo.fullPath);
    //             let ast = parser.parse(content, {
    //                 sourceType: "module",
    //                 allowImportExportEverywhere: true,

    //                 plugins: ["jsx", "typescript", "estree"]
    //             });

    //             console.log(ast);
    //         },
    //         () => {
    //             // format processed exports and write to stdout
    //         }
    //     );
    // } catch (e) {
    //     console.error(e);
    //     console.error("could not find any files");
    // }
})();
