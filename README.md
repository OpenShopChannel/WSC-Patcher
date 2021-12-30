# WSC-Patcher

WSC-Patcher applies patches to the Wii Shop Channel, such as the ability to use your own servers and certificates.
This is useful for research and development of services utilizing the WSC.

It is important to read the following so that you will have a usable WAD available upon patch completion.

## Setup
You will need an externally resolvable domain with four subdomains:
 - `oss-auth`, utilized for the Wii Shop's main HTML
 - `ecs`, utilized for ticket syncing and title purchases
 - `ias`, utilized for user registration
 - `ccs`, utilized to download titles

The domain must be equal to or smaller than `shop.wii.com` in length, so 12 characters.

If you do not plan to interact with EC, and plan to solely utilize HTML/JS components of the Wii Shop Channel, only configuring `oss-auth` is acceptable.

If you do not have a domain available, you are welcome to utilize `a.taur.cloud` as the base domain!
This domain resolves to `127.0.0.1`, usable within Dolphin.
It is guaranteed `oss-auth`, `ecs`, `ias`, `cas` (cataloguing, within DLC titles), and `ccs`/`ucs` (cached/uncached content servers) are available.

You may additionally choose to specify a root certificate you already have configured on a server. If so, please provide the public key of the CA within the file `output/root.crt` in DER form.
If not, one will be generated for you.

## Operation
Invoke WSC-Patcher similar to the following:
```
./WSC-Patcher <base domain>
```

Throughout its operation, the patcher will perform the following:
 - Version 20 (latest, as of writing) of the Wii Shop Channel will be downloaded to `cache/original.wad`.
 - If `output/root.cer` is not present, a 1024-bit (RSA), SHA-1 CA certificate will be generated.
   - At the same time, `*.<basedomain>` will be issued for ease of use. See `output/server.pem` and `output/server.key` for usage with nginx or similar servers.
 - Modifications are made to the application's main `.arc` (within content index 2) to permit Opera loading the base domain, and the customized certificates.
 - Patches to the application's main dol are also performed. Please see `docs/patch_<name>.md` for more information on what these contain.
 - The patched WAD is written to disk.
 