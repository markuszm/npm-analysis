import * as fs from "fs";
import * as process from "process";

import readdirp from "readdirp";

import * as parser from "./parser";
import { Traversal } from "./traversal";
import { Export } from "./model";
import { TernClient } from "./ternClient";

let debug = false;

// parse command arguments
const args: Array<string> = process.argv.slice(2);
const path = args[0];
if (args.length > 1 && args[1] === "debug") {
    debug = true;
}

const stats = fs.statSync(path);

const ternClient = new TernClient(debug);

if (stats.isDirectory()) {
    try {
        const definedExports: Array<Export> = [];
        readdirp(
            {
                root: path,
                fileFilter: ["*.ts", "*.js", "*.jsx"],
                directoryFilter: ["!.git", "!node_modules", "!assets"]
            },
            (fileInfo: any) => {
                if (debug) console.log("now parsing:", fileInfo.fullPath);

                const content = fs.readFileSync(fileInfo.fullPath, "utf-8");
                ternClient.addFile(fileInfo.name, fileInfo.fullPath);
                const traverse = new Traversal(ternClient, fileInfo.name, debug);
                try {
                    const ast = parser.parseAst(content);
                    definedExports.push(...traverse.traverseAst(ast));
                } catch (e) {
                    if (debug) {
                        console.error(e)
                    }
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
} else {
    const content = fs.readFileSync(path, "utf-8");
    ternClient.addFile(path, path);
    let ast = parser.parseAst(content);
    const traverse = new Traversal(ternClient, path, debug);
    const definedExports = traverse.traverseAst(ast);
    console.log(JSON.stringify(definedExports));
}
