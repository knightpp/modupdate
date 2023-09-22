{
  buildGoModule,
  lib,
  nix-gitignore,
  self,
}: let
  path = builtins.path {
    name = "modupdate";
    path = ./.;
  };
  gitRev =
    if (self ? rev)
    then self.rev
    else "dirty";
in
  buildGoModule {
    name = "modupdate";
    src = nix-gitignore.gitignoreSource [] path;
    vendorSha256 = "sha256-XU4kLbEAPCL8mNk4omk2OIijKdiiAsJKBfoKkJJfHkI=";

    ldflags = [
      "-s"
      "-w"
      "-X 'main.version=${self.shortRev or ""}'"
      "-X 'main.vcsCommit=${gitRev}'"
    ];

    meta = with lib; {
      description = "Tool to update direct dependencies in go.mod";
      homepage = "https://github.com/knightpp/modupdate";
      license = licenses.mit;
      maintainers = with maintainers; [knightpp];
    };
  }
