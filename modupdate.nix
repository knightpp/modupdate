{
  buildGoModule,
  lib,
}:
buildGoModule {
  name = "modupdate";
  src = builtins.path {
    name = "modupdate";
    path = ./.;
    filter = path: type:
      builtins.elem (/. + path) [
        ./go.mod
        ./go.sum
        ./main.go
        ./version.go
      ];
  };
  vendorSha256 = "sha256-QWeXMgZUEg0fAyc4DS3QaWqv2VxDznhUP9Fl8GOuIYk=";

  ldflags = ["-s" "-w"];

  meta = with lib; {
    description = "Tool to update direct dependencies in go.mod";
    homepage = "https://github.com/knightpp/modupdate";
    license = licenses.mit;
    maintainers = with maintainers; [knightpp];
  };
}
