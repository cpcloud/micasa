{
  description = "micasa Go development environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable-small";
    flake-utils.url = "github:numtide/flake-utils";
    git-hooks.url = "github:cachix/git-hooks.nix";
  };

  outputs = { nixpkgs, flake-utils, git-hooks, ... }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };

        micasa = pkgs.buildGoModule {
          pname = "micasa";
          version = "0.1.0";
          src = ./.;
          subPackages = [ "cmd/micasa" ];
          vendorHash = null;
          env = {
            CGO_ENABLED = 0;
          };
        };

        preCommit = git-hooks.lib.${system}.run {
          src = ./.;
          hooks = {
            golines = {
              enable = true;
              settings.flags =
                "--base-formatter=${pkgs.gofumpt}/bin/gofumpt "
                + "--max-len=100";
            };
            golangci-lint.enable = true;
            gotest.enable = true;
          };
        };

        root = pkgs.buildEnv {
          name = "micasa-root";
          paths = [ micasa ];
          pathsToLink = [ "/bin" ];
        };
      in
      {
        checks = {
          pre-commit = preCommit;
        };

        devShells.default =
          let
            inherit (preCommit) enabledPackages shellHook;
          in
          pkgs.mkShell {
            inherit shellHook;
            packages = [
              pkgs.gopls
              pkgs.git
            ] ++ enabledPackages;
          };

        packages = {
          inherit micasa;
          default = micasa;
          micasa-container = pkgs.dockerTools.buildImage {
            name = "micasa";
            tag = "latest";
            copyToRoot = root;
            config = {
              Entrypoint = [ "/bin/micasa" ];
            };
          };
        };

        apps = {
          default = flake-utils.lib.mkApp { drv = micasa; };
        };

        formatter = pkgs.nixpkgs-fmt;
      }
    );
}
