# Patch: Change EC Configuration Path

## Motivation
We want to store our own custom credentials for the Wii Shop Channel.
Between provisioning tickets, identifying users, and permitting download, this is essential to our operation.

However, this file likely contains legitimate identifiers to Nintendo's services. We do not want to have users unable to download purchases of any kind should they wish.

While Nintendo's servers may reprovision credentials as part of syncing, we do not want to risk the possibility of such not occurring for whatever reason.

## Explanation
The Wii Shop Channel provides a file at `/title/00010002/48414241/data/ec.cfg`. It is utilized by every application that invokes EC - DLC applications, rentals, the Wii Shop, so forth.

This file contains persistent data, such as the provisioned account ID and device token, utilized for authentication. It additionally persists values such as the last sync time.

Other identifiers can be set via `ec.getPersistentValue(name)` or `ec.setPersistentValue(name, value)` within [ECommerceInterface](https://docs.oscwii.org/wii-shop-channel/js/ec/ecommerceinterface).

It is quite important that our changes do not modify or overwrite the user's existing credentials.

We evaluated several solutions:
  - Not using account identifiers, and accepting any existing identifiers that hit our server.
    - While most likely possible, it would be nice to not need to do so. Additionally for new clients, credentials we generate will not be accepted by Nintendo later.
  - Similar to the above, utilizing our own set of identifiers via the exposed JS APIs with custom value names.
    - This would prove complicated for similar reasons to the above.
  - Modifying the configuration file's name.
    - This is the most concise solution, and was chosen.

## Execution
We replace the only instance of `ec.cfg` with `osc.cfg`, where OSC refers to the Open Shop Channel.

A similar change will be necessary for any DLC titles wishing to take advantage of the Open Shop Channel's backend.