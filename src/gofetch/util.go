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
)

type pathFilter func(string)(bool)

/**
 * Match VCS files
 */
func vcsFileFilter(p string) bool {
  n := path.Base(p)
  return n == ".git" || n == ".svn" || n == ".hg" || n == ".bzr"
}

/**
 * Delete files in a directory hierarchy which match the provided filter.
 */
func prunePath(dir string, filter pathFilter, rec bool) error {
  
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
  }
  items, err := file.Readdir(0)
  if err != nil && err != io.EOF {
    return err
  }
  
  for _, e := range items {
    abs := path.Join(dir, e.Name())
    if filter(abs) {
      fmt.Printf("PRUNE: %v\n", abs)
      err = os.RemoveAll(abs)
    }else if e.IsDir() && rec {
      err = prunePath(abs, filter, rec)
    }
    if err != nil {
      return err
    }
  }
  
  return nil
}