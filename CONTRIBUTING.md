# Contributing Guide

This is part of the [Porter][porter] project. If you are a new contributor,
check out our [New Contributor Guide][new-contrib]. The Porter [Contributing
Guide][contrib] also has lots of information about how to interact with the
project.

[porter]: https://github.com/deislabs/porter
[new-contrib]: https://porter.sh/contribute
[contrib]: https://github.com/deislabs/porter/blob/main/CONTRIBUTING.md

---

* [Initial setup](#initial-setup)
* [Makefile explained](#makefile-explained)

---

# Initial setup

You need to have [porter installed](https://porter.sh/install) first. Then run
`make build install`. This will build and install the plugin into your porter
home directory.

## Makefile explained

Here are the most common Makefile tasks:

* `build` builds the plugin.
* `install` installs the plugin into **~/.porter/plugins**.
* `test-unit` runs the unit tests.
* `verify-vendor` cleans up packr generated files and verifies that dep's Gopkg.lock 
   and vendor/ are up-to-date. Use this makefile target instead of running 
   dep check manually.
