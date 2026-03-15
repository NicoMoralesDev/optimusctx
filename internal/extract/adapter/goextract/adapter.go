package goextract

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	sitter "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_go "github.com/tree-sitter/tree-sitter-go/bindings/go"

	"github.com/niccrow/optimusctx/internal/extract"
	"github.com/niccrow/optimusctx/internal/repository"
)

const (
	adapterName    = "tree-sitter-go"
	grammarVersion = "v0.25.0"
)

var language = sitter.NewLanguage(tree_sitter_go.Language())

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
	parser := sitter.NewParser()
	defer parser.Close()

	if err := parser.SetLanguage(language); err != nil {
		return extract.Result{}, fmt.Errorf("set language: %w", err)
	}

	tree := parser.ParseCtx(ctx, req.Content, nil)
	if tree == nil {
		return extract.Result{
			CoverageState:  repository.ExtractionCoverageStateFailed,
			CoverageReason: repository.ExtractionCoverageReasonAdapterError,
		}, nil
	}
	defer tree.Close()

	root := tree.RootNode()
	state := walkState{
		source:               req.Content,
		cursor:               root.Walk(),
		typeStableKeysByName: make(map[string]string),
	}
	defer state.cursor.Close()

	state.collect(root, "")

	parserErrorCount := int64(countErrorNodes(root))
	result := extract.Result{
		ParserErrorCount: parserErrorCount,
		HasErrorNodes:    root.HasError(),
		Symbols:          state.symbols,
	}

	switch {
	case root.HasError() && len(state.symbols) == 0:
		result.CoverageState = repository.ExtractionCoverageStateFailed
		result.CoverageReason = repository.ExtractionCoverageReasonParseError
	case root.HasError():
		result.CoverageState = repository.ExtractionCoverageStatePartial
		result.CoverageReason = repository.ExtractionCoverageReasonParseError
	default:
		result.CoverageState = repository.ExtractionCoverageStateSupported
	}

	return result, nil
}

type walkState struct {
	source               []byte
	cursor               *sitter.TreeCursor
	symbols              []repository.SymbolRecord
	typeStableKeysByName map[string]string
}

func (s *walkState) collect(node *sitter.Node, lexicalParent string) {
	if node == nil {
		return
	}

	switch node.Kind() {
	case "package_clause":
		if symbol, ok := s.packageSymbol(node); ok {
			s.symbols = append(s.symbols, symbol)
		}
	case "const_spec":
		s.symbols = append(s.symbols, s.valueSymbols(node, "const", lexicalParent)...)
	case "var_spec":
		s.symbols = append(s.symbols, s.valueSymbols(node, "var", lexicalParent)...)
	case "type_spec":
		symbol, nestedParent, ok := s.typeSymbols(node, lexicalParent)
		if ok {
			s.symbols = append(s.symbols, symbol...)
			lexicalParent = nestedParent
		}
	case "field_declaration":
		s.symbols = append(s.symbols, s.fieldSymbols(node, lexicalParent)...)
	case "method_elem":
		if symbol, ok := s.interfaceMethodSymbol(node, lexicalParent); ok {
			s.symbols = append(s.symbols, symbol)
		}
	case "function_declaration":
		if symbol, ok := s.functionSymbol(node); ok {
			s.symbols = append(s.symbols, symbol)
		}
	case "method_declaration":
		if symbol, ok := s.methodSymbol(node); ok {
			s.symbols = append(s.symbols, symbol)
		}
	}

	for _, child := range node.NamedChildren(s.cursor) {
		if !s.shouldTraverse(node.Kind(), child.Kind()) {
			continue
		}

		childParent := lexicalParent
		if node.Kind() == "type_spec" {
			switch child.Kind() {
			case "struct_type", "interface_type":
				if container := s.lastSymbolStableKey(); container != "" {
					childParent = container
				}
			}
		}
		s.collect(&child, childParent)
	}
}

