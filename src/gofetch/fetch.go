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
  "path"
)

/**
 * Fetch a package
 */
func packageRepo(pkg, base string) (string, os.FileInfo, *repoRoot, error) {
  
  repo, err := repoRootForImportPath(pkg, secure)
  if err != nil {
    return "", nil, nil, fmt.Errorf("could not determine repo root: %v\n", err)
  }
  
  output := path.Join(base, repo.root)
  info, err := os.Stat(output)
  if err != nil && !os.IsNotExist(err) {
    return "", nil, nil, fmt.Errorf("could not read directory: %v\n", err)
  }
  
  return output, info, repo, nil
}

/**
 * Fetch a package
 */
func fetchPackage(output string, info os.FileInfo, repo *repoRoot) error {
  var err error
  
  if info == nil {
    base := path.Dir(output)
    
    err = os.MkdirAll(base, os.ModeDir | 0755)
    if err != nil {
      return fmt.Errorf("could not create directory: %v\n", err)
    }
    
    err = repo.vcs.create(output, repo.repo)
    if err != nil {
      return fmt.Errorf("could not create repo: %v\n", err)
    }
    
  }else if buildU {
    
    err = repo.vcs.download(output)
    if err != nil {
      return fmt.Errorf("could not update directory: %v\n", err)
    }
    
  }else{
    
    if buildV {
      fmt.Printf("%v: %v exists\n", cmd, repo.root)
    }
    
  }
  
  return nil
}

/**
 * Infer package dependencies
 */
func packageDeps(dir string) ([]string, error) {
  
  imp, err := importsForSourceDir(dir, looksLikeADomainNameFilter)
  if err != nil {
    return nil, fmt.Errorf("could not infer dependencies: %v\n", err)
  }
  
  return imp, nil
}
