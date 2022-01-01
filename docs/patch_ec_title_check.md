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
  - Is the channel a service title?  (`00010100`)

If any of these are true, installation of the title is permitted.
Otherwise, installation is forbidden.

Two other functions perform these four conditionals as well: `ec::isManagedTitle`, and `ec::isManagedTicket`. They are invoked when retrieving title information and when deleting titles.

It was identified that upon unregistration, a function named `ec::removeAllTitles` is called. This function loops through all installed titles, checking whether they are managed titles. If the title is managed, its ticket is removed and its title contents are deleted.

Simply returning that all titles are managed exposes a large risk of deletion. Several options were discussed on how to approach this:
  - Have all titles and tickets be managed, and ensure that the user is never unregistered
    - While possible, not worth the risk.
  - Possibly add hidden titles as managed (since we'll have the installation stub there)
    - While far more safe than the first bullet point, we will still delete all NAND channels upon unregistration.
  - Nullify the deletion function on unregister and hope that nothing else is like this
    - This approach was chosen. This patch set will be refined if other mass-deletion functions are identified.

## Execution
This behavior is not ideal. Four patches are applied:
  - `ec::allowDownloadByApp`, `ec::isManagedTitle`, and `ec::isManagedTicket` are all patched to immediately return `1`, or true.
  - `ec::removeAllTitles` immediately returns `0`, preventing all damage. Its return value is seemingly the amount of titles remaining.