func (s *walkState) shouldTraverse(parentKind string, childKind string) bool {
	switch parentKind {
	case "const_spec", "var_spec", "package_clause", "function_declaration", "method_declaration", "field_declaration", "method_elem":
		return false
	case "type_spec":
		return childKind == "struct_type" || childKind == "interface_type"
	default:
		return true
	}
}

func (s *walkState) packageSymbol(node *sitter.Node) (repository.SymbolRecord, bool) {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		nameNode = firstNamedChildOfKind(node, s.cursor, "package_identifier")
	}
	if nameNode == nil || node.HasError() {
		return repository.SymbolRecord{}, false
	}
	return s.newSymbol("package", nameNode.Utf8Text(s.source), node, nameNode, "", ""), true
}

func (s *walkState) valueSymbols(node *sitter.Node, kind string, parentStableKey string) []repository.SymbolRecord {
	if node.HasError() {
		return nil
	}

	nameNodes := node.ChildrenByFieldName("name", s.cursor)
	symbols := make([]repository.SymbolRecord, 0, len(nameNodes))
	for _, nameNode := range nameNodes {
		name := nameNode.Utf8Text(s.source)
		if name == "" {
			continue
		}
		symbols = append(symbols, s.newSymbol(kind, name, &nameNode, &nameNode, parentStableKey, name))
	}
	return symbols
}

func (s *walkState) typeSymbols(node *sitter.Node, parentStableKey string) ([]repository.SymbolRecord, string, bool) {
	if node.HasError() {
		return nil, "", false
	}

	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil, "", false
	}
	name := nameNode.Utf8Text(s.source)
	typeSymbol := s.newSymbol("type", name, node, nameNode, parentStableKey, name)
	s.typeStableKeysByName[name] = typeSymbol.StableKey

	typeNode := node.ChildByFieldName("type")
	switch {
	case typeNode == nil:
		return []repository.SymbolRecord{typeSymbol}, "", true
	case typeNode.Kind() == "struct_type":
		structSymbol := s.newSymbol("struct", name, typeNode, nameNode, typeSymbol.StableKey, name)
		return []repository.SymbolRecord{typeSymbol, structSymbol}, structSymbol.StableKey, true
	case typeNode.Kind() == "interface_type":
		interfaceSymbol := s.newSymbol("interface", name, typeNode, nameNode, typeSymbol.StableKey, name)
		return []repository.SymbolRecord{typeSymbol, interfaceSymbol}, interfaceSymbol.StableKey, true
	default:
		return []repository.SymbolRecord{typeSymbol}, "", true
	}
}

func (s *walkState) fieldSymbols(node *sitter.Node, parentStableKey string) []repository.SymbolRecord {
	if node.HasError() || parentStableKey == "" {
		return nil
	}

	nameNodes := node.ChildrenByFieldName("name", s.cursor)
	if len(nameNodes) == 0 {
		typeNode := node.ChildByFieldName("type")
		if typeNode == nil || typeNode.HasError() {
			return nil
		}
		name := normalizeTypeName(typeNode.Utf8Text(s.source))
		if name == "" {
			return nil
		}
		return []repository.SymbolRecord{s.newSymbol("field", name, node, typeNode, parentStableKey, name)}
	}

	symbols := make([]repository.SymbolRecord, 0, len(nameNodes))
	for _, nameNode := range nameNodes {
		name := nameNode.Utf8Text(s.source)
		if name == "" {
			continue
		}
		symbols = append(symbols, s.newSymbol("field", name, node, &nameNode, parentStableKey, name))
	}
	return symbols
}

func (s *walkState) interfaceMethodSymbol(node *sitter.Node, parentStableKey string) (repository.SymbolRecord, bool) {
	if node.HasError() || parentStableKey == "" {
		return repository.SymbolRecord{}, false
	}

	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return repository.SymbolRecord{}, false
	}
	name := nameNode.Utf8Text(s.source)
	if name == "" {
		return repository.SymbolRecord{}, false
	}
	return s.newSymbol("method", name, node, nameNode, parentStableKey, name), true
}

