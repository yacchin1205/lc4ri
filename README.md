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

`lc4ri` builds the notebook from `.lc4ri/notebook/Dockerfile`, so that directory
is the local customization point for the notebook image.

`lc4ri enable nbsearch` adds `.lc4ri/compose.nbsearch.yaml` and
`.lc4ri/notebook/nbsearch/build.sh` to enable the notebook-side extension during
the notebook image build. Feature compose fragments matching
`.lc4ri/compose.*.yaml` are loaded automatically.

To disable NBSearch:

```console
$ lc4ri disable nbsearch
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
- `lc4ri enable nbsearch`
- `lc4ri disable nbsearch`
- `lc4ri ssh keygen`
- `lc4ri ssh authorize`
