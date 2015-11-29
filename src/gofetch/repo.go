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

type repoInfo struct {
  Output  string
  Stat    os.FileInfo
  Repo    *repoRoot
}

var repoCache = make(map[string]repoInfo)

/**
 * Fetch a package
 */
func packageRepo(pkg string, remap map[string]string, base string) (string, os.FileInfo, *repoRoot, error) {
  
  if remap != nil {
    if v, ok := remap[pkg]; ok {
      pkg = v
    }
  }
  
  cached, ok := repoCache[pkg]
  if ok {
    return cached.Output, cached.Stat, cached.Repo, nil
  }
  
  repo, err := repoRootForImportPath(pkg, secure)
  if err != nil {
    return "", nil, nil, fmt.Errorf("could not determine repo root: %v\n", err)
  }
  
  output := path.Join(base, repo.root)
  info, err := os.Stat(output)
  if err != nil && !os.IsNotExist(err) {
    return "", nil, nil, fmt.Errorf("could not read directory: %v\n", err)
  }
  
  cached = repoInfo{output, info, repo}
  repoCache[pkg]        = cached
  repoCache[repo.root]  = cached
  
  return output, info, repo, nil
}
