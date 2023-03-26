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
sudo mtx -f /dev/sg0 status
sudo mt -f /dev/st0 status
```
