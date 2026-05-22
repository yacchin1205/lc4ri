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
4. make `notebooks/` available from the notebook UI.

Then stop it when you are done:

```console
$ lc4ri down
```

If the environment is already running, open it in your browser:

```console
$ lc4ri open
```

`lc4ri open` is expected to open the running environment in your browser.

`lc4ri` keeps its local runtime files under `.lc4ri/` in the current
directory and uses `notebooks/` as the notebook root:

```text
.
  .ssh/
  .lc4ri/
    compose.yaml
    .env
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
- `lc4ri open`
- `lc4ri down`
- `lc4ri status`
