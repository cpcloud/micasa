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
      self,
      nixpkgs,
      flake-utils,
      git-hooks,
      ...
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs { inherit system; };
        version = builtins.replaceStrings [ "\n" "\r" ] [ "" "" ] (builtins.readFile ./VERSION);

        micasa = pkgs.buildGoModule {
          pname = "micasa";
          inherit version;
          src = ./.;
          subPackages = [ "cmd/micasa" ];
          vendorHash = "sha256-JkhY04nP3aWuk7YvF3dkK4ZfRxGfdv7hgK0YlX4Dg5w=";
          env.CGO_ENABLED = 0;
          preCheck = ''
            export HOME="$(mktemp -d)"
          '';
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
              excludes = ["LICENSE" "flake\\.lock" "go\\.sum" "\\.json$" "^docs/content/"];
              language = "system";
              pass_filenames = true;
            };
          };
        };

        # Fontconfig for VHS recordings using Hack Nerd Font.
        # JetBrains Mono's variable font files cause xterm.js in Chromium to
        # miscalculate cell width, producing visible letter-spacing gaps.
        # Hack Nerd Font renders correctly and includes icon glyphs.
        vhsFontsConf = pkgs.makeFontsConf {
          fontDirectories = [ "${pkgs.nerd-fonts.hack}/share/fonts/truetype" ];
        };

        vhsFontSetup = ''
          export FONTCONFIG_FILE="${vhsFontsConf}"
        '';

        deadcode = pkgs.buildGoModule {
          pname = "deadcode";
          version = "0.42.0";
          src = pkgs.fetchFromGitHub {
            owner = "golang";
            repo = "tools";
            rev = "v0.42.0";
            hash = "sha256-0RiinnIocPaj8Z5jtYGkbFiRf1BXyap4Z8e/sw2FBgg=";
          };
          subPackages = [ "cmd/deadcode" ];
          vendorHash = "sha256-oYmM+5lNmlP2i78NsG3v4WRhAUbiwS+EFkiicI6MKXA=";
          doCheck = false;
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
              pkgs.osv-scanner
              pkgs.git
              pkgs.hugo
              pkgs.vhs
            ]
            ++ enabledPackages;
          };

        packages = {
          inherit micasa;
          default = micasa;
          docs = pkgs.writeShellApplication {
            name = "micasa-docs";
            runtimeInputs = [ pkgs.hugo ];
            text = ''
              mkdir -p docs/static/images
              cp images/favicon.svg docs/static/images/favicon.svg
              cp images/demo.webp docs/static/images/demo.webp
              rm -rf website
              hugo --source docs --destination ../website
            '';
          };
          website = pkgs.writeShellApplication {
            name = "micasa-website";
            runtimeInputs = [ pkgs.hugo ];
            text = ''
              mkdir -p docs/static/images
              cp images/favicon.svg docs/static/images/favicon.svg
              cp images/demo.webp docs/static/images/demo.webp

              _port=$((RANDOM % 10000 + 30000))
              printf 'http://localhost:%s\n' "$_port"
              exec hugo server --source docs --buildDrafts --port "$_port" --bind 0.0.0.0 &>/dev/null
            '';
          };
          # Records any VHS tape and converts the GIF output to WebP
          record-tape = pkgs.writeShellApplication {
            name = "record-tape";
            runtimeInputs = [
              micasa
              pkgs.vhs
              pkgs.nerd-fonts.hack
              pkgs.libwebp
            ];
            text = ''
              if [[ $# -ne 1 ]]; then
                echo "usage: record-tape <tape-file>" >&2
                exit 1
              fi

              tape="$1"

              gif_path=$(grep -m1 '^Output ' "$tape" | awk '{print $2}')
              if [[ -z "$gif_path" || "$gif_path" != *.gif ]]; then
                echo "error: tape must contain an Output directive ending in .gif" >&2
                exit 1
              fi

              webp_path="''${gif_path%.gif}.webp"

              ${vhsFontSetup}

              mkdir -p "$(dirname "$gif_path")"
              vhs "$tape"
              gif2webp -m 6 "$gif_path" -o "$webp_path"
              rm "$gif_path"
            '';
          };
          record-demo = pkgs.writeShellApplication {
            name = "record-demo";
            runtimeInputs = [ self.packages.${system}.record-tape ];
            text = ''
              record-tape docs/tapes/demo.tape
            '';
          };
          # Captures a single VHS tape to a WebP screenshot: capture-one <tape-file>
          capture-one = pkgs.writeShellApplication {
            name = "capture-one";
            runtimeInputs = [
              micasa
              pkgs.vhs
              pkgs.nerd-fonts.hack
              pkgs.imagemagick
            ];
            text = ''
              if [[ $# -ne 1 ]]; then
                echo "usage: capture-one <tape-file>" >&2
                exit 1
              fi

              tape="$1"
              name="$(basename "$tape" .tape)"
              OUT="docs/static/images"
              mkdir -p "$OUT"

              ${vhsFontSetup}

              tmpdir=$(mktemp -d)
              trap 'rm -rf "$tmpdir"' EXIT

              vhs "$tape"

              # Extract last frame from GIF as lossless WebP
              magick "$OUT/$name.gif" -coalesce "$tmpdir/frame-%04d.png"
              last=$(printf '%s\n' "$tmpdir/frame"-*.png | sort -t- -k2 -n | tail -1)
              magick "$last" -quality 100 -define webp:lossless=true "$OUT/$name.webp"
              rm -f "$OUT/$name.gif"

              echo "$name -> $OUT/$name.webp"
            '';
          };

          # Captures VHS tapes in parallel: capture-screenshots [name ...]
          capture-screenshots = pkgs.writeShellApplication {
            name = "capture-screenshots";
            runtimeInputs = [
              self.packages.${system}.capture-one
              pkgs.fd
              pkgs.parallel
            ];
            text = ''
              TAPES="docs/tapes"

              if [[ $# -gt 0 ]]; then
                # Named tapes in parallel
                printf '%s\n' "$@" | parallel --bar capture-one "$TAPES/{}.tape"
                exit
              fi

              # All tapes in parallel (skip demo and using-* animated tapes)
              ntapes=$(fd -e tape --exclude demo.tape --exclude 'using-*.tape' . "$TAPES" | wc -l)
              nprocs=$(nproc)
              jobs=$(( ntapes < nprocs ? ntapes : nprocs ))
              fd -e tape --exclude demo.tape --exclude 'using-*.tape' -0 . "$TAPES" \
                | parallel -0 -j"$jobs" --bar capture-one {}
            '';
          };
          # Records all animated demo tapes (using-*) in parallel
          record-animated = pkgs.writeShellApplication {
            name = "record-animated";
            runtimeInputs = [
              self.packages.${system}.record-tape
              pkgs.fd
              pkgs.parallel
            ];
            text = ''
              TAPES="docs/tapes"
              ntapes=$(fd -g 'using-*.tape' . "$TAPES" | wc -l)
              ntapes=$(fd -g 'using-*.tape' . "$TAPES" | wc -l)
              if [[ "$ntapes" -eq 0 ]]; then
                echo "no using-*.tape files found in $TAPES" >&2
                exit 1
              fi
              nprocs=$(nproc)
              jobs=$(( ntapes < nprocs ? ntapes : nprocs ))
              fd -g 'using-*.tape' -0 . "$TAPES" \
                | parallel -0 -j"$jobs" --bar record-tape {}
            '';
          };
          run-deadcode = pkgs.writeShellApplication {
            name = "run-deadcode";
            runtimeInputs = [ deadcode pkgs.go ];
            runtimeEnv.CGO_ENABLED = "0";
            text = ''
              export GOCACHE="''${GOCACHE:-$(mktemp -d)}"
              export GOMODCACHE="''${GOMODCACHE:-$(mktemp -d)}"
              deadcode -test ./...
            '';
          };
          run-osv-scanner = pkgs.writeShellApplication {
            name = "run-osv-scanner";
            runtimeInputs = [ pkgs.osv-scanner ];
            text = ''
              osv-scanner scan --config osv-scanner.toml --recursive .
            '';
          };
          run-pre-commit =
            let
              runtimePath = pkgs.lib.makeBinPath (
                [ pkgs.go pkgs.git ] ++ preCommit.enabledPackages
              );
            in
            pkgs.writeShellScriptBin "run-pre-commit" ''
              export PATH="${runtimePath}:$PATH"
              ${preCommit.shellHook}
              pre-commit run --all-files
            '';
          micasa-container = pkgs.dockerTools.buildImage {
            name = "micasa";
            tag = "latest";
            copyToRoot = root;
            config = {
              Entrypoint = [ "/bin/micasa" ];
              Labels = {
                "org.opencontainers.image.title" = "micasa";
                "org.opencontainers.image.description" =
                  "Terminal UI for managing home projects and maintenance";
                "org.opencontainers.image.source" =
                  "https://github.com/cpcloud/micasa";
                "org.opencontainers.image.url" = "https://micasa.dev";
                "org.opencontainers.image.documentation" =
                  "https://micasa.dev/docs/";
                "org.opencontainers.image.licenses" = "Apache-2.0";
              };
            };
          };
        };

        apps = {
          default = flake-utils.lib.mkApp { drv = micasa; };
          website = flake-utils.lib.mkApp { drv = self.packages.${system}.website; };
          record-tape = flake-utils.lib.mkApp { drv = self.packages.${system}.record-tape; };
          record-demo = flake-utils.lib.mkApp { drv = self.packages.${system}.record-demo; };
          docs = flake-utils.lib.mkApp { drv = self.packages.${system}.docs; };
          capture-one = flake-utils.lib.mkApp { drv = self.packages.${system}.capture-one; };
          capture-screenshots = flake-utils.lib.mkApp { drv = self.packages.${system}.capture-screenshots; };
          record-animated = flake-utils.lib.mkApp { drv = self.packages.${system}.record-animated; };
          deadcode = flake-utils.lib.mkApp { drv = self.packages.${system}.run-deadcode; };
          osv-scanner = flake-utils.lib.mkApp { drv = self.packages.${system}.run-osv-scanner; };
          pre-commit = flake-utils.lib.mkApp { drv = self.packages.${system}.run-pre-commit; };
        };

        formatter = pkgs.nixpkgs-fmt;
      }
    );
}
