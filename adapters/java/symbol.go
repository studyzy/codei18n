package java

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

// resolveSymbol 通过查找注释节点的下一个命名兄弟节点来确定符号路径
// 对于 Java，符号路径格式为：package.ClassName#methodName 或 package.ClassName#fieldName
func resolveSymbol(node *sitter.Node, src []byte, packageName string) string {
	// 查找下一个命名兄弟节点以附加注释
	next := node.NextNamedSibling()
	if next == nil {
		// 如果没有下一个兄弟节点，可能是文件级注释
		return ""
	}

	return getSymbolPath(next, src, packageName)
}

// getSymbolPath 构建给定节点的完整符号路径
func getSymbolPath(node *sitter.Node, src []byte, packageName string) string {
	// 构建符号路径：从当前节点向上遍历，收集类名、方法名等
	var parts []string
	var memberName string
	var isMember bool

	current := node
	for current != nil {
		switch current.Type() {
		case "class_declaration", "interface_declaration", "enum_declaration":
			// 顶级类型声明
			name := getChildContent(current, "name", src)
			if name != "" {
				parts = append([]string{name}, parts...)
			}
		case "method_declaration", "constructor_declaration":
			// 方法或构造器
			if current == node {
				memberName = getChildContent(current, "name", src)
				isMember = true
			}
		case "field_declaration":
			// 字段声明
			if current == node {
				// 字段声明可能包含多个变量声明器
				declarator := current.ChildByFieldName("declarator")
				if declarator != nil && declarator.Type() == "variable_declarator" {
					memberName = getChildContent(declarator, "name", src)
					isMember = true
				}
			}
		}
		current = current.Parent()
	}

	// 构建最终路径
	var result []string
	if packageName != "" {
		result = append(result, packageName)
	}
	result = append(result, parts...)

	// 如果是成员（方法或字段），使用 # 连接
	if isMember && memberName != "" {
		if len(result) > 0 {
			result[len(result)-1] = result[len(result)-1] + "#" + memberName
		} else {
			result = append(result, memberName)
		}
	}

	return strings.Join(result, ".")
}

// getChildContent 获取节点指定字段的内容
func getChildContent(node *sitter.Node, fieldName string, src []byte) string {
	child := node.ChildByFieldName(fieldName)
	if child != nil {
		return child.Content(src)
	}
	return ""
}

// extractPackageName 从根节点提取包名
func extractPackageName(root *sitter.Node, src []byte) string {
	for i := uint32(0); i < root.NamedChildCount(); i++ {
		child := root.NamedChild(int(i))
		if child.Type() == "package_declaration" {
			// package_declaration 包含一个 scoped_identifier 或 identifier
			for j := uint32(0); j < child.NamedChildCount(); j++ {
				nameNode := child.NamedChild(int(j))
				if nameNode.Type() == "scoped_identifier" || nameNode.Type() == "identifier" {
					return nameNode.Content(src)
				}
			}
		}
	}
	return ""
}
