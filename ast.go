// Ast parsing file
package ast

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

type Expr struct {
	Name string
	Val  interface{}
}

// struct describe
type StructDesc struct {
	Name string `json:"name"` // struct name eg:UserFav

	Fields  []*StructFieldDesc `json:"fields"`  // field
	Comment string             `json:"comment"` // comment
}

// struct field describe
type StructFieldDesc struct {
	Name    string            `json:"name"` // name
	Type    string            `json:"type"`
	Tags    map[string]string `json:"tag"` // tag
	Comment string            `json:"comment"`
}

// ast file desc
type FileDesc struct {
	Package     string            `json:"package"`      // package name
	Imports     map[string]string `json:"imports"`      // import；eg：commonmodel:/common/model.go
	StructDescs []*StructDesc     `json:"struct_descs"` // structs；

	ReferencedFileDescs []*FileDesc `json:"referenced_file_descs"` // Referenced file;Imports key:FileDesc
}

// ast helper
type AstHelper struct {
	astFile  *ast.File
	FileDesc *FileDesc
	FilePath string
}

func NewDefAstHelper() *AstHelper {
	return &AstHelper{}
}

func NewAstHelper(filePath string) *AstHelper {
	if filePath == "" {
		panic("filePath not empty")
	}

	fileSet := token.NewFileSet()
	f, err := parser.ParseFile(fileSet, filePath, nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}
	if f == nil {
		panic("parser file err")
	}
	return &AstHelper{
		astFile:  f,
		FilePath: filePath,
	}
}

func (h *AstHelper) GetAstFile() *ast.File {
	return h.astFile
}

func (h *AstHelper) GetFileDesc() (*FileDesc, error) {
	var err error
	pk, err := h.getPackage()
	if err != nil {
		return nil, err
	}
	imports, err := h.getImports()
	if err != nil {
		return nil, err
	}
	structDescs, err := h.getStructDescs()
	if err != nil {
		return nil, err
	}
	referencedFileDesc, err := h.getReferencedFileDesc(imports)
	if err != nil {
		//return nil, err
	}
	var fileDesc = &FileDesc{
		Package:             pk,
		Imports:             imports,
		StructDescs:         structDescs,
		ReferencedFileDescs: referencedFileDesc,
	}
	return fileDesc, nil
}

// get package name
func (h *AstHelper) getPackage() (string, error) {
	astFile := h.GetAstFile()

	if astFile.Name == nil {
		return "", nil
	}
	return astFile.Name.Name, nil
}

// get import;
func (h *AstHelper) getImports() (map[string]string, error) {
	var m = map[string]string{}
	var err error
	_ = err
	if len(h.GetAstFile().Imports) <= 0 {
		return m, nil
	}
	for _, importSpec := range h.GetAstFile().Imports {
		if importSpec.Path == nil {
			return m, nil
		}
		var aliasName = ""
		var importPath = strings.ReplaceAll(importSpec.Path.Value, "\"", "")
		if importSpec == nil || importSpec.Name == nil || importSpec.Name.Name == "" {
			aliasName = importPath
		} else {
			aliasName = importSpec.Name.Name
		}
		m[aliasName] = importPath
	}
	return m, nil
}

func (h *AstHelper) getStructDescs() ([]*StructDesc, error) {
	astFile := h.GetAstFile()
	if astFile.Scope == nil {
		return nil, ErrInvalidEmptyBody
	}
	var structDescs = make([]*StructDesc, 0)
	for objName, obj := range astFile.Scope.Objects {
		_ = objName
		if obj == nil {
			continue
		}
		if obj.Kind != ast.Typ {
			continue
		}
		if obj.Decl == nil {
			continue
		}
		typeSpec, ok := obj.Decl.(*ast.TypeSpec)
		if !ok {
			continue
		}
		if typeSpec.Type == nil {
			continue
		}
		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			continue
		}
		var fields = make([]*StructFieldDesc, 0)
		if structType.Fields != nil {
			for _, astField := range structType.Fields.List {
				var fieldName = ""
				if len(astField.Names) > 0 && astField.Names[0] != nil {
					fieldName = astField.Names[0].Name
				}
				comment := h.getComment(astField.Comment)
				if h.getComment(astField.Doc) != "" {
					comment += " " + h.getComment(astField.Doc)
				}

				if astField.Type == nil {
					continue
				}
				// field type
				var fieldType = ""
				switch d := astField.Type.(type) {
				case *ast.SelectorExpr:
					// embedded
					if d.X == nil || d.Sel == nil {
						break
					}
					ident, ok := d.X.(*ast.Ident)
					if !ok {
						break
					}
					// eg commonmodel.BaseField
					alias := ident.Name    // commonmodel
					fieldType = d.Sel.Name // BaseField
					if alias != "" && fieldType != "" {
						fieldType = fmt.Sprintf("%s.%s", alias, fieldType)
					}
					if fieldName == "" {
						fieldName = fieldType
					}
				case *ast.Ident:
					// go normal type
					fieldType = d.Name
				case *ast.ArrayType:
					// array or slice
					if d.Elt == nil {
						break
					}
					switch dd := d.Elt.(type) {
					case *ast.InterfaceType:
						fieldType = "[]interface{}"
					case *ast.Ident:
						fieldType = "[]" + dd.Name
					}

				}
				var tags = map[string]string{}
				if astField.Tag != nil {
					var tag = strings.ReplaceAll(astField.Tag.Value, "`", "")
					tags = h.splitFieldTags(tag)
				}
				var field = &StructFieldDesc{
					Name:    fieldName,
					Type:    fieldType,
					Tags:    tags,
					Comment: comment,
				}
				fields = append(fields, field)
			}
		}
		comment := h.getComment(typeSpec.Comment)
		if h.getComment(typeSpec.Doc) != "" {
			comment += " " + h.getComment(typeSpec.Doc)
		}
		var structDesc = &StructDesc{
			Name:    objName,
			Fields:  fields,
			Comment: comment,
		}
		structDescs = append(structDescs, structDesc)
	}
	return structDescs, nil
}

