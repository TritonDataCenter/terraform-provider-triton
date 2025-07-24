{
  description = "triton-go development environment";

  inputs.nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
  inputs.devshell.url = "github:numtide/devshell";
  inputs.flake-parts.url = "github:hercules-ci/flake-parts";

  outputs = inputs@{ self, flake-parts, devshell, nixpkgs, }:
    flake-parts.lib.mkFlake { inherit inputs; } {
      imports = [
        devshell.flakeModule
      ];

      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];

      perSystem = { pkgs, ... }: {
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
            nodePackages.triton
            terraform
          ];
          bash.extra = ''
            export GOPATH=~/.local/share/go
            export PATH=$GOPATH/bin:$PATH
          '';
        };
      };
    };
}
