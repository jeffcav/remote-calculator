package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

// TreeNode Implements the node of a tree
type TreeNode struct {
	Value string `json:"value" yaml:"value"`
	Left  *Tree  `json:"left" yaml:"left"`
	Right *Tree  `json:"right" yaml:"right"`
}

// Tree Represents a tree
type Tree struct {
	Node TreeNode `json:"node" yaml:"node"`
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

func runServer(useYAML *bool) {
	listen, err := net.Listen("tcp", ":10011")
	if err != nil {
		fmt.Printf("Server failed. Err: %v\n", err)
	}

	defer listen.Close()

	for {
		conn, _ := listen.Accept()
		defer conn.Close()

		var exprAST Tree
		if *useYAML {
			buff := make([]byte, 4*1024) //4KB
			n, _ := conn.Read(buff)
			err = yaml.Unmarshal(buff[:n], &exprAST)
		} else {
			encodedExpr := json.NewDecoder(conn)
			encodedExpr.Decode(&exprAST)
		}

		// Compute expression - Result will be an AST with a single node
		result := strconv.Itoa(computeExpression(&exprAST))
		resultAST := Tree{TreeNode{Value: result}}

		// Encode result AST to the choosen format (JSON or YAML)
		var encodedResult []byte
		if *useYAML {
			encodedResult, _ = yaml.Marshal(resultAST)
		} else {
			encodedResult, _ = json.Marshal(resultAST)
		}

		// Send result in JSON format
		conn.Write([]byte(encodedResult))
	}
}

func printInfo(useYAML *bool) {
	fmt.Printf("\nRemote calculator with - ")
	if *useYAML {
		fmt.Printf("YAML\n\n")
	} else {
		fmt.Printf("JSON\n\n")
	}

	fmt.Printf("WRONG Expression:   10+4/2 (Without spaces)\n")
	fmt.Printf("CORRECT Expression: 10 + 4 / 2 (With spaces)\n\n")
}

func main() {
	useYAML := flag.Bool("yaml", false, "Use YAML format")
	flag.Parse()

	fmt.Print("Starting server... ")
	go runServer(useYAML)
	time.Sleep(1 * time.Second)
	fmt.Println("OK")

	printInfo(useYAML)

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

		// Build Abstract Syntax Tree (AST) from expression
		exprAST := buildASTFromString(expr)

		// Encode AST to the choosen format (JSON or YAML)
		var encodedExpr []byte
		if *useYAML {
			encodedExpr, err = yaml.Marshal(exprAST)
			if err != nil {
				fmt.Printf("YAML marshall error %v:\n", err)
			}
		} else {
			encodedExpr, _ = json.Marshal(exprAST)
			if err != nil {
				fmt.Printf("JSON marshall error %v:\n", err)
			}
		}

		// Send expression in JSON/YAML format
		//fmt.Println(string([]byte(encodedExpr)))
		conn.Write([]byte(encodedExpr))

		// Read result in JSON/YAML format and decode to AST
		var resultAST Tree
		if *useYAML {
			buff := make([]byte, 4*1024) //4KB
			n, _ := conn.Read(buff)
			yaml.Unmarshal(buff[:n], &resultAST)
		} else {
			encodedResult := json.NewDecoder(conn)
			encodedResult.Decode(&resultAST)
		}

		// Print result
		fmt.Printf("Result: %s\n\n", resultAST.Node.Value)
	}
}
