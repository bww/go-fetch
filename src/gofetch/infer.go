// 
// Go Fetch Dependencies
// Copyright (c) 2015 Brian W. Wolter, All rights reserved.
// 
// Redistribution and use in source and binary forms, with or without modification,
// are permitted provided that the following conditions are met:
// 
//   * Redistributions of source code must retain the above copyright notice, this
//     list of conditions and the following disclaimer.
// 
//   * Redistributions in binary form must reproduce the above copyright notice,
//     this list of conditions and the following disclaimer in the documentation
//     and/or other materials provided with the distribution.
//     
//   * Neither the names of Brian W. Wolter nor the names of the contributors may
//     be used to endorse or promote products derived from this software without
//     specific prior written permission.
//     
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED.
// IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT,
// INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING,
// BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF
// LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE
// OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED
// OF THE POSSIBILITY OF SUCH DAMAGE.
// 

package main

import (
  "os"
  "io"
  "fmt"
  "path"
  "regexp"
  "strings"
  "go/token"
  "go/parser"
)

var domainPrefixRegex = regexp.MustCompile("^[a-zA-Z0-9]([a-zA-Z0-9\\-]*[a-zA-Z0-9])\\.([a-zA-Z0-9]{2,20})")
var privatePathRegex  = regexp.MustCompile("(^|\\/)([_].*|Godep|third_party|pkg)($|\\/)")

var defaultExcludePackages = []string{
  "golang.org/x/tools/cmd/fiximports/testdata",
  "golang.org/x/tools/go/loader/testdata",
  "camlistore.org/depcheck",
}

/**
 * Import inference options
 */
type inferOptions struct {
  ExcludeFilter pathFilter
  ListPaths, ListPackages bool
}

/**
 * Exclude a package?
 */
func excludePackage(n string, s []string) bool {
  for _, e := range s {
    if strings.HasPrefix(strings.ToLower(n), strings.ToLower(e)) {
      return true
    }
  }
  return false
}

/**
 * Exclude imports that don't start with what looks like a "domain.name"
 */
func looksLikeADomainNameFilter(n string) bool {
  return domainPrefixRegex.MatchString(n) && !privatePathRegex.MatchString(n) && !excludePackage(n, defaultExcludePackages)
}

/**
 * Exclude sources that look private (e.g., start with '.', '_'; are a directory known
 * to be used by a dependency manager; are a test file (with suffix '_test.go'); or are
 * otherwise known to be problematic for some reason)
 */
func looksPrivateSourceFilter(n string) bool {
  ts := "_test.go"
  
  base := path.Base(n)
  switch {
    case len(base) < 1 || base[0] == '.' || base[0] == '_':
      return false
    case len(base) > len(ts) && strings.EqualFold(base[:len(base) - len(ts)], ts):
      return false
    case strings.EqualFold(base, "vendor"):
      return false
    case strings.EqualFold(base, "Godep"):
      return false
    case strings.EqualFold(base, "third_party"):
      return false
    case strings.EqualFold(base, "pkg"):
      return false
  }
  
  for _, e := range defaultExcludePackages {
    if strings.HasSuffix(n, e) {
      return false
    }
  }
  
  return true
}

/**
 * Imports
 */
func importsForSourceDir(dir string, filter pathFilter, opts inferOptions) ([]string, error) {
  
  set := make(map[string]struct{})
  err := importsForSourceDirInc(set, dir, true, filter, opts)
  if err != nil {
    return nil, err
  }
  
  imp := make([]string, len(set))
  i := 0
  for k, _ := range set {
    imp[i] = k
    i++
  }
  
  return imp, nil
}

/**
 * Incremental imports
 */
func importsForSourceDirInc(imp map[string]struct{}, dir string, rec bool, filter pathFilter, opts inferOptions) error {
  
  name := path.Base(dir)
  if len(name) < 1 || name[0] == '.' {
    return nil
  }
  
  info, err := os.Stat(dir)
  if err != nil {
    return err
  }
  if !info.IsDir() {
    return fmt.Errorf("Path is not a directory: %v", dir)
  }
  
  file, err := os.Open(dir)
  if err != nil {
    return err
  }else{
    defer file.Close()
  }
  
  items, err := file.Readdir(0)
  if err != nil && err != io.EOF {
    return err
  }
  
  fset := token.NewFileSet()
  for _, e := range items {
    name = e.Name()
    abs := path.Join(dir, name)
    if opts.ExcludeFilter != nil && !opts.ExcludeFilter(abs) {
      continue
    }
    if !e.IsDir() {
      if e.Size() == 0 {
        continue // ignore empty files
      }
      if strings.EqualFold(path.Ext(name), ".go") {
        fset.AddFile(abs, -1, int(e.Size()))
      }
    }else if rec {
      err := importsForSourceDirInc(imp, abs, rec, filter, opts)
      if err != nil {
        return err
      }
    }
  }
  
  pkgs, err := parser.ParseDir(fset, dir, nil, parser.ImportsOnly)
  if err != nil {
    return err
  }
  
  for _, e := range pkgs {
    if e.Files != nil {
      for _, f := range e.Files {
        if f.Imports != nil {
          for _, v := range f.Imports {
            if lit := v.Path.Value; len(lit) > 2 {
              str := lit[1:len(lit)-1]
              if filter == nil || filter(str) {
                imp[str] = struct{}{}
              }
            }
          }
        }
      }
    }
  }
  
  return nil
}

