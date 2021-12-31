# Patch: Overwrite IOS Syscall for ES

## Motivation
When installing a custom title via the Wii Shop Channel, ES expects valid signatures and certificates.
Most homebrew-based WAD installers apply IOS patches or manually insert contents to their proper place.
However, as we use Nintendo's official code, we lack this luxury.

When EC finishes downloading a title, it calls [ES](https://wiibrew.org/wiki//dev/es) within IOS to handle title installation.
Along the process, the WAD's certificate is verified, and content signatures are verified.
This is done via `IOSC_VerifyPublicKeySign`, an [IOS syscall](https://wiibrew.org/wiki/IOS/Syscalls) available for ES.

As we are not Nintendo and lack their private key, we cannot sign officially, and fail all checks.
We need to bypass this.

## Explanation
It was determined that the best solution was to overwrite `IOSC_VerifyPublicKeySign` itself.

The Wii Shop Channel uses IOS 56. Across a Wii and a vWii, analysis of the IOSP module (roughly its primary kernel)
shows that the `IOSC_VerifyPublicKeySign` syscall has its function present at `0x13a73ad4` within MEM2. (Dolphin silently ignores this write.)

The function - in ARM THUMB mode - looks similar to the following (as copied from Ghidra):
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


This leaves us with the challenge of writing 4 bytes - `20 00 47 70` - to `0xd3a73ad4` as identified previously.

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

We consolidate these four functions into a single function at `0x80014420` - our own `textinput::EventObserver::doNothing`, if you will.
We must additionally update references to this single at two separate virtual tables:
  - `textinput::EventObserver` at `0x802f7a9`
  - `ipl::keyboard::EventObserver` at `0x802f8418`

---
Next, we need to devise a way to have the channel overwrite IOS memory.

We have carved out our own space at `0x80014428` to put a function.
Thankfully, the operation is fairly simple:
 - Write to [`MEM_PROT`](https://wiibrew.org/wiki/Hardware/Hollywood_Registers) and disable it. It is on by default.
   - We use `0xcd8b420a` instead of `0x0d8b420a` as that appears to be where it is mapped for us.
   - Dolphin appears to silently ignore this MMIO access. One day, we may want to not apply the patch should we be able to open `/dev/dolphin`.
 - Write `20 00 47 70` as described above to `0x13a73ad4` to negate `IOSC_VerifyPublicKeySign`.
   - Again, we must actually use `0xd3a73ad4` due to mapping.
   - Dolphin once again appears to ignore this, thankfully.
 - Clear cache
   - TODO: Is this actually functional?

We write and apply the following PowerPC assembly to achieve this task:
```asm
overwriteIOSPatch:
  ; Load 0x0d8b420a, location of MEM_PROT, to r9.
  lis r9, 0xcd8b
  ori r9, r9, 0x420a
  ; We wish to write 0x2 in order to disable.
  li r10, 0x2

  ; And... write!
  sth r10, 0x0(r9)
  eieio
    
  ; Load 0xd3a73ad4, location of of IOSC_VerifyPublicKeySig, to r9.
  lis r9, 0xd3a7
  ori r9, r9, 0x73ad4
  ; 0x20004770 represents our actual patch.
  lis r10, 0x2000
  ori r10, r10, 0x4770

  ; And... write.
  stw r10, 0x0(r9)

  ; Clear cache
  dcbi 0, r10
  blr
```

---
Finally, we need to determine the best way to call our custom patching function.
Using the aforementioned symbols we find `ES_InitLib`, called once during initialization to open a handle with `/dev/es`.

We insert a call to our function in its epilog, immediately before loading the previous LR from stack and branching back.
This makes its flow roughly the following pseudocode:
```c
int ES_InitLib() {
  int fd = 0;
  if (ES_HANDLE < 0) {
    ES_HANDLE = IOS_Open("/dev/es", 0);
    if (ES_HANDLE < 0) {
      fd = ES_HANDLE;
    }
  }
  
  // Our custom code
  overwriteIOSMemory();
  
  return fd;
}
```
