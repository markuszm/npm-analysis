let Debug = require("debug");

class SampleClass {
    constructor() {
        let self = this;

        self.state = require("./state");

        self.log = Debug("off-the-record");

        const message = "some Message";

        self.log("message", message);

        self.state.addMessage(message);
    }
}
