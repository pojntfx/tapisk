# tapisk

![Logo](./docs/logo-readme.png)

Expose a tape drive as a block device.

## Overview

🚧 This project is a work-in-progress! Instructions will be added as soon as it is usable. 🚧

## Contributing

To contribute, please use the [GitHub flow](https://guides.github.com/introduction/flow/) and follow our [Code of Conduct](./CODE_OF_CONDUCT.md).

To build and start a development version of tapisk locally, run the following:

```shell
# (Optional) install the MHVTL virtual tape library to test tapisk without having to use a physical tape drive
git clone https://github.com/markh794/mhvtl.git /tmp/mhvtl
cd /tmp/mhvtl

cd kernel

make
sudo make install

sudo modprobe mhvtl

cd ..

make
sudo make install

sudo systemctl daemon-reload
sudo systemctl enable --now mhvtl.target

sudo usermod -a -G tape ${USER}
newgrp tape

# Install tapisk
$ git clone https://github.com/pojntfx/tapisk.git
$ cd tapisk

# If you have a tape library
$ lsscsi -g # Find your tape library (`/dev/sgX`)
$ mtx -f /dev/sg1 status
$ mtx -f /dev/sg1 load 1 # Load a tape into your drive

$ lsscsi -g # Find your tape drive (`/dev/nstX`)
$ mt -f /dev/nst1 status
$ sudo mt -f /dev/nst1 stsetoptions scsi2logical # Enable the `tell` syscall on your tape drive
$ mt -f /dev/nst1 setblk 4096 # Set the block size
$ mt -f /dev/nst1 rewind
$ mt -f /dev/nst1 weof # Erase the tape
$ mt -f /dev/nst1 rewind

$ make
$ sudo make install
$ sudo rm -f /tmp/tapisk.db && sudo tapisk --drive /dev/nst1 --index /tmp/tapisk.db # Start the block device; the block device path (/dev/nbdX) will be logged to stdout

# In another terminal
$ sudo mkfs.ext4 /dev/nbd0 # Format the tape
$ sudo time sync -f ~/Downloads/mnt; sudo umount ~/Downloads/mnt; sudo rm -rf ~/Downloads/mnt && sudo mkdir -p ~/Downloads/mnt && sudo mount /dev/nbd0 ~/Downloads/mnt && sudo chown -R "${USER}" ~/Downloads/mnt # Mount the tape to ~/Downloads/mnt

$ cat ~/Downloads/mnt/test; echo "Current date: $(date)" | tee ~/Downloads/mnt/test && cat ~/Downloads/mnt/test; sync -f ~/Downloads/mnt/test # Test the filesystem by reading & writing the current date to a file
```

Have any questions or need help? Chat with us [on Matrix](https://matrix.to/#/#tapisk:matrix.org?via=matrix.org)!

## License

tapisk (c) 2023 Felicitas Pojtinger and contributors

SPDX-License-Identifier: AGPL-3.0
