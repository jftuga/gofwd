# gofwd
A cross-platform TCP port forwarder with Duo 2FA and Geographic IP integration.

Stand-alone, single-file executables for Windows, MacOS, and Linux can be downloaded from [Releases](https://github.com/jftuga/gofwd/releases).


## Usage

```
usage: gofwd [<flags>]

Flags:
      --help               Show context-sensitive help (also try --help-long and
                           --help-man).
  -i, --int                list local interface IP addresses
  -f, --from=FROM          from address:port
  -t, --to=TO              to address:port
      --examples           show command line example and then exit
      --version            show version and then exit
      --city=CITY          only accept incoming connections that originate from
                           given city
      --region=REGION      only accept incoming connections that originate from
                           given region (eg: state)
      --country=COUNTRY    only accept incoming connections that originate from
                           given 2 letter country abbreviation
  -l, --loc=LOC            only accept incoming connections from within a
                           geographic radius given in LAT,LON
  -d, --distance=DISTANCE  only accept incoming connections from within the
                           distance (in miles)
      --duo=DUO            path to duo ini config file and duo username; format:
                           filename:user (see --examples)
```

## Examples

```
+-------------------------------------------------------------------+-------------------------------------------------------------------------+
|                              EXAMPLE                              |                                 COMMAND                                 |
+-------------------------------------------------------------------+-------------------------------------------------------------------------+
| get the local IP address *(run this first)*, eg: 1.2.3.4          | gofwd -i                                                                |
| forward from a bastion host to an internal server                 | gofwd -f 1.2.3.4:22 -t 192.168.192.1.1:22                               |
| allow only if the remote IP is within 50 miles of this host       | gofwd -f 1.2.3.4:22 -t 192.168.192.1.1:22 -d 50                         |
| allow only if remote IP is located in Denver, CO                  | gofwd -f 1.2.3.4:22 -t 192.168.192.1.1:22 -city Denver -region Colorado |
| allow only if remote IP is located in Canada                      | gofwd -f 1.2.3.4:22 -t 192.168.192.1.1:22 -country CA                   |
| allow only if remote IP is located within 75 miles of Atlanta, GA | gofwd -f 1.2.3.4:22 -t 192.168.192.1.1:22 -l 33.756529,-84.400996 -d 75 |
|     to get Latitude, Longitude use https://www.latlong.net/       |                                                                         |
| allow only for a successful two-factor duo auth for 'testuser'    | gofwd -f 1.2.3.4:22 -t 192.168.192.1.1:22 --duo duo.ini:testuser        |
+-------------------------------------------------------------------+-------------------------------------------------------------------------+
```

## Two Factor Authentication (2FA) via Duo

### Basic Setup
* https://duo.com/
* `gofwd` will only work with a single Duo user; therefore, only one person will be able to access the resource behind `gofwd`.
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
* Create a `duo.ini` file with the **user name** as an ini section heading (created under *Basic Setup*)
* * Use the **Integration Key**, **Secret Key**, and **API HostName** in your .ini file.
* * Example: [duo-example.ini](https://github.com/jftuga/gofwd/blob/master/duo-example.ini)

### Running with Duo
* See the *Examples* section to see how to run `gofwd` with duo authentication enabled


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
