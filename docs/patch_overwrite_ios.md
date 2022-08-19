# Patch: Overwrite IOS Syscall for ES

## Motivation
When installing a custom title via the Wii Shop Channel, ES expects valid signatures and certificates.
Most homebrew-based WAD installers apply IOS patches or manually insert contents to their proper place.
However, as we use Nintendo's official code, we lack this luxury.

When EC finishes downloading a title, it calls [ES](https://wiibrew.org/wiki//dev/es) within IOS to handle title installation.
Along the process, the WAD's certificate is verified, and content signatures are verified.
This is done via `IOSC_VerifyPublicKeySign`, an [IOS syscall](https://wiibrew.org/wiki/IOS/Syscalls) available for ES.

As we are not Nintendo and lack their private key, we cannot sign officially, and fail all checks. This is not ideal.

Additionally, on a vWii, ES ensures that titles installed are not in group `00000001` (titles such as IOS),
`00010002` (system titles, such as the Wii Shop Channel itself), or `00010008` (hidden titles).
If such titles are installed, ES returns -1017.
As we heavily utilize hidden titles for Homebrew title delivery, this is likewise not ideal.

We need to bypass this.

## Explanation
For the first issue, it was determined that the best solution was to overwrite `IOSC_VerifyPublicKeySign` itself.

The Wii Shop Channel uses IOS 56. Across a Wii and a vWii, analysis of the IOSP module (roughly its primary kernel)
shows that the `IOSC_VerifyPublicKeySign` syscall has its function present at `0x13a73ad4` within MEM2. (Dolphin silently ignores this write.)

The function - in ARM Thumb mode - looks similar to the following (as copied from Ghidra):
```
                     **************************************************************
                     *                                                            *
                     *  FUNCTION                                                  *
                     **************************************************************
                     int __stdcall IOSC_VerifyPublicKeySign(uint8_t * inputDa
                       assume LRset = 0x0
                       assume TMode = 0x1
     int               r0:4           <RETURN>
     uint8_t *         r0:4           inputData
     u32               r1:4           inputSize
     int               r2:4           publicHandle
     uint8_t *         r3:4           signData
     undefined4        Stack[-0x24]:4 local_24                                XREF[1]:     13a73b54(R)  
     undefined4        Stack[-0x28]:4 local_28                                XREF[2]:     13a73ae4(W), 
                                                                                           13a73b8a(R)  
     undefined4        Stack[-0x2c]:4 local_2c                                XREF[2]:     13a73b68(W), 
                                                                                           13a73b8c(W)  
                     IOSC_VerifyPublicKeySign+1                      XREF[0,1]:   ffff9580(*)  
                     IOSC_VerifyPublicKeySign
13a73ad4 b5 f0           push       { r4, r5, r6, r7, lr }
13a73ad6 46 57           mov        r7, r10
13a73ad8 46 4e           mov        r6, r9
13a73ada 46 45           mov        r5, r8
13a73adc b4 e0           push       { r5, r6, r7 }
```

Thankfully, IOS utilizes [TPCS](https://developer.arm.com/documentation/espc0002/1-0), the Thumb Procedure Call Standard.
This states that `r0` is utilized as the first return value, and Ghidra accurately identifies this.

Therefore, we are able to simply return `IOS_SUCCESS` (or 0) to negate the entire check:
```
                     IOSC_VerifyPublicKeySign
13a73ad4 20 00           mov        r0, #0x0
13a73ad6 47 70           bx         lr
```

This leaves us with the challenge of writing 4 bytes - `20 00 47 70` - to `0x13a73ad4` as identified previously.

We next investigate patches needed for a vWii. In pseudocode, Nintendo's check is similar to this:
```c
if (((title_upper == 0x00000001) || (title_upper == 0x00010002)) || (title_upper == 0x00010008)) {
  return -1017;
}
```
In assembly, this often results in three conditionals, all branching to returning an error code on failure, or continuing on.

It was determined that three functions have this logic:
 - `ES_AddTicket` at `0x20102102`
 - `ES_AddTitleStart` at `0x20103240`
 - `ES_AddContentStart` at `0x20103564`

 The easiest patch for them is to simply `b`ranch over these conditionals. As mentioned previously, each instruction in Thumb is two bytes. We therefore write a branch over the conditional immediately followed by a `nop` in order to write 4 bytes.

 However, for alignment purposes, we choose to write at `0x20102100` for `ES_AddTicket`, having our 4 bytes consist of the original instruction previously and then our branch.

## Execution
Ideally, we would not rely on a hardcoded address for this patch, and would instead choose to iterate through memory.
However, space constraints for our patch made this difficult. We chose to hardcode the address.
(Further testing may result in this patch having a different method of application.)

By default, the PowerPC core (Broadway) has memory protections enabled, preventing from us editing IOS's memory in MEM2.
We need to apply several patches to achieve our goal.

---
First, we need to obtain access to overwriting IOS memory.
We set the Access Rights field in the WAD's Title Metadata (or `.tmd`) to 0x00000003, permitting memory access. We will use this with `MEM2_PROT` later.

This is thankfully a very quick fix.

---
Second, we need to find space to put our own custom function within the binary.

Via symbols within the main ARC, we find a C++ class named `textinput::EventObserver` with 4 functions in a row that
immediately `blr` - returning with no other logic:
 - At `0x80014420`, `textinput::EventObserver::onSE`
 - At `0x80014430`, `textinput::EventObserver::onEvent`
 - At `0x80014440`, `textinput::EventObserver::onCommand`
 - At `0x80014450`, `textinput::EventObserver::onInput`

An unrelated function resides at `0x80014460`, so we cannot continue. However, we do not need to.

We consolidate these four functions into a single function at `0x800143f0` - our own `textinput::EventObserver::doNothing`, if you will. 

We then find three related functions before our other vtable members:
 - At `0x800143f0`, `textinput::EventObserver::onOutOfLength` - printing `OutOfLength!`
 - At `0x800143F4`, `textinput::EventObserver::onCancel` - printing `Cancel!`
 - At `0x800143F8`, `textinput::EventObserver::onOK` - printing `OK!`

It was determined that these functions are never called, as their logic is overridden elsewhere. Additionally, the output from OSReport falls nowhere on production systems by default. We coalsce them into the aforementioned `doNothing`.

We shift the `blr` up from `0x80014420` in order to permit space for our patching function.
This all requires us to update references to this single function at two separate virtual tables:
  - `textinput::EventObserver` at `0x802f7a9`
  - `ipl::keyboard::EventObserver` at `0x802f8418`

---

Going forward, we need to permit writing our patches. While apparently unnecessary on a physical Wii, a vWii is quite persistent on what memory ranges we are permitted to access.

For reference purposes, the PowerPC provides a set of "block address translation" (hereforth known as BAT) registers. This permits the processor to map physical memory to another location with parameters such as size and access controls. Per table 7-8, "Upper BAT Register Block Size Mask Encodings" in *PowerPC Microprocessor Family: The Programming Environments, Rev 0.1*, we learn that bits 19 to 29 specify the range mapped (referred to as the block length, or BL).

After observation, it was determined Nintendo sets these relevant BAT layouts by default in both the Wii's and vWii's NANDLoader:
 - `DBAT4`, mapping an unknown lower (likely `0x10000000`) with upper `0x900003ff`.
 - `DBAT6`, mapping `0x12000000` to `0x920001ff`.
 - `DBAT7`, mapping `0x13000000` to `0x930000ff`.

Only 8 bits are set for `DBAT7U`, providing us with only 8 Mb of range (`0x93000000` to `0x93800000`).

IOS' executable code resides in the higher range of MEM2 (for our purposes, `0x939f0000` to `0x93a80000`), falling out of our accessible range.

We must update DBAT7U to contain `0x930001ff`, allowing us 16 Mb - from `0x93000000` to `0x94000000`, technically. We do not exceed `0x93a80000` for our purposes.

---
Next, we need to devise a way to have the channel overwrite IOS memory.

We have carved out our own space at `0x800143f0` to put a function.
Thankfully, the operation is fairly simple:
 - Write `2` to [`MEM_PROT`](https://wiibrew.org/wiki/Hardware/Hollywood_Registers) and disable it. It is on by default.
   - We use `0xcd8b420a` instead of `0x0d8b420a` as that appears to be where it is mapped for us.
   - Dolphin appears to silently ignore this MMIO access. One day, we may want to not apply any patches should we be able to open `/dev/dolphin`.
 - Set `DBAT7U` to `0x930001ff` so that we may match.
 - Write `20 00 47 70` as described above to `0x13a73ad4` to negate `IOSC_VerifyPublicKeySign`.
   - Again, we must actually use `0x93a73ad4` due to mapping.
   - Dolphin once again appears to ignore this, thankfully.
 - Read `0x0d8005a0` and shift it 16 bits in order to determine what console were running on. Per [WiiUBrew](https://wiiubrew.org/wiki/Hardware/Latte_registers), this is `0xcafe`, even in vWii mode.
 - Compare `0xcafe` to our read value. If it is not, we cease patching. Otherwise, we continue.
 - Write our vWii EC patches.
   - These are all branches over.

In order to fully utilize space, we utilize another empty spot on in the binary - a "patch table", if you will. Results are read off one register for brevity.

Please find the full source of this patch under "Insert overwriteIOSMemory" within `patch_overwrite.go`. It is too long to include in full, but is commented.

---
Finally, we need to determine the best way to call our custom patching function.
While debugging issues earlier with memory - throwing DSI exceptions - it was discovered that the Wii Shop Channel has a rudimentary exception handler built in, much like that of [Mario Kart Wii](https://youtu.be/p6RLSFQeeRM?t=130) or [New Super Mario Bros. Wii](https://youtu.be/fnRdSmbOvkw?t=81). It is likely that these are the same implementation, using Nintendo's in-house IPL library (potentially EGG?).

As we wish to show the user that we have crashed, we apply our patch within the epilog of constructor, replacing its `blr` with a simple `b`ranch and utilizing our own. Should our patches crash, we are assured that the handler is registered to catch them.

It was additionally determined that, quite similar to that of MKW or NSMBW, a combination is required to access the handler upon crash. It was thought that the code (on a Wii Remote) would be 2, B, B, Right, Plus, Left, Minus, B; however, this appears to be incorrect. Instead of spending engineering effort to determine the correct code, the check for a combination was simply removed.