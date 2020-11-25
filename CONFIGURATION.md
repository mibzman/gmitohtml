`gmitohtml` loads its configuration from `~/.config/gmitohtml/config.yaml` by
default. You may specify a different location via the `--config` argument.

# Configuration options

## Client certificates

Client certificates may be specified via the `Certs` option.

To generate a client certificate, run the following:

```bash
openssl req -x509 -out localhost.crt -keyout localhost.key \
  -newkey rsa:2048 -nodes -sha256 \
  -subj '/CN=localhost' -extensions EXT -config <( \
   printf "[dn]\nCN=localhost\n[req]\ndistinguished_name = dn\n[EXT]\nsubjectAltName=DNS:localhost\nkeyUsage=digitalSignature\nextendedKeyUsage=serverAuth")
```

Files `localhost.crt` and `localhost.key` are generated. Rename these files to
match the domain where the certificate will be used.

## Allow file:// access

By default, local files are not served by gmitohtml. When executed with the
`--allow-file` argument, local files may be accessed via `file://`.

For example, to view `/home/dioscuri/sites/gemlog/index.gmi`, navigate to
`file:///home/dioscuri/sites/gemlog/index.gmi`. 
 
# Example config.yaml

```yaml
certs:
  astrobotany.mozz.us:
    cert: /home/dioscuri/.config/gmitohtml/astrobotany.mozz.us.crt
    key: /home/dioscuri/.config/gmitohtml/astrobotany.mozz.us.crt
  gemini.rocks:
    cert: /home/dioscuri/.config/gmitohtml/gemini.rocks.crt
    key: /home/dioscuri/.config/gmitohtml/gemini.rocks.key

```
