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

var cmd = path.Base(os.Args[0])
var cmdline = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

/**
 * Init
 */
func init() {
  cmdline.BoolVar (&optVerbose, "verbose",  false,   "Be verbose.")
  cmdline.BoolVar (&optDebug,   "debug",    false,   "Be even more verbose.")
}

/**
 * Print usage
 */
func usage() {
  fmt.Printf("usage: %v (fetch|infer) [-...] package1, [package2, ...]\n", cmd)
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
    case strings.HasPrefix("infer", act):
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
  for _, e := range cmdline.Args() {
    err := inferInc(noted, *fSource, []string{e}, opts)
    if err != nil {
      fmt.Printf("%v: %v", cmd, err)
      return
    }
  }
  
 }

/**
 * Process packages
 */
func inferInc(noted map[string]struct{}, srcbase string, pkgs []string, opts inferOptions) error {
  for _, e := range pkgs {
    if _, ok := noted[e]; ok {
      continue
    }else{
      noted[e] = struct{}{}
    }
    
    // find our repo
    dir, info, _, err := packageRepo(e, srcbase)
    if err != nil {
      return err
    }
    if info == nil {
      continue
    }
    
    // infer dependencies
    deps, err := packageDeps(dir, opts)
    if err != nil {
      return err
    }
    
    // list them
    for _, d := range deps {
      if _, ok := noted[d]; !ok {
        if opts.ListPaths {
          fmt.Printf(" + %v\n", path.Join(srcbase, d))
        }else if opts.ListPackages {
          fmt.Printf(" + %v\n", d)
        }
      }
    }
    
    // recurse to dependencies
    err = inferInc(noted, srcbase, deps, opts)
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
  
  opts := fetchOptions{
    AllowUpdate: *fUpdate,
    StripVCS: !*fKeepVCS,
    InferOptions: inferOptions{
      ExcludeFilter: looksPrivateSourceFilter,
    },
  }
  
  noted := make(map[string]struct{})
  err := fetchInc(noted, cmdline.Args(), *fOutput, opts)
  if err != nil {
    fmt.Printf("%v: %v", cmd, err)
    return
  }
  
}

/**
 * Process packages
 */
func fetchInc(noted map[string]struct{}, pkgs []string, outbase string, opts fetchOptions) error {
  for _, e := range pkgs {
    
    // find our repo
    dir, info, repo, err := packageRepo(e, outbase)
    if err != nil {
      return err
    }
    
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
    err = fetchInc(noted, deps, outbase, opts)
    if err != nil {
      return err
    }
    
  }
  return nil
}
