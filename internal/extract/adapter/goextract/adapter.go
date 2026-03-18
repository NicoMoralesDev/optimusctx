package goextract

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"go/ast"
	"go/parser"
	"go/scanner"
	"go/token"
	"strconv"
	"strings"

	"github.com/niccrow/optimusctx/internal/extract"
	"github.com/niccrow/optimusctx/internal/repository"
)

const (
	adapterName    = "tree-sitter-go"
	grammarVersion = "v0.25.0"
)

type Adapter struct{}

func New() *Adapter {
	return &Adapter{}
}

func (a *Adapter) Name() string {
	return adapterName
}

func (a *Adapter) Language() string {
	return "go"
}

func (a *Adapter) GrammarVersion() string {
	return grammarVersion
}

func (a *Adapter) Extract(ctx context.Context, req extract.Request) (extract.Result, error) {
	select {
	case <-ctx.Done():
		return extract.Result{}, ctx.Err()
	default:
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, req.Candidate.Path, req.Content, parser.AllErrors)
	parserErrorCount := int64(countParseErrors(err))
	result := extract.Result{
		ParserErrorCount: parserErrorCount,
		HasErrorNodes:    parserErrorCount > 0,
	}
	if file == nil {
		result.CoverageState = repository.ExtractionCoverageStateFailed
		result.CoverageReason = repository.ExtractionCoverageReasonParseError
		return result, nil
	}

	state := newWalkState(fset, file, req.Content)
	state.collectFile(file)
	result.Symbols = state.symbols
	meaningfulSymbols := countMeaningfulSymbols(state.symbols)

	switch {
	case result.HasErrorNodes && meaningfulSymbols == 0:
		result.Symbols = nil
		result.CoverageState = repository.ExtractionCoverageStateFailed
		result.CoverageReason = repository.ExtractionCoverageReasonParseError
	case result.HasErrorNodes:
		result.CoverageState = repository.ExtractionCoverageStatePartial
		result.CoverageReason = repository.ExtractionCoverageReasonParseError
	default:
		result.CoverageState = repository.ExtractionCoverageStateSupported
	}

	return result, nil
}

type walkState struct {
	source               []byte
	locator              locator
	symbols              []repository.SymbolRecord
	typeStableKeysByName map[string]string
}

func newWalkState(fset *token.FileSet, file *ast.File, source []byte) walkState {
	state := walkState{
		source:               source,
		locator:              newLocator(fset, file, source),
		typeStableKeysByName: make(map[string]string),
	}
	state.collectTypeStableKeys(file)
	return state
}

func (s *walkState) collectTypeStableKeys(file *ast.File) {
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}
		for _, spec := range gen.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok || typeSpec.Name == nil {
				continue
			}
			name := typeSpec.Name.Name
			if name == "" {
				continue
			}
			specSpan := s.locator.span(typeSpec.Pos(), typeSpec.End())
			s.typeStableKeysByName[name] = stableKey("type", name, specSpan.StartByte, specSpan.EndByte)
		}
	}
}

func (s *walkState) collectFile(file *ast.File) {
	if symbol, ok := s.packageSymbol(file); ok {
		s.symbols = append(s.symbols, symbol)
	}

	for _, decl := range file.Decls {
		switch node := decl.(type) {
		case *ast.GenDecl:
			s.collectGenDecl(node)
		case *ast.FuncDecl:
			if node.Recv == nil {
				if symbol, ok := s.functionSymbol(node); ok {
					s.symbols = append(s.symbols, symbol)
				}
				continue
			}
			if symbol, ok := s.methodSymbol(node); ok {
				s.symbols = append(s.symbols, symbol)
			}
		}
	}
}

