provider "docker" {
  host = coalesce(
    var.docker_host,
    var.operating_system == "Windows" ? "npipe:////.//pipe//docker_engine" : "unix:///var/run/docker.sock",
  )
}

provider "local" {}
