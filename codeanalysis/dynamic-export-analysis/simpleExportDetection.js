const readdirp = require("readdirp");

const process = require("process");
const path = require("path");
const fs = require("fs");

const ANALYSIS_NAME = "dynamic-exports";
const ANALYSIS_VERSION = "0.1.0";

// argument parsing
const args = process.argv.slice(2);
if (args.length === 0) {
    console.error("missing package argument");
    process.exit(1);
}
const packageName = args[0];
const isR2C = args[1];

// ~ Model ~
let exportedFunctions = [];

// ~ Tracking of exports ~

// requiring package and retrieving function definitions
const imported = require(packageName);

// filters out default prototype functions existing on classes
const defaultPrototypeFunctions = [
    "constructor",
    "__defineGetter__",
    "__defineSetter__",
    "hasOwnProperty",
    "__lookupGetter__",
    "__lookupSetter__",
    "isPrototypeOf",
    "propertyIsEnumerable",
    "toString",
    "valueOf",
    "toLocaleString"
];

let props = Object.getOwnPropertyNames(imported);
// also tracks functions on class instance by getting the prototype of the object see https://stackoverflow.com/questions/31054910/get-functions-methods-of-a-class
if (typeof imported === "object") {
    props.push(
        ...Object.getOwnPropertyNames(Object.getPrototypeOf(imported)).filter(
            prop => !defaultPrototypeFunctions.includes(prop)
        )
    );
}
if (typeof imported === "function") {
    const functionBody = imported.toString();
    exportedFunctions.push({
        Name: "default",
        InternalName: imported.name,
        Contents: functionBody
    });
}
for (let prop of props) {
    try {
        if (typeof imported[prop] === "function") {
            const functionBody = imported[prop].toString();
            exportedFunctions.push({
                Name: prop,
                InternalName: imported[prop].name,
                Contents: functionBody.replace(/\s/g, ""),
                Locations: []
            });
        }
    } catch (error) {}
}

// ~ Searching for file location of exported method ~

readdirp(
    {
        root: path.join(__dirname, "node_modules", packageName),
        fileFilter: ["*.js"],
        directoryFilter: ["!.git", "!node_modules", "!assets"]
    },
    fileInfo => {
        const fullPath = fileInfo.fullPath;
        const fileContents = fs.readFileSync(fullPath, "utf8").replace(/\s/g, "");
        for (let exportedFunction of exportedFunctions) {
            if (exportedFunction.Contents === "") {
                continue;
            }
            let functionIndex = fileContents.indexOf(exportedFunction.Contents);
            if (functionIndex !== -1) {
                exportedFunction.Locations.push({ File: fileInfo.path, Index: functionIndex });
            }
        }
    },
    () => {
        // ~ JSON output ~
        if (isR2C) {
            const r2cResults = [];
            for (let exportedFunction of exportedFunctions) {
                r2cResults.push({
                    check_id: "export",
                    file: packageName,
                    extra: {
                        Name: exportedFunction.Name,
                        InternalName: exportedFunction.InternalName,
                        Locations: exportedFunction.Locations
                    }
                });
            }
            const output = {
                spec_version: "0.1.0",
                name: ANALYSIS_NAME,
                version: ANALYSIS_VERSION,
                results: r2cResults
            };
            console.log(JSON.stringify(output));
        } else {
            const output = [];
            for (let exportedFunction of exportedFunctions) {
                output.push({
                    Name: exportedFunction.Name,
                    InternalName: exportedFunction.InternalName,
                    Locations: exportedFunction.Locations
                });
            }
            console.log(JSON.stringify(output));
        }
    }
);