func (h *AstHelper) getNameFromAstExpr(expr ast.Expr) Expr {
	if expr == nil {
		return Expr{}
	}
	var res = Expr{}
	switch d := expr.(type) {
	case *ast.Ident:
		res.Name = d.Name
	case *ast.ArrayType:
		if d.Elt != nil {
			res = h.getNameFromAstExpr(d.Elt)
			res.Name = fmt.Sprintf("[]%s", res.Name)
		}
	case *ast.StarExpr:
		// ptr type
		if d.X != nil {
			res = h.getNameFromAstExpr(d.X)
			res.Name = fmt.Sprintf("*%s", res.Name)
		}
	case *ast.BasicLit:
		res.Name = strings.ToLower(d.Kind.String())
		res.Val = strings.ReplaceAll(d.Value, "\"", "")
	case *ast.CompositeLit:
		switch dd := d.Type.(type) {
		case *ast.Ident:
			res.Name = dd.Name
			res.Val = fmt.Sprintf("%s{}", res.Name)
		case *ast.ArrayType:
			switch ddd := dd.Elt.(type) {
			case *ast.Ident:
				res.Name = ddd.Name
				res.Val = fmt.Sprintf("[]%s{}", res.Name)
			case *ast.StarExpr:
				ident := ddd.X.(*ast.Ident)
				res.Name = ident.Name
				res.Val = fmt.Sprintf("[]*%s{}", res.Name)
			}

		}

	case *ast.UnaryExpr:

		switch dd := d.X.(type) {
		case *ast.CompositeLit:
			ident := dd.Type.(*ast.Ident)
			res.Name = ident.Name
			res.Val = fmt.Sprintf("%s%s{}", d.Op.String(), res.Name)

		}

	}
	return res
}

func (h *AstHelper) getComment(comment *ast.CommentGroup) string {
	if comment == nil {
		return ""
	}
	if len(comment.List) == 0 {
		return ""
	}
	var text string
	for _, c := range comment.List {
		text += fmt.Sprintf("%s;", strings.ReplaceAll(c.Text, "// ", ""))
	}

	return text

}

// string to uint64
func (h *AstHelper) stringToUint64(in string) (uint64, error) {
	if in == "" {
		return 0, nil
	}
	parseInt, err := strconv.ParseInt(in, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint64(parseInt), nil
}

func (h *AstHelper) getReferencedFileDesc(imports map[string]string) ([]*FileDesc, error) {
	var fileDescs = make([]*FileDesc, 0)
	for _, importPath := range imports {
		// check
		splits := strings.Split(importPath, "/")
		if len(splits) < 2 {
			continue
		}
		var localReference = true
		for _, spilt := range splits[:2] {
			if !strings.Contains(h.GetFilePath(), spilt) {
				localReference = false
				break
			}
		}
		if !localReference {
			continue
		}

		// get full filePath
		fileDir := path.Join(h.GetFilePath()[:strings.Index(h.GetFilePath(), splits[0])], importPath)

		//
		files, err := h.readDirFiles(fileDir)
		if err != nil {
			return nil, err
		}
		for _, filePath := range files {
			if filePath == h.GetFilePath() {
				continue
			}
			fileDesc, err := NewAstHelper(filePath).GetFileDesc()
			if err != nil {
				continue
			}
			fileDescs = append(fileDescs, fileDesc)
		}

	}
	return fileDescs, nil

}

func (a *AstHelper) readDirFiles(fileDir string) ([]string, error) {
	var files []string
	var walkFunc = func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return nil
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		//fmt.Printf("%s\n", path)
		return nil
	}
	err := filepath.Walk(fileDir, walkFunc)
	if err != nil {
		return nil, err
	}
	return files, nil
}
func (h *AstHelper) GetFilePath() string {
	return h.FilePath
}
func (h *AstHelper) splitFieldTags(tag string) map[string]string {
	// json:"card_id" gorm:"not null;primary_key;"
	// {"json:"card_id", "gorm:"not null;primary_key;"}
	// {"json:"card_id", "gorm:"not null;primary_key;"}
	// {"json:"card_id", "gorm:"not null;primary_key;"}

	if tag == "" {
		return map[string]string{}
	}
	var res = map[string]string{}
	for i := 0; i <= 10; i++ {
		tag = strings.ReplaceAll(tag, "\"  ", "\" ")
	}
	// 构造为json结构

	tag = fmt.Sprintf("{\"%s}", tag)
	tag = strings.ReplaceAll(tag, "\" ", "\", \"")
	tag = strings.ReplaceAll(tag, ":\"", "\":\"")

	_ = json.Unmarshal([]byte(tag), &res)
	return res
}
