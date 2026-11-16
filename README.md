# lc4ri

`lc4ri` is a small command-line tool for starting a local notebook workspace for
*Literate Computing for Reproducible Infrastructure*.

It is inspired by `OperationHub`, but it is intentionally much simpler:

- single-user only
- local machine only
- bound to `localhost`
- centered on the current working directory
- easy to start and stop with one command

## Concept

`OperationHub` is built on JupyterHub for multi-user operation on a shared
server.

`lc4ri` takes the same basic idea, but removes the server-side concerns:

- no multi-user management
- no host-level user synchronization
- no external authentication
- no public network exposure by default

Instead, `lc4ri` prepares a local Jupyter-based environment backed by Docker,
mounts your current directory as the notebook user's home directory, and opens a
notebook workspace for just one user on the same machine.

## Installation

Download the binary for your OS and architecture from the
[Releases page](https://github.com/yacchin1205/lc4ri/releases), rename it to
`lc4ri`, make it executable, and place it on your `PATH`.

On macOS or Linux, after downloading and renaming the binary:

```console
$ chmod +x lc4ri
$ sudo mv lc4ri /usr/local/bin/
```

On Windows, download the executable for your architecture, rename it to
`lc4ri.exe`, and place it on your `PATH`.

Checksums for each release are published as `checksums.txt` on the same
Releases page.

## Typical Usage

Run `lc4ri up` in a project directory:

```console
$ cd my-project
$ lc4ri up
```

This is expected to:

1. start a local notebook environment,
2. mount the current directory as `/home/$USER`,
3. bind the service to `localhost`,
4. open the notebook UI in your browser,
5. make `notebooks/` available from the notebook UI.

To start the environment without opening a browser:

```console
$ lc4ri up --no-open
```

`lc4ri up` builds the local notebook image and pulls missing base images by
default. To refresh base images before starting:

```console
$ lc4ri up --pull
```

Enable NBSearch before starting the environment when notebook search is needed:

```console
$ lc4ri enable nbsearch
```

Enable Mynerva before starting the environment when the notebook assistant is
needed:

```console
$ lc4ri enable mynerva
```

`lc4ri` builds the notebook from `.lc4ri/notebook/Dockerfile`.

To edit the notebook image Dockerfile:

```console
$ lc4ri image edit
```

This opens `.lc4ri/notebook/Dockerfile` with `$VISUAL`, `$EDITOR`, or `vi`.
The default Dockerfile marks the lc4ri-managed section with `DO NOT EDIT`.

`lc4ri enable nbsearch` adds `.lc4ri/compose.nbsearch.yaml` and
`.lc4ri/notebook/nbsearch/build.sh` to enable the notebook-side extension during
the notebook image build. Feature compose fragments matching
`.lc4ri/compose.*.yaml` are loaded automatically.

`lc4ri enable mynerva` adds `.lc4ri/notebook/mynerva/build.sh` to enable the
JupyterLab extension during the notebook image build. It installs
`jupyter_mynerva` from PyPI with pip; the package dependency installs
`nblibram`, whose wheels include the executable. The default install
does not require Node.js or Go. See the upstream
[Mynerva installation instructions](https://github.com/NII-cloud-operation/jupyter-mynerva#install)
and [nblibram installation instructions](https://github.com/NII-cloud-operation/nblibram#getting-started).
It also adds
`.lc4ri/compose.mynerva.yaml` with `MYNERVA_SECRET_KEY` for encrypted API key
storage.

To disable NBSearch:

```console
$ lc4ri disable nbsearch
```

To disable Mynerva:

```console
$ lc4ri disable mynerva
```

Then stop it when you are done:

```console
$ lc4ri down
```

If the environment is already running, open it in your browser:

```console
$ lc4ri open
```

`lc4ri open` is expected to open the running environment in your browser.

Generate an SSH key inside the running notebook environment:

```console
$ lc4ri ssh keygen
```

`lc4ri ssh keygen` prompts for a non-empty passphrase and then runs
`ssh-keygen` in the notebook container. Additional `ssh-keygen` options can be
passed after `keygen`, except `-N`.

Authorize the notebook environment's public key on a remote SSH target:

```console
$ lc4ri ssh authorize hostname
```

`lc4ri ssh authorize` appends the selected public key to the remote
`~/.ssh/authorized_keys` for the user that the SSH connection logs in as. Use
`--pubkey` to select the public key when needed. Use `-i` and `-p` to choose the
host-side SSH identity file and port used for the authorization connection.

`lc4ri` keeps its local runtime files under `.lc4ri/` in the current
directory and uses `notebooks/` as the notebook root:

```text
.
  .ssh/
  .lc4ri/
    compose.yaml
    compose.nbsearch.yaml
    .env
    notebook/
      Dockerfile
      nbsearch/
        build.sh
  notebooks/
```

## Design Goals

- Make local notebook environments easy to reproduce.
- Keep the interface small and predictable.
- Prefer local safety over remote accessibility.
- Stay close to Docker and Jupyter primitives.
- Be simple enough for day-to-day project use.

## Non-Goals

At least initially, `lc4ri` is not intended to provide:

- multi-user features
- shared server management
- TLS termination
- public hosting
- enterprise authentication
- JupyterHub-style administration

## Requirement

- `docker compose` is available on the host machine.

## Available Commands

`lc4ri` provides a minimal local notebook workflow:

- `lc4ri up`
- `lc4ri up --no-open`
- `lc4ri up --pull`
- `lc4ri open`
- `lc4ri down`
- `lc4ri status`
- `lc4ri logs`
- `lc4ri enable nbsearch`
- `lc4ri disable nbsearch`
- `lc4ri enable mynerva`
- `lc4ri disable mynerva`
- `lc4ri image edit`
- `lc4ri ssh keygen`
- `lc4ri ssh authorize`
