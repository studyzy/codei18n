package java

import (
	"context"
	"os"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/java"
	"github.com/studyzy/codei18n/core/domain"
)

// Adapter 实现 Java 语言的 LanguageAdapter 接口
// 使用 Tree-sitter 解析 Java 源代码并提取注释及其上下文
type Adapter struct {
	parser *sitter.Parser
}

// NewAdapter 创建一个新的 Java 适配器实例
// 初始化 Tree-sitter 解析器并加载 Java 语法
func NewAdapter() *Adapter {
	p := sitter.NewParser()
	p.SetLanguage(java.GetLanguage())
	return &Adapter{parser: p}
}

// Language 返回语言标识符 ("java")
func (a *Adapter) Language() string {
	return "java"
}

// Parse 解析提供的 Java 源代码并提取注释
// 返回包含注释文本、位置和上下文符号的 domain.Comment 列表
// 如果 src 为 nil，则从文件路径读取源代码
func (a *Adapter) Parse(file string, src []byte) ([]*domain.Comment, error) {
	if src == nil {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		src = data
	}

	// Parse the source code using Tree-sitter
	tree, err := a.parser.ParseCtx(context.Background(), nil, src)
	if err != nil {
		return nil, err
	}
	defer tree.Close()

	return a.extractComments(tree.RootNode(), src, file)
}

// extractComments 从语法树中提取注释
func (a *Adapter) extractComments(root *sitter.Node, src []byte, file string) ([]*domain.Comment, error) {
	// 首先提取包名
	packageName := extractPackageName(root, src)

	// 创建查询
	q, err := sitter.NewQuery([]byte(javaCommentQuery), java.GetLanguage())
	if err != nil {
		return nil, err
	}
	defer q.Close()

	qc := sitter.NewQueryCursor()
	defer qc.Close()

	qc.Exec(q, root)

	var comments []*domain.Comment

	for {
		m, ok := qc.NextMatch()
		if !ok {
			break
		}

		for _, c := range m.Captures {
			node := c.Node
			text := node.Content(src)

			// 规范化文本并识别注释类型
			cType := domain.CommentTypeLine
			normalized := text

			if strings.HasPrefix(text, "//") {
				cType = domain.CommentTypeLine
				normalized = strings.TrimPrefix(text, "//")
			} else if strings.HasPrefix(text, "/*") {
				if strings.HasPrefix(text, "/**") {
					cType = domain.CommentTypeDoc
					normalized = strings.TrimPrefix(text, "/**")
				} else {
					cType = domain.CommentTypeBlock
					normalized = strings.TrimPrefix(text, "/*")
				}
				normalized = strings.TrimSuffix(normalized, "*/")
			}
			normalized = strings.TrimSpace(normalized)

			// 解析符号路径
			symbol := resolveSymbol(node, src, packageName)

			comment := &domain.Comment{
				File:     file,
				Language: "java",
				Symbol:   symbol,
				Range: domain.TextRange{
					StartLine: int(node.StartPoint().Row) + 1,
					StartCol:  int(node.StartPoint().Column) + 1,
					EndLine:   int(node.EndPoint().Row) + 1,
					EndCol:    int(node.EndPoint().Column) + 1,
				},
				SourceText: text,
				Type:       cType,
			}

			comments = append(comments, comment)
		}
	}

	return comments, nil
}
