<!-- Copyright 2026 Phillip Cloud -->
<!-- Licensed under the Apache License, Version 2.0 -->

# micasa SSH Service {#module-services-micasa}

Serves [micasa](https://micasa.dev) over SSH. Users connect with
`ssh micasa@<host>` and land directly in the terminal UI.

## Quick Start {#module-services-micasa-quickstart}

```nix
services.micasa = {
  enable = true;
  package = inputs.micasa.packages.${pkgs.system}.default;
  authorizedKeys = [ "ssh-ed25519 AAAA..." ];
};
```

## Security {#module-services-micasa-security}

The module creates a locked-down system user with key-only authentication.
All SSH forwarding and tunneling is disabled. The database directory is
`0700` and files are created with umask `0077`.
