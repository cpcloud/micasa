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
        version = "0.1.0";

        micasa = pkgs.buildGoModule {
          pname = "micasa";
          inherit version;
          src = ./.;
          subPackages = [ "cmd/micasa" ];
          vendorHash = "sha256-NfXXLvq0MOU1vmOyEWBBEbl7Faf4o1WfKIrg36bu2OE=";
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
              pkgs.gopls
              pkgs.git
              pkgs.tokei
              pkgs.fd
              pkgs.ripgrep-all
              pkgs.hugo
            ]
            ++ enabledPackages;
          };

        packages = {
          inherit micasa;
          default = micasa;
          website = pkgs.writeShellScriptBin "micasa-website" ''
            rm -rf website/docs
            ${pkgs.hugo}/bin/hugo --source docs --baseURL /docs/ --destination ../website/docs >/dev/null 2>&1
            ${pkgs.python3}/bin/python3 -m http.server 0 -d website
          '';
          docs = pkgs.writeShellScriptBin "micasa-docs" ''
            ${pkgs.hugo}/bin/hugo server --source docs --baseURL /docs/ --bind 0.0.0.0
          '';
          record-demo = pkgs.writeShellApplication {
            name = "record-demo";
            runtimeInputs = [
              micasa
              pkgs.asciinema
              pkgs.asciinema-agg
              pkgs.tmux
              pkgs.fontconfig
              pkgs.jetbrains-mono
            ];
            text = ''
              CAST_FILE="$(mktemp --suffix=.cast)"
              GIF_FILE="images/demo.gif"
              SESSION="micasa-demo"
              COLS=132
              ROWS=42

              cleanup() {
                tmux kill-session -t "$SESSION" 2>/dev/null || true
              }
              trap cleanup EXIT
              cleanup

              send() { tmux send-keys -t "$SESSION" "$@"; }
              pause() { sleep "''${1:-0.5}"; }

              tmux new-session -d -s "$SESSION" -x "$COLS" -y "$ROWS"

              send "TERM=xterm-256color asciinema rec --cols $COLS --rows $ROWS --overwrite '$CAST_FILE' -c 'micasa --demo'" Enter
              pause 3

              # Projects tab
              pause 2
              send j; pause 0.6
              send j; pause 0.6
              send j; pause 0.6
              send j; pause 1

              send l; pause 0.5
              send l; pause 0.5
              send s; pause 1.2
              send s; pause 1.5

              # Maintenance tab
              send Tab; pause 2
              send j; pause 0.5
              send j; pause 0.5

              send l; pause 0.4
              send l; pause 0.4
              send l; pause 0.4
              send l; pause 0.4
              send l; pause 1

              send Enter; pause 3
              send Escape; pause 1.5

              # Appliances tab
              send Tab; pause 2
              send j; pause 0.5
              send j; pause 0.5
              send j; pause 1

              # House profile
              send H; pause 3
              send H; pause 1.5

              # Help overlay
              send ?; pause 4
              send Escape; pause 1

              # Quit
              send q; pause 1

              echo "Converting to GIF..."
              FONT_FILE=$(fc-list : file family | grep -i "JetBrains Mono" | head -1 | cut -d: -f1)
              FONT_DIR=$(dirname "$FONT_FILE")
              agg --font-dir "$FONT_DIR" \
                  --font-family "JetBrains Mono" \
                  --theme dracula \
                  --cols "$COLS" --rows "$ROWS" \
                  "$CAST_FILE" "$GIF_FILE"

              echo "Done: $CAST_FILE and $GIF_FILE"
            '';
          };
          capture-screenshots = pkgs.writeShellApplication {
            name = "capture-screenshots";
            runtimeInputs = [
              micasa
              pkgs.asciinema
              pkgs.asciinema-agg
              pkgs.tmux
              pkgs.imagemagick
            ];
            text = ''
              OUT="docs/static/images"
              mkdir -p "$OUT"
              SESSION="micasa-screenshots"
              COLS=120
              ROWS=36

              cleanup() {
                tmux kill-session -t "$SESSION" 2>/dev/null || true
              }
              trap cleanup EXIT

              send() { tmux send-keys -t "$SESSION" "$@"; }
              pause() { sleep "''${1:-0.5}"; }

              FONT_DIR="${pkgs.jetbrains-mono}/share/fonts/truetype"
              AGG_OPTS=(
                --font-dir "$FONT_DIR"
                --font-family "JetBrains Mono"
                --theme dracula
                --cols "$COLS" --rows "$ROWS"
                --no-loop
              )

              # Capture a single screenshot:
              #   capture <name> <setup_delay> <key_sequence_fn>
              # Records a short asciinema cast, converts to GIF, extracts last frame as PNG.
              capture() {
                local name="$1"
                local setup_delay="$2"
                local cast_file gif_file png_file
                cast_file="$(mktemp --suffix=.cast)"
                gif_file="$(mktemp --suffix=.gif)"
                png_file="$OUT/$name.png"

                cleanup

                tmux new-session -d -s "$SESSION" -x "$COLS" -y "$ROWS"
                send "TERM=xterm-256color asciinema rec --cols $COLS --rows $ROWS --overwrite '$cast_file' -c 'micasa --demo'" Enter
                pause "$setup_delay"

                # Run the key sequence (caller sets this up via the third arg eval)
                eval "$3"

                # Hold final state briefly so agg captures it clearly
                pause 1.5

                # Quit micasa and stop recording
                send q
                pause 1

                echo "  rendering $name..."
                agg "''${AGG_OPTS[@]}" "$cast_file" "$gif_file" 2>/dev/null

                # Extract last frame from GIF as PNG
                convert "''${gif_file}[-1]" "$png_file"

                rm -f "$cast_file" "$gif_file"
                echo "  -> $png_file"
              }

              echo "Capturing screenshots..."

              # 1. Dashboard (appears on launch)
              capture "dashboard" 4 "true"

              # 2. Projects table (D dismisses dashboard without switching tabs)
              capture "projects" 4 '
                send D; pause 1.5
                send j; pause 0.3
                send j; pause 0.3
                send j; pause 0.3
              '

              # 3. Quotes table
              capture "quotes" 4 '
                send D; pause 0.5
                send Tab; pause 1.5
                send j; pause 0.3
                send j; pause 0.3
              '

              # 4. Maintenance table
              capture "maintenance" 4 '
                send D; pause 0.5
                send Tab; pause 0.5
                send Tab; pause 1.5
                send j; pause 0.3
                send j; pause 0.3
              '

              # 5. Service log drilldown (maintenance > Log column > enter)
              capture "service-log" 4 '
                send D; pause 0.5
                send Tab; pause 0.5
                send Tab; pause 1
                send l; pause 0.3
                send l; pause 0.3
                send l; pause 0.3
                send l; pause 0.3
                send l; pause 0.3
                send l; pause 0.3
                send l; pause 0.5
                send Enter; pause 1.5
              '

              # 6. Appliances table
              capture "appliances" 4 '
                send D; pause 0.5
                send Tab; pause 0.5
                send Tab; pause 0.5
                send Tab; pause 1.5
                send j; pause 0.3
                send j; pause 0.3
              '

              # 7. House profile expanded
              capture "house-profile" 4 '
                send D; pause 0.5
                send H; pause 1.5
              '

              # 8. Help overlay
              capture "help" 4 '
                send D; pause 0.5
                send "?"; pause 1.5
              '

              # 9. Sorting indicators
              capture "sorting" 4 '
                send D; pause 0.5
                send l; pause 0.3
                send l; pause 0.3
                send s; pause 0.8
                send l; pause 0.3
                send s; pause 0.8
              '

              echo ""
              echo "Done! Screenshots in $OUT/"
              ls -la "$OUT/"*.png
            '';
          };
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
                  "https://github.com/cpcloud/micasa";
                "org.opencontainers.image.url" = "https://micasa.dev";
                "org.opencontainers.image.licenses" = "Apache-2.0";
              };
            };
          };
        };

        apps = {
          default = flake-utils.lib.mkApp { drv = micasa; };
          website = flake-utils.lib.mkApp { drv = self.packages.${system}.website; };
          record-demo = flake-utils.lib.mkApp { drv = self.packages.${system}.record-demo; };
          docs = flake-utils.lib.mkApp { drv = self.packages.${system}.docs; };
          capture-screenshots = flake-utils.lib.mkApp { drv = self.packages.${system}.capture-screenshots; };
        };

        formatter = pkgs.nixpkgs-fmt;
      }
    );
}
