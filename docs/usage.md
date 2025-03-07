Usage
=====

This document describes how to use [actionlint](..).

## `actionlint` command

With no argument, actionlint finds all workflow files in the current repository and checks them.

```sh
actionlint
```

When paths to YAML workflow files are given as arguments, actionlint checks them.

```sh
actionlint path/to/workflow1.yaml path/to/workflow2.yaml
```

When `-` argument is given, actionlint reads inputs from stdin and checks it as workflow source.

```sh
cat path/to/workflow.yaml | actionlint -
```

To know all flags and options, see an output of `actionlint -h` or [the online command manual][cmd-manual].

### Ignore some errors

To ignore some errors, `-ignore` option offers to filter errors by messages using regular expression. The option is repeatable.

```sh
actionlint -ignore 'label ".+" is unknown' -ignore '".+" is potentially untrusted'
```

`-shellcheck` and `-pyflakes` specifies file paths of executables. Setting empty string to them disables `shellcheck` and
`pyflakes` rules. As a bonus, disabling them makes actionlint much faster Since these external linter integrations spawn many
processes.

```sh
actionlint -shellcheck= -pyflakes=
```

<a name="format"></a>
### Format error messages

`-format` option can flexibly format error messages with [Go template syntax][go-template].

Before explaining the formatting details, let's see some examples.

#### Example: Serialized into JSON

```sh
actionlint -format '{{json .}}'
```

Output:

```
[{"message":"unexpected key \"branch\" for ...
```

#### Example: Markdown

````sh
actionlint -format '{{range $err := .}}### Error at line {{$err.Line}}, col {{$err.Column}} of `{{$err.Filepath}}`\n\n{{$err.Message}}\n\n```\n{{$err.Snippet}}\n```\n\n{{end}}'
````

Output:

````markdown
### Error at line 21, col 20 of `test.yaml`

property "platform" is not defined in object type {os: string}

```
          key: ${{ matrix.platform }}-node-${{ hashFiles('**/package-lock.json') }}
                   ^~~~~~~~~~~~~~~
```
````

#### Example: Serialized in [JSON Lines][jsonl]

```sh
actionlint -format '{{range $err := .}}{{json $err}}{{end}}'
```

Output:

```
{"message":"unexpected key \"branch\" for ...
{"message":"character '\\' is invalid for branch ...
{"message":"label \"linux-latest\" is unknown. ...
```

#### Example: [Error annotation][ga-annotate-error] on GitHub Actions

````sh
actionlint -format '{{range $err := .}}::error file={{$err.Filepath}},line={{$err.Line}},col={{$err.Column}}::{{$err.Message}}%0A```%0A{{replace $err.Snippet "\\n" "%0A"}}%0A```\n{{end}}' -ignore 'SC2016:'
````

Output:

<img src="https://github.com/rhysd/ss/blob/master/actionlint/ga-annotate.png?raw=true" alt="annotations on GitHub Actions" width="731" height="522"/>

