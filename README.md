# gofwd
A TCP port forwarder with Duo and Geographic IP integration

## Usage

```
usage: gofwd [<flags>]

Flags:
      --help               Show context-sensitive help (also try --help-long and
                           --help-man).
  -i, --int                list local interface IP addresses
  -f, --from=FROM          from address:port
  -t, --to=TO              to address:port
      --examples           show command line example
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
| to get Latitude, Longitude use https://www.latlong.net/           |                                                                         |
+-------------------------------------------------------------------+-------------------------------------------------------------------------+
```

## Duo Auth API (work in progress)
* You will need to create a Duo account.  The free tier supports 10 users.
* https://duo.com/docs/authapi

## Acknowledgments
Some code was adopted from [The little forwarder that could](https://github.com/kintoandar/fwd/)
