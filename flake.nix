{
  description = "triton terraform provider";

  inputs.nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
  inputs.devshell.url = "github:numtide/devshell";
  inputs.devshell.inputs.nixpkgs.follows = "nixpkgs";
  inputs.flake-parts.url = "github:hercules-ci/flake-parts";

  outputs = inputs@{ self, flake-parts, devshell, nixpkgs, }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      imports = [
        devshell.flakeModule
      ];

      systems = [
        "x86_64-linux"
        "aarch64-linux"
      ];

      perSystem = { pkgs, system, ... }: {
        # https://github.com/hercules-ci/flake-parts/pull/129
        # This sets `pkgs` to a nixpkgs with allowUnfree option set.
        _module.args.pkgs = import nixpkgs {
          inherit system;
          config.allowUnfree = true;
        };

        devshells.default = {
          # Add additional packages you'd like to be available in your devshell
          # PATH here
          packages = with pkgs; [
            go
            goreleaser
            errcheck
            go-tools
            gnumake
            golangci-lint
            gopls
            opentofu
            terraform
            nodePackages.triton
          ];
          bash.extra = ''
            export GOPATH=~/.local/share/go
            export PATH=$GOPATH/bin:$PATH
          '';
        };
      };
    };
}
