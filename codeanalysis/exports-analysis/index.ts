import * as fs from "fs";
import * as process from "process";

import readdirp from "readdirp";

import * as parser from "./parser";
import * as traversal from "./traversal";

let debug = false;

// parse command arguments
const args: Array<string> = process.argv.slice(2);
const path = args[0];
if (args.length > 1 && args[1] === "debug") {
    debug = true;
}

try {
    const definedExports: Array<traversal.Export> = [];
    readdirp(
        { root: path, fileFilter: ["*.ts", "*.js", "*.jsx"], directoryFilter: [ '!.git', '!node_modules', '!assets' ] },
        fileInfo => {
            const content = fs.readFileSync(fileInfo.fullPath, "utf-8");
            try {
                let ast = parser.parseAst(content);
                definedExports.push(...traversal.traverseAst(ast, debug));
            } catch (e) {
                // ignore errors in parsing for now
            }
        },
        // after file processing
        () => {
            console.log(JSON.stringify(definedExports));
        }
    );
} catch (err) {
    console.error(err);
    console.error("could not find any files");
}
