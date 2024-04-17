# DNS

## Commands in container
```bash
$ apk add bind-tools strace tmux    # install tools in alpine linux
$ host api.pubnative.net            # resolve a domain name
$ dig api.pubnative.net             # resolve a domain but show more details
$ ping -c1 api.pubnative.net        # resolve a domain name to IP and ping this IP
$ strace ping -c1 api.pubnative.net
    # Check what it's going on under the hood of ping.
    #
    # strace runs a provided command and prints all system calls with all their
    # parameters and results. A system call is a way for a process to ask
    # a (linux) kernel to do something useful, like open a file or send a UDP
    # packet. There is a very good indepth video about system calls and signals,
    # if you want to learn more: https://www.youtube.com/watch?v=d0gS5TXarXc
    #
    # In our case we care that the process opens /etc/hosts and /ets/resolv.conf,
    # gets the DNS server configuration from the latter, creates a UDP socket,
    # and starts to send UDP DNS Requests to the DNS server.
$ tcpdump -nn 'port 53'
    # tcpdump is a tool that intercepts network packets (not only TCP) so that
    # one can figure out what's going on.
    #
    # By default, it will show all packets, which can be slow and spammy.
    # That's why one usually provides a filter, 'port 53' in this case, because
    # DNS uses port 53 and I care about DNS in this demo.
$ tcpdump -nn 'port 53' -A
    # Usually tcpdump prints a short summary of each packet, but one can display
    # the full content with -A, 'A' stands for ASCII. Instead of -A,
    # you can pass -X ('X' stands for heX) and it will print both ASCII and HEX,
    # if you're into that sort of things.
$ tcpdump -nn 'port 53' -w dump.pcap
    # Instead of printing packets to a terminal, write them to a file. It will
    # keep writing until you stop tcpdump with ^c (Ctrl-C).
$ tcpdump -nn 'port 53' -c 10 -w dump.pcap
    # Same as previous, but stop once 10 packets are written to a file.
$ kubectl --context <cluster> cp <container>:dump.pcap dump.pcap
    # Copy a file from the <container> in kubernetes to your local machine, so that
    # you can check packet capture file with Wireshark. This command you
    # supposed to run locally, not in a container.
    #
    # This will work, if you haven't changed a directory before running tcpdump.
    # If you have, then you would need to provide a full path to the file, e.g.
$ kubectl --context <cluster> cp <container>:/root/dump.pcap dump.pcap
    # Copy a file with a full path specified.
```

## Let's make our own DNS server

```bash
$ go run dnsserver.go
```

This will start a DNS server on port 5354, but you need # to have go installed
(`brew install golang`).

```bash
$ dig @127.0.0.1 -p 5354 alvo.me
    # @-symbol specifies what DNS server to use instead of the one specified in 
    # /etc/resolv.conf
```

This is really small and shabby just to illustrate the point. But if you add a few
things here and there (most importantly SOA and NS records), you can use it for
a real production domain (but please don't!). 

Now, you can get 2 server boxes with 2 publicly accessible IP addresses, run
dnsserver.go on each (or better a real
[DNS server](https://en.wikipedia.org/wiki/Comparison_of_DNS_server_software)),
then come to your DNS registrar and tell them yours IP addresses. Eventually, 'root'
DNS servers will learn that your domain is served by the provided IP addresses.

Instead of running your own DNS server, you can get it as a service, like
AWS Route53, Google Cloud DNS, Cloudflare, or GoDaddy.

Anyway, when it's properly set, anyone on the internet can discover your DNS
server and domains that it servers.

## Trace a domain name from the root server
So, how exactly a random guy on the internet will figure out how to reach to
your DNS server? Normally, people use a DNS server from their Internet Service
Provider (ISP) or a one from google (8.8.8.8) or cloudflare (1.1.1.1). This is
a different kind of DNS server, it's called a recursor.

Let's say you sent `api.pubnative.net` address (A) query to a recursor, here is
what happens in the most simplistic scenario.

Each recursor knows IP addresses for the root servers. Those IP addresses are
sort of set in stone and they are not supposed to change any time soon. Root
servers are maintained by different organizations, check
[IANA's page](https://www.iana.org/domains/root/servers)
for more information. From this page we take any IP address and start looking
for an IP of `api.pubnative.net`.

```bash
$ dig @193.0.14.129 api.pubnative.net NS
    # Ask a RIPE-operated DNS root server about what DNS server is
    # responsible for api.pubnative.net. Here we're asking for
    # a NS (Name Server) record.
    #
    # Root server doesn't know about api.pubnative.net, but it knows about
    # who is managing .net zone and returns a bunch of addresses, so we can
    # ask them instead. Let's pick one (192.5.6.30) and ask them the same
    # question:
$ dig @192.5.6.30 api.pubnative.net NS
    # Since this server serves .net zone, it knows that pubnative.net
    # is served by Route53, in particular by ns-649.awsdns-17.net. that has
    # IP 205.251.194.137.
$ dig @205.251.194.137 api.pubnative.net NS
    # Here we're checking, if api.pubnative.net is served by some other server.
    # In this case, it doesn't, so we know that api.pubnative.net is served by
    # this server.
    #
    # The recursion process is stopped, and we can finally get IP address:
$ dig @205.251.194.137 api.pubnative.net A
    # Let's say we've got 192.158.1.38 as a response
```

Now recursor can respond 192.158.1.38 back to the user. Recursor caches all
query responses for some time (specified in the Time-To-Live (TTL) field in DNS
response), so that it doesn't have to repeat this recursion process all the time.

If you don't like your ISP or google, you can setup your own recursor at home,
because the protocols are open and IPs are public. The only problem is that
recursors are doing some caching, which becomes more efficient, if many users
use the same recursor. If everyone would run their own recursors, the load on
root and top level domain name servers will increase significantly and organizations
that operate them will not be happy.

## Links
* https://datatracker.ietf.org/doc/html/rfc1035
* https://www.netmeister.org/blog/dns-size.html
* https://www.youtube.com/watch?v=d0gS5TXarXc
* https://www.iana.org/domains/root/servers
