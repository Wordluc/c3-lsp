package server

import (
	"log"
	"strconv"
	"strings"

	"github.com/pherrymason/c3-lsp/internal/c3c"
	"github.com/pherrymason/c3-lsp/internal/lsp/project_state"
	"github.com/pherrymason/c3-lsp/pkg/cast"
	"github.com/pherrymason/c3-lsp/pkg/fs"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

func (s *Server) RunDiagnostics(state *project_state.ProjectState, notify glsp.NotifyFunc, delay bool) {
	if !s.options.Diagnostics.Enabled {
		return
	}

	runDiagnostics := func() {
		out, stdErr, err := c3c.CheckC3ErrorsCommand(s.options.C3, s.tempDir)
		log.Println("output:", out.String())
		log.Println("output:", stdErr.String())
		if err == nil {
			s.clearOldDiagnostics(s.state, notify)
			return
		}

		log.Println("Diagnostics report:", err)
		errorsInfo, diagnosticsDisabled := extractErrorDiagnostics(stdErr.String())

		if diagnosticsDisabled {
			s.options.Diagnostics.Enabled = false
			s.clearOldDiagnostics(s.state, notify)
			return
		}
		var newDiagnostics map[string][]protocol.Diagnostic = make(map[string][]protocol.Diagnostic, 0)
		for _, errInfo := range errorsInfo {
			lenTemp := len(s.tempDir)
			errInfo.File = s.state.GetProjectRootURI() + "/" + errInfo.File[lenTemp:]
			newDiagnostics[errInfo.File] = append(newDiagnostics[errInfo.File], errInfo.Diagnostic)
			state.SetDocumentDiagnostics(errInfo.File, newDiagnostics[errInfo.File])
		}
		//TODO see if can be improved
		for key, value := range newDiagnostics {
			if len(value) == 0 {
				continue
			}
			notify(
				protocol.ServerTextDocumentPublishDiagnostics,
				protocol.PublishDiagnosticsParams{
					URI:         fs.ConvertPathToURI(key, s.options.C3.StdlibPath),
					Diagnostics: value,
				})
		}
	}

	if delay {
		s.diagnosticDebounced(runDiagnostics)
	} else {
		runDiagnostics()
	}
}

type ErrorInfo struct {
	File       string
	Diagnostic protocol.Diagnostic
}

func extractErrorDiagnostics(output string) ([]ErrorInfo, bool) {
	errorsInfo := []ErrorInfo{}
	diagnosticsDisabled := false

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		// > LSPERR|error|"/<path>>/test.c3"|13|47|"Expected ';'"
		if strings.HasPrefix(line, "> LSPERR") {
			parts := strings.Split(line, "|")
			if parts[1] == "error" {
				if len(parts) != 6 {
					// Disable future diagnostics, looks like c3c is an old version.
					diagnosticsDisabled = true
				} else {
					errorLine, err := strconv.Atoi(parts[3])
					if err != nil {
						continue
					}
					errorLine -= 1
					character, err := strconv.Atoi(parts[4])
					if err != nil {
						continue
					}
					character -= 1

					errorsInfo = append(errorsInfo, ErrorInfo{
						File: strings.Trim(parts[2], `"`),
						Diagnostic: protocol.Diagnostic{
							Range: protocol.Range{
								Start: protocol.Position{Line: protocol.UInteger(errorLine), Character: protocol.UInteger(character)},
								End:   protocol.Position{Line: protocol.UInteger(errorLine), Character: protocol.UInteger(99)},
							},
							Severity: cast.ToPtr(protocol.DiagnosticSeverityError),
							Source:   cast.ToPtr("c3c build --lsp"),
							Message:  parts[5],
						},
					})
				}
			}
			break
		}
	}

	return errorsInfo, diagnosticsDisabled
}

func (s *Server) clearOldDiagnostics(state *project_state.ProjectState, notify glsp.NotifyFunc) {
	for k := range state.GetDocumentDiagnostics() {
		go notify(protocol.ServerTextDocumentPublishDiagnostics,
			protocol.PublishDiagnosticsParams{
				URI:         fs.ConvertPathToURI(k, s.options.C3.StdlibPath),
				Diagnostics: []protocol.Diagnostic{},
			})
	}
	state.ClearDocumentDiagnostics()
}

func hasDiagnosticForFile(file string, errorsInfo []ErrorInfo) bool {
	for _, v := range errorsInfo {
		if file == v.File {
			return true
		}
	}

	return false
}