func (s *walkState) collectGenDecl(node *ast.GenDecl) {
	switch node.Tok {
	case token.CONST:
		for _, spec := range node.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			s.symbols = append(s.symbols, s.valueSymbols(valueSpec, "const", "")...)
		}
	case token.VAR:
		for _, spec := range node.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			s.symbols = append(s.symbols, s.valueSymbols(valueSpec, "var", "")...)
		}
	case token.TYPE:
		for _, spec := range node.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			symbols, nestedParent, ok := s.typeSymbols(typeSpec, "")
			if !ok {
				continue
			}
			s.symbols = append(s.symbols, symbols...)
			switch typed := typeSpec.Type.(type) {
			case *ast.StructType:
				s.symbols = append(s.symbols, s.fieldSymbols(typed.Fields, nestedParent)...)
			case *ast.InterfaceType:
				s.symbols = append(s.symbols, s.interfaceMethodSymbols(typed.Methods, nestedParent)...)
			}
		}
	}
}

func (s *walkState) packageSymbol(node *ast.File) (repository.SymbolRecord, bool) {
	if node == nil || node.Name == nil || node.Name.Name == "" {
		return repository.SymbolRecord{}, false
	}
	return s.newSymbol(
		"package",
		node.Name.Name,
		s.locator.span(node.Package, node.Name.End()),
		s.locator.span(node.Name.Pos(), node.Name.End()),
		"",
		"",
		token.NoPos,
	), true
}

func (s *walkState) valueSymbols(node *ast.ValueSpec, kind string, parentStableKey string) []repository.SymbolRecord {
	if node == nil {
		return nil
	}

	symbols := make([]repository.SymbolRecord, 0, len(node.Names))
	for _, nameNode := range node.Names {
		if nameNode == nil || nameNode.Name == "" {
			continue
		}
		nameSpan := s.locator.span(nameNode.Pos(), nameNode.End())
		symbols = append(symbols, s.newSymbol(kind, nameNode.Name, nameSpan, nameSpan, parentStableKey, nameNode.Name, token.NoPos))
	}
	return symbols
}

func (s *walkState) typeSymbols(node *ast.TypeSpec, parentStableKey string) ([]repository.SymbolRecord, string, bool) {
	if node == nil || node.Name == nil || node.Name.Name == "" {
		return nil, "", false
	}

	name := node.Name.Name
	typeSymbol := s.newSymbol(
		"type",
		name,
		s.locator.span(node.Pos(), node.End()),
		s.locator.span(node.Name.Pos(), node.Name.End()),
		parentStableKey,
		name,
		token.NoPos,
	)

	switch typeNode := node.Type.(type) {
	case *ast.StructType:
		structSymbol := s.newSymbol(
			"struct",
			name,
			s.locator.span(typeNode.Pos(), typeNode.End()),
			s.locator.span(node.Name.Pos(), node.Name.End()),
			typeSymbol.StableKey,
			name,
			token.NoPos,
		)
		return []repository.SymbolRecord{typeSymbol, structSymbol}, structSymbol.StableKey, true
	case *ast.InterfaceType:
		interfaceSymbol := s.newSymbol(
			"interface",
			name,
			s.locator.span(typeNode.Pos(), typeNode.End()),
			s.locator.span(node.Name.Pos(), node.Name.End()),
			typeSymbol.StableKey,
			name,
			token.NoPos,
		)
		return []repository.SymbolRecord{typeSymbol, interfaceSymbol}, interfaceSymbol.StableKey, true
	default:
		return []repository.SymbolRecord{typeSymbol}, "", true
	}
}

