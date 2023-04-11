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

# Adjust `/dev/*` to your own system's values

mtx -f /dev/sg1 status
mtx -f /dev/sg1 load 1
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

## Local Development

```shell
lsscsi -g # Find the tape library you want to use (here: `/dev/sg2`)

mtx -f /dev/sg2 status
mtx -f /dev/sg2 load 1 # Loads tape drive (here: `/dev/nst4`)
mt -f /dev/nst4 setblk 512
mt -f /dev/nst4 status

mt -f /dev/nst4 rewind && mt -f /dev/nst4 erase && rm -f /tmp/tapisk.db && go run . --dev /dev/nst4 --cache /tmp/tapisk.db

go install github.com/pojntfx/go-nbd/cmd/go-nbd-example-client@latest
sudo umount ~/Downloads/mnt; sudo $(which go-nbd-example-client) --file /dev/nbd0

sudo mkfs.ext4 /dev/nbd0

sudo umount ~/Downloads/mnt; sudo rm -rf ~/Downloads/mnt && sudo mkdir -p ~/Downloads/mnt && sudo mount /dev/nbd0 ~/Downloads/mnt && sudo cat ~/Downloads/mnt/test; echo "Current date: $(date)" | sudo tee ~/Downloads/mnt/test && sudo cat ~/Downloads/mnt/test && sudo sync -f ~/Downloads/mnt/test
```
