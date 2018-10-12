import * as fs from "fs";

import readdirp from "readdirp";

import * as parser from "./parser";
import { Traversal } from "./traversal";
import { Export } from "./model";
import { TernClient } from "./ternClient";
import * as path from "path";

let debug = false;

// parse command arguments
const args: Array<string> = process.argv.slice(2);
const entryPath = args[0];
if (args.length > 1 && args[1] === "debug") {
    debug = true;
}

const stats = fs.statSync(entryPath);

const ternClient = new TernClient(debug);

function getFileNameInsidePackage(fileInfo: any) {
    const fullPath: string = fileInfo.fullPath;
    const regexFileName = /(?:\/?.+)(?:\/package\/)(.+)/;
    if (fullPath.indexOf("package") != -1) {
        let [, fileName]: RegExpMatchArray = fullPath.match(regexFileName) || [];
        if (!fileName || fileName === "") {
            return trimExt(fileInfo.name);
        }
        return trimExt(fileName);
    }

    return trimExt(fileInfo.name);
}

function trimExt(fileName: string): string {
    return fileName.replace(path.extname(fileName), "");
}

if (stats.isDirectory()) {
    try {
        const definedExports: Array<Export> = [];
        readdirp(
            {
                root: entryPath,
                fileFilter: function(fileInfo: any) {
                    const fileName = fileInfo.name;
                    const ext = path.extname(fileName);
                    switch (ext) {
                        case ".js":
                        case ".ts":
                        case ".jsx":
                            return true;
                        case "":
                            const file = fs.readFileSync(fileInfo.fullPath, { encoding: "utf8" });
                            return file.startsWith("#!/usr/bin/env node");
                        default:
                            return false;
                    }
                },
                directoryFilter: ["!.git", "!node_modules", "!assets"]
            },
            (fileInfo: any) => {
                if (debug) console.log("now parsing:", fileInfo.fullPath);

                const content = fs.readFileSync(fileInfo.fullPath, "utf-8");
                const fileNameInsidePackage = getFileNameInsidePackage(fileInfo);
                ternClient.addFile(fileNameInsidePackage, fileInfo.fullPath);
                const traverse = new Traversal(ternClient, fileNameInsidePackage, debug);
                try {
                    const ast = parser.parseAst(content);
                    definedExports.push(...traverse.traverseAst(ast));
                } catch (e) {
                    if (debug) {
                        console.error(e);
                    }
                    // ignore errors in parsing for now
                }
                ternClient.delFile(fileNameInsidePackage);
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
    const content = fs.readFileSync(entryPath, "utf-8");
    ternClient.addFile(entryPath, entryPath);
    let ast = parser.parseAst(content);
    const traverse = new Traversal(ternClient, entryPath, debug);
    const definedExports = traverse.traverseAst(ast);
    console.log(JSON.stringify(definedExports));
}
