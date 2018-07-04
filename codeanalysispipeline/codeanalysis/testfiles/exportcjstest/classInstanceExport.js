class Calculator {
    add(a, b) {
        return a + b;
    }
    substract(a, b) {
        return a - b;
    }
}

class AdvancedCalculator extends Calculator {
    multiply(a, b) {
        return a * b;
    }
    divide(a, b) {
        return a / b;
    }
}

module.exports = new AdvancedCalculator();
