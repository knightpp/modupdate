{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-23.05";
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
          go_1_20
          gotools # Go tools like goimports, godoc, and others
        ];
      };
    });

    packages =
      forAllSystems
      ({pkgs}: let
        path = builtins.path {
          name = "modupdate";
          path = ./.;
        };
        gitRev =
          if (self ? rev)
          then self.rev
          else "dirty";
        modupdate = pkgs.buildGoModule {
          name = "modupdate";
          src = pkgs.nix-gitignore.gitignoreSource [] path;
          vendorSha256 = "sha256-XU4kLbEAPCL8mNk4omk2OIijKdiiAsJKBfoKkJJfHkI=";

          ldflags = [
            "-s"
            "-w"
            "-X 'main.version=${self.shortRev or ""}'"
            "-X 'main.vcsCommit=${gitRev}'"
          ];

          meta = with pkgs.lib; {
            description = "Tool to update direct dependencies in go.mod";
            homepage = "https://github.com/knightpp/modupdate";
            license = licenses.mit;
            maintainers = with maintainers; [knightpp];
          };
        };
        container =
          # docker run --rm -i --tty -v (pwd):/src modupdate
          pkgs.dockerTools.buildImage {
            name = "modupdate";
            tag = "latest";
            # created = "now"; # if you want correct timestamp
            copyToRoot = pkgs.buildEnv {
              name = "modupdate-root";
              paths = [
                modupdate
                pkgs.go
                pkgs.cacert # x509 certificates to pull from https
              ];
              pathsToLink = [
                "/bin"
                "/etc/ssl" # include x509 certificates
              ];
            };
            config = {
              Cmd = ["modupdate"];
              WorkingDir = "/src";
            };
          };
      in {
        default = modupdate;
        # NOTE: Do not use this, it's just an example for my own use
        inherit container;
      });
  };
}
