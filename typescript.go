package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/label"
	"github.com/bazelbuild/bazel-gazelle/language"
	"github.com/bazelbuild/bazel-gazelle/repo"
	"github.com/bazelbuild/bazel-gazelle/resolve"
	"github.com/bazelbuild/bazel-gazelle/rule"
)

type tsLang struct{}

func NewLanguage() language.Language {
	return &tsLang{}
}

// Kinds returns a map of maps rule names (kinds) and information on how to
// match and merge attributes that may be found in rules of those kinds. All
// kinds of rules generated for this language may be found here.
func (tsLang) Kinds() map[string]rule.KindInfo {
	return map[string]rule.KindInfo{
		"ts_project": {
			MergeableAttrs: map[string]bool{
				"srcs": true, "deps": true},
		},
	}
}

// Loads returns .bzl files and symbols they define. Every rule generated by
// GenerateRules, now or in the past, should be loadable from one of these
// files.
func (tsLang) Loads() []rule.LoadInfo {
	return []rule.LoadInfo{
		{
			Name: "@npm//@bazel/typescript:index.bzl",
			Symbols: []string{
				"ts_project",
			},
		},
	}
}

// GenerateRules extracts build metadata from source files in a directory.
// GenerateRules is called in each directory where an update is requested
// in depth-first post-order.
//
// args contains the arguments for GenerateRules. This is passed as a
// struct to avoid breaking implementations in the future when new
// fields are added.
//
// A GenerateResult struct is returned. Optional fields may be added to this
// type in the future.
//
// Any non-fatal errors this function encounters should be logged using
// log.Print.
func (tsLang) GenerateRules(args language.GenerateArgs) language.GenerateResult {
	var res language.GenerateResult
	re := regexp.MustCompile(`(?m)^import(?:.*?)from '(.*?)';?$`)
	pkgRule := rule.NewRule("ts_project", "ts_project")
	allDeps := []string{}
	for _, f := range args.RegularFiles {
		log.Println("processing file", path.Join(args.Dir, f))
		if filepath.Ext(f) != ".ts" {
			continue
		}
		content, err := ioutil.ReadFile(path.Join(args.Dir, f))
		if err != nil {
			log.Println("error reading file " + f)
			log.Println(err)
		}
		contentStr := string(content)
		log.Println("content", contentStr)
		lines := strings.Split(contentStr, "\n")
		for _, line := range lines {
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				allDeps = append(allDeps, matches[1])
			}
		}
	}
	log.Println("all deps", allDeps)
	if len(allDeps) > 0 {
		pkgRule.SetAttr("deps", allDeps)
		res.Gen = []*rule.Rule{pkgRule}
		res.Imports = []interface{}{"foo"}
	}
	return res
}

func (tsLang) Name() string { return "gazelle_typescript" }

func (tsLang) RegisterFlags(fs *flag.FlagSet, cmd string, c *config.Config) {}

func (tsLang) CheckFlags(fs *flag.FlagSet, c *config.Config) error { return nil }

func (tsLang) Configure(c *config.Config, rel string, f *rule.File) {
}

func (tsLang) Embeds(r *rule.Rule, from label.Label) []label.Label { return nil }

// Lists files in the given directory with the given extension
func listFilesWithExtension(dir string, ext string) []string {
	dirFile, err := os.Open(dir)
	if err != nil {
		log.Fatal(err)
	}
	files, err := dirFile.Readdir(-1)
	dirFile.Close()
	if err != nil {
		log.Fatal(err)
	}
	theFiles := make([]string, 0, len(files))
	for _, file := range files {
		if filepath.Ext(file.Name()) == ext {
			theFiles = append(theFiles, path.Join(dir, file.Name()))
		}
	}
	return theFiles
}

func (*tsLang) Fix(c *config.Config, f *rule.File) {
	if !c.ShouldFix {
		return
	}
	cabalFiles := listFilesWithExtension(filepath.Dir(f.Path), ".ts")
	if len(cabalFiles) == 0 || f == nil {
		return
	}
}

func (tsLang) Imports(c *config.Config, r *rule.Rule, f *rule.File) []resolve.ImportSpec {
	return []resolve.ImportSpec{}
}

func (tsLang) KnownDirectives() []string {
	return nil
}

func (tsLang) Resolve(c *config.Config, ix *resolve.RuleIndex, rc *repo.RemoteCache, r *rule.Rule, imports interface{}, from label.Label) {

}
