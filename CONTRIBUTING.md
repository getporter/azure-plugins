# Contributing Guide

This is part of the [Porter][porter] project. If you are a new contributor,
check out our [New Contributor Guide][new-contrib]. The Porter [Contributing
Guide][contrib] also has lots of information about how to interact with the
project.

[porter]: https://github.com/getporter/porter
[new-contrib]: https://porter.sh/contribute
[contrib]: https://porter.sh/src/CONTRIBUTING.md

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

## Debugging

Porter plugins can be debugged using [delve](https://github.com/go-delve/delve) so before attempting to debug [install](https://github.com/go-delve/delve/tree/master/Documentation/installation) delve. Debugging in porter is initiated by setting the following environment varibles prior to starting porter:

* `PORTER_RUN_PLUGIN_IN_DEBUGGER` should be set to the name of the plugin to be debugged (e.g. `secrets.azure.keyvault` to debug the azure secrets plugin)
* `PORTER_DEBUGGER_PORT` should be set to the port number where the delve API will listen, if not set it defaults to `2345`
* `PORTER_PLUGIN_WORKING_DIRECTORY` should be the local path to the source code for the plugin being debugged. (e.g. for the Azure plugin this value would be `<PATH>/porter-azure-plugins/cmd/azure`

Porter will only attach the debugger to an existing binary in **~/.porter/plugins**. It is therefore necessary to build and install the plugin before debugging, use `make debug` to do this.

### Debugging Scenarios

There are 2 different debugging scenarios:

#### 1. Debug the plugin only. 
1. Run `make debug`
1. Set the environment variables as described above.
1. Set the [appropriate configuration in config.toml](https://porter.sh/plugins/azure/).
1. Run porter with a command that invokes the plugin e.g. `porter install debugtest --tag getporter/plugins-tutorial:v0.1.0 --debug -c plugins-tutorial`
1. Connect the debugger to the delve API (by default this will be listening at 127.0.0.1:2345), e.g using delve `dlv connect 127.0.0.1:2345`

#### 2. Debug porter and the plugin simultaneously.
1. Run `make debug`
1. Set the environment variables as described above.
1. Set the [appropriate configuration in config.toml](https://porter.sh/plugins/azure/).
1. Run porter under the debugger with a command that invokes the plugin e.g. `dlv exec <path-to-porter> -- install debugtest --tag getporter/plugins-tutorial:v0.1.0 --debug -c plugins-tutorial` . Note that for this to work porter should be compiled so that the debugger can launch the binary successfully, the simplest way to do this is to build with no options.
1. Using a seperate terminal session connect the debugger to the delve API (by default this will be listening at 127.0.0.1:2345), e.g using delve `dlv connect 127.0.0.1:2345`

### Debugging Using Visual Studio Code

Visual Studio Code tasks, scripts and launch configurations are included in the `.vscode` directory

1. Debug just the plugin.
      1. Edit the porter command by editing the arguments property in the `RunPorter` task in `.vscode/tasks.json`
      1. Set the active launch configuration to `Debug Plugin`
      1. Press F5 to begin debugging.

2. Debug both porter and the plugin simultaneously.
      1. Follow the steps outlined [above](debug-the-plugin-only). (VS Code can be used to debug porter rather than using dlv command line.)
      1. Instead of the final step set the active configuration to `Attach To Plugin` 
      1. Press F5 to begin debugging.
