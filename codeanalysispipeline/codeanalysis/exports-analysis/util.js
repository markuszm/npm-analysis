exports.isDirectAssignment = left =>
    (left.type === "MemberExpression" &&
        left.object.name === "module" &&
        left.property.name === "exports") ||
    (left.type === "Identifier" && left.name === "exports");

exports.isPropertyAssignment = left =>
    left.type === "MemberExpression" &&
    ((left.object.type === "MemberExpression" &&
        left.object.object.name === "module" &&
        left.object.property.name === "exports") ||
        (left.object.type === "Identifier" && left.object.name === "exports"));
