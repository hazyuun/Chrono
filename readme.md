# Chrono
<p align="center">
    <img src="assets/logo.png" width="200"/>
</p>
<p align="center">
    <strong>Chrono</strong>
</p>
<p align="center">
    A git time machine
</p>
<div align="center">

<img alt="GitHub" src="https://img.shields.io/github/license/hazyuun/chrono?color=green&style=flat-square">

<img alt="GitHub go.mod Go version" src="https://img.shields.io/github/go-mod/go-version/hazyuun/chrono?style=flat-square">

</div>

Chrono is a tool that automatically commits in a temporary branch in your git repository every time an event occurs
(events are customizable), So that you can always rollback to a specific point in time if anything goes wrong. You can squash merge all the temporary commits into one once you are done.

## Disclaimer
This is still in early development stages, if you are going to use it or test it, please use caution.
Use at your own risk, I am NOT responsible for any of your acts.

## Workflow
### Create a chrono branch and checkout to it
Chrono can only push to its own branch in order to avoid messing up your repository, so start by making a branch for chrono
```bash
$ git branch chrono
```
git checkout to it
```bash
$ git checkout chrono
```
 
### Start a chrono session
Start a chrono session using 
```bash
$ chrono session start
```
from now on, chrono will be automatically committing changes whenever an event occurs, events are customizable using a `chrono.yaml` file (see below for details)

### Squash merge and delete the chrono branch
When done, you can merge (A squash merge is recommended) the chrono branch to your original branch (let's call it original_branch) 
```bash
$ git checkout original_branch
```
```bash
$ git merge --squash chrono
```
Then if everything is as expected, you can commit the merge and delete the chrono branch
```bash
$ git commit -m "Your commit message"
```

```bash
$ git branch -D chrono
```

## Config file
Put a file named `chrono.yml` in the root of your repository, here is an example config file
```yaml

# Events when to automatically commit
events:
    # This triggers every amount of minutes
    - periodic:

        # Every 60 seconds
        period: 60

        # Commit those files
        files: ["src/", "file.txt"] 

    # This triggers every file save
    - save:

        # Those files will be committed once they're saved
        files: ["notes.txt"]
        
# Use files: ["."] if you want all files to be commited
```

If you want to exclude some files when using `files: ["."]`, just use your regular .gitignore file

## Contributions
Pull requests and issues are welcome !