func (s *walkState) fieldSymbols(fields *ast.FieldList, parentStableKey string) []repository.SymbolRecord {
	if fields == nil || parentStableKey == "" {
		return nil
	}

	symbols := make([]repository.SymbolRecord, 0, len(fields.List))
	for _, field := range fields.List {
		if field == nil {
			continue
		}
		fieldSpan := s.locator.span(field.Pos(), field.End())
		if len(field.Names) == 0 {
			name := normalizeTypeName(s.locator.sourceText(field.Type.Pos(), field.Type.End()))
			if name == "" {
				continue
			}
			nameSpan := s.locator.span(field.Type.Pos(), field.Type.End())
			symbols = append(symbols, s.newSymbol("field", name, fieldSpan, nameSpan, parentStableKey, name, token.NoPos))
			continue
		}

		for _, nameNode := range field.Names {
			if nameNode == nil || nameNode.Name == "" {
				continue
			}
			symbols = append(symbols, s.newSymbol(
				"field",
				nameNode.Name,
				fieldSpan,
				s.locator.span(nameNode.Pos(), nameNode.End()),
				parentStableKey,
				nameNode.Name,
				token.NoPos,
			))
		}
	}
	return symbols
}

func (s *walkState) interfaceMethodSymbols(methods *ast.FieldList, parentStableKey string) []repository.SymbolRecord {
	if methods == nil || parentStableKey == "" {
		return nil
	}

	symbols := make([]repository.SymbolRecord, 0, len(methods.List))
	for _, method := range methods.List {
		if method == nil || len(method.Names) == 0 {
			continue
		}
		methodSpan := s.locator.span(method.Pos(), method.End())
		for _, nameNode := range method.Names {
			if nameNode == nil || nameNode.Name == "" {
				continue
			}
			symbols = append(symbols, s.newSymbol(
				"method",
				nameNode.Name,
				methodSpan,
				s.locator.span(nameNode.Pos(), nameNode.End()),
				parentStableKey,
				nameNode.Name,
				token.NoPos,
			))
		}
	}
	return symbols
}

func (s *walkState) functionSymbol(node *ast.FuncDecl) (repository.SymbolRecord, bool) {
	if node == nil || node.Name == nil || node.Name.Name == "" || node.Body == nil {
		return repository.SymbolRecord{}, false
	}
	return s.newSymbol(
		"function",
		node.Name.Name,
		s.locator.span(node.Pos(), node.End()),
		s.locator.span(node.Name.Pos(), node.Name.End()),
		"",
		node.Name.Name,
		bodyPos(node.Body),
	), true
}

func (s *walkState) methodSymbol(node *ast.FuncDecl) (repository.SymbolRecord, bool) {
	if node == nil || node.Name == nil || node.Name.Name == "" || node.Recv == nil || node.Body == nil {
		return repository.SymbolRecord{}, false
	}

	name := node.Name.Name
	receiverType := receiverTypeName(node.Recv, s.locator)
	qualifiedName := name
	parentStableKey := ""
	if receiverType != "" {
		qualifiedName = receiverType + "." + name
		parentStableKey = s.typeStableKeysByName[receiverType]
	}

	return s.newSymbol(
		"method",
		name,
		s.locator.span(node.Pos(), node.End()),
		s.locator.span(node.Name.Pos(), node.Name.End()),
		parentStableKey,
		qualifiedName,
		bodyPos(node.Body),
	), true
}

func (s *walkState) newSymbol(kind string, name string, span sourceSpan, nameSpan sourceSpan, parentStableKey string, qualifiedName string, bodyStart token.Pos) repository.SymbolRecord {
	symbol := repository.SymbolRecord{
		StableKey:       stableKey(kind, qualifiedName, span.StartByte, span.EndByte),
		ParentStableKey: parentStableKey,
		Kind:            kind,
		Name:            name,
		QualifiedName:   qualifiedName,
		IsExported:      isExported(name),
	}

	symbol.StartByte = span.StartByte
	symbol.EndByte = span.EndByte
	symbol.StartRow = span.StartRow
	symbol.StartColumn = span.StartColumn
	symbol.EndRow = span.EndRow
	symbol.EndColumn = span.EndColumn
	symbol.NameStartByte = nameSpan.StartByte
	symbol.NameEndByte = nameSpan.EndByte

	if parentStableKey == "" {
		symbol.Depth = 0
	} else {
		symbol.Depth = 1
	}

	if bodyStart.IsValid() {
		symbol.SignatureStartByte = symbol.StartByte
		symbol.SignatureEndByte = s.locator.offset(bodyStart)
	}

	return symbol
}

