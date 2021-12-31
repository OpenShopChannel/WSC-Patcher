# Patch: Load Custom CA within IOS

## Motivation
When interacting with the EC library via JavaScript,
certain functions will remotely POST SOAP requests to the configured server.

We already have a custom CA certificate loaded and trusted within Opera.
However, Opera utilizes its own SSL library with its own separate trust store as IOS only permits loading 3 certificate "slots" at a time for usage.
(Opera's credits cite it utilizing OpenSSL, however it is most likely RSA's [BSAFE](https://en.wikipedia.org/wiki/BSAFE)
as many strings match.)

As EC utilizes NHTTP - utilizing IOS's SSL trust store - we need to additionally load our CA certificate into IOS.

## Explanation
NHTTP is the name of Nintendo's first-party HTTP client on the Wii and other platforms of the time.
All SSL functionality present is via IOS's [`/dev/net/ssl`](https://wiibrew.org/wiki//dev/net/ssl).

Via symbols within the main ARC, we are able to see what functions it has available.

We are able to identify a function named `ipl::netconnect::NetConnect::connectThread` with two tasks:
initializing sockets (i.e. awaiting DHCP), and a second function named `NHTTPi_Startup`.

Among other tasks, this NHTTP-specific startup function creates a separate thread to manage HTTP tasks.
This thread manages things such as sending headers in time, properly sending data, and keeping track of sockets requests.

When a request is ready to connect to a socket, the aptly-named function `NHTTPi_SocConnect` is invoked.
It checks whether the current NHTTP object has SSL enabled, and if so, utilizes `NHTTPi_SocSSLConnect` going forward.

As part of `/dev/net/ssl` as noted above, a few wrapping functions are available from Nintendo's first-party SSL SDK library.
These are utilized by `NHTTPi_SocSSLConnect`.

The following things occur within:
 - `SSLNew`, opening a new SSL socket and returning its file descriptor.
 - The NHTTP object is checked on whether the built-in client certificate in IOS should be loaded.
   - If so, `SSLSetBuiltinClientCert` is called, loading the certificate to a specified index.
   - If not, and a client certificate is present, `SSLSetClientCert` is called, again loading to a specified index.
 - The NHTTP object is checked on whether the built-in Nintendo CA certificate should be loaded.
   - If so, `SSLSetBuiltinRootCA` is called, loading its certificate to a specified index.
   - If not, and a CA certificate is present, `SSLSetRootCA` is called to load the certificate at a given buffer.
 - `SSLConnect` is called.
 - `SSLDoHandshake` is called, optionally throwing an error if present.
 - The function returns, and HTTP request data is written.

## Execution
We have no interest in ever utilizing the built-in root CA. We can edit out this logic, freeing room to only load our root certificate. 

First, we need to determine an appropriate place to put our certificate within the binary itself.
(It has been discussed to potentially utilize `ARCOpen` and similar functions, placing our certificate there.
However, this is tricky, and has not been attempted.)

Within a data segment in our binary, we find 928 consecutive empty bytes of space, starting at `0x802e97b8`.

> This appears to be utilized for a Unicode to Shift-JIS conversion table, utilized for the keyboard.
> 
> It's unclear on why so much data is continuously empty - presumably this is an unallocated region in UTF-16.
> Should this ever cause a conflict with Japanese users entering text, we should switch to another method.
> However, this is unlikely, as Nintendo handles keyboard text as a wchar_t (UTF-16 BE, following Windows' sizing).

Second, we need to modify the flow of checks for `SSLSetBuiltinRootCA`.

The following pseudocode represents the original flow, starting at `0x800acad0`:
```c
int NHTTPi_SocSSLConnect(/* parameters */) {
  // Assuming the NHTTP object type is named NHTTPInternals
  NHTTPInternals* internals;

  if (internals->ca_cert == NULL) {
    int result = SSLSetBuiltinRootCA(internals->ssl_fd, internals->ssl_index);
    if (result != 0) {
      return -1004;
    }
  } else {
    int result = SSLSetRootCA(internals->ssl_fd, internals->ca_cert, internals->cert_length;
    if (iVar4 != 0) {
      return -1004;
    }
  }
}
```

We are able to remove the conditional branch entirely and directly call. Note that `{certLen}` is a placeholder value found by the patcher throughout operation.
```asm
; Our certificate is present at 0x802e97b8 as identified above.
; r4 is the second parameter of SSLSetRootCA, the ca_cert pointer.
lis r4, 0x802e
ori r4, r4, 0x97b8

; r5 is the third parameter of SSLSetRootCA, the cert_length field.
li r5, {certLen}

; r3 is the first parameter of SSLSetRootCA, the ssl_fd.
; We load it exactly as Nintendo does.
lwz r3, 0xac(r28)

; We then continue with logic as Nintendo does.
bl SSLSetRootCA

; Error checking
cmpwi r3, 0x0
beq CONTINUE_CONNECTING

; It has failed.
li r3, 1004
b FUNCTION_EPILOG
```

The remaining bytes after this function can simply be `nop`'d out, and original function flow continues.