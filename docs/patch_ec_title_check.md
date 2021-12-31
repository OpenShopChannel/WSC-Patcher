# Patch: Negate EC Title Check

## Motivation
A check on the title type is present, preventing installation of `00010008` (hidden) titles. We would like to do so.

## Explanation
Via symbols within the main ARC, we are able to see function names.

Prior to downloading a title in three scenarios - normal downloading, gifting, or purchasing - EC runs a function called `ec::allowDownloadByApp`.

Within this, four conditions are checked:
  - Is the channel a downloadable title/NAND title? (`00010001`)
  - Is the channel a game channel? This checks two types: 
    - `00010000`, typically used for discs
    - `00010004`.
  - Is the channel a "service title"?  (`00010100`)
    - Name taken from `ec::isServiceTitle`.

If any of these are true, installation of the title is permitted.
Otherwise, installation is forbidden.

## Execution
This behavior is not ideal. `ec::allowDownloadByApp` is patched to immediately return `1`, or true.

In the future, `ec::isManagedTitle` and `ec::isManagedTicket` may wish to be patched as well due to similar reasons.