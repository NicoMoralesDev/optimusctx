package app

import (
	"errors"
	"fmt"

	"github.com/niccrow/optimusctx/internal/repository"
)

type BenchmarkSuiteRequest struct {
	SuiteID      string
	SuitePath    string
	SuitesDir    string
	FixturesRoot string
}

type BenchmarkRunner struct {
	LoadSuiteFile             func(string) (repository.BenchmarkSuiteDefinition, error)
	LoadSuiteFiles            func(string) ([]repository.BenchmarkSuiteDefinition, error)
	ValidateFixtureReferences func([]repository.BenchmarkSuiteDefinition, string) error
}

func NewBenchmarkRunner() BenchmarkRunner {
	return BenchmarkRunner{
		LoadSuiteFile:             repository.LoadBenchmarkSuite,
		LoadSuiteFiles:            repository.LoadBenchmarkSuites,
		ValidateFixtureReferences: repository.ValidateBenchmarkFixtureReferences,
	}
}

func (r BenchmarkRunner) LoadSuite(request BenchmarkSuiteRequest) (repository.BenchmarkSuiteDefinition, error) {
	r = r.withDefaults()

	if request.SuitePath != "" {
		if request.SuiteID != "" || request.SuitesDir != "" {
			return repository.BenchmarkSuiteDefinition{}, errors.New("suitePath cannot be combined with suiteID or suitesDir")
		}
		suite, err := r.LoadSuiteFile(request.SuitePath)
		if err != nil {
			return repository.BenchmarkSuiteDefinition{}, err
		}
		if err := r.ValidateFixtureReferences([]repository.BenchmarkSuiteDefinition{suite}, request.FixturesRoot); err != nil {
			return repository.BenchmarkSuiteDefinition{}, err
		}
		return suite, nil
	}

	if request.SuiteID == "" {
		return repository.BenchmarkSuiteDefinition{}, errors.New("suiteID or suitePath is required")
	}
	if request.SuitesDir == "" {
		return repository.BenchmarkSuiteDefinition{}, errors.New("suitesDir is required when selecting by suiteID")
	}

	suites, err := r.LoadSuiteFiles(request.SuitesDir)
	if err != nil {
		return repository.BenchmarkSuiteDefinition{}, err
	}
	if err := r.ValidateFixtureReferences(suites, request.FixturesRoot); err != nil {
		return repository.BenchmarkSuiteDefinition{}, err
	}
	for _, suite := range suites {
		if suite.ID == request.SuiteID {
			return suite, nil
		}
	}
	return repository.BenchmarkSuiteDefinition{}, fmt.Errorf("benchmark suite %q not found", request.SuiteID)
}

func (r BenchmarkRunner) LoadSuites(request BenchmarkSuiteRequest) ([]repository.BenchmarkSuiteDefinition, error) {
	r = r.withDefaults()
	if request.SuitesDir == "" {
		return nil, errors.New("suitesDir is required")
	}
	suites, err := r.LoadSuiteFiles(request.SuitesDir)
	if err != nil {
		return nil, err
	}
	if err := r.ValidateFixtureReferences(suites, request.FixturesRoot); err != nil {
		return nil, err
	}
	return suites, nil
}

func (r BenchmarkRunner) withDefaults() BenchmarkRunner {
	if r.LoadSuiteFile == nil {
		r.LoadSuiteFile = repository.LoadBenchmarkSuite
	}
	if r.LoadSuiteFiles == nil {
		r.LoadSuiteFiles = repository.LoadBenchmarkSuites
	}
	if r.ValidateFixtureReferences == nil {
		r.ValidateFixtureReferences = repository.ValidateBenchmarkFixtureReferences
	}
	return r
}
