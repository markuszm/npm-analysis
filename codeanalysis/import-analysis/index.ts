import * as fs from "fs";
import * as process from "process";

import readdirp from "readdirp";

import * as parser from "./parser";
import { Traversal } from "./traversal";
import { Import } from "./model";

let debug = false;

// parse command arguments
const args: Array<string> = process.argv.slice(2);
const path = args[0];
if (args.length > 1 && args[1] === "debug") {
    debug = true;
}

const stats = fs.statSync(path);

if (stats.isDirectory()) {
    try {
        const definedImports: Array<Import> = [];
        readdirp(
            {
                root: path,
                fileFilter: ["*.ts", "*.js", "*.jsx"],
                directoryFilter: ["!.git", "!node_modules", "!assets"]
            },
            (fileInfo: any) => {
                if (debug) console.log("now parsing:", fileInfo.fullPath);

                const content = fs.readFileSync(fileInfo.fullPath, "utf-8");
                const traverse = new Traversal(debug);
                try {
                    const ast = parser.parseAst(content);
                    definedImports.push(...traverse.traverseAst(ast));
                } catch (e) {
                    if (debug) {
                        console.error(e)
                    }
                    // ignore errors in parsing for now
                }
            },
            // after file processing
            () => {
                console.log(JSON.stringify(definedImports));
            }
        );
    } catch (err) {
        console.error(err);
        console.error("could not find any files");
    }
} else {
    const content = fs.readFileSync(path, "utf-8");
    let ast = parser.parseAst(content);
    const traverse = new Traversal(debug);
    const definedImports = traverse.traverseAst(ast);
    console.log(JSON.stringify(definedImports));
}
