const fs = require("fs");

function run() {
  let [packedResults, testResults, analyzerResults] = [
    "com.returntocorp.npm-packer",
    "com.returntocorp.js-test-finder",
    "de.tu-darmstadt.callgraph"
  ].map(analyzer =>
    JSON.parse(fs.readFileSync(`/analysis/inputs/${analyzer}.json`, "utf8"))
  );
  let packedFiles = new Set(
    packedResults.results
      .filter(result => result.check_id === "npm-packed")
      .map(result => result.path)
  );
  let testFiles = new Set(testResults.results.map(result => result.path));
  return {
    results: analyzerResults.results.filter(
      result => packedFiles.has(result.path) && !testFiles.has(result.path)
    )
  };
}

try {
  console.log(JSON.stringify(run(), null, "\t"));
} catch (e) {
  console.log(
    JSON.stringify(
      {
        errors: [
          {
            message: e,
            data: {
              name: e.name,
              position: {
                file: e.fileName,
                line: e.lineNumber,
                column: e.columnNumber
              },
              stack: e.stack
            }
          }
        ]
      },
      null,
      "\t"
    )
  );
  process.exit(1);
}
