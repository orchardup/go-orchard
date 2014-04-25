Change log
==========

2.0.6 (2014-04-25)
------------------

- Added `orchard ip` command for printing a host's IP address to stdout
- New Orchard API URL

2.0.5 (2014-03-05)
------------------

- `orchard proxy` prints instructions for setting the DOCKER_HOST environment variable

2.0.4 (2014-02-18)
------------------

- `orchard hosts rm -f` can be used to bypass confirmation
- Fix: `orchard proxy` wasn't picking up -H argument

2.0.3 (2014-02-18)
------------------

- API token can be set with ORCHARD_API_TOKEN environment variable
- `orchard proxy` can be given an explicit URL to listen on (either `tcp://` or `unix://`)

2.0.2 (2014-02-18)
------------------

- Sensible error in 'orchard docker' if host isn't running

2.0.1 (2014-02-18)
------------------

- Fix: host certificate was incorrect

2.0.0 (2014-02-16)
------------------

Initial release.