func stableKey(kind string, qualifiedName string, startByte int64, endByte int64) string {
	payload := kind + ":" + qualifiedName + ":" + formatInt(startByte) + ":" + formatInt(endByte)
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}

func countParseErrors(err error) int {
	if err == nil {
		return 0
	}

	switch typed := err.(type) {
	case scanner.ErrorList:
		return len(typed)
	case *scanner.ErrorList:
		return len(*typed)
	default:
		return 1
	}
}

func receiverTypeName(receivers *ast.FieldList, locator locator) string {
	if receivers == nil || len(receivers.List) == 0 {
		return ""
	}

	field := receivers.List[0]
	if field == nil || field.Type == nil {
		return ""
	}
	return normalizeTypeName(locator.sourceText(field.Type.Pos(), field.Type.End()))
}

func countMeaningfulSymbols(symbols []repository.SymbolRecord) int {
	count := 0
	for _, symbol := range symbols {
		if symbol.Kind == "package" {
			continue
		}
		count++
	}
	return count
}

func normalizeTypeName(raw string) string {
	cleaned := strings.TrimSpace(raw)
	cleaned = strings.Trim(cleaned, "()")
	cleaned = strings.TrimPrefix(cleaned, "*")
	cleaned = strings.TrimPrefix(cleaned, "[]")
	cleaned = strings.TrimPrefix(cleaned, "...")
	if idx := strings.Index(cleaned, "["); idx >= 0 {
		cleaned = cleaned[:idx]
	}
	if idx := strings.LastIndex(cleaned, "."); idx >= 0 {
		cleaned = cleaned[idx+1:]
	}
	return strings.TrimSpace(cleaned)
}

func isExported(name string) bool {
	if name == "" {
		return false
	}
	first := rune(name[0])
	return strings.ToUpper(string(first)) == string(first) && strings.ToLower(string(first)) != string(first)
}

type locator struct {
	file   *token.File
	source []byte
}

type sourceSpan struct {
	StartByte   int64
	EndByte     int64
	StartRow    int64
	StartColumn int64
	EndRow      int64
	EndColumn   int64
}

func newLocator(fset *token.FileSet, file *ast.File, source []byte) locator {
	if file == nil {
		return locator{source: source}
	}

	tokenFile := fset.File(file.Package)
	if tokenFile == nil {
		tokenFile = fset.File(file.Pos())
	}
	return locator{file: tokenFile, source: source}
}

func (l locator) span(start token.Pos, end token.Pos) sourceSpan {
	if l.file == nil || !start.IsValid() || !end.IsValid() {
		return sourceSpan{}
	}

	startPos := l.file.Position(start)
	endPos := l.file.Position(end)
	return sourceSpan{
		StartByte:   l.offset(start),
		EndByte:     l.offset(end),
		StartRow:    int64(startPos.Line - 1),
		StartColumn: int64(startPos.Column - 1),
		EndRow:      int64(endPos.Line - 1),
		EndColumn:   int64(endPos.Column - 1),
	}
}

func (l locator) offset(pos token.Pos) int64 {
	if l.file == nil || !pos.IsValid() {
		return 0
	}

	offset := l.file.Offset(pos)
	switch {
	case offset < 0:
		return 0
	case offset > len(l.source):
		return int64(len(l.source))
	default:
		return int64(offset)
	}
}

func (l locator) sourceText(start token.Pos, end token.Pos) string {
	startOffset := l.offset(start)
	endOffset := l.offset(end)
	if endOffset < startOffset {
		return ""
	}
	return string(l.source[startOffset:endOffset])
}

func bodyPos(block *ast.BlockStmt) token.Pos {
	if block == nil {
		return token.NoPos
	}
	return block.Pos()
}

func formatInt(value int64) string {
	return strconv.FormatInt(value, 10)
}
