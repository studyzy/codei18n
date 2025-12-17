package rust

import (
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

// ResolveSymbolPath resolves the semantic symbol path for a given node by traversing up the syntax tree.
// It constructs a path string like "mod_name::StructName::method_name".
// Supported nodes: function_item, struct_item, enum_item, trait_item, mod_item, impl_item.
func ResolveSymbolPath(node *sitter.Node, src []byte) string {
	var parts []string
	curr := node

	for curr != nil {
		name := ""

		switch curr.Type() {
		case "function_item":
			name = getChildContent(curr, "name", src)
		case "struct_item":
			name = getChildContent(curr, "name", src)
		case "enum_item":
			name = getChildContent(curr, "name", src)
		case "trait_item":
			name = getChildContent(curr, "name", src)
		case "mod_item":
			name = getChildContent(curr, "name", src)
		case "impl_item":
			// Handle impl blocks: impl Foo or impl Bar for Foo
			typeName := getChildContent(curr, "type", src)
			traitName := getChildContent(curr, "trait", src)
			if traitName != "" {
				name = fmt.Sprintf("impl<%s for %s>", traitName, typeName)
			} else if typeName != "" {
				name = fmt.Sprintf("impl<%s>", typeName)
			}
		}

		if name != "" {
			parts = append([]string{name}, parts...) // Prepend
		}
		curr = curr.Parent()
	}

	return strings.Join(parts, "::")
}

func getChildContent(node *sitter.Node, fieldName string, src []byte) string {
	child := node.ChildByFieldName(fieldName)
	if child != nil {
		return child.Content(src)
	}
	return ""
}
