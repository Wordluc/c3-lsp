package context

import (
	"github.com/pherrymason/c3-lsp/internal/lsp/project_state"
	"github.com/pherrymason/c3-lsp/pkg/symbols"
	sitter "github.com/smacker/go-tree-sitter"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

type CursorContext struct {
	Position symbols.Position
	DocURI   protocol.DocumentUri

	IsLiteral          bool
	IsIdentifier       bool
	IsModuleIdentifier bool
}

func BuildFromDocumentPosition(
	position protocol.Position,
	docURI protocol.DocumentUri,
	state *project_state.ProjectState,
) CursorContext {
	context := CursorContext{
		Position: symbols.NewPositionFromLSPPosition(position),
		DocURI:   docURI,
	}

	doc := state.GetDocument(docURI)
	tree := doc.ContextSyntaxTree
	root := tree.RootNode()

	// Search sitter.Node where cursor is currently
	node := root.NamedDescendantForPointRange(
		sitter.Point{Row: uint32(position.Line), Column: uint32(position.Character)},
		sitter.Point{Row: uint32(position.Line), Column: uint32(position.Character + 1)},
	)

	if node == nil {
		// Could not find node in document.
		return context
	}

	//s := fmt.Sprintf("Node found. Type: %s. Content: %s", node.Type(), node.Content([]byte(doc.SourceCode.Text)))
	//could it be done better?
	var findTypeIdent func(*sitter.Node) *sitter.Node
	findTypeIdent = func(n *sitter.Node) *sitter.Node {
		if n.Type() == "type_ident" {
			return n
		}
		for i := 0; i < int(n.ChildCount()); i++ {
			result := findTypeIdent(n.Child(i))
			if result != nil {
				return result
			}
		}
		return nil
	}
	switch node.Type() {
	case "integer_literal":
		context.IsLiteral = true
	case "real_literal":
		context.IsLiteral = true
	case "char_literal":
		context.IsLiteral = true
	case "string_literal":
		context.IsLiteral = true
	case "raw_string_literal":
		context.IsLiteral = true
	case "string_expr":
		context.IsLiteral = true
	case "bytes_expr":
		context.IsLiteral = true
	case "ident":
		context.IsIdentifier = true
		if node.Parent().Type() == "module_resolution" {
			context.IsModuleIdentifier = true
		}
	case "initializer_list":
		//	initializerList := node.Parent().Parent()
		//	typeNode := findTypeIdent(initializerList)
		//	typeValue := (typeNode.Content([]byte(doc.SourceCode.Text)))
		//I have the type of struct how i use it?

	}

	return context
}
