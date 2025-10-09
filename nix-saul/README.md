# Using Nix BTW

**TLDR**: install with `nix profile install github:DeprecatedLuar/better-curl-saul?dir=nix-saul`, `saul` command will be available in your shell

---

## Using Nix Flakes for development
If you have flakes enabled, you can use these commands in the root of the repo:

- **Build**
  ```bash
  nix build ./nix-saul
  ```
  The built binary will be available in `./result/bin/saul`.

- **Run**
  ```bash
  nix run ./nix-saul
  ```
  This runs the `saul` binary directly.

- **Install to your user profile**
  ```bash
  nix profile install ./nix-saul
  ```
  The binary will be available in your `$PATH`.

- **Enter Dev Shell**
  ```bash
  nix develop ./nix-saul
  ```
  This starts a shell with all Go dependencies available.

---

## Run Directly From GitHub (No Clone Needed)

You can build, run and install the binary directly from GitHub using Nix flakes:

- **Build**
  ```bash
  nix build github:DeprecatedLuar/better-curl-saul?dir=nix-saul
  ```
- **Run**
  ```bash
  nix run github:DeprecatedLuar/better-curl-saul?dir=nix-saul
  ```
- **Install to your user profile**
  ```bash
  nix profile install github:DeprecatedLuar/better-curl-saul?dir=nix-saul
  ```
