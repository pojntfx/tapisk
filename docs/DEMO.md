# Demo

## Virtual Tape Library/Drive Setup

```shell
git clone https://github.com/markh794/mhvtl.git
cd mhvtl/

cd kernel

make
sudo make install

sudo modprobe mhvtl

cd ..

make
sudo make install

sudo systemctl daemon-reload
sudo systemctl enable --now mhvtl.target

lsscsi -g

sudo usermod -a -G tape ${USER}
newgrp tape

mtx -f /dev/sg1 status
mtx -f /dev/sg1 load 1 # Loads into `/dev/st6`
mt -f /dev/st6 status

mt -f /dev/nst6 setblk 512

tar cvf /dev/st6 .
tar tvf /dev/st6

mt -f /dev/nst6 rewind
mt -f /dev/nst6 tell

tar cvf /dev/nst6 .
mt -f /dev/nst6 tell

mt -f /dev/nst6 rewind
tar tvf /dev/nst6
mt -f /dev/nst6 tell

mt -f /dev/nst6 rewind
mt -f /dev/nst6 erase
mt -f /dev/nst6 rewind
```
