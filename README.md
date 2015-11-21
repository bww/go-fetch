# Go Fetch

Go Fetch can be used to download dependency packages recursively into your project. It is intended to be used with the [Go 1.5 Vendor Experiment](https://docs.google.com/document/d/1Bz5-UB7g2uPBdOx-rw5t9MxJwkfpx90cqG9AFL0JAYo/edit) `vendor` package feature.

Go Fetch is not a dependency manager. Unlike the [many](https://github.com/tools/godep), [many](https://github.com/niemeyer/gopkg), [many](https://github.com/gpmgo/gopm), [many](https://github.com/mattn/gom), [many](https://github.com/nitrous-io/goop) Go dependency managers out there, Go Fetch does not attempt to manage anything for you. It doesn't expect your project to have any particular layout, it doesn't litter your codebase with configuration files or create special directories.

Go Fetch makes it simple to download packages along with their dependencies and it makes it simple for you to selectively update those packages. It's kind of like if `go get` was designed for vendoring.

## The Gist

Go Fetch will download packages specified on the command line using similar semantics to `go get` (from which it borrows a fair amount of source). Given a package, Go fetch will:

1. Figure out which repository that package belongs to,
2. Download the repository source into the directory you indicate (or `$PWD`),
3. Parse the downloaded Go source files to discover packages imported by those files,
4. Recursively fetch imported packages which have "domain-y" prefixes (e.g., `github.com/...`, `bitbucket.com/...`, etc),

When scanning for imports Go Fetch makes efforts to avoid private-looking files and packages, including: directories known to be used by dependency managers (`Godep`, etc), hidden files, and files prefixed with `_`.

By default, Go Fetch will strip VCS files when it downloads packages (that is: `.git`, `.hg`, `.svn`, `.bzr`). This is done so that it's easy to commit downloaded package sources into your own repository under a `vendor` package. (If you insist, this behavior can be disabled by passing `-keep-vcs`).

## Commands

Go Fetch supports a couple commands, each of which has options that can be listed by running `gofetch {command} -h` (replacing `{command}` with the actual command name).

* `fetch` – Download packages and dependencies.
* `scan` – Scan a codebase for imported packages and print them to standard output.

## Examples

### Fetch a Package

Fetch the package `github.com/stretchr/testify/assert` and it's dependencies and put them all in the directory `./vendor`. If the package and its dependencies already exists nothing is done.

	$ gofetch fetch -output vendor github.com/stretchr/testify/assert

Which produces the output:

	+ github.com/stretchr/testify/assert (github.com/stretchr/testify)
	+ github.com/stretchr/objx
	+ github.com/davecgh/go-spew/spew (github.com/davecgh/go-spew)
	+ github.com/pmezard/go-difflib/difflib (github.com/pmezard/go-difflib)

### Update an Existing Package

Later, should you want to update a package, provide the `-update` flag to the `fetch` command. In this case the most current version of the package and it's dependencies will be re-fetched from their respective repositories.

	$ gofetch fetch -update -output vendor github.com/stretchr/testify/assert

Which produces the same output as above:

	+ github.com/stretchr/testify/assert (github.com/stretchr/testify)
	+ github.com/stretchr/objx
	+ github.com/davecgh/go-spew/spew (github.com/davecgh/go-spew)
	+ github.com/pmezard/go-difflib/difflib (github.com/pmezard/go-difflib)

### Scan for Imported Packages

To scan a codebase for packages imported by a specific package, you can use the `scan` command.

	$ gofetch scan -source vendor github.com/stretchr/testify/assert

Which produces the output:

	github.com/pmezard/go-difflib/difflib
	github.com/stretchr/testify/mock
	github.com/stretchr/objx
	github.com/stretchr/testify/assert
	github.com/stretchr/testify/require
	github.com/stretchr/testify/http
	github.com/davecgh/go-spew/spew
	github.com/davecgh/go-spew/spew/testdata

In all cases more than one package may be provided in which case the operation is performed on all the arguments.

## Support

Go Fetch is mainly tested on OS X and should work on Linux/UNIX systems. Probably not so hot on Windows.
