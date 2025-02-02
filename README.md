<p align="center">
  <img alt="bulk PR creator" alt="Bulk PR creator" src="https://imgur.com/GkKyjKT.png" width="300">
</p>

<p align="center">
  Bulk PR creator for Github and Github Enterprise.
  <br>
  Edit, delete files in multiple repos at once & create PRs.
  <br>
  And NO, it won't delete, or force push to any existing branch.
</p>
<p align="center">
  Useful for tech leads, who work on many repos and many orgs, and has annoying change to make to all repos.
  <br>
  <code>bpr --org=username --cmd="echo .DS_STORE >> .gitignore"</code>
</p>



## Install

```sh
curl -sL https://raw.githubusercontent.com/kevincobain2000/bpr/master/install.sh | sh
```

or via go

```sh
go install github.com/kevincobain2000/bpr@latest
```

## Basic Usage

```sh
# Get your token from github.com/settings/tokens
export GITHUB_TOKEN=your_github_token

# Personal repos
# replace from_text with to_text in all files
bpr --org=username --cmd='sed -i "" "s/from_text/to_text/g"'

# Org repos
# replace from_text with to_text in all files with .go extension
bpr --org=your_org --cmd='sed -i "" "s/from_text/to_text/g" *.go'

# Org repos
# remove all files with test in their name
bpr --org=your_org --cmd='rm ./**/*test*'

# Org repos
# update all workflow files with new node version
bpr --org=your_org --cmd='sed -i "" "s/node: 12/node: 20/g" *.yml'

# Org repos
# Just these repos, csv format
bpr --org=your_org --repos=ui,web --cmd='sed -i "" "s/node: 12/node: 14/g" *.yml'

# Org repos
# Enterprise users
bpr --org=your_org --base-url=ghe.company.com --cmd='echo .DS_STORE >> .gitignore'

# Above are the examples on how to use bulk PR creator (bpr) by giving it a command as cmd
# Write a new command to: <<your GPT prompt>>
```

## Advanced Usage

```sh
  -base-url string
    	GitHub base URL (default "github.com")
  -cmd string
    	action command to run (required)
  -default-branch string
    	Where the PR will be created to (empty for default)
  -dry
    	dry run
  -log-level int
    	log level (0=info, -4=debug, 4=warn, 8=error)
  -org string
    	GitHub organization name or your own username (required)
  -parallel int
    	number of parallel requests (default 10)
  -pr-body string
    	pull request body (default "BPR: bulk PR changes")
  -pr-branch string
    	pull request branch (default "bpr-<random>")
  -pr-commit-msg string
    	pull request commit message (default "BPR: bulk PR changes")
  -pr-title string
    	pull request title (default "BPR: bulk PR changes")
  -repos string
    	comma-separated list of repositories (empty for all)
  -token string
    	GITHUB_TOKEN via env or flag
  -version
    	print version and exit
```


## CHANGELOG

- **v1.0.0** - Initial release
