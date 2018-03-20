# reg

[![Travis CI](https://travis-ci.org/genuinetools/reg.svg?branch=master)](https://travis-ci.org/genuinetools/reg)

Docker registry v2 command line client.

- [Installation](#installation)
- [Usage](#usage)
- [Auth](#auth)
- [List Repositories and Tags](#list-repositories-and-tags)
- [Get a Manifest](#get-a-manifest)
- [Download a Layer](#download-a-layer)
- [Delete an Image](#delete-an-image)
- [Vulnerability Reports](#vulnerability-reports)
- [Testing](#testing)

## Installation

#### Binaries

- **darwin** [386](https://github.com/genuinetools/reg/releases/download/v0.13.0/reg-darwin-386) / [amd64](https://github.com/genuinetools/reg/releases/download/v0.13.0/reg-darwin-amd64)
- **linux** [386](https://github.com/genuinetools/reg/releases/download/v0.13.0/reg-linux-386) / [amd64](https://github.com/genuinetools/reg/releases/download/v0.13.0/reg-linux-amd64) / [arm](https://github.com/genuinetools/reg/releases/download/v0.13.0/reg-linux-arm) / [arm64](https://github.com/genuinetools/reg/releases/download/v0.13.0/reg-linux-arm64)
- **windows** [386](https://github.com/genuinetools/reg/releases/download/v0.13.0/reg-windows-386) / [amd64](https://github.com/genuinetools/reg/releases/download/v0.13.0/reg-windows-amd64)

#### Via Go

```bash
$ go get github.com/genuinetools/reg
```

## Usage

```console
$ reg
NAME:
   reg - Docker registry v2 client.

USAGE:
   reg [global options] command [command options] [arguments...]

VERSION:
   version v0.13.0, build 3b7dafb

AUTHOR:
   The Genuinetools Authors <no-reply@butts.com>

COMMANDS:
     delete, rm       delete a specific reference of a repository
     layer, download  download a layer for the specific reference of a repository
     list, ls         list all repositories
     manifest         get the json manifest for the specific reference of a repository
     tags             get the tags for a repository
     vulns            get a vulnerability report for the image from CoreOS Clair
     help, h          Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --debug, -d                 run in debug mode
   --insecure, -k              do not verify tls certificates
   --force-non-ssl, -f         force allow use of non-ssl
   --username value, -u value  username for the registry
   --password value, -p value  password for the registry
   --registry value, -r value  URL to the private registry (ex. r.j3ss.co) (default: "https://registry-1.docker.io") [$REG_REGISTRY]
   --help, -h                  show help
   --version, -v               print the version
```

Note that the `--registry` can be set by an environment variable `REG_REGISTRY`, so you can set this in your shell login scripts.
Specifying the registry on the command-line will override an environment variable setting.

## Auth

`reg` will automatically try to parse your docker config credentials, but if
not present, you can pass through flags directly.

## List Repositories and Tags

**Repositories**

```console
# this command might take a while if you have hundreds of images like I do
$ reg -r r.j3ss.co ls
Repositories for r.j3ss.co
REPO                  TAGS
awscli                latest
beeswithmachineguns   latest
camlistore            latest
chrome                beta, latest, stable
...
```

**Tags**

```console
$ reg tags tor-browser
alpha
hardened
latest
stable
```

## Get a Manifest

```console
$ reg manifest htop
{
   "schemaVersion": 1,
   "name": "htop",
   "tag": "latest",
   "architecture": "amd64",
   "fsLayers": [
     {
       "blobSum": "sha256:a3ed95caeb02ffe68cdd9fd84406680ae93d633cb16422d00e8a7c22955b46d4"
     },
     ....
   ],
   "history": [
     ....
   ]
 }
```

## Download a Layer

```console
$ reg layer -o chrome@sha256:a3ed95caeb02ffe68cdd9fd84406680ae93d633cb16422d00e8a7c22955b46d4
OR
$ reg layer chrome@sha256:a3ed95caeb0.. > layer.tar
```


## Delete an Image

```console
$ reg rm chrome@sha256:a3ed95caeb02ffe68cdd9fd84406680ae93d633cb16422d00e8a7c22955b46d4
Deleted chrome@sha256:a3ed95caeb02ffe68cdd9fd84406680ae93d633cb16422d00e8a7c22955b46d4
```

## Vulnerability Reports

```console
$ reg vulns --clair https://clair.j3ss.co chrome
Found 32 vulnerabilities
CVE-2015-5180: [Low]

https://security-tracker.debian.org/tracker/CVE-2015-5180
-----------------------------------------
CVE-2016-9401: [Low]
popd in bash might allow local users to bypass the restricted shell and cause a use-after-free via a crafted address.
https://security-tracker.debian.org/tracker/CVE-2016-9401
-----------------------------------------
CVE-2016-3189: [Low]
Use-after-free vulnerability in bzip2recover in bzip2 1.0.6 allows remote attackers to cause a denial of service (crash) via a crafted bzip2 file, related to block ends set to before the start of the block.
https://security-tracker.debian.org/tracker/CVE-2016-3189
-----------------------------------------
CVE-2011-3389: [Medium]
The SSL protocol, as used in certain configurations in Microsoft Windows and Microsoft Internet Explorer, Mozilla Firefox, Google Chrome, Opera, and other products, encrypts data by using CBC mode with chained initialization vectors, which allows man-in-the-middle attackers to obtain plaintext HTTP headers via a blockwise chosen-boundary attack (BCBA) on an HTTPS session, in conjunction with JavaScript code that uses (1) the HTML5 WebSocket API, (2) the Java URLConnection API, or (3) the Silverlight WebClient API, aka a "BEAST" attack.
https://security-tracker.debian.org/tracker/CVE-2011-3389
-----------------------------------------
CVE-2016-5318: [Medium]
Stack-based buffer overflow in the _TIFFVGetField function in libtiff 4.0.6 and earlier allows remote attackers to crash the application via a crafted tiff.
https://security-tracker.debian.org/tracker/CVE-2016-5318
-----------------------------------------
CVE-2016-9318: [Medium]
libxml2 2.9.4 and earlier, as used in XMLSec 1.2.23 and earlier and other products, does not offer a flag directly indicating that the current document may be read but other files may not be opened, which makes it easier for remote attackers to conduct XML External Entity (XXE) attacks via a crafted document.
https://security-tracker.debian.org/tracker/CVE-2016-9318
-----------------------------------------
CVE-2015-7554: [High]
The _TIFFVGetField function in tif_dir.c in libtiff 4.0.6 allows attackers to cause a denial of service (invalid memory write and crash) or possibly have unspecified other impact via crafted field data in an extension tag in a TIFF image.
https://security-tracker.debian.org/tracker/CVE-2015-7554
-----------------------------------------
Unknown: 2
Negligible: 23
Low: 3
Medium: 3
High: 1
```

## Testing

If you plan on contributing you should be able to run the tests locally. The
tests run for CI via docker-in-docker. But running locally with `go test`, you
need to make one modification to your docker daemon config so that you can talk
to the local registry for the tests.

Add the flag `--insecure-registry localhost:5000` to your docker daemon,
documented [here](https://docs.docker.com/registry/insecure/) for testing
against an insecure registry.

OR run `make dind dtest` to avoid having to change your local docker config and
to run the tests as docker-in-docker.
