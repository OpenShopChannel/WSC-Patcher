# Patch: Change Base Domain

## Motivation
Ranging from HTML/JS/etc loaded from the browser portion of the shop channel to ES underneath, traffic points to Nintendo's servers without other modifications.

We need to have all traffic directed to our controlled servers.

## Explanation
With Opera's `myfilter.ini`, we simply insert our custom URL to the allowed domains on the filter list.

Unfortunately, we cannot simply `s/shop.wii.com/$USER_BASE_DOMAIN/` in the main DOL, no matter how appealing as that sounds. Most prominently, we need to handle padding for domain names shorter than `shop.wii.com`.

Thankfully, for all types present, we can replace the domain name replaced and add padding to match if necessary.

We can identify five types of URLs within the main DOL to patch:
  - `https://oss-auth.shop.wii.com/startup?initpage=showManual&titleId=`
    - For an unknown reason, this specific URL (alongside several others) is present 9 times within the main DOL. It is only referenced once. The other 8 occurrences appear to be directly after the data segment for other JS plugins.
  - `https://oss-auth.shop.wii.com/oss/getLog`
    - It is unclear on what this is for, as it is never requested whatsoever or accessed during normal runtime. Perhaps an earlier version of the Wii Shop Channel posted `ec.getLog()` to it.
    - It appears directly after the data for `wiiShop`'s JS plugin, implying it goes unused. The similarly-named `getLogUrl` Setter within `wiiShop` appears to be stubbed out.
  - `.shop.wii.com`
    - This suffix is compared on all pages. If the loaded page's domain does not match, most EC functionality is disabled.
  - `https://oss-auth.shop.wii.com`
    - Similar to the first, this appears 9 times and is only referenced once.
  - `https://ecs.a.taur.cloud/ecs/services/ECommerceSOAP`
    - Appears to be present for `GetECConfig`, which is most likely not called within the Wii Shop channel. Instead, `ECommerceInterface#setWebSvcUrls` is preferred.

## Execution
We simply iterate through these 5 types of URLs, replace the domain, and pad if appropriate. Doing so allows us to not fragment the rest of the URL with null bytes should padding be added.