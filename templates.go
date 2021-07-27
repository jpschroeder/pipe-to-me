package main

import "text/template"

func templates() *template.Template {
	tmpl := template.Must(template.New("").Parse(""))

	template.Must(tmpl.New("home").Parse(`pipeto.me(1)                     PIPE TO ME                         pipeto.me(1)

NAME
    pipeto.me: streaming data over http

SYNOPSIS
    Randomly generated pipe address:                  {{ .URL }}

EXAMPLES
    Pipe example:

    (chrome/firefox): {{ .URL }}
    (terminal2)$ echo hello world | curl -T- {{ .URL }}

    Input example:

    (chrome/firefox): {{ .URL }}
    (terminal)$ curl -T- {{ .URL }}
                hello world<enter>

    File transfer example:

    (terminal1)$ curl {{ .URL }} > output.txt
    (terminal2)$ cat input.txt | curl -T- {{ .URL }}

    Chat example(curl>=7.68):

    (terminal1)$ curl -T. -u user1: {{ .URL }}?mode=interactive
    (terminal2)$ curl -T. -u user2: {{ .URL }}?mode=interactive
                 hello world<enter>

DESCRIPTION
    Data is not buffered or stored in any way.
    Data is not retrievable after it has been delivered.

    Maximum upload size: {{ .MaxUploadMb }} MB
    Not allowed: anything illegal, malicious, inappropriate, etc.

    This is a personal project and makes no guarantees on:
    reliability, performance, privacy, etc.

    Default Mode:

    If data is sent to the pipe when no receivers are listening, 
    it will be dropped and is not retrievable.

    Fail Mode: 

    $ curl -T- {{ .URL }}?mode=fail
    In this mode, a send request will fail if no receivers are listening.

    Block Mode:

    $ curl -T- --expect100-timeout 86400 {{ .URL }}?mode=block
    In this mode, a send request will wait to send data until a receiver connects.

    Interactive Mode:

    $ curl -T. -u <username>: {{ .URL }}?mode=interactive
    In this mode the system will append the username to messages.
    The system will also send connected and disconnected notifications.

SEE ALSO
    Demo: https://raw.githubusercontent.com/jpschroeder/pipe-to-me/master/demo.gif
    Source: https://github.com/jpschroeder/pipe-to-me
	`))

	template.Must(tmpl.New("stats").Parse(`pipeto.me(1)                     PIPE TO ME                         pipeto.me(1)

STATISTICS

    Connected Pipes:        {{ .Active.PipeCount }}
    Connected Receivers:    {{ .Active.ReceiverCount }}
    Connected Senders:      {{ .Active.SenderCount }}
    Connected Sent:         {{ .Active.BytesSent }} ({{ .Active.MegaBytesSent }} MB)

    Total Pipes:            {{ .Global.PipeCount }}
    Total Receivers:        {{ .Global.ReceiverCount }}
    Total Senders:          {{ .Global.SenderCount }}
    Total Sent:             {{ .Global.BytesSent }} ({{ .Global.MegaBytesSent }} MB)
	`))

	return tmpl
}
