package base

import (
	"fmt"
	"github.com/ElaraLang/elara/interpreter"
	"github.com/ElaraLang/elara/lexer"
	"github.com/ElaraLang/elara/parserlegacy"
	"os"
	"time"
)

func Execute(fileName *string, code string, scriptMode bool) (results []*interpreter.Value, lexTime, parseTime, execTime time.Duration) {
	start := time.Now()
	result := lexer.Lex(code)
	lexTime = time.Since(start)

	start = time.Now()
	psr := parserlegacy.NewParser(result)
	parseRes, errs := psr.Parse()
	parseTime = time.Since(start)

	if len(errs) != 0 {
		file := "Unknown File"
		if fileName != nil {
			file = *fileName
		}
		_, _ = os.Stderr.WriteString(fmt.Sprintf("Syntax Errors found in %s: \n", file))
		for _, err := range errs {
			_, _ = os.Stderr.WriteString(fmt.Sprintf("%s\n", err))
		}
		return []*interpreter.Value{}, lexTime, parseTime, time.Duration(-1)
	}

	start = time.Now()
	evaluator := interpreter.NewInterpreter(parseRes)

	results = evaluator.Exec(scriptMode)
	execTime = time.Since(start)
	return results, lexTime, parseTime, execTime
}
