# reg-server

A static UI for a docker registry. Comes with vulnerability scanning if you
have a [CoreOS Clair](https://github.com/coreos/clair) server set up.

Demo at [r.j3ss.co](https://r.j3ss.co).

## Usage

```console
$ reg-server -h
NAME:
   reg-server - Docker registry v2 static UI server.

USAGE:
   reg-server [global options] command [command options] [arguments...]

VERSION:
   v0.1.0

AUTHOR:
   The Genuinetools Authors <no-reply@butts.com>

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --debug, -d                 run in debug mode
   --username value, -u value  username for the registry
   --password value, -p value  password for the registry
   --registry value, -r value  URL to the private registry (ex. r.j3ss.co)
   --insecure, -k              do not verify tls certificates of registry
   --port value                port for server to run on (default: "8080")
   --cert value                path to ssl cert
   --key value                 path to ssl key
   --interval value            interval to generate new index.html's at (default: "1h")
   --clair value               url to clair instance
   --help, -h                  show help
   --version, -v               print the version
```

## Screenshots

![home.png](home.png)

![vuln.png](vuln.png)
