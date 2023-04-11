# go-ast
A simpler and easier to understand ast toolset

## What is Go-Ast?

Ast is commonly used for project code generation, but its syntax is very difficult to understand, so we have written a more understandable and easy-to-use ast toolset

## Install

~~~go
go get github.com/cxiaoji/go-ast
~~~

## Examples

~~~go
import "fmt"

func main() {
	var filePath = "./test/person.go"
	h := NewAstHelper(filePath)
	fileDesc ,err:= h.GetFileDesc()
	if err != nil {
		return
	}
	fmt.Println(fileDesc.Package) // package name
	fmt.Println(fileDesc.StructDescs) // all struct 
}
~~~