To include newlines in the annotation body, it prints `%0A`. (ref [actions/toolkit#193](https://github.com/actions/toolkit/issues/193)).
And it suppresses `SC2016` shellcheck rule error since it complains about the template argument.

Basically it is more recommended to use [Problem Matchers](#problem-matchers) or reviewdog as explained in
['Tools integration' section](#tools-integ) below.

#### Formatting syntax

In [Go template syntax][go-template], `.` within `{{ }}` means the target object. Here, the target object is a sequence of error
objects.

The sequence can be traversed with `range` statement, which is like `for ... = range ... {}` in Go.

```
{{range $err = .}} this part iterates error objects with the iteration variable $err {{end}}
```

The error object has the following fields.

| Field               | Description                                        | Example                                                          |
|---------------------|----------------------------------------------------|------------------------------------------------------------------|
| `{{$err.Message}}`  | Body of error message                              | `property "platform" is not defined in object type {os: string}` |
| `{{$err.Snippet}}`  | Code snippet to indicate error position            | `          node_version: 16.x\n          ^~~~~~~~~~~~~`          |
| `{{$err.Kind}}`     | Name of rule the error belongs to                  | `expression`                                                     |
| `{{$err.Filepath}}` | Canonical relative file path of the error position | `.github/workflows/ci.yaml`                                      |
| `{{$err.Line}}`     | Line number of the error position (1-based)        | `21`                                                             |
| `{{$err.Column}}`   | Column number of the error position (1-based)      | `20`                                                             |

For example, the following simple iteration body

```
line is {{$err.Line}}, col is {{$err.Column}}, message is {{$err.Message | printf "%q"}}
```

will produce output like below.

```
line is 21, col is 20, message is "property \"platform\" is not defined in object type {os: string}"
```

In `{{ }}` placeholder, input can be piped and action can be used to transform texts. In above example, the message is piped with
`|` and transformed with `printf "%q"`. Most useful action would be `json` as we already used it in the above JSON example. It
serializes the given object into JSON string followed by newline character.

Note that special characters escaped with back slash like `\n` in the format string are automatically unespcaed.

### Exit status

`actionlint` command exits with one of the following exit statuses.

| Status | Description                                             |
|--------|---------------------------------------------------------|
| `0`    | The command ran successfully and no problem was found   |
| `1`    | The command ran successfully and some problem was found |
| `2`    | The command failed due to invalid command line option   |
| `3`    | The command failed due to some fatal error              |

<a name="on-github-actions"></a>
## Use actionlint on GitHub Actions

Preparing `actionlint` executable with the download script is recommended. See [the instruction](install.md#download-script) for
more details. It sets an absolute file path of downloaded executable to `executable` output in order to use the executable in the
following steps easily.

Here is an example of simple workflow to run actionlint on GitHub Actions. Please ensure `shell: bash` since the default
shell for Windows runners is `pwsh`.

```yaml
name: Lint GitHub Actions workflows
on: [push, pull_request]

jobs:
  actionlint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Download actionlint
        id: get_actionlint
        run: bash <(curl https://raw.githubusercontent.com/rhysd/actionlint/main/scripts/download-actionlint.bash)
        shell: bash
      - name: Check workflow files
        run: ${{ steps.get_actionlint.outputs.executable }} -color
        shell: bash
```

Or simply download the executable and run it in one step:

```yaml
- name: Check workflow files
  run: |
    bash <(curl https://raw.githubusercontent.com/rhysd/actionlint/main/scripts/download-actionlint.bash)
    ./actionlint -color
  shell: bash
```

If you want to enable [shellcheck integration](checks.md#check-shellcheck-integ), install `shellcheck` command. Note that
shellcheck is [pre-installed on Ubuntu worker][preinstall-ubuntu].

If you want to [annotate errors][ga-annotate-error] from actionlint on GitHub, consider to use
[Problem Matchers](#problem-matchers).

## Online playground

Thanks to WebAssembly, actionlint playground is available on your browser. It never sends any data to outside of your browser.

https://rhysd.github.io/actionlint/

Paste your workflow content to the code editor at left pane. It automatically shows the results at right pane. When editing
the workflow content in the code editor, the results will be updated on the fly. Clicking an error message in the results
table moves a cursor to position of the error in the code editor.

<a name="docker"></a>
## [Docker][docker] image

[Official Docker image][docker-image] is available. The image contains `actionlint` executable and all dependencies (shellcheck
and pyflakes).

Available tags are:

- `actionlint:latest`: Latest stable version of actionlint. This image is recommended.
- `actionlint:{version}`: Specific version of actionlint. (e.g. `actionlint:1.6.1`)

Just run the image with `docker run`:

```sh
docker run --rm rhysd/actionlint:latest -version
```

To check all workflows in your repository, mount your repository's root directory as a volume and run actionlint in the mounted
directory. When you are at a root directory of your repository:

```sh
docker run --rm -v $(pwd):/repo --workdir /repo rhysd/actionlint:latest -color
```

To check a file with actionlint in a Docker container, pass the file content via stdin and use `-` argument:

```sh
cat /path/to/workflow.yml | docker run --rm -i rhysd/actionlint:latest -color -
```

Or mount the workflows directory and pass the paths as arguments:

```sh
docker run --rm -v /path/to/workflows:/workflows rhysd/actionlint:latest -color /workflows/ci.yml
```

## Using actionlint from Go program

Go APIs are available. See [the Go API document](api.md) for more details.


<a name="tools-integ"></a>
## Tools integration

### reviewdog

[reviewdog][] is an automated review tool for various code hosting services. It officially [supports actionlint][reviewdog-actionlint].
You can check errors from actionlint easily with inline review comments at pull request review.

The usage is easy. Run `reviewdog/action-actionlint` action in your workflow as follows.

```yaml
name: reviewdog
on: [pull_request]
jobs:
  actionlint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: reviewdog/action-actionlint@v1
```

<a name="problem-matchers"></a>
### Problem Matchers

[Problem Matchers][problem-matchers] is a feature to extract GitHub Actions annotations from terminal outputs of linters.

Copy [actionlint-matcher.json][actionlint-matcher] to `.github/actionlint-matcher.json` in your repository.

Then enable the matcher using `add-matcher` command before running `actionlint` in the step of your workflow.

```yaml
- name: Check workflow files
  run: |
    echo "::add-matcher::.github/actionlint-matcher.json"
    bash <(curl https://raw.githubusercontent.com/rhysd/actionlint/main/scripts/download-actionlint.bash)
    ./actionlint -color
  shell: bash
```

Note that `-color` option is not available because ANSI color sequences prevent the matcher from matching the error messages.

When you change your workflow and the changed line causes a new error, CI will annotate the diff with the extracted error message.

<img src="https://github.com/rhysd/ss/blob/master/actionlint/problem-matcher.png?raw=true" alt="annotation by Problem Matchers" width="715" height="221"/>

### super-linter

[super-linter][] is a Bash script for a simple combination of various linters, provided by GitHub. It has support for actionlint.
Running super-linter in your repository automatically runs actionlint.

Note that currently super-linter does not provide a way to pass additional command line options. So it is not possible to ignore
errors by `-ignore` option with super-linter. See [github/super-linter#1853](https://github.com/github/super-linter/pull/1853) or
[github/super-linter#1667](https://github.com/github/super-linter/pull/1667) for more details.

### pre-commit

[pre-commit][] is a framework for managing and maintaining multi-language Git pre-commit hooks.

actionlint is available as a pre-commit hook.

Once the `actionlint` binary is installed locally, add this to your `.pre-commit-config.yaml` in
your repository:

```yaml
---
repos:
  - repo: https://github.com/rhysd/actionlint
    rev: main
    hooks:
      - id: actionlint
```

or alternatively, run actionlint with Docker:

```yaml
---
repos:
  - repo: https://github.com/rhysd/actionlint
    rev: main
    hooks:
      - id: actionlint-docker
```

---

[Checks](checks.md) | [Installation](install.md) | [Configuration](config.md) | [Go API](api.md) | [References](reference.md)

[reviewdog-actionlint]: https://github.com/reviewdog/action-actionlint
[reviewdog]: https://github.com/reviewdog/reviewdog
[cmd-manual]: https://rhysd.github.io/actionlint/usage.html
[go-template]: https://pkg.go.dev/text/template
[ga-annotate-error]: https://docs.github.com/en/actions/reference/workflow-commands-for-github-actions#setting-an-error-message
[jsonl]: https://jsonlines.org/
[problem-matchers]: https://github.com/actions/toolkit/blob/master/docs/problem-matchers.md
[super-linter]: https://github.com/github/super-linter
[actionlint-matcher]: https://raw.githubusercontent.com/rhysd/actionlint/main/.github/actionlint-matcher.json
[preinstall-ubuntu]: https://github.com/actions/virtual-environments/blob/main/images/linux/Ubuntu2004-README.md
[pre-commit]: https://pre-commit.com
[docker]: https://www.docker.com/
[docker-image]: https://hub.docker.com/r/rhysd/actionlint
