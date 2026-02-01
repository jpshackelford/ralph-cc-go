// Package cshmgen implements the Cshmgen pass: Clight → Csharpminor
// This file handles program-level translation.
package cshmgen

import (
	"github.com/raymyers/ralph-cc/pkg/clight"
	"github.com/raymyers/ralph-cc/pkg/csharpminor"
)

// TranslateProgram translates a complete Clight program to Csharpminor.
func TranslateProgram(prog *clight.Program) *csharpminor.Program {
	result := &csharpminor.Program{}

	// Build global variable set
	globals := make(map[string]bool)
	for _, g := range prog.Globals {
		globals[g.Name] = true
	}

	// Translate global variables
	for _, g := range prog.Globals {
		size := sizeofType(g.Type)
		result.Globals = append(result.Globals, csharpminor.VarDecl{
			Name: g.Name,
			Size: size,
		})
	}

	// Translate functions
	for _, fn := range prog.Functions {
		csharpFn := translateFunction(&fn, globals)
		result.Functions = append(result.Functions, csharpFn)
	}

	return result
}

// translateFunction translates a single function from Clight to Csharpminor.
func translateFunction(fn *clight.Function, globals map[string]bool) csharpminor.Function {
	// Build expression and statement translators
	exprTr := NewExprTranslator(globals)
	stmtTr := NewStmtTranslator(exprTr)

	// Build signature
	sig := csharpminor.Sig{
		Return: fn.Return,
	}
	for _, p := range fn.Params {
		sig.Args = append(sig.Args, p.Type)
	}

	// Translate locals
	var locals []csharpminor.VarDecl
	for _, l := range fn.Locals {
		size := sizeofType(l.Type)
		locals = append(locals, csharpminor.VarDecl{
			Name: l.Name,
			Size: size,
		})
	}

	// Build parameter names
	var params []string
	for _, p := range fn.Params {
		params = append(params, p.Name)
	}

	// Translate body
	body := stmtTr.TranslateStmt(fn.Body)

	return csharpminor.Function{
		Name:   fn.Name,
		Sig:    sig,
		Params: params,
		Locals: locals,
		Temps:  fn.Temps,
		Body:   body,
	}
}
