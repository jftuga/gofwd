# gofwd


## Description

`gofwd` is a cross-platform TCP port forwarder with Duo 2FA and Geographic IP integration. Its use case is to help protect services when using a VPN is not possible. Before a connection is forwarded, the remote IP address is geographically checked against city, region (state), and/or country.  Distance (in miles) can also be used.  If this condition is satisfied, a Duo 2FA request can then be sent to a mobile device. The connection is only forwarded after Duo has verified the user.

Stand-alone, single-file executables for Windows, MacOS, and Linux can be downloaded from [Releases](https://github.com/jftuga/gofwd/releases).


## Usage

```
usage: gofwd [<flags>]

Flags:
      --help                Show context-sensitive help (also try --help-long
                            and --help-man).
  -i, --int                 list local interface IP addresses
  -f, --from=FROM           from IP address:port
  -t, --to=TO               to IP address:port
      --examples            show command line example and then exit
      --version             show version and then exit
      --city=CITY           only accept incoming connections that originate from
                            given city
      --region=REGION       only accept incoming connections that originate from
                            given region (eg: state)
      --country=COUNTRY     only accept incoming connections that originate from
                            given 2 letter country abbreviation
  -l, --loc=LOC             only accept from within a geographic radius; format:
                            LATITUDE,LONGITUDE (use with --distance)
  -d, --distance=DISTANCE   only accept from within a given distance (in miles)
      --duo=DUO             path to duo ini config file and duo username;
                            format: filename:user (see --examples)
      --duo-cache-time=120  number of seconds to cache a successful Duo
                            authentication (default is 120)
  -p, --private             allow RFC1918 private addresses for the remote IP

```


## Examples

```
+-------------------------------------------------------------------+-----------------------------------------------------------------------------+
|                              EXAMPLE                              |                                   COMMAND                                   |
+-------------------------------------------------------------------+-----------------------------------------------------------------------------+
| get the local IP address *(run this first)*, eg: 1.2.3.4          | gofwd -i                                                                    |
| forward from a bastion host to an internal server                 | gofwd -f 1.2.3.4:22 -t 192.168.1.1:22                                       |
| allow only if the remote IP is within 50 miles of this host       | gofwd -f 1.2.3.4:22 -t 192.168.1.1:22 -d 50                                 |
| allow only if remote IP is located in Denver, CO                  | gofwd -f 1.2.3.4:22 -t 192.168.1.1:22 -city Denver -region Colorado         |
| allow only if remote IP is located in Canada                      | gofwd -f 1.2.3.4:22 -t 192.168.1.1:22 -country CA                           |
| allow only if remote IP is located within 75 miles of Atlanta, GA | gofwd -f 1.2.3.4:22 -t 192.168.1.1:22 -l 33.756529,-84.400996 -d 75         |
|     to get Latitude, Longitude use https://www.latlong.net/       |                                                                             |
| allow only for a successful two-factor duo auth for 'testuser'    | gofwd -f 1.2.3.4:22 -t 192.168.1.1:22 --duo duo.ini:testuser                |
| allow only after both Geo IP and Duo are verified                 | gofwd -f 1.2.3.4:22 -t 192.168.1.1:22 --region Texas --duo duo.ini:testuser |
+-------------------------------------------------------------------+-----------------------------------------------------------------------------+
```


## Two Factor Authentication (2FA) via Duo

### Basic Setup
* https://duo.com/
* `gofwd` will only work with a single Duo user; therefore, only one person will be able to access the resource residing behind `gofwd`.
* * Multiple `gofwd` instantiations can be used for different users.
* * The .ini configuration file supports multiple users *(see below)*.
* You will need to create a Duo account.  The free tier supports 10 users.
* Create a user and set their status to `Require two-factor authentication`. This is the default.
* * You should also add an email address and phone number.
* Install the Duo app on to your mobile device.

### Application Setup
* On the Duo website, click on Applications.
* Protect an Application
* Select `Partner Auth API`
* Under `Settings`, give your application a name such as `gofwd ssh` or `gofwd rdp`.
* Create a `duo.ini` file with the **user name** as an ini section heading (the one that you just created under *Basic Setup*)
* * Use the **Integration Key**, **Secret Key**, and **API HostName** to configure your .ini file.
* * Example: [duo-example.ini](https://github.com/jftuga/gofwd/blob/master/duo-example.ini)

### Running with Duo
* Add the ``--duo`` command line option
* * See the *Examples* section to see how to run `gofwd` with duo authentication enabled

## Docker

### Example Helper Scripts
* To build an image: [docker_build_image.sh](https://github.com/jftuga/gofwd/blob/master/docker_build_image.sh)
* * Your image will need to include these files used for DNS resolution:
* * * etc/ssl/certs/ca-bundle.cr
* * * etc/ssl/certs/ca-bundle.trust.crt
* To run the built image: [docker_start_gofwd.sh](https://github.com/jftuga/gofwd/blob/master/docker_start_gofwd.sh) *(Edit first)*
* To use `gofwd.exe` in Docker under Windows, consider using the [Microsoft Windows Nano Server](https://hub.docker.com/_/microsoft-windows-nanoserver) for containers.
* To use `gofwd` in Docker under Linux, consider using the [Scratch](https://hub.docker.com/_/scratch) for the container.

### Static Compilation - Docker Only
* Your version of `gofwd` will need to be statically compiled:

| Platform | Command
----------|-----
| windows | go build -tags netgo -ldflags "-extldflags -static"
| linux/bsd | go build -tags netgo -ldflags '-extldflags "-static" -s -w'
| macos | go build -ldflags '-s -extldflags "-sectcreate __TEXT __info_plist Info.plist"'
| android | go build -ldflags -s

**NOTE:** *I have not been able to test all of these*

## Docker Example
```
docker run -d --restart unless-stopped -p 4567:4567
    -v /home/ec2-user/duo.ini:/duo.ini \
    jftuga:gofwd:v050.1 -f 1.2.3.4:4567 -t 192.168.1.1:22 \
    --duo /duo.ini:jftuga -l `39.858706,-104.670732` -d 80
```

| Explanation | Parameter |
--------------|------------
| detach and run Docker in daemon mode | -d
| restart container unless explicitly stopped | --restart unless-stopped
| redirect external TCP port to internal TCP port | -p 4567:4567
| ini file is located on the host here: `/home/ec2-user/duo.ini` | -v `/home/ec2-user/duo.ini`:/duo.ini
| ini file is mounted inside the container here: `/duo.ini` | -v /home/ec2-user/duo.ini:/`duo.ini`
| container name and tag | jftuga:gofwd:v050.1
| external service is `1.2.3.4` on port `4567` | -f 1.2.3.4:4567 
| internal service is `192.168.1.1` on port `22` | -t 192.168.1.1:22
| duo config file is mounted within the container | --duo `/duo.ini`:jftuga
| duo user name | --duo /duo.ini:`jftuga`
| location: Denver, CO with coordinates of | -l 39.858706,-104.670732
| distance: `80 miles` from Denver | -d 80


**Note:** if you are running in a NAT environment, such as AWS, then you will need to append the `-p` option to allow RFC1918 private IPv4 addresses.


## chroot environment
* Please review chroot_start_gofwd.sh


## Acknowledgments
Some code was adopted from [The little forwarder that could](https://github.com/kintoandar/fwd/)

Other Go code used:

* Logging: https://go.uber.org/zap
* Command line arguments: https://gopkg.in/alecthomas/kingpin.v2
* Output tables: https://github.com/olekukonko/tablewriter
* Ini file: https://gopkg.in/ini.v1
* Network interfaces: https://github.com/jftuga/nics
* IP info: https://github.com/jftuga/ipinfo
* Duo API: https://github.com/duosecurity/duo_api_golang/authapi

## Future Work
* [Run the Docker daemon as a non-root user - Rootless Mode](https://docs.docker.com/engine/security/rootless/)
* [Docker Tips: Running a Container With a Non Root User](https://medium.com/better-programming/running-a-container-with-a-non-root-user-e35830d1f42a)
