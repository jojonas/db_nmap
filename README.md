# db_nmap

`db_nmap` and `db_import` are glue commands between [Nmap](https://nmap.org/) and [Metasploit](https://www.metasploit.com/):

- `db_nmap` is a wrapper around Nmap that inserts Nmap's results into the Metasploit PostgreSQL database, right after they are finished scanning.
- `db_import` is a standalone program that takes an Nmap result XML document and inserts the results into the Metasploit PostgreSQL daabase.

After importing the results, they can be inspected with the Metasploit console commands `services` and `hosts`.

Both commands are actually standalone implementations of the corresponding commands in Metasploit, which are documented [here (`db_nmap`)](https://www.offensive-security.com/metasploit-unleashed/port-scanning/) and [here (`db_import`)](https://www.offensive-security.com/metasploit-unleashed/using-databases/).

## Example

On my local system, I am currently only running a PostgreSQL database (for Metasploit). I can scan `localhost` with the following command:

    $ db_nmap -sV 127.0.0.1
    INFO[2021-10-19 18:47:52] Connected to Metasploit PostgreSQL database "msf" at 127.0.0.1:5432 as user "jonas"
    DEBU[2021-10-19 18:47:52] ID of Metasploit workspace "default": 1
    DEBU[2021-10-19 18:47:52] Running "/usr/bin/nmap -sV 127.0.0.1 -oX /dev/fd/3" ...
    Starting Nmap 7.92 ( https://nmap.org ) at 2021-10-19 18:47 CEST
    Nmap scan report for localhost (127.0.0.1)
    Host is up (0.0000040s latency).
    Not shown: 999 closed tcp ports (reset)
    PORT STATE SERVICE VERSION
    5432/tcp open postgresql PostgreSQL DB 9.6.0 or later
    1 service unrecognized despite returning data. [...]

    Service detection performed. Please report any incorrect results at https://nmap.org/submit/ .
    Nmap done: 1 IP address (1 host up) scanned in 6.48 seconds
    DEBU[2021-10-19 18:47:59] Inserted/updated host 127.0.0.1 (localhost).
    DEBU[2021-10-19 18:47:59] Inserted/updated service tcp:5432 (PostgreSQL DB).
    DEBU[2021-10-19 18:47:59] XML document complete.
    INFO[2021-10-19 18:47:59] Wrapper stats: registered 1 hosts with 1 services.

Note that `db_nmap` used the `nmap` command found in my `PATH`.
After scanning, the results can be retrieved from the Metasploit database:

    $ msfconsole
    [...]

    msf6 > hosts -c address,name

    Hosts
    =====

    address    name
    -------    ----
    127.0.0.1  localhost

    msf6 > services
    Services
    ========

    host       port  proto  name        state  info
    ----       ----  -----  ----        -----  ----
    127.0.0.1  5432  tcp    postgresql  open   PostgreSQL DB

## Passing options

Options, such as the database host, port, user and password can be passed through PostgreSQL's default environment variables documented [here](https://www.postgresql.org/docs/current/libpq-envars.html), e.g.:

    $ PGUSER=metasploit PGPASSWORD=secret db_nmap -sV 127.0.0.1

In addition, the Metasploit workspace name can be set with the environment variable `MSF_WORKSPACE` (default: `default`):

    $ MSF_WORKSPACE=project2 db_nmap -sV 127.0.0.1

## Building

The project is implemented in Go and can be built as follows:

    go build ./cmd/db_nmap
    go build ./cmd/db_import
