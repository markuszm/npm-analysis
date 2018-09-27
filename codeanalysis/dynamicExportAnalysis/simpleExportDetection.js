const process = require("process");

// argument parsing
const args = process.argv.slice(2);
if (args.length === 0) {
    console.error("missing package argument");
    process.exit(1);
}
const packageName = args[0];

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
                Contents: imported[prop].toString()
            });
        }
    } catch (error) {}
}

// JSON output
console.log(JSON.stringify(exportedFunctions));
