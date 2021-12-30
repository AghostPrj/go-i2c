I2C-bus interaction of peripheral sensors with Raspberry PI embedded linux or respective clones
==============================================================================================


[![MIT License](http://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)

This library written in [Go programming language](https://golang.org/) intended to activate and interact with the I2C
bus by reading and writing data.

Compatibility
-------------

Tested on Orange Pi Zero Plus 2 H5.

Golang usage
------------

```go
func main() {
// Create new connection to I2C bus on 2 line with address 0x27
i2c, err := i2c.NewI2C(0x27, 2)
if err != nil { log.Fatal(err) }
// Free I2C connection on exit
defer i2c.Close()
....
// Here goes code specific for sending and reading data
// to and from device connected via I2C bus, like:
_, err := i2c.Write([]byte{0x1, 0xF3})
if err != nil { log.Fatal(err) }
....
}
```

Tutorial
--------

In its turn, go-i2c use [logrus](https://github.com/sirupsen/logrus) library to output debug and other notification's
lines which produce all necessary levels of logging.


Troubleshooting
--------------

- *How to obtain fresh Golang installation to RPi device (either any RPi clone):*
  If your RaspberryPI golang installation taken by default from repository is outdated, you may consider to install
  actual golang manually from official Golang [site](https://golang.org/dl/). Download tar.gz file containing armv6l in
  the name. Follow installation instructions.

- *How to enable I2C bus on RPi device:*
  If you employ RaspberryPI, use raspi-config utility to activate i2c-bus on the OS level. Go to "Interfacing Options"
  menu, to active I2C bus. Probably you will need to reboot to load i2c kernel module. Finally you should have device
  like /dev/i2c-1 present in the system.

- *How to find I2C bus allocation and device address:*
  Use i2cdetect utility in format "i2cdetect -y X", where X may vary from 0 to 5 or more, to discover address occupied
  by peripheral device. To install utility you should run
  `apt install i2c-tools` on debian-kind system. `i2cdetect -y 1` sample output:
  ```
       0  1  2  3  4  5  6  7  8  9  a  b  c  d  e  f
  00:          -- -- -- -- -- -- -- -- -- -- -- -- --
  10: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- --
  20: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- --
  30: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- --
  40: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- --
  50: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- --
  60: -- -- -- -- -- -- -- -- -- -- -- -- -- -- -- --
  70: -- -- -- -- -- -- 76 --    
  ```

License
-------

Go-i2c is licensed under MIT License.
