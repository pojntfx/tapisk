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
mtx -f /dev/sg1 load 1 # Loads into `/dev/st4`
mt -f /dev/st4 status

mt -f /dev/nst4 setblk 512

tar cvf /dev/st4 .
tar tvf /dev/st4

mt -f /dev/nst4 rewind
mt -f /dev/nst4 tell

tar cvf /dev/nst4 .
mt -f /dev/nst4 tell

mt -f /dev/nst4 rewind
tar tvf /dev/nst4
mt -f /dev/nst4 tell

mt -f /dev/nst4 rewind
mt -f /dev/nst4 erase
mt -f /dev/nst4 rewind
```
