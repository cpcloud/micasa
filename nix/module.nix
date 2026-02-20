# Copyright 2026 Phillip Cloud
# Licensed under the Apache License, Version 2.0

# NixOS module that serves micasa over SSH.
#
# Users connect with `ssh micasa@<host>` and land directly in the TUI.
# All forwarding and tunneling is disabled for the service user.
#
# Usage in a NixOS configuration:
#
#   services.micasa = {
#     enable = true;
#     package = inputs.micasa.packages.${pkgs.system}.default;
#     authorizedKeys = [ "ssh-ed25519 AAAA..." ];
#   };
{
  config,
  lib,
  pkgs,
  ...
}:
let
  cfg = config.services.micasa;
in
{
  options.services.micasa = {
    enable = lib.mkEnableOption "micasa SSH service";

    package = lib.mkOption {
      type = lib.types.package;
      description = "The micasa package to use.";
    };

    user = lib.mkOption {
      type = lib.types.str;
      default = "micasa";
      description = "User account for the micasa SSH service.";
    };

    group = lib.mkOption {
      type = lib.types.str;
      default = "micasa";
      description = "Group for the micasa service user.";
    };

    dataDir = lib.mkOption {
      type = lib.types.path;
      default = "/var/lib/micasa";
      description = "Directory where the micasa database is stored.";
    };

    authorizedKeys = lib.mkOption {
      type = lib.types.listOf lib.types.str;
      default = [ ];
      description = "SSH public keys authorized to access micasa.";
    };
  };

  config = lib.mkIf cfg.enable {
    users.users.${cfg.user} = {
      isSystemUser = true;
      inherit (cfg) group;
      home = cfg.dataDir;
      createHome = true;
      shell = pkgs.bashInteractive;
      openssh.authorizedKeys.keys = cfg.authorizedKeys;
    };

    users.groups.${cfg.group} = { };

    services.openssh.extraConfig = lib.mkAfter ''
      Match User ${cfg.user}
        ForceCommand ${cfg.package}/bin/micasa --db ${cfg.dataDir}/micasa.db
        AllowTcpForwarding no
        AllowAgentForwarding no
        AllowStreamLocalForwarding no
        X11Forwarding no
        PermitTunnel no
        PermitUserEnvironment no
        PermitTTY yes
    '';
  };
}
