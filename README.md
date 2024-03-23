<pre style="text-align:center;"><p align="center">
 ▄█        ▄█   ▄█         ▄█      ███         ▄█    █▄    
███        ███  ███        ███  ▀█████████▄    ███    ███   
███        ███▌ ███        ███▌    ▀███▀▀██    ███    ███   
███        ███▌ ███        ███▌     ███   ▀   ▄███▄▄▄▄███▄▄ 
███        ███▌ ███        ███▌     ███      ▀▀███▀▀▀▀███▀  
███        ███  ███        ███      ███        ███    ███   
███▌    ▄  ███  ███▌    ▄  ███      ███        ███    ███   
       █████▄▄██  █▀   █████▄▄██  █▀      ▄████▀      ███    █▀           

v1.0.0

<b>Use at your own risk</b>. Criticism and suggestions are welcome.
</pre>         

LILITH is a stream cipher written in pure Go. It implements the algorithm described in the paper [Kashmar, Dr & Ismail, Eddie Shahril. (2017). Blostream: A high speed stream cipher. Journal of Engineering Science and Technology. 12. 1111-1128](https://www.researchgate.net/publication/316942854_Blostream_A_high_speed_stream_cipher).

## Installation

You'll need [Go](https://go.dev/) to build this cipher.

```bash
git clone https://github.com/pre-eth/lilith.git
cd lilith
go build
./lilith -h
```

## Command Line

```
                                           [OPTIONS]
  -d	Decrypt the provided file.
  -e	Encrypt the provided file.
  -f string
    	File name where input is read from
  -j	Save decrypted output as a JPEG
  -n string
    	File name containing 96-bit nonce. Must be a binary file.
  -o string
    	File name where output is written to.
  -p	Save decrypted output as a PNG
  -q	Quick mode - reduce interface FX
  -r	Round trip operation - encrypt and then decrypt
  -s string
    	File name containing 128-bit seed. Must be a binary file.
  -t	Save decrypted output as a text file
  -v	Version of this software (1.0.0)
```

## Using In Your Project

`go get github.com/pre-eth/lilith@v1.0.0`