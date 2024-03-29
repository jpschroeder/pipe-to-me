
PIPE TO ME
==========

Stream data over http using curl.  Hosted at [pipeto.me](https://pipeto.me/)

- No javascript
- Go standard library only

![](demo.gif)

```
pipeto.me(1)                     PIPE TO ME                         pipeto.me(1)

NAME
    pipeto.me: streaming data over http

SYNOPSIS
    Randomly generated pipe address:                  https://pipeto.me/<key>

EXAMPLES
    Pipe example:

    (chrome/firefox): https://pipeto.me/<key>
    (terminal2)$ echo hello world | curl -T- https://pipeto.me/<key>

    Input example:

    (chrome/firefox): https://pipeto.me/<key>
    (terminal)$ curl -T- https://pipeto.me/<key>
                hello world<enter>

    File transfer example:

    (terminal1)$ curl https://pipeto.me/<key> > output.txt
    (terminal2)$ cat input.txt | curl -T- https://pipeto.me/<key>

    Chat example(curl>=7.68):

    (terminal1)$ curl -T. -u user1: https://pipeto.me/<key>?mode=interactive
    (terminal2)$ curl -T. -u user2: https://pipeto.me/<key>?mode=interactive
                 hello world<enter>

DESCRIPTION
    Data is not buffered or stored in any way.
    Data is not retrievable after it has been delivered.

    Maximum upload size: 64 MB
    Not allowed: anything illegal, malicious, inappropriate, etc.

    This is a personal project and makes no guarantees on:
    reliability, performance, privacy, etc.

    Default Mode:

    If data is sent to the pipe when no receivers are listening, 
    it will be dropped and is not retrievable.

    Fail Mode: 

    $ curl -T- https://pipeto.me/<key>?mode=fail
    In this mode, a send request will fail if no receivers are listening.

    Block Mode:

    $ curl -T- --expect100-timeout 86400 https://pipeto.me/<key>?mode=block
    In this mode, a send request will wait to send data until a receiver connects.

    Interactive Mode:

    $ curl -T. -u <username>: https://pipeto.me/<key>?mode=interactive
    In this mode the system will append the username to messages.
    The system will also send connected and disconnected notifications.

SEE ALSO
    Demo: https://raw.githubusercontent.com/jpschroeder/pipe-to-me/master/demo.gif
    Source: https://github.com/jpschroeder/pipe-to-me
```

## Installation

With go installed:
```shell
go get -u github.com/jpschroeder/pipe-to-me
```

## Usage

```shell
pipe-to-me -h
  -baseurl string
        the base url of the service
         (default "http://localhost:8080/")
  -httpaddr string
        the address/port to listen on for http
        use :<port> to listen on all addresses
         (default "localhost:8080")
```

## Building

In order to build the project, just use:
```shell
go build
```

## Deploying

You can build the project under linux (or Windows Subsystem for Linux) and just copy the executable to your server.

You can then run the program directly or use systemd to install it as a service and keep it running.

Customize the `scripts/pipe-to-me.service` file in the repo for your server and copy it to `/lib/systemd/system/pipe-to-me.service` (ubuntu)

Start the app with: `systemctl start pipe-to-me`  
Enable it on boot with: `systemctl enable pipe-to-me`  
Check it's status with: `systemctl status pipe-to-me`  
See standard output/error with: `journalctl -f -u pipe-to-me`

### NGINX

You can host the application using go directly, or you can listen on a local port and use nginx to proxy connections to the app.

Make sure that nginx is installed with: `apt-get install nginx`

Customize `scripts/pipe-to-me.nginx.conf` and copy it to `/etc/nginx/sites-available/pipe-to-me.nginx.conf`

Remove the default website configuration: `rm /etc/nginx/sites-enabled/default`

Enable the go proxy: `ln -s /etc/nginx/sites-available/pipe-to-me.nginx.conf /etc/nginx/sites-enabled/pipe-to-me.nginx.conf`

Restart nginx to pick up the changes: `systemctl restart nginx`

### NGINX HTTPS

If running as a stand-alone go application, you can use the built-in https support.  When running behind a proxy, you should enable https in nginx and forward to the localhost http address.

Install the letsencrypt client with: 

```shell
add-apt-repository ppa:certbot/certbot
apt-get install python-certbot-nginx
```

Generate and install a certificate with: `certbot --nginx -d pipeto.me`

The certificate should auto-renew when necessary.

### Docker

A prepared Docker image is available [here](https://hub.docker.com/r/jpschroeder/pipe-to-me).

`docker run --name pipe-to-me -p 80:8080 jpschroeder/pipe-to-me -baseurl http://localhost/`

You can build the image yourself by running `docker-compose up --build`.
