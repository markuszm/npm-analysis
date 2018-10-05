const process = require("process");

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

// model
let exportedFunctions = [];

// requiring package and retrieving function definitions
const imported = require(packageName);

const props = Object.getOwnPropertyNames(imported);
for (let prop of props) {
    try {
        if (typeof imported[prop] === "function") {
            exportedFunctions.push({
                Name: prop,
                InternalName: imported[prop].name,
                Contents: Buffer.from(imported[prop].toString()).toString('base64')
            });
        }
    } catch (error) {}
}

// JSON output
if (isR2C) {
    r2cResults = [];
    for (let exportedFunction of exportedFunctions) {
        r2cResults.push({
            check_id: "export",
            file: packageName,
            extra: exportedFunction
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
    console.log(JSON.stringify(exportedFunctions));
}
