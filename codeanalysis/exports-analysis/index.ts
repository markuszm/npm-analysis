import * as fs from "fs";
import * as process from "process";

import readdirp from "readdirp";

import * as parser from "./parser";
import * as traversal from "./traversal";

let debug = false;

// parse command arguments
const args: Array<string> = process.argv.slice(2);
const command = args[0];
const path = args[1];
if (args.length > 2 && args[2] === "debug") {
    debug = true;
}

switch (command) {
    case "file":
        const content = fs.readFileSync(path, "utf-8");
        let ast = parser.parseAst(content);
        const definedExports = traversal.traverseAst(ast, debug);
        console.log(JSON.stringify(definedExports));
        break;
    case "folder":
        try {
            const definedExports: Array<traversal.Export> = [];
            readdirp(
                { root: path, fileFilter: ["*.ts", "*.js", "*.jsx"] },
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
        break;
    default:
        console.log("wrong command - try file or folder");
        break;
}
