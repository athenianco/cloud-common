package codegen

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func Generate(fname string) error {
	path := filepath.Dir(fname)
	dir, err := os.Open(path)
	if err != nil {
		return err
	}
	defer dir.Close()

	names, err := dir.Readdirnames(-1)
	if err != nil {
		return err
	}
	g := new(Generator)
	for _, name := range names {
		if !strings.HasSuffix(name, ".go") {
			continue
		}
		if err := g.ProcessFile(filepath.Join(path, name)); err != nil {
			return err
		}
	}
	_, err = g.Generate(fname)
	return err
}

type FuncKind int

func (k FuncKind) Trigger() string {
	switch k {
	case HTTPFunc:
		return "http"
	case PubSubFunc:
		return "pubsub"
	}
	return ""
}

const (
	Invalid = FuncKind(iota)
	HTTPFunc
	PubSubFunc
)

type Func struct {
	Type string
	Init bool
	Kind FuncKind
}

type Generator struct {
	Pkg   string
	Funcs map[string]*Func
}

func (g *Generator) ProcessFile(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	fs := token.NewFileSet()
	f, err := parser.ParseFile(fs, path, data, parser.ParseComments)
	if err != nil {
		return err
	}
	if g.Pkg == "" {
		g.Pkg = f.Name.Name
	}
	return g.processAST(f)
}

func (g *Generator) processAST(f *ast.File) error {
	for _, d := range f.Decls {
		switch d := d.(type) {
		case *ast.FuncDecl:
			if err := g.processFunc(d); err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *Generator) funcByType(r *ast.FieldList) *Func {
	e := r.List[0].Type
	if p, ok := e.(*ast.StarExpr); ok {
		e = p.X
	}
	id := e.(*ast.Ident)
	k := id.Name
	f := g.Funcs[k]
	if f == nil {
		if g.Funcs == nil {
			g.Funcs = make(map[string]*Func)
		}
		f = &Func{Type: k}
		g.Funcs[k] = f
	}
	return f
}

func (g *Generator) processFunc(d *ast.FuncDecl) error {
	if d.Recv == nil || len(d.Recv.List) == 0 {
		return nil
	}
	// TODO: check signatures
	switch d.Name.Name {
	case "Init":
		f := g.funcByType(d.Recv)
		f.Init = true
	case "ServeHTTP":
		f := g.funcByType(d.Recv)
		f.Kind = HTTPFunc
	case "HandleMessage":
		f := g.funcByType(d.Recv)
		f.Kind = PubSubFunc
	}
	return nil
}

func (g *Generator) GenerateFunc(w io.Writer, f *Func) error {
	switch f.Kind {
	case HTTPFunc:
		return g.generateHTTPFunc(w, f)
	case PubSubFunc:
		return g.generatePubSubFunc(w, f)
	}
	return nil
}

func (g *Generator) generateHTTPFunc(w io.Writer, f *Func) error {
	t := "fnc" + f.Type
	_, err := fmt.Fprintf(w, `
var _ funcs.WebhookHandler = &%s{}

var %s struct{
	once sync.Once
	h %s
}

// %sFunc is an auto-generate wrapper for a cloud function. See %s for details.
func %sFunc(w http.ResponseWriter, r *http.Request) {
	defer report.Flush(3*time.Second)
	defer sentry.RecoverAndPanic(r.Context())
	f := &%s
	f.once.Do(func(){
		if err := f.h.Init(); err != nil {
			panic(err)
		}
		enabled, err := service.Register(r.Context())
		if err != nil {
			panic(err)
		}
		if !enabled {
			panic(report.NewIgnoredError(errors.New("disabled")))
		}
	})
	ctx, cancel := common.EnsureTimeout(r.Context())
	defer cancel()
	r = r.WithContext(ctx)
	f.h.ServeHTTP(w, r)
}
`,
		f.Type,
		t, f.Type,
		f.Type, f.Type,
		f.Type,
		t,
	)
	return err
}

func (g *Generator) generatePubSubFunc(w io.Writer, f *Func) error {
	t := "fnc" + f.Type
	_, err := fmt.Fprintf(w, `
var _ funcs.PubSubHandler = &%s{}

var %s struct{
	once sync.Once
	h %s
}

// %sFunc is an auto-generate wrapper for a cloud function. See %s for details.
func %sFunc(ctx context.Context, msg *pubsub.Message) error {
	defer report.Flush(3*time.Second)
	defer sentry.RecoverAndPanic(ctx)
	f := &%s
	f.once.Do(func(){
		if err := f.h.Init(); err != nil {
			panic(err)
		}
		enabled, err := service.Register(ctx)
		if err != nil {
			panic(err)
		}
		if !enabled {
			panic(report.NewIgnoredError(errors.New("disabled")))
		}
	})
	ctx, cancel := common.EnsureTimeout(ctx)
	defer cancel()
	return f.h.HandleMessage(ctx, msg)
}
`,
		f.Type,
		t, f.Type,
		f.Type, f.Type,
		f.Type,
		t,
	)
	return err
}

func (g *Generator) Generate(out string) ([]string, error) {
	var funcs []*Func
	for _, v := range g.Funcs {
		if v.Init && v.Kind != Invalid {
			funcs = append(funcs, v)
		}
	}
	sort.Slice(funcs, func(i, j int) bool {
		return funcs[i].Type < funcs[j].Type
	})
	if len(funcs) == 0 {
		return nil, nil
	}
	dir := filepath.Dir(out)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil && !os.IsExist(err) {
		return nil, err
	}
	f, err := os.Create(out)
	if err != nil {
		return nil, err
	}
	var w io.Writer = f
	fmt.Fprintf(w, `// Code generated by funcgen. DO NOT EDIT.

package %s

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	// https://github.com/GoogleCloudPlatform/functions-framework-go/issues/30#issuecomment-648528715
	_ "github.com/GoogleCloudPlatform/functions-framework-go/funcframework"

	// always keep those imports to register monitoring services
	_ "github.com/athenianco/cloud-common/report/prometheus"
	_ "github.com/athenianco/cloud-common/report/sentry"

	common "github.com/athenianco/cloud-common"
	"github.com/athenianco/cloud-common/funcs"
	"github.com/athenianco/cloud-common/pubsub"
	"github.com/athenianco/cloud-common/report"
	"github.com/athenianco/cloud-common/report/sentry"
	"github.com/athenianco/cloud-common/service"
)

var (
	_ = http.StatusOK
	_ = common.EnsureTimeout
	_ context.Context
	_ *pubsub.Message
	_ report.Reporter
	_ = service.Register
	_ funcs.WebhookHandler
	_ time.Time
)
`, g.Pkg)
	var (
		objs      []interface{}
		funcNames []string
	)
	for _, f := range funcs {
		if err := g.GenerateFunc(w, f); err != nil {
			return nil, err
		}
		fName := f.Type + "Func"
		funcNames = append(funcNames, fName)
		objs = append(objs, map[string]string{
			"name":    fName,
			"trigger": f.Kind.Trigger(),
		})
	}
	if err := f.Close(); err != nil {
		return nil, err
	}
	data, err := json.MarshalIndent(objs, "", "\t")
	if err != nil {
		return nil, err
	}
	err = ioutil.WriteFile(filepath.Join(dir, "cloudfuncs.json"), data, 0644)
	if err != nil {
		return nil, err
	}
	return funcNames, nil
}
