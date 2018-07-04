import { Pattern, MemberExpression, Identifier } from "estree";

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
