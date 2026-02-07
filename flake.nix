# Copyright 2026 Phillip Cloud
# Licensed under the Apache License, Version 2.0

{
  description = "micasa Go development environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable-small";
    flake-utils.url = "github:numtide/flake-utils";
    git-hooks.url = "github:cachix/git-hooks.nix";
  };

  outputs =
    {
      nixpkgs,
      flake-utils,
      git-hooks,
      ...
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs { inherit system; };
        version = "0.1.0";

        micasa = pkgs.buildGoModule {
          pname = "micasa";
          inherit version;
          src = ./.;
          subPackages = [ "cmd/micasa" ];
          vendorHash = "sha256-m7+UP6UEOYrLz4xz0YAVvzVieGFOpibPx9UEKT9roDk=";
          env.CGO_ENABLED = 0;
          ldflags = [
            "-X main.version=${version}"
          ];
        };

        licenseCheck = pkgs.writeShellScript "license-check" ''
          head=${pkgs.coreutils}/bin/head
          sed=${pkgs.gnused}/bin/sed
          grep=${pkgs.gnugrep}/bin/grep
          basename=${pkgs.coreutils}/bin/basename
          date=${pkgs.coreutils}/bin/date

          year=$($date +%Y)
          owner="Phillip Cloud"
          spdx="Licensed under the Apache License, Version 2.0"

          comment_prefix() {
            case "$1" in
              *.go|go.mod)  echo "//" ;;
              *.nix|*.yml|*.yaml|*.sh|.envrc|.gitignore) echo "#" ;;
              *.md)         echo "md" ;;
              *)            echo "#" ;;
            esac
          }

          status=0
          for f in "$@"; do
            name=$($basename "$f")
            pfx=$(comment_prefix "$name")

            if [ "$pfx" = "md" ]; then
              line1="<!-- Copyright $year $owner -->"
              line2="<!-- $spdx -->"
              year_pat="<!-- Copyright [0-9]\{4\} $owner -->"
            else
              line1="$pfx Copyright $year $owner"
              line2="$pfx $spdx"
              year_pat="$pfx Copyright [0-9]\{4\} $owner"
            fi

            first=$($head -n1 "$f")
            second=$($sed -n '2p' "$f")

            # Already correct
            if [ "$first" = "$line1" ] && [ "$second" = "$line2" ]; then
              continue
            fi

            # Header present with stale year -- bump it
            if echo "$first" | $grep -q "^$year_pat$" \
               && [ "$second" = "$line2" ]; then
              $sed -i "1s|$year_pat|$line1|" "$f"
              echo "bumped year in $f"
              continue
            fi

            # No header -- insert it
            $sed -i "1i\\$line1\n$line2\n" "$f"
            echo "added license header to $f"
            status=1
          done
          exit $status
        '';

        preCommit = git-hooks.lib.${system}.run {
          src = ./.;
          hooks = {
            golines = {
              enable = true;
              settings.flags = "--base-formatter=${pkgs.gofumpt}/bin/gofumpt " + "--max-len=100";
            };
            golangci-lint.enable = true;
            license-header = {
              enable = true;
              name = "license-header";
              entry = "${licenseCheck}";
              files = "\\.(go|nix|ya?ml|sh|md)$|^\\.envrc$|\\.gitignore$|^go\\.mod$";
              excludes = ["LICENSE" "flake\\.lock" "go\\.sum" "\\.json$"];
              language = "system";
              pass_filenames = true;
            };
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
            CGO_ENABLED = "0";
            packages = [
              pkgs.go
              pkgs.gopls
              pkgs.git
              pkgs.tokei
              pkgs.fd
              pkgs.ripgrep-all
            ]
            ++ enabledPackages;
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
              Labels = {
                "org.opencontainers.image.description" =
                  "Terminal UI for managing home projects and maintenance";
                "org.opencontainers.image.source" =
                  "https://github.com/micasa/micasa";
                "org.opencontainers.image.url" = "https://micasa.dev";
                "org.opencontainers.image.licenses" = "Apache-2.0";
              };
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
