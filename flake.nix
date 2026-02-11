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
          vendorHash = "sha256-okcV/Lpj3DU2gmZxpNqpeD3MK16Oqxeao74Teolovrs=";
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
              pkgs.git
              pkgs.hugo
              pkgs.vhs
            ]
            ++ enabledPackages;
          };

        packages = {
          inherit micasa;
          default = micasa;
          build-docs = pkgs.writeShellApplication {
            name = "micasa-build-docs";
            runtimeInputs = [ pkgs.hugo ];
            text = ''
              mkdir -p website/images docs/static/images
              cp images/favicon.svg website/images/favicon.svg
              cp images/demo.gif website/images/demo.gif
              cp images/favicon.svg docs/static/images/favicon.svg
              rm -rf website/docs
              hugo --source docs --baseURL /docs/ --destination ../website/docs
            '';
          };
          website = pkgs.writeShellApplication {
            name = "micasa-website";
            runtimeInputs = [
              self.packages.${system}.build-docs
              pkgs.python3
            ];
            text = ''
              micasa-build-docs >/dev/null 2>&1
              python3 -m http.server 0 -d website
            '';
          };
          record-demo = pkgs.writeShellApplication {
            name = "record-demo";
            runtimeInputs = [
              micasa
              pkgs.vhs
              pkgs.jetbrains-mono
            ];
            text = ''
              TMPDIR=$(mktemp -d)
              trap 'rm -rf "$TMPDIR"' EXIT

              # Make JetBrains Mono visible to Chrome inside VHS
              FC_CONF="$TMPDIR/fonts.conf"
              cat > "$FC_CONF" <<FCXML
              <?xml version="1.0"?>
              <!DOCTYPE fontconfig SYSTEM "urn:fontconfig:fonts.dtd">
              <fontconfig>
                <include>/etc/fonts/fonts.conf</include>
                <dir>${pkgs.jetbrains-mono}/share/fonts</dir>
                <cachedir>$TMPDIR/fc-cache</cachedir>
              </fontconfig>
              FCXML
              export FONTCONFIG_FILE="$FC_CONF"

              vhs docs/tapes/demo.tape
              echo "Done: images/demo.gif"
            '';
          };
          # Captures a single VHS tape to PNG: capture-one <tape-file>
          capture-one = pkgs.writeShellApplication {
            name = "capture-one";
            runtimeInputs = [
              micasa
              pkgs.vhs
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

              TMPDIR=$(mktemp -d)
              trap 'rm -rf "$TMPDIR"' EXIT

              # Make JetBrains Mono visible to Chrome inside VHS
              FC_CONF="$TMPDIR/fonts.conf"
              cat > "$FC_CONF" <<FCXML
              <?xml version="1.0"?>
              <!DOCTYPE fontconfig SYSTEM "urn:fontconfig:fonts.dtd">
              <fontconfig>
                <include>/etc/fonts/fonts.conf</include>
                <dir>${pkgs.jetbrains-mono}/share/fonts</dir>
                <cachedir>$TMPDIR/fc-cache</cachedir>
              </fontconfig>
              FCXML
              export FONTCONFIG_FILE="$FC_CONF"

              vhs "$tape"

              # Extract last frame from GIF as PNG
              magick "$OUT/$name.gif" -coalesce "$TMPDIR/$name-frame-%04d.png"
              last=$(printf '%s\n' "$TMPDIR/$name-frame"-*.png | sort -t- -k3 -n | tail -1)
              mv "$last" "$OUT/$name.png"
              rm -f "$OUT/$name.gif"

              echo "$name -> $OUT/$name.png"
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

              # All tapes in parallel (skip demo -- it has its own command)
              ntapes=$(fd -e tape --exclude demo.tape . "$TAPES" | wc -l)
              nprocs=$(nproc)
              jobs=$(( ntapes < nprocs ? ntapes : nprocs ))
              fd -e tape --exclude demo.tape -0 . "$TAPES" \
                | parallel -0 -j"$jobs" --bar capture-one {}

              echo ""
              echo "Done! Screenshots in docs/static/images/"
              ls -la docs/static/images/*.png
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
          record-demo = flake-utils.lib.mkApp { drv = self.packages.${system}.record-demo; };
          build-docs = flake-utils.lib.mkApp { drv = self.packages.${system}.build-docs; };
          capture-one = flake-utils.lib.mkApp { drv = self.packages.${system}.capture-one; };
          capture-screenshots = flake-utils.lib.mkApp { drv = self.packages.${system}.capture-screenshots; };
          pre-commit = flake-utils.lib.mkApp { drv = self.packages.${system}.run-pre-commit; };
        };

        formatter = pkgs.nixpkgs-fmt;
      }
    );
}
