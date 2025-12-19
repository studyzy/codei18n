package typescript

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

// resolveSymbol determines the symbol path for a comment node by looking at its following sibling.
func resolveSymbol(node *sitter.Node, src []byte) string {
	// Look for the next named sibling to attach the comment to
	next := node.NextNamedSibling()
	if next == nil {
		return ""
	}

	return getSymbolPath(next, src)
}

func getSymbolPath(node *sitter.Node, src []byte) string {
	parts := []string{}
	current := node

	// Build path from current node up to root, but here we usually just want to know
	// "what is this node?" and maybe its parent context if it's a method.
	// For MVP, let's implement basic identification of the immediate node.
	// Advanced path building (e.g. Class.Method) usually requires traversing UP parents.

	// Strategy:
	// 1. Identify the name of the current node.
	// 2. Traverse parents to prepend context (Class, Module, Interface).

	name := getNodeName(current, src)
	if name != "" {
		parts = append(parts, name)
	}

	// Traverse up to find parents
	parent := current.Parent()
	for parent != nil {
		parentName := ""
		switch parent.Type() {
		case "class_declaration", "interface_declaration", "module_declaration", "class":
			parentName = getNodeName(parent, src)
		}

		if parentName != "" {
			// Prepend parent name
			parts = append([]string{parentName}, parts...)
		}
		parent = parent.Parent()
	}

	return strings.Join(parts, ".")
}

func getNodeName(node *sitter.Node, src []byte) string {
	switch node.Type() {
	case "function_declaration", "generator_function_declaration":
		// (function_declaration name: (identifier) @name)
		return getChildContent(node, "name", src)

	case "method_definition":
		// (method_definition name: (property_identifier) @name)
		return getChildContent(node, "name", src)

	case "class_declaration", "class":
		// (class_declaration name: (identifier) @name)
		return getChildContent(node, "name", src)

	case "interface_declaration":
		// (interface_declaration name: (type_identifier) @name)
		return getChildContent(node, "name", src)

	case "type_alias_declaration":
		// (type_alias_declaration name: (type_identifier) @name)
		return getChildContent(node, "name", src)

	case "variable_declarator":
		// (variable_declarator name: (identifier) @name value: (arrow_function))
		// Check if it's an arrow function or function expression
		val := node.ChildByFieldName("value")
		if val != nil && (val.Type() == "arrow_function" || val.Type() == "function_expression") {
			return getChildContent(node, "name", src)
		}
		// Also handle simple variable assignments if needed, but usually we care about functions/classes
		return getChildContent(node, "name", src)

	case "lexical_declaration":
		// const x = ...; lexical_declaration contains variable_declarator
		// We might need to look inside if the comment is above 'const'
		if node.NamedChildCount() > 0 {
			firstChild := node.NamedChild(0)
			if firstChild.Type() == "variable_declarator" {
				return getNodeName(firstChild, src)
			}
		}

	case "export_statement":
		// export const x = ...; or export function ...
		// We need to look at the declaration being exported
		if node.NamedChildCount() > 0 {
			decl := node.NamedChild(0) // declaration
			return getNodeName(decl, src)
		}
	}

	return ""
}

func getChildContent(node *sitter.Node, fieldName string, src []byte) string {
	child := node.ChildByFieldName(fieldName)
	if child != nil {
		return child.Content(src)
	}
	return ""
}
