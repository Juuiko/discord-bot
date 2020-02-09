package bot

import (
	"errors"
	"strconv"
)

func hasPrecedence(op1 string, op2 string) bool {
	if op2 == "(" || op2 == ")" {
		return false
	}
	if (op1 == "*" || op1 == "/") && (op2 == "+" || op2 == "-") {
		return false
	}
	return true
}

func applyOp(op string, b int64, a int64) (int64, error) {
	switch op {
	case "+":
		return a + b, nil
	case "-":
		return a - b, nil
	case "*":
		return a * b, nil
	case "/":
		if b == 0 {
			return b, errors.New("cannot divide by 0")
		}
		return a / b, nil
	}
	return 0, nil
}

func calculate(tokens []string) (int64, error) {
	var valueST []int64
	var opST []string
	for i := 1; i <= len(tokens)-1; i++ {
		if tokens[i] == "(" {
			opST = append(opST, tokens[i])
		} else if tokens[i] == ")" {
			for opST[len(opST)-1] != "(" {
				pop1 := opST[len(opST)-1]
				pop2 := valueST[len(valueST)-1]
				pop3 := valueST[len(valueST)-2]
				opST = opST[:len(opST)-1]          // Pop opST
				valueST = valueST[:len(valueST)-2] //Pop valuesST x2
				newValue, err := applyOp(pop1, pop2, pop3)
				if err != nil {
					return 0, err
				}
				valueST = append(valueST, newValue)
			}
			opST = opST[:len(opST)-1]
		} else if tokens[i] == "+" || tokens[i] == "-" || tokens[i] == "*" || tokens[i] == "/" {
			for len(opST) > 0 && hasPrecedence(tokens[i], opST[len(opST)-1]) {
				pop1 := opST[len(opST)-1]
				pop2 := valueST[len(valueST)-1]
				pop3 := valueST[len(valueST)-2]
				opST = opST[:len(opST)-1]          // Pop opST
				valueST = valueST[:len(valueST)-2] //Pop valuesST x2
				newValue, err := applyOp(pop1, pop2, pop3)
				if err != nil {
					return 0, err
				}
				valueST = append(valueST, newValue)
			}
			opST = append(opST, tokens[i])
		} else {
			integer, err := strconv.ParseInt(tokens[i], 10, 64)
			if err != nil {
				return 0, errors.New("input error")
			}
			valueST = append(valueST, integer)
		}
	}
	for len(opST) > 0 {
		pop1 := opST[len(opST)-1]
		pop2 := valueST[len(valueST)-1]
		pop3 := valueST[len(valueST)-2]
		opST = opST[:len(opST)-1]          // Pop opST
		valueST = valueST[:len(valueST)-2] //Pop valuesST x2
		newValue, err := applyOp(pop1, pop2, pop3)
		if err != nil {
			return 0, err
		}
		valueST = append(valueST, newValue)
	}
	return valueST[len(valueST)-1], nil
}
