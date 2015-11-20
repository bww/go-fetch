# Go Fetch

Go Fetch can be used to copy dependencies recursively into a project, it is intended to be used with the "Go 1.5 Vendor Experiment" `vendor` package.

Go Fetch is not a dependency manager. Unlike the [many](https://github.com/tools/godep), [many](https://github.com/niemeyer/gopkg), [many](https://github.com/gpmgo/gopm), [many](https://github.com/mattn/gom), [many](https://github.com/nitrous-io/goop) Go dependency managers out there, Go Fetch does not attempt to manage anything for you. It makes it simple to download packages along with their dependencies in a manner suitable for use in a Go 1.5 `vendor` package.

## What it does

Go Fetch will download packages with similar semantics to `go get` (from which it borrows a fair amount of source):

1. Given a package (e.g., `github.com/gorilla/websocket`) Go Fetch will download the package source to the directory you indicate,
2. It will parse all the Go source files in that package and inspect the packages imported by those files,
3. For every import that has a prefix which looks "domain-y" (e.g., `github.com/...`, `bitbucket.com/...`, etc) it will recursively fetch that package as well, and so on.

When scanning for imports Go Fetch makes efforts to avoid private-looking files and packages, including: directories known to be used by dependency managers (`Godep`, etc), special packages (`vendor`, `internal`), hidden files, and files prefixed with `_`.

## VCS File Stripping

By default, Go Fetch will strip VCS files when it downloads packages (that is: `.git`, `.hg`, `.svn`, `.bzr`). This is done so that it's straightforward to commit the downloaded package sources into your own repository under your `vendor` package. As a result packages cannot be updated by downloading incremental changes, they must be re-fetched in full. In practice this is generally not much of an inconvenience (it can be disabled with `-keep-vcs` if you insist).
