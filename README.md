# tapisk

![Logo](./docs/logo-readme.png)

Expose a tape drive as a block device.

## Overview

🚧 This project is a work-in-progress! Instructions will be added as soon as it is usable. 🚧

## Contributing

To contribute, please use the [GitHub flow](https://guides.github.com/introduction/flow/) and follow our [Code of Conduct](./CODE_OF_CONDUCT.md).

To build and start a development version of tapisk locally, run the following:

```shell
$ git clone https://github.com/pojntfx/tapisk.git
$ cd tapisk

# If you have a tape library
$ lsscsi -g # Find your tape library (`/dev/sgX`)
$ mtx -f /dev/sg1 load 1 # Load a tape into your drive

$ lsscsi -g # Find your tape drive (`/dev/nstX`)
$ mt -f /dev/nst3 setblk 512 # Set the block size
$ mt -f /dev/nst3 rewind
$ mt -f /dev/nst3 erase # Erase the tape

$ rm -f /tmp/tapisk.db && go run . --dev /dev/nst3 --cache /tmp/tapisk.db # Start the NBD server

# In another terminal
$ sudo umount ~/Downloads/mnt; go build -o /tmp/go-nbd-example-client ./cmd/go-nbd-example-client/ && sudo /tmp/go-nbd-example-client --file /dev/nbd1 # Start the NBD client

# In another terminal
$ sudo mkfs.ext4 /dev/nbd1 # Format the tape
$ sudo sync -f ~/Downloads/mnt/test; sudo umount ~/Downloads/mnt; sudo rm -rf ~/Downloads/mnt && sudo mkdir -p ~/Downloads/mnt && sudo mount /dev/nbd1 ~/Downloads/mnt && sudo chown -R "${USER}" ~/Downloads/mnt # Mount the tape to ~/Downloads/mnt

$ cat ~/Downloads/mnt/test; echo "Current date: $(date)" | tee ~/Downloads/mnt/test && cat ~/Downloads/mnt/test; sync -f ~/Downloads/mnt/test # Test the filesystem by reading & writing the current date to a file
```

Have any questions or need help? Chat with us [on Matrix](https://matrix.to/#/#tapisk:matrix.org?via=matrix.org)!

## License

tapisk (c) 2023 Felicitas Pojtinger and contributors

SPDX-License-Identifier: AGPL-3.0
