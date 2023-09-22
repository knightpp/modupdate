{
  dockerTools,
  buildEnv,
  modupdate,
  go,
  cacert,
}:
# docker run --rm -i --tty -v (pwd):/src modupdate
dockerTools.buildImage {
  name = "modupdate";
  tag = "latest";
  # created = "now"; # if you want correct timestamp
  copyToRoot = buildEnv {
    name = "modupdate-root";
    paths = [
      modupdate
      go
      cacert # x509 certificates to pull from https
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
}
