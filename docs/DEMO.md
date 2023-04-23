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
go install github.com/pojntfx/go-nbd/cmd/go-nbd-example-client@latest

lsscsi -g # Find the tape library you want to use (here: `/dev/sg2`)

mtx -f /dev/sg2 status
mtx -f /dev/sg2 load 1 # Loads tape drive (here: `/dev/nst4`)
mt -f /dev/nst4 setblk 512
mt -f /dev/nst4 status

mt -f /dev/nst4 rewind && mt -f /dev/nst4 erase && rm -f /tmp/tapisk.db && go run . --dev /dev/nst4 --index /tmp/tapisk.db

sudo umount ~/Downloads/mnt; sudo $(which go-nbd-example-client) --file /dev/nbd0

sudo mkfs.ext4 /dev/nbd0

sync -f ~/Downloads/mnt/; sudo umount ~/Downloads/mnt; sudo rm -rf ~/Downloads/mnt && sudo mkdir -p ~/Downloads/mnt && sudo mount /dev/nbd0 ~/Downloads/mnt && sudo chown ${USER} -R ~/Downloads/mnt && cat ~/Downloads/mnt/test; echo "Current date: $(date)" | tee ~/Downloads/mnt/test && cat ~/Downloads/mnt/test
```

## Using a Separate Journal Device

```shell
go install github.com/pojntfx/go-nbd/cmd/go-nbd-example-server-file@latest
go install github.com/pojntfx/go-nbd/cmd/go-nbd-example-client@latest

rm -f /tmp/disk.img && truncate -s 10G /tmp/disk.img && go-nbd-example-server-file --file /tmp/disk.img

sudo $(which go-nbd-example-client) --file /dev/nbd0

mt -f /dev/nst4 rewind && mt -f /dev/nst4 erase && rm -f /tmp/tapisk.db && go run . --dev /dev/nst4 --index /tmp/tapisk.db --laddr ':10810'

sudo umount ~/Downloads/mnt; sudo $(which go-nbd-example-client) --raddr 'localhost:10810' --file /dev/nbd1

sudo mke2fs -O journal_dev /dev/nbd0

sudo mkfs.ext4 -J device=/dev/nbd0 /dev/nbd1

sudo tune2fs -l /dev/nbd1 | grep -i journal

sudo blkid | grep 'f19ccdb6-b828-4fe9-a65f-ec2910e56b95'

sync -f ~/Downloads/mnt; sudo umount ~/Downloads/mnt; sudo rm -rf ~/Downloads/mnt && sudo mkdir -p ~/Downloads/mnt && sudo mount /dev/nbd1 ~/Downloads/mnt && sudo chown ${USER} -R ~/Downloads/mnt && cat ~/Downloads/mnt/test; echo "Current date: $(date)" | tee ~/Downloads/mnt/test && cat ~/Downloads/mnt/test && sync -f ~/Downloads/mnt
```
