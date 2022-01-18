# go-dura
[Tim Kellogg's Dura](https://github.com/tkellogg/dura) but written in Go

I encourage anyone with Rust knowledge to contribute to his project and use this great tool.

The main purpose for this effort is to provide the same (or at least very similar) functionality as Tim Kellogg's Dura but in a version written in Golang. 

Initial working version is here! (Bugs expected!)
Pull requests, discussions, and collaborations are welcomed and appreciated.

## Dependencies
go-dura requires libgit2 to be installed in order to function. As with all software there are varying versions and the Go package [git2go](https://github.com/libgit2/git2go) used by go-dura has different versions to match. 
The current go-dura source uses git2go v33, but if you have an incompatible/older version of libgit2 installed and can't update you can look [here](https://github.com/libgit2/git2go) to find an appropriate version and change 
go.mod file and import statements accordingly. go-dura has not been tested extensively even with v33 so no guarantees on older git2go versions.

## Building
To build go-dura simply clone the repository, enter the project folder and run:

    go build -o dura .

## Configuration
go-dura uses [sp13/cobra](https://github.com/spf13/cobra) & [sp13/viper](https://github.com/spf13/viper) for its CLI and configuration management respectively. 
In an attempt to keep the configuration "simple" with people coming from or going to [tkellogg/dura](https://github.com/tkellogg/dura) go-dura uses TOML format for its configuration and has a few extra config options (more to come). 

go-dura defaults to looking into $HOME for its configuration file (.go-dura.toml) but the config home path can be set using the environment variable DURA_CONFIG_HOME to the directory desired.

### Options
#### dura.sleep_seconds (optional)
This is an integer value used to set the sleep time (seconds) between captures during a Dura serve loop, defaults to 5 seconds, Dura will set to default value if value less than 1 second is provided.

#### commit.author (optional)
Author name used as the name in the git signature. If not provided and dura.exclude_git_config is false, Dura will default to the repository's default signature name.

#### commit.email (optional)
Email to be used in the git signature. If not provided and dura.exclude_git_config is false, Dura will default to repository's default signature email.

#### commit.exclude_git_config (optional)
Boolean value indicating whether Dura should ignore any default git configuration settings (such as using a repositories default signature). Default is false.

#### repos
A map of Go type map\[string\]WatchConfig representing all the repositories that Dura will watch for changes and make continuous commits.
The map keys are absolute paths to local git repository folders. Values represent watch configurations with properties: include, exclude and max depth. 
The include and exclude properties are string slices representing gitignore strings which are used in filtering watched files/folders. 
The max depth property is used to control recursion depth.

This configuration property can be set manually through editing the configuration file but is mutated using the Dura CLI watch & unwatch routines.

### Example
An example of a go-dura configuration file would be:

    [dura]
    sleep_seconds=10

    [commit]
    author="Apogee"
    email="apogeesystemsllc@gmail.com"
    exclude_git_config=true

    [repos]
    [repos."/path/to/some/repo"]
    include=["**/src","configs/*",/exe]
    exclude=["**/*.log"]
    max_depth=255

## Usage
Presently the go-dura CLI is not extensive and most commands are self-explanatory, however I'll provide a brief description and usage here, as these commands mature more detail will be added.

### dura capture
This command executes a one-off capture call to the provided repository. The underlying routine represents the action taken by Dura at steady intervals when running the serve command. 
If differences are detected in the repository and the repository and files match all other criteria a Dura commit (and optionally a branch) will be created.

#### Example

    dura capture /home/apogee/go/src/myrepo

### dura watch
This command adds the given repositories to the Dura configuration file. You may optionally specify a comma-separated list of gitignore strings to include (--include, -i) or exclude (--exclude, -e) matching file/folder patterns from the watch.
Additionally, you may specify a recursion max depth (--max-depth, -d). The max depth value must be between 0-255, if an invalid value is provided Dura sets the value back to the default (255).

#### Example

    dura watch /home/apogee/go/src/myrepo
    dura watch /home/apogee/go/src/myrepo /path/to/another/repo --include="/src/**,**/*.log" -e "**/*.exe,**/*.test" --max-depth=200

### dura unwatch
This command removes the given repositories from the Dura configuration, once removed, the repositories will no longer be watched by Dura but past commits will be persisted. 

#### Example

    dura unwatch /home/apogee/go/src/myrepo
    dura unwatch /home/apogee/go/src/myrepo /path/to/some/other/repo

### dura kill
This command is currently not implemented but will serve to kill the Dura daemon process.

#### Example

    dura kill

### dura serve
This is the heart of the Dura CLI, once called Dura will enter an infinite for-loop sleeping for dura.sleep_seconds seconds before looping through all the watched repositories and calling a capture on each one. 
This can be ran in the background or left to log in the terminal.

#### Example

    # running background/daemon process
    dura serve &
    
    # Foreground
    dura serve