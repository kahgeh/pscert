# Purpose

Retrieve valid certificates from aws parameter store and save them to `/etc/letsencrypt/live/[domain name]`.

If certificates do not exist or is out of date ( specified with parameter `--valid-days` ), it will call letsencrypt certbot to generate certificate.

The newly generated certificate is save to parameter store.  

