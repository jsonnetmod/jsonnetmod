package modfile

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
)

func LoadModFile(dir string, m *ModFile) (bool, error) {
	f := filepath.Join(dir, ModFilename)

	if m.Comments == nil {
		m.Comments = map[string][]string{}
	}

	if m.Replace == nil {
		m.Replace = map[PathIdentity]PathIdentity{}
	}

	if m.Require == nil {
		m.Require = map[string]Require{}
	}

	if data, err := ioutil.ReadFile(f); err != nil {
		if !os.IsNotExist(err) {
			return false, err
		}
	} else {
		node, err := jsonnet.SnippetToAST(f, string(data))
		if err != nil {
			return false, err
		}

		if o, ok := node.(*ast.DesugaredObject); ok {
			for _, f := range o.Fields {
				fieldName := f.Name.(*ast.LiteralString).Value

				switch fieldName {
				case "module":
					if v, ok := f.Body.(*ast.LiteralString); ok {
						m.Module = v.Value
					} else {
						return false, fmt.Errorf("%s must be a string value of %s", fieldName, ModFilename)
					}
				case "jpath":
					if v, ok := f.Body.(*ast.LiteralString); ok {
						m.JPath = v.Value
					} else {
						return false, fmt.Errorf("%s must be a string value of %s", fieldName, ModFilename)
					}
				case "replace":
					if v, ok := f.Body.(*ast.DesugaredObject); ok {
						if err := rangeField(v, func(name string, value string, hidden bool, f ast.DesugaredObjectField) error {
							from, err := ParsePathIdentity(name)
							if err != nil {
								return nil
							}

							to, err := ParsePathIdentity(value)
							if err != nil {
								return nil
							}

							if to.Path == "" {
								to.Path = from.Path
							}

							m.Replace[*from] = *to
							m.Comments["replace:"+name] = pickNodeComponents(f.Name)

							return nil
						}); err != nil {
							return false, err
						}
					} else {
						return false, fmt.Errorf("%s must be a object of %s", fieldName, ModFilename)
					}
				case "require":
					if v, ok := f.Body.(*ast.DesugaredObject); ok {
						if err := rangeField(v, func(name string, value string, hidden bool, f ast.DesugaredObjectField) error {
							m.Require[name] = Require{
								Version:  value,
								Indirect: hidden,
							}
							m.Comments["require:"+name] = pickNodeComponents(f.Name)
							return nil
						}); err != nil {
							return false, err
						}
					} else {
						return false, fmt.Errorf("%s must be a object of %s", fieldName, ModFilename)
					}
				}
			}
		} else {
			return false, fmt.Errorf("invalid %s", ModFilename)
		}

		return true, nil
	}

	return false, nil
}

func pickNodeComponents(node ast.Node) []string {
	comments := make([]string, 0)

	if f := node.OpenFodder(); f != nil {
		for _, fe := range *f {
			comments = append(comments, fe.Comment...)
		}
	}

	return comments
}

func rangeField(o *ast.DesugaredObject, each func(name string, value string, hidden bool, f ast.DesugaredObjectField) error) error {
	for _, f := range o.Fields {
		key, ok := f.Name.(*ast.LiteralString)
		if !ok {
			return fmt.Errorf("%s should be a string", f.Name)
		}

		value, ok := f.Body.(*ast.LiteralString)
		if !ok {
			return fmt.Errorf("%s should be a string", f.Body)
		}
		if err := each(key.Value, value.Value, f.Hide == ast.ObjectFieldHidden, f); err != nil {
			return err
		}
	}
	return nil
}
