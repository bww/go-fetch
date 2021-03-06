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
  "fmt"
  "flag"
  "path"
  "strings"
)

var go15VendorExperiment bool

var buildV bool // -v flag
var buildX bool // -x flag

var optVerbose bool
var optDebug bool
var optMapPackages stringList

var cmd = path.Base(os.Args[0])
var cmdline = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

/**
 * Init
 */
func init() {
  cmdline.BoolVar (&optVerbose,     "verbose",  false,  "Be verbose.")
  cmdline.BoolVar (&optDebug,       "debug",    false,  "Be even more verbose.")
  cmdline.Var     (&optMapPackages, "map",              "Explicitly map a package to its root (e.g., 'github.com/a/b/c/d=github.com/a/b'). This can be used to correct for broken or badly behaving repos.")
}

/**
 * Print usage
 */
func usage() {
  fmt.Printf("usage: %v (fetch|scan) [-options] package1 [package2 ...]\n", cmd)
}

/**
 * You know what it does
 */
func main() {
  
  if len(os.Args) < 2 {
    usage()
    return
  }
  
  go15VendorExperiment = os.Getenv("GO15VENDOREXPERIMENT") != ""
  
  act := os.Args[1]
  switch {
    case strings.HasPrefix("fetch", act):
      fetch(os.Args[2:])
    case strings.HasPrefix("scan", act):
      infer(os.Args[2:])
    default:
      fmt.Printf("error: no such command %q\n", act)
      usage()
      return
  }
  
}

/**
 * Infer imports
 */
func infer(args []string) {
  
  fSource   := cmdline.String ("source",  os.Getenv("PWD"),   "The directory in which package sources are found.")
  fListPath := cmdline.Bool   ("paths",   false,              "List paths instead of packages.")
  cmdline.Parse(args)
  
  opts := inferOptions{
    ExcludeFilter: looksPrivateSourceFilter,
  }
  if *fListPath {
    opts.ListPaths = true
  }else{
    opts.ListPackages = true
  }
  
  noted := make(map[string]struct{})
  listed := make(map[string]struct{})
  for _, e := range cmdline.Args() {
    err := inferInc(noted, listed, *fSource, []string{e}, nil, opts)
    if err != nil {
      fmt.Printf("%v: %v\n", cmd, err)
      return
    }
  }
  
 }

/**
 * Process packages
 */
func inferInc(noted, listed map[string]struct{}, srcbase string, pkgs []string, remap map[string]string, opts inferOptions) error {
  for _, e := range pkgs {
    
    // find our repo
    dir, info, _, err := packageRepo(e, remap, srcbase)
    if err == errRepoRootNotFound {
      dir, srcbase = e, e
      info, err = os.Stat(dir)
      if err != nil {
        if os.IsNotExist(err) {
          return fmt.Errorf("no such package or path: %v\n", e)
        }else{
          return fmt.Errorf("could not read directory: %v\n", err)
        }
      }
    }else if err != nil {
      return fmt.Errorf("%v: %v", e, err)
    }
    if info == nil {
      continue
    }
    
    // make sure we haven't already visited this repo
    if _, ok := noted[dir]; ok {
      continue
    }else{
      noted[dir] = struct{}{}
    }
    
    // infer dependencies
    deps, err := packageDeps(dir, opts)
    if err != nil {
      return err
    }
    
    // list them
    for _, d := range deps {
      if _, ok := listed[d]; ok {
        continue
      }
      if opts.ListPaths {
        fmt.Printf("%v\n", path.Join(srcbase, d))
      }else if opts.ListPackages {
        fmt.Printf("%v\n", d)
      }
      listed[d] = struct{}{}
    }
    
    // recurse to dependencies
    err = inferInc(noted, listed, srcbase, deps, remap, opts)
    if err != nil {
      return err
    }
    
  }
  return nil
}

/**
 * Fetch packages
 */
func fetch(args []string) {
  
  fOutput   := cmdline.String ("output",    os.Getenv("PWD"),  "The directory in which to write packages.")
  fUpdate   := cmdline.Bool   ("update",    false,             "Update packages if they have already been downloaded. When combined with -s packages are remoted and re-fetched.")
  fKeepVCS  := cmdline.Bool   ("keep-vcs",  false,             "Retain VCS files from downloaded packages (.git, .svn, .hg, .bzr).")
  cmdline.Parse(args)
  
  mapPackages := make(map[string]string)
  if optMapPackages != nil {
    for _, e := range optMapPackages {
      p := strings.Split(e, "=")
      if len(p) != 2 {
        fmt.Printf("%v: invalid package mapping: %v", cmd, e)
        return
      }
      mapPackages[p[0]] = p[1]
    }
  }
  
  opts := fetchOptions{
    AllowUpdate: *fUpdate,
    StripVCS: !*fKeepVCS,
    InferOptions: inferOptions{
      ExcludeFilter: looksPrivateSourceFilter,
    },
  }
  
  noted := make(map[string]struct{})
  err := fetchInc(noted, cmdline.Args(), mapPackages, *fOutput, opts)
  if err != nil {
    fmt.Printf("%v: %v\n", cmd, err)
    return
  }
  
}

/**
 * Process packages
 */
func fetchInc(noted map[string]struct{}, pkgs []string, remap map[string]string, outbase string, opts fetchOptions) error {
  for _, e := range pkgs {
    
    // find our repo
    dir, info, repo, err := packageRepo(e, remap, outbase)
    if err != nil {
      return err
    }
    
    // make sure we haven't already visited this repo
    if _, ok := noted[dir]; ok {
      continue
    }else{
      noted[dir] = struct{}{}
    }
    
    if repo.root != e {
      fmt.Printf(" + %v (%v)\n", e, repo.root)
    }else{
      fmt.Printf(" + %v\n", e)
    }
    
    // if we're stripping VCS files we cannot update, we must delete and re-fecth
    if info != nil && opts.AllowUpdate && opts.StripVCS {
      err = os.RemoveAll(dir)
      if err != nil {
        return err
      }
      info = nil
    }
    
    // if we're not only listing packages, actually fetch them
    err = fetchPackage(dir, info, repo, opts)
    if err != nil {
      return err
    }
    
    // if we're stripping VCS files, do that
    if opts.StripVCS {
      err = prunePath(dir, vcsFileFilter, true)
      if err != nil {
        return err
      }
    }
    
    // infer dependencies
    deps, err := packageDeps(dir, opts.InferOptions)
    if err != nil {
      return err
    }
    
    // recurse to dependencies
    err = fetchInc(noted, deps, remap, outbase, opts)
    if err != nil {
      return err
    }
    
  }
  return nil
}
