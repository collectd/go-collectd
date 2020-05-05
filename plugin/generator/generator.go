package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"
	"text/template"
)

var functions = []Function{
	{
		Name: "plugin_register_complex_read",
		Args: []Argument{
			{"group", "meta_data_t *"},
			{"name", "char const *"},
			{"callback", "plugin_read_cb"},
			{"interval", "cdtime_t"},
			{"ud", "user_data_t *"},
		},
		Ret: "int",
	},
	{
		Name: "plugin_register_write",
		Args: []Argument{
			{"name", "char const *"},
			{"callback", "plugin_write_cb"},
			{"ud", "user_data_t *"},
		},
		Ret: "int",
	},
	{
		Name: "plugin_register_shutdown",
		Args: []Argument{
			{"name", "char const *"},
			{"callback", "plugin_shutdown_cb"},
		},
		Ret: "int",
	},
	{
		Name: "plugin_register_log",
		Args: []Argument{
			{"name", "char const *"},
			{"callback", "plugin_log_cb"},
			{"ud", "user_data_t *"},
		},
		Ret: "int",
	},
	{
		Name: "plugin_dispatch_values",
		Args: []Argument{
			{"vl", "value_list_t const *"},
		},
		Ret: "int",
	},
	{
		Name: "plugin_get_interval",
		Ret:  "cdtime_t",
	},
	{
		Name: "meta_data_create",
		Ret:  "meta_data_t *",
	},
	{
		Name: "meta_data_destroy",
		Args: []Argument{
			{"md", "meta_data_t *"},
		},
		Ret: "void",
	},
	{
		Name: "meta_data_toc",
		Args: []Argument{
			{"md", "meta_data_t *"},
			{"toc", "char ***"},
		},
		Ret: "int",
	},
	{
		Name: "meta_data_type",
		Args: []Argument{
			{"md", "meta_data_t *"},
			{"key", "char const *"},
		},
		Ret: "int",
	},
	{
		Name: "meta_data_add_boolean",
		Args: []Argument{
			{"md", "meta_data_t *"},
			{"key", "char const *"},
			{"value", "bool"},
		},
		Ret: "int",
	},
	{
		Name: "meta_data_add_double",
		Args: []Argument{
			{"md", "meta_data_t *"},
			{"key", "char const *"},
			{"value", "double"},
		},
		Ret: "int",
	},
	{
		Name: "meta_data_add_signed_int",
		Args: []Argument{
			{"md", "meta_data_t *"},
			{"key", "char const *"},
			{"value", "int64_t"},
		},
		Ret: "int",
	},
	{
		Name: "meta_data_add_unsigned_int",
		Args: []Argument{
			{"md", "meta_data_t *"},
			{"key", "char const *"},
			{"value", "uint64_t"},
		},
		Ret: "int",
	},
	{
		Name: "meta_data_add_string",
		Args: []Argument{
			{"md", "meta_data_t *"},
			{"key", "char const *"},
			{"value", "char const *"},
		},
		Ret: "int",
	},
	{
		Name: "meta_data_get_boolean",
		Args: []Argument{
			{"md", "meta_data_t *"},
			{"key", "char const *"},
			{"value", "bool *"},
		},
		Ret: "int",
	},
	{
		Name: "meta_data_get_double",
		Args: []Argument{
			{"md", "meta_data_t *"},
			{"key", "char const *"},
			{"value", "double *"},
		},
		Ret: "int",
	},
	{
		Name: "meta_data_get_signed_int",
		Args: []Argument{
			{"md", "meta_data_t *"},
			{"key", "char const *"},
			{"value", "int64_t *"},
		},
		Ret: "int",
	},
	{
		Name: "meta_data_get_unsigned_int",
		Args: []Argument{
			{"md", "meta_data_t *"},
			{"key", "char const *"},
			{"value", "uint64_t *"},
		},
		Ret: "int",
	},
	{
		Name: "meta_data_get_string",
		Args: []Argument{
			{"md", "meta_data_t *"},
			{"key", "char const *"},
			{"value", "char **"},
		},
		Ret: "int",
	},
}

const ptrTmpl = "static {{.Ret}} (*{{.Name}}_ptr)({{.ArgsTypes}});\n"

const wrapperTmpl = `{{.Ret}} {{.Name}}_wrapper({{.ArgsStr}}) {
	LOAD({{.Name}});
	{{if not .IsVoid}}return {{end}}(*{{.Name}}_ptr)({{.ArgsNames}});
}

`

type Argument struct {
	Name string
	Type string
}

func (a Argument) String() string {
	return a.Type + " " + a.Name
}

type Function struct {
	Name string
	Args []Argument
	Ret  string
}

func (f Function) ArgsTypes() string {
	if len(f.Args) == 0 {
		return "void"
	}

	var args []string
	for _, a := range f.Args {
		args = append(args, a.Type)
	}

	return strings.Join(args, ", ")
}

func (f Function) ArgsNames() string {
	var args []string
	for _, a := range f.Args {
		args = append(args, a.Name)
	}

	return strings.Join(args, ", ")
}

func (f Function) ArgsStr() string {
	if len(f.Args) == 0 {
		return "void"
	}

	var args []string
	for _, a := range f.Args {
		args = append(args, a.String())
	}

	return strings.Join(args, ", ")
}

func (f Function) IsVoid() bool {
	return f.Ret == "void"
}

type byName []Function

func (f byName) Len() int           { return len(f) }
func (f byName) Less(i, j int) bool { return f[i].Name < f[j].Name }
func (f byName) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }

func main() {
	var rawC bytes.Buffer
	fmt.Fprint(&rawC, `#include "plugin.h"
#include <stdlib.h>
#include <stdbool.h>
#include <dlfcn.h>

#define LOAD(f)                                                             \
  if (f##_ptr == NULL) {                                                    \
    void *hnd = dlopen(NULL, RTLD_LAZY);                                    \
    f##_ptr = dlsym(hnd, #f);                                               \
    dlclose(hnd);                                                           \
  }

`)

	sort.Sort(byName(functions))

	t, err := template.New("ptr").Parse(ptrTmpl)
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range functions {
		if err := t.Execute(&rawC, f); err != nil {
			log.Fatal(err)
		}
	}

	fmt.Fprintln(&rawC)

	t, err = template.New("wrapper").Parse(wrapperTmpl)
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range functions {
		if err := t.Execute(&rawC, f); err != nil {
			log.Fatal(err)
		}
	}

	var fmtC bytes.Buffer

	cmd := exec.Command("clang-format")
	cmd.Stdin = &rawC
	cmd.Stdout = &fmtC
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}

	fmt.Print(`// +build go1.5,cgo

package plugin // import "collectd.org/plugin"

// #cgo CPPFLAGS: -DHAVE_CONFIG_H
// #cgo LDFLAGS: -ldl
`)
	s := bufio.NewScanner(&fmtC)
	for s.Scan() {
		fmt.Println("//", s.Text())
	}
	fmt.Println(`import "C"`)
}
