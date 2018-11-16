import { parse } from "@babel/parser";

export function parseAst(content: string): any {
    return parse(content, {
        sourceType: "module",
        allowImportExportEverywhere: true,

        plugins: ["jsx", "typescript", "estree"]
    });
}
