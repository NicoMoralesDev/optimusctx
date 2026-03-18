//go:build !cgo
// +build !cgo

package goextract

import (
	"context"

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

func (a *Adapter) Extract(context.Context, extract.Request) (extract.Result, error) {
	return extract.Result{
		CoverageState:  repository.ExtractionCoverageStateUnsupported,
		CoverageReason: repository.ExtractionCoverageReasonAdapterError,
	}, nil
}
