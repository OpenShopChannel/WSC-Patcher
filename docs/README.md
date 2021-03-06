# Docs
This directory contains documentation about given patches applied.

Contents:
 - [`opcacrt6.yml`](opcacrt6.yml): A [Kaitai](https://kaitai.io) structure describing a very basic `opcacrt6.dat`.
It does not attempt to handle things such as client certificates or user passwords.
 - [`patch_overwrite_ios.md`](patch_overwrite_ios.md): An explanation over why and how IOS is patched for operation of the Wii Shop Channel.
 - [`patch_custom_ca_ios.md`](patch_custom_ca_ios.md): The logistics of inserting our custom CA into IOS as well for EC usage.
 - [`patch_base_domain.md`](patch_base_domain.md): Information about what URLs are present within the main DOL and information about patching them.
 - [`patch_ec_title_check.md`](patch_ec_title_check.md): Information about title checks run by EC, and why they were negated.
 - [`patch_ec_cfg_path.md`](patch_ec_cfg_path.md): An explanation over why the path of `ec.cfg` is changed.