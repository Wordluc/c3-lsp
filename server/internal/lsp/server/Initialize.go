package server

import (
	"os"
	"os/exec"

	"github.com/pherrymason/c3-lsp/pkg/cast"
	"github.com/pherrymason/c3-lsp/pkg/document"
	"github.com/pherrymason/c3-lsp/pkg/fs"
	"github.com/pherrymason/c3-lsp/pkg/utils"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

// Support "Hover"
func (s *Server) Initialize(serverName string, serverVersion string, capabilities protocol.ServerCapabilities, context *glsp.Context, params *protocol.InitializeParams) (any, error) {
	//capabilities := handler.CreateServerCapabilities()

	change := protocol.TextDocumentSyncKindIncremental
	capabilities.TextDocumentSync = protocol.TextDocumentSyncOptions{
		OpenClose: cast.ToPtr(true),
		Change:    &change,
		Save:      cast.ToPtr(true),
	}
	capabilities.DeclarationProvider = true
	capabilities.CompletionProvider = &protocol.CompletionOptions{
		TriggerCharacters: []string{".", ":"},
	}
	capabilities.SignatureHelpProvider = &protocol.SignatureHelpOptions{
		TriggerCharacters:   []string{"(", ","},
		RetriggerCharacters: []string{")"},
	}
	capabilities.Workspace = &protocol.ServerCapabilitiesWorkspace{
		FileOperations: &protocol.ServerCapabilitiesWorkspaceFileOperations{
			DidDelete: &protocol.FileOperationRegistrationOptions{
				Filters: []protocol.FileOperationFilter{{
					Pattern: protocol.FileOperationPattern{
						Glob: "**/*.{c3,c3i}",
					},
				}},
			},
			DidRename: &protocol.FileOperationRegistrationOptions{
				Filters: []protocol.FileOperationFilter{{
					Pattern: protocol.FileOperationPattern{
						Glob: "**/*.{c3,c3i}",
					},
				}},
			},
		},
	}

	if params.RootURI != nil {
		s.state.SetProjectRootURI(utils.NormalizePath(*params.RootURI))
		path, _ := fs.UriToPath(*params.RootURI)
		s.loadServerConfigurationForWorkspace(path)
		s.indexWorkspace()

		s.RunDiagnostics(s.state, context.Notify, false)
	}

	pathTemp, err := os.MkdirTemp(os.TempDir(), "tempGolang")
	if err != nil {
		panic(err)
	}
	// completely arbitrary paths
	oldDir := s.state.GetProjectRootURI() + "/."
	newDir := pathTemp

	cmd := exec.Command("cp", "--recursive", oldDir, newDir)
	println("--------cloning---------")
	println("oldDir ", oldDir)
	println("newDir ", newDir)
	println("--------tempReady---------")
	err = cmd.Run()
	if err != nil {
		panic(err)
	}
	s.tempDir = pathTemp
	if params.Capabilities.TextDocument.PublishDiagnostics.RelatedInformation == nil || *params.Capabilities.TextDocument.PublishDiagnostics.RelatedInformation == false {
		s.options.Diagnostics.Enabled = false
	}

	return protocol.InitializeResult{
		Capabilities: capabilities,
		ServerInfo: &protocol.InitializeResultServerInfo{
			Name:    serverName,
			Version: &serverVersion,
		},
	}, nil
}

func (h *Server) indexWorkspace() {
	path := h.state.GetProjectRootURI()
	files, _ := fs.ScanForC3(fs.GetCanonicalPath(path))

	for _, filePath := range files {
		content, _ := os.ReadFile(filePath)
		doc := document.NewDocumentFromString(filePath, string(content))
		h.state.RefreshDocumentIdentifiers(&doc, h.parser)
	}
}
