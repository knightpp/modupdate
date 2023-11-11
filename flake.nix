{
  nixConfig = {
    extra-substituters = "https://modupdate.cachix.org";
    extra-trusted-public-keys = "modupdate.cachix.org-1:1gFM53SV2wsCnKk8nUVnuk23vu6PcXkyeAUN0p4Y4+M=";
  };

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs = {
    self,
    nixpkgs,
  }: let
    # Systems supported
    allSystems = [
      "x86_64-linux"
      "aarch64-linux"
      "x86_64-darwin"
      "aarch64-darwin"
    ];

    # Helper to provide system-specific attributes
    forAllSystems = f:
      nixpkgs.lib.genAttrs allSystems (system:
        f {
          pkgs = import nixpkgs {inherit system;};
        });
  in {
    # Development environment output
    devShells = forAllSystems ({pkgs}: {
      default = pkgs.mkShell {
        # The Nix packages provided in the environment
        packages = with pkgs; [
          go_1_21
          go-tools
          gopls
          delve
          gomodifytags
        ];
      };
    });

    packages =
      forAllSystems
      ({pkgs}: rec {
        default = pkgs.callPackage ./modupdate.nix {};

        # NOTE: Do not use this, it's just an example for my own use
        container = pkgs.callPackage ./container.nix {
          modupdate = default;
        };
      });
  };
}
