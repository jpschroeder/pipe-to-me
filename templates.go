package main

import "text/template"

func templates() *template.Template {
	tmpl := template.Must(template.New("").Parse(""))

	template.Must(tmpl.New("home").Parse(`
pipeto.me(1)                     PIPE TO ME                         pipeto.me(1)

NAME
    pipeto.me: streaming data over http

SYNOPSIS
    Randomly generated pipe address:                  {{ .Url }}

EXAMPLES
    Input example:

    (chrome/firefox): {{ .Url }}
    (terminal)$ curl -T- {{ .Url }}
                hello world<enter>

    Pipe example:

    (terminal1)$ curl {{ .Url }}
    (terminal2)$ echo hello world | curl -T- {{ .Url }}

    Chat example(curl>=7.68):

    (terminal1)$ curl -T. {{ .Url }}
    (terminal2)$ curl -T. {{ .Url }}
                 hello world<enter>

    File transfer example:

    (terminal1)$ curl {{ .Url }} > output.txt
    (terminal2)$ cat input.txt | curl -T- {{ .Url }}

    Watch log example:

    (chrome/firefox): {{ .Url }}
    (terminal)$ tail -f logfile | curl -T- {{ .Url }}

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

    $ curl -T- {{ .Url }}?mode=fail
    In this mode, a send request will fail if no receivers are listening.

    Block Mode:

    $ curl -T- --expect100-timeout 86400 {{ .Url }}?mode=block
    In this mode, a send request will wait to send data until a receiver connects.

SEE ALSO
    Demo: https://raw.githubusercontent.com/jpschroeder/pipe-to-me/master/demo.gif
    Source: https://github.com/jpschroeder/pipe-to-me
	`))

	template.Must(tmpl.New("stats").Parse(`
pipeto.me(1)                     PIPE TO ME                         pipeto.me(1)

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