func (s *walkState) functionSymbol(node *sitter.Node) (repository.SymbolRecord, bool) {
	if node.HasError() {
		return repository.SymbolRecord{}, false
	}

	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return repository.SymbolRecord{}, false
	}
	name := nameNode.Utf8Text(s.source)
	return s.newSymbol("function", name, node, nameNode, "", name), true
}

func (s *walkState) methodSymbol(node *sitter.Node) (repository.SymbolRecord, bool) {
	if node.HasError() {
		return repository.SymbolRecord{}, false
	}

	nameNode := node.ChildByFieldName("name")
	receiverNode := node.ChildByFieldName("receiver")
	if nameNode == nil || receiverNode == nil {
		return repository.SymbolRecord{}, false
	}

	name := nameNode.Utf8Text(s.source)
	receiverType := receiverTypeName(receiverNode, s.source)
	qualifiedName := name
	parentStableKey := ""
	if receiverType != "" {
		qualifiedName = receiverType + "." + name
		parentStableKey = s.typeStableKeysByName[receiverType]
	}

	return s.newSymbol("method", name, node, nameNode, parentStableKey, qualifiedName), true
}

func (s *walkState) newSymbol(kind string, name string, spanNode *sitter.Node, nameNode *sitter.Node, parentStableKey string, qualifiedName string) repository.SymbolRecord {
	symbol := repository.SymbolRecord{
		StableKey:       stableKey(kind, qualifiedName, spanNode),
		ParentStableKey: parentStableKey,
		Kind:            kind,
		Name:            name,
		QualifiedName:   qualifiedName,
		IsExported:      isExported(name),
	}

	if spanNode != nil {
		symbol.StartByte = int64(spanNode.StartByte())
		symbol.EndByte = int64(spanNode.EndByte())
		symbol.StartRow = int64(spanNode.StartPosition().Row)
		symbol.StartColumn = int64(spanNode.StartPosition().Column)
		symbol.EndRow = int64(spanNode.EndPosition().Row)
		symbol.EndColumn = int64(spanNode.EndPosition().Column)
	}
	if nameNode != nil {
		symbol.NameStartByte = int64(nameNode.StartByte())
		symbol.NameEndByte = int64(nameNode.EndByte())
	}

	if parentStableKey == "" {
		symbol.Depth = 0
	} else {
		symbol.Depth = 1
	}

	if bodyNode := spanNode.ChildByFieldName("body"); bodyNode != nil {
		symbol.SignatureStartByte = symbol.StartByte
		symbol.SignatureEndByte = int64(bodyNode.StartByte())
	}

	return symbol
}

func (s *walkState) lastSymbolStableKey() string {
	if len(s.symbols) == 0 {
		return ""
	}
	return s.symbols[len(s.symbols)-1].StableKey
}

func stableKey(kind string, qualifiedName string, node *sitter.Node) string {
	payload := fmt.Sprintf("%s:%s:%d:%d", kind, qualifiedName, node.StartByte(), node.EndByte())
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}

func countErrorNodes(node *sitter.Node) int {
	if node == nil {
		return 0
	}

	count := 0
	if node.IsError() || node.IsMissing() {
		count++
	}
	cursor := node.Walk()
	defer cursor.Close()
	for _, child := range node.NamedChildren(cursor) {
		count += countErrorNodes(&child)
	}
	return count
}

func firstNamedChildOfKind(node *sitter.Node, cursor *sitter.TreeCursor, kind string) *sitter.Node {
	for _, child := range node.NamedChildren(cursor) {
		if child.Kind() == kind {
			return &child
		}
	}
	return nil
}

func receiverTypeName(node *sitter.Node, source []byte) string {
	if node == nil {
		return ""
	}

	switch node.Kind() {
	case "type_identifier", "qualified_type", "generic_type":
		return normalizeTypeName(node.Utf8Text(source))
	case "identifier":
		return ""
	}

	cursor := node.Walk()
	defer cursor.Close()
	for _, child := range node.NamedChildren(cursor) {
		if name := receiverTypeName(&child, source); name != "" {
			return name
		}
	}
	return ""
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
