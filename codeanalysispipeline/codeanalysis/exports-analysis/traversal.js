const util = require("./util");
const traverse = require("@babel/traverse").default;

class Export {
    constructor(type, identifier) {
        this.type = type;
        this.identifier = identifier;
    }
}

exports.traverseAst = (ast, debug) => {
    const definedExports = [];

    traverse(ast, {
        // TODO: retrieve full method signature by also visiting function declarations
        AssignmentExpression(path) {
            const node = path.node;
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
                        break;
                    case "ClassExpression":
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
                                definedExports.push(
                                    new Export("method", element.key.name)
                                );
                            }
                        }
                    default:
                        break;
                }
            }

            if (util.isPropertyAssignment(left)) {
                definedExports.push(new Export("method", left.property.name));
                if (debug) {
                    console.log(
                        `Found export with name: ${left.property.name}`
                    );
                }
            }
        }
    });

    return definedExports;
};
