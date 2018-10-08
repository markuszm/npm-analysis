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

let props = Object.getOwnPropertyNames(imported);
// also tracks functions on class instance by getting the prototype of the object see https://stackoverflow.com/questions/31054910/get-functions-methods-of-a-class
if (typeof imported === "object") {
    props.append(...Object.getOwnPropertyNames(Object.getPrototypeOf(imported)))
}
if(typeof imported === "function") {
    const functionBody = imported[prop].toString();
    exportedFunctions.push({
        Name: "default",
        InternalName: imported.name,
        Contents: functionBody
    });
}
// TODO: find location via searching files
for (let prop of props) {
    try {
        if (typeof imported[prop] === "function") {
            const functionBody = imported[prop].toString();
            exportedFunctions.push({
                Name: prop,
                InternalName: imported[prop].name,
                Contents: functionBody
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
