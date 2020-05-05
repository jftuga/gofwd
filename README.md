# gofwd


## Description

`gofwd` is a cross-platform TCP port forwarder with Duo 2FA and Geographic IP integration. Its use case is to help protect services when using a VPN is not possible. Before a connection is forwarded, the remote IP address is geographically checked against city, region (state), and/or country.  Distance (in miles) can also be used.  If this condition is satisfied, a Duo 2FA request can then be sent to a mobile device. The connection is only forwarded after Duo has verified the user.

Stand-alone, single-file executables for Windows, MacOS, and Linux can be downloaded from [Releases](https://github.com/jftuga/gofwd/releases).


## Use case

The standard `Duo 2FA` Windows RDP implementation issues the the second factor request after a RDP client has connected and issued a valid username / password combination.  The RDP port is always open to the Internet, which is a potential security issue.

`gofwd` uses `Duo 2FA` *before* forwarding the RDP connection to an internal system.  The big caveat with `gofwd` is that it only works well in single-user scenarios.  However, being able to access your home lab remotely fits in well with this.

`gofwd` can also be uses to protect SSH such as an AWS EC2 instance or Digital Ocean Droplet.

Both RDP to a home computer and remote SSH access work reliably well.  On a home network, `gofwd` can be run on a Raspberry Pi that forwards the connection to a Windows 10 system once Duo authentication is successful.  It can also run from within a Docker container for added security.

The Geo-IP feature is nice because it limits who can initiate a `Duo 2FA` request. If someone tries to connect to your RDP port but is not within the defined geo-ip fence, then a `Duo 2FA` will not be sent to your phone.  **Running on a non-standard port for RDP is recommended to limit the number of 2FA requests.**

 For example, you could use an 50 mile radius from your residence and you will probably not receive a `Duo 2FA` request from another person or bot.  Be aware that some mobile operators might issue you an IP address that is further away than expected.  The geo-ip fence can alternatively be defined based on city, region (state) and/or country or by using latitude, longitude coordinates. `gofwd` uses https://ipinfo.io/ to get this information in real time.

The overall elegance of this solution is that no additional software is needed.  As long as you are within your predefined geo-ip location, have your phone, and know your hostname/ip address (and port number), then you will be able to access your system remotely.

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
* To use `gofwd` in Docker under Linux, consider starting with [Scratch](https://hub.docker.com/_/scratch) for the container.

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
    --duo /duo.ini:jftuga -l 39.858706,-104.670732 -d 80
```

| Explanation | Parameter |
--------------|------------
| detach and run Docker in daemon mode | -d
| restart container (on error) unless explicitly stopped | --restart unless-stopped
| redirect external TCP port to internal TCP port | -p 4567:4567
| ini file is located on the host here: `/home/ec2-user/duo.ini` | -v `/home/ec2-user/duo.ini`:/duo.ini
| ini file is mounted inside the container here: `/duo.ini` | -v /home/ec2-user/duo.ini:/`duo.ini`
| container name and tag | jftuga:gofwd:v050.1
| external service is `1.2.3.4` on port `4567` | -f 1.2.3.4:4567 
| internal service is `192.168.1.1` on port `22` | -t 192.168.1.1:22
| duo config file is mounted within the container | --duo `/duo.ini`:jftuga
| duo user name | --duo /duo.ini:`jftuga`
| location: use coordinates for Denver, CO | -l 39.858706,-104.670732
| distance: `80 miles` from Denver | -d 80


**Note:** if you are running in a NAT environment, such as AWS, then you will need to include the `-p` option to allow RFC1918 private IPv4 addresses.


## chroot environment
* Please review [chroot_start_gofwd.sh](https://github.com/jftuga/gofwd/blob/master/chroot_start_gofwd.sh)


## Acknowledgments
* Some code was adopted from [The little forwarder that could](https://github.com/kintoandar/fwd/)
* `gofwd`uses https://ipinfo.io/ to get Geo IP information in real time

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
