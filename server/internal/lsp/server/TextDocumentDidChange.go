package server

import (
	"fmt"
	"os"

	"github.com/pherrymason/c3-lsp/pkg/utils"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func (s *Server) TextDocumentDidChange(context *glsp.Context, params *protocol.DidChangeTextDocumentParams) error {
	s.state.UpdateDocument(params.TextDocument.URI, params.ContentChanges, s.parser)

	doc := s.state.GetDocument(utils.NormalizePath(params.TextDocument.URI))
	if doc == nil {
		return fmt.Errorf("No file %v found to write in temp directory", doc.URI)
	}
	lenRoot := len(s.state.GetProjectRootURI())
	pathFile := doc.URI[lenRoot+1:]
	if err := os.WriteFile(s.tempDir+"/"+pathFile, []byte(doc.SourceCode.Text), os.ModeTemporary); err != nil {
		return err
	}
	s.RunDiagnostics(s.state, context.Notify, true)

	return nil
}
