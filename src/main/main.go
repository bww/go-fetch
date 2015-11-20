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

var buildA bool // -a flag
var buildN bool // -n flag
var buildV bool // -v flag
var buildX bool // -x flag
var buildI bool // -i flag

/**
 * 
 */
func main() {
  cmd := path.Base(os.Args[0])
  pwd := os.Getenv("PWD")
  
  cmdline   := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
  fOutput   := cmdline.String   ("o",  pwd,     "The directory in which to write packages")
  fUpdate   := cmdline.Bool     ("u",  false,   "Update the package if it has already been downloaded")
  fListOnly := cmdline.Bool     ("l",  false,   "Do not update packages; only list imports if a package exists")
  fVerbose  := cmdline.Bool     ("v",  false,   "Be more verbose")
  cmdline.Parse(os.Args[1:])
  
  go15VendorExperiment = os.Getenv("GO15VENDOREXPERIMENT") != ""
  buildV = *fVerbose
  
  args := cmdline.Args()
  for _, e := range args {
    var err error
    
    // Analyze the import path to determine the version control system,
    // repository, and the import path for the root of the repository.
    rr, err := repoRootForImportPath(e, secure)
    if err != nil {
      fmt.Printf("%v: could not determine repo root: %v\n", cmd, err)
      return
    }
    
    output := path.Join(*fOutput, rr.root)
    info, err := os.Stat(output)
    if err != nil && !os.IsNotExist(err) {
      fmt.Printf("%v: could not read directory: %v\n", cmd, err)
      return
    }
    
    if buildV{
      fmt.Printf("%v: %v -> %v\n", cmd, rr.repo, output)
    }
    
    if !*fListOnly {
      if info == nil {
        base := path.Dir(output)
        err = os.MkdirAll(base, os.ModeDir | 0755)
        if err != nil {
          fmt.Printf("%v: could not create directory: %v\n", cmd, err)
          return
        }
        err = rr.vcs.create(output, rr.repo)
        if err != nil {
          fmt.Printf("%v: could not create repo: %v\n", cmd, err)
          return
        }
      }else if *fUpdate{
        err = rr.vcs.download(output)
        if err != nil {
          fmt.Printf("%v: could not update directory: %v\n", cmd, err)
          return
        }
      }else{
        if buildV {
          fmt.Printf("%v: %v exists\n", cmd, rr.root)
        }
        continue
      }
    }
    
    imp, err := importsForSourceDir(output, looksLikeADomainNameFilter)
    if err != nil {
      fmt.Printf("%v: could not infer dependencies: %v\n", cmd, err)
      return
    }
    
    fmt.Println(strings.Join(imp, "\n"))
  }
  
}
