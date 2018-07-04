const parser = require("@babel/parser");

exports.parseAst = content =>
    parser.parse(content, {
        sourceType: "module",
        allowImportExportEverywhere: true,

        plugins: ["jsx", "typescript", "estree"]
    });
