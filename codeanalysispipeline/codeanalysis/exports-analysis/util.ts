import { BaseFunction, ClassBody, Expression, Identifier, MemberExpression, ObjectExpression, Pattern } from "estree";
import { Export, Function } from "./traversal";

export function isDirectAssignment(left: Pattern): boolean {
    switch (left.type) {
        case "MemberExpression":
            const leftExpr = left as MemberExpression;
            return (
                leftExpr.object.type === "Identifier" &&
                leftExpr.property.type === "Identifier" &&
                leftExpr.object.name === "module" &&
                leftExpr.property.name === "exports"
            );
        case "Identifier":
            const leftId = left as Identifier;
            return leftId.name === "exports";

        default:
            return false;
    }
}

export function isPropertyAssignment(left: Pattern): boolean {
    if (left.type === "MemberExpression") {
        const leftExpr = left as MemberExpression;
        switch (leftExpr.object.type) {
            case "MemberExpression":
                const leftInner = leftExpr.object as MemberExpression;
                return (
                    leftInner.object.type === "Identifier" &&
                    leftInner.property.type === "Identifier" &&
                    leftInner.object.name === "module" &&
                    leftInner.property.name === "exports"
                );
            case "Identifier":
                const leftId = leftExpr.object as Identifier;
                return leftId.name === "exports";
            default:
                return false;
        }
    }
    return false;
}

export function patternToString(pattern: Pattern): string {
    // TODO: better handling for other patterns
    switch (pattern.type) {
        case "AssignmentPattern":
            return `${patternToString(pattern.left)}=${expressionToString(pattern.right)}`;
        case "Identifier":
            return pattern.name;
        case "MemberExpression":
            return expressionToString(pattern);
        default:
            return pattern.type;
    }
}

export function expressionToString(expression: Expression): string {
    // TODO: better handling for other expressions
    switch (expression.type) {
        case "MemberExpression":
            if (expression.object.type === "Super") {
                return `${expressionToString(expression)}.${expressionToString(
                    expression.property
                )}`;
            } else {
                return `${expressionToString(expression.object)}.${expressionToString(
                    expression.property
                )}`;
            }
        case "Identifier":
            return expression.name;
        default:
            return expression.type;
    }
}

export function createMethodSignatureString(id: string, params: Array<String>) {
    let signature = id;
    if (params.length > 0) {
        signature += "(";
        for (let i = 0; i < params.length; i++) {
            signature += i < params.length - 1 ? params[i] + "," : params[i];
        }
        signature += ")";
    } else {
        signature += "()";
    }
    return signature;
}

export function extractFunctionInfo(id: Identifier | null | undefined, baseFunc: BaseFunction): Function {
    const functionName = id ? id.name : "default function";
    const params = baseFunc.params;

    const paramsToString: Array<string> = [];
    for (let param of params) {
        paramsToString.push(patternToString(param));
    }
    return new Function(functionName, paramsToString);
}

export function extractMethodsFromClassBody(body: ClassBody): Array<string> {
    const methods: Array<string> = [];
    const bodyElements = body.body;
    for (let element of bodyElements) {
        if (element.type === "MethodDefinition") {
            const classMethod = extractFunctionInfo(element.value.id, element.value);
            if (element.key.type === "Identifier") {
                const methodSignature = createMethodSignatureString(
                    element.key.name,
                    classMethod.params
                );
                methods.push(methodSignature);
            }
        }
    }
    return methods;
}

export function extractExportsFromObject(object: ObjectExpression): Array<Export> {
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