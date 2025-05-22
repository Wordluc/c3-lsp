package server

import (
	"os"

	"github.com/pherrymason/c3-lsp/pkg/document"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func (h *Server) TextDocumentDidOpen(context *glsp.Context, params *protocol.DidOpenTextDocumentParams) error {
	/*
		doc, err := h.documents.Open(*params, context.Notify)
		if err != nil {
			//glspServer.Log.Debug("Could not open file document.")
			return err
		}

		if doc != nil {
			h.state.RefreshDocumentIdentifiers(doc, h.parser)
		}
	*/

	langID := params.TextDocument.LanguageID
	if langID != "c3" {
		return nil
	}

	doc := document.NewDocumentFromDocURI(params.TextDocument.URI, params.TextDocument.Text, params.TextDocument.Version)
	h.state.RefreshDocumentIdentifiers(doc, h.parser)
	lenRoot := len(h.state.GetProjectRootURI())
	pathFile := doc.URI[lenRoot+1:]
	if err := os.WriteFile(h.tempDir+"/"+pathFile, []byte(doc.SourceCode.Text), 0644); err != nil {
		return err
	}

	return nil
}
