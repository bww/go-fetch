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
var buildU bool // -u flag
var buildL bool // -l flag

var cmd string

/**
 * 
 */
func main() {
  cmd = path.Base(os.Args[0])
  
  cmdline   := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
  fOutput   := cmdline.String   ("o",  os.Getenv("PWD"),  "The directory in which to write packages")
  fUpdate   := cmdline.Bool     ("u",  false,             "Update the package if it has already been downloaded")
  fListOnly := cmdline.Bool     ("l",  false,             "Do not update packages; only list imports if a package exists")
  fVerbose  := cmdline.Bool     ("v",  false,             "Be more verbose")
  cmdline.Parse(os.Args[1:])
  
  go15VendorExperiment = os.Getenv("GO15VENDOREXPERIMENT") != ""
  buildV = *fVerbose
  buildU = *fUpdate
  buildL = *fListOnly
  
  args := cmdline.Args()
  for _, e := range args {
    
    dir, info, repo, err := packageRepo(e, *fOutput)
    if err != nil {
      fmt.Printf("%v: %v", cmd, err)
      return
    }
    
    if !buildL {
      err = fetchPackage(dir, info, repo)
      if err != nil {
        fmt.Printf("%v: %v", cmd, err)
        return
      }
    }
    
    deps, err := packageDeps(dir)
    if err != nil {
      fmt.Printf("%v: %v", cmd, err)
      return
    }
    
    if buildL {
      fmt.Println(strings.Join(deps, "\n"))
    }
    
  }
  
}
