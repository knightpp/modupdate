let
  pkgs = import <nixpkgs> { };
in
pkgs.callPackage ./modupdate.nix { }
