import {
    ArrayExpression,
    ArrayPattern,
    ArrowFunctionExpression,
    AwaitExpression,
    BaseFunction,
    BinaryExpression,
    ClassBody,
    ClassExpression,
    ConditionalExpression,
    Expression,
    FunctionExpression,
    Identifier,
    Literal,
    LogicalExpression,
    MemberExpression,
    MetaProperty,
    NewExpression,
    ObjectExpression,
    Pattern,
    SequenceExpression,
    SpreadElement,
    Super,
    TaggedTemplateExpression,
    TemplateLiteral,
    ThisExpression,
    UnaryExpression,
    UpdateExpression,
    YieldExpression
} from "estree";
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
    if (!pattern) {
        return "null";
    }

    let patternString = "";
    switch (pattern.type) {
        case "ArrayPattern":
            for (let i = 0; i < pattern.elements.length; i++) {
                const element = pattern.elements[i];
                patternString += patternToString(element);
                if (i < pattern.elements.length - 1) {
                    patternString += ",";
                }
            }
            break;
        case "AssignmentPattern":
            patternString = `${patternToString(pattern.left)}=${expressionToString(pattern.right)}`;
            break;
        case "Identifier":
            patternString = pattern.name;
            break;
        case "MemberExpression":
            patternString = expressionToString(pattern);
            break;
        case "ObjectPattern":
            for (let i = 0; i < pattern.properties.length; i++) {
                const prop = pattern.properties[i];
                patternString += patternToString(prop.value);
                if (i < pattern.properties.length - 1) {
                    patternString += ",";
                }
            }
            break;
        case "RestElement":
            patternString += "..." + patternToString(pattern.argument);
            break;
    }
    return patternString;
}

export function expressionToString(expression: Expression): string {
    if (!expression) {
        return "null";
    }

    switch (expression.type) {
        case "ArrayExpression":
            let arrayString = "[";
            for (let i = 0; i < expression.elements.length; i++) {
                const element = expression.elements[i];
                arrayString +=
                    element && element.type === "SpreadElement"
                        ? "..." + expressionToString(element.argument)
                        : expressionToString(element);
                if (i < expression.elements.length - 1) {
                    arrayString += ",";
                }
            }
            arrayString += "]";
            return arrayString;
        case "ObjectExpression":
            let propertiesString = "{";
            for (let i = 0; i < expression.properties.length; i++) {
                const property = expression.properties[i];
                propertiesString += `${expressionToString(property.key)}:${
                    (property.value && property.value.type === "AssignmentPattern") ||
                    property.value.type === "ObjectPattern" ||
                    property.value.type === "ArrayPattern" ||
                    property.value.type === "RestElement"
                        ? patternToString(property.value)
                        : expressionToString(property.value)
                }`;
                if (i < expression.properties.length - 1) {
                    propertiesString += ",";
                }
            }
            propertiesString += "}";
            return propertiesString;
        case "FunctionExpression":
            let functionString = `${expression.async ? "async " : ""}function(`;
            for (let i = 0; i < expression.params.length; i++) {
                let param = expression.params[i];
                functionString += patternToString(param);
                if (i < expression.params.length - 1) {
                    functionString += ",";
                }
            }
            functionString += ") {...}";
            return functionString;
        case "ArrowFunctionExpression":
            let arrowFunctionString = `${expression.async ? "async " : ""}(`;
            for (let i = 0; i < expression.params.length; i++) {
                let param = expression.params[i];
                arrowFunctionString += patternToString(param);
                if (i < expression.params.length - 1) {
                    arrowFunctionString += ",";
                }
            }
            arrowFunctionString += ") => {...}";
            return arrowFunctionString;
        case "YieldExpression":
            return `yield ${expression.argument ? expressionToString(expression.argument) : ""}`;
        case "UnaryExpression":
            return expression.operator + expressionToString(expression.argument);
        case "UpdateExpression":
            return expression.prefix
                ? expression.operator + expressionToString(expression.argument)
                : expressionToString(expression.argument) + expression.operator;
        case "BinaryExpression":
            return `${expressionToString(expression.left)} ${
                expression.operator
            } ${expressionToString(expression.right)}`;
        case "AssignmentExpression":
            return `${
                expression.left.type === "MemberExpression"
                    ? expressionToString(expression.left)
                    : patternToString(expression.left)
            } ${expression.operator} ${expressionToString(expression.right)}`;
        case "LogicalExpression":
            return `${expressionToString(expression.left)} ${
                expression.operator
            } ${expressionToString(expression.right)}`;
        case "ConditionalExpression":
            return `${expressionToString(expression.test)} ? \
             ${expressionToString(expression.consequent)} \ 
             : ${expressionToString(expression.alternate)}`;
        case "CallExpression":
            return methodExpressionToString(expression.callee, expression.arguments);
        case "NewExpression":
            return "new " + methodExpressionToString(expression.callee, expression.arguments);
        case "SequenceExpression":
            let sequenceString = "";
            for (let i = 0; i < expression.expressions.length; i++) {
                let expr = expression.expressions[i];
                sequenceString += expressionToString(expr);
                if (i < expression.expressions.length - 1) {
                    sequenceString += ",";
                }
            }
            return sequenceString;
        case "TemplateLiteral":
            let templateString = "";
            let exprIndex = 0;
            for (let quasi of expression.quasis) {
                templateString += quasi;
                if (exprIndex < expression.expressions.length) {
                    templateString += expressionToString(expression.expressions[exprIndex]);
                }
            }
            return templateString;
        case "TaggedTemplateExpression":
            return expression.type.toString();
        case "ClassExpression":
            const className = expression.id ? expression.id.name : "";
            return "class " + className;
        case "MetaProperty":
            return `${expression.meta.name}.${expression.property.name}`;
        case "AwaitExpression":
            return "await " + expressionToString(expression.argument);
        case "ThisExpression":
            return "this";
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
        case "Literal":
            return expression.value ? expression.value.toString() : "null";
        case "Identifier":
            return expression.name;
    }
}

function methodExpressionToString(
    callee: Expression | Super,
    args: Array<Expression | SpreadElement>
): string {
    const calleeString = callee.type === "Super" ? "super" : expressionToString(callee);
    let argumentsString = "";
    for (let i = 0; i < args.length; i++) {
        let arg = args[i];
        argumentsString +=
            arg.type === "SpreadElement"
                ? "..." + expressionToString(arg.argument)
                : expressionToString(arg);
        if (i < args.length - 1) {
            argumentsString += ",";
        }
    }
    return `${calleeString}(${argumentsString})`;
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

export function extractFunctionInfo(
    id: Identifier | null | undefined,
    baseFunc: BaseFunction
): Function {
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
