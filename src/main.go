package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// TreeNode Implements the node of a tree
type TreeNode struct {
	Value string `json:"value"`
	Left  *Tree  `json:"left"`
	Right *Tree  `json:"right"`
}

// Tree Represents a tree
type Tree struct {
	Node TreeNode `json:"node"`
}

func tokenize(expr string) []string {
	toks := strings.Split(expr, " ")
	return toks
}

func printAST(tree *Tree) {
	if tree.Node.Left != nil {
		printAST(tree.Node.Left)
	}

	println(tree.Node.Value)

	if tree.Node.Right != nil {
		printAST(tree.Node.Right)
	}
}

func computeExpression(tree *Tree) int {
	if tree.Node.Left == nil && tree.Node.Right == nil {
		val, _ := strconv.Atoi(tree.Node.Value)
		return val
	}

	left := computeExpression(tree.Node.Left)
	right := computeExpression(tree.Node.Right)

	switch tree.Node.Value {
	case "+":
		return left + right
	case "-":
		return left - right
	case "*":
		return left * right
	case "/":
		return left / right
	default:
		fmt.Println("Expression error")
		return 0
	}
}

func buildASTFromString(expr string) *Tree {
	var root *Tree = nil
	var higherPrecedenceOperation bool = false
	toks := tokenize(expr)

	for _, tok := range toks {
		node := TreeNode{Value: tok}

		// Insert operation
		if tok == "+" || tok == "-" || tok == "*" || tok == "/" {
			if root == nil {
				root = &Tree{Node: node}
			} else {
				// Apply precedence
				higherPrecedenceOperation = (tok == "*" || tok == "/") && (root.Node.Value == "+" || root.Node.Value == "-")
				if higherPrecedenceOperation {
					node.Left = root.Node.Right
					root.Node.Right = &Tree{Node: node}
				} else {
					node.Left = root
					root = &Tree{Node: node}
				}
			}
		} else { // Insert operand
			if root == nil {
				root = &Tree{Node: node}
			} else {
				if higherPrecedenceOperation && root.Node.Right != nil {
					root.Node.Right.Node.Right = &Tree{Node: node}
				} else {
					root.Node.Right = &Tree{Node: node}
				}
				higherPrecedenceOperation = false
			}
		}
	}

	return root
}

func runServer() {
	listen, err := net.Listen("tcp", ":10011")
	if err != nil {
		fmt.Printf("Server failed. Err: %v\n", err)
	}

	fmt.Println("Server started")
	defer listen.Close()

	for {
		conn, _ := listen.Accept()
		defer conn.Close()

		// Read Expression
		encodedExpr := json.NewDecoder(conn)
		var exprAST Tree
		encodedExpr.Decode(&exprAST)

		// Compute and enconde result to JSON format
		result := strconv.Itoa(computeExpression(&exprAST))
		resultAST := Tree{TreeNode{Value: result}}
		encodedResult, _ := json.Marshal(&resultAST)

		// Send result in JSON format
		conn.Write([]byte(encodedResult))
	}
}

func main() {
	go runServer()
	time.Sleep(1 * time.Second)

	fmt.Printf("REMOTE CALCULATOR\n\n")
	fmt.Printf("Insert <SPACE> between operands and operations.\n")
	fmt.Printf("WRONG Expression: 10+4/2\n")
	fmt.Printf("CORRECT Expression: 10 + 4 / 2\n\n")

	for {
		// Connect to the calculator server
		conn, err := net.Dial("tcp", "127.0.0.1:10011")
		if err != nil {
			fmt.Printf("Did not connect: %v\n", err)
		}

		defer conn.Close()

		// Read expression
		fmt.Print("Expression: ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		expr := scanner.Text()

		// Build AST from expression and encode to JSON
		exprAST := buildASTFromString(expr)
		encodedExpr, _ := json.Marshal(exprAST)

		// Send expression in JSON format
		conn.Write([]byte(encodedExpr))

		// Read result in JSON format
		encodedResult := json.NewDecoder(conn)

		// Convert result to AST
		var resultAST Tree
		encodedResult.Decode(&resultAST)

		// Print result
		fmt.Printf("Result: %s\n", resultAST.Node.Value)
	}
}
