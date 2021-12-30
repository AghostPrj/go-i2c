// Package i2c provides low level control over the Linux i2c bus.
//
// Before usage you should load the i2c-dev kernel module
//
//      sudo modprobe i2c-dev
//
// Each i2c bus can address 127 independent i2c devices, and most
// Linux systems contain several buses.

package i2c

import (
	"encoding/hex"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"syscall"
	"time"
)

const (
	// DefaultReadDelay 默认的写入指令与读取数据之间的延迟
	// DefaultReadDelay Default delay between write cmd and read from bus
	DefaultReadDelay = 0
)

// I2C 此结构体用于存储访问的i2c设备的信息
//
// I2C represents a connection to I2C-device.
type I2C struct {
	addr uint8
	bus  int
	rc   *os.File
}

// NewI2C 此方法用于打开一个i2c句柄
// bus为系统的i2c接口的id
// addr为要访问的设备在bus上的地址
//
// NewI2C opens a connection for I2C-device.
// SMBus (System Management Bus) protocol over I2C
// supported as well: you should preliminary specify
// register address to read from, either write register
// together with the data in case of write operations.
func NewI2C(addr uint8, bus int) (*I2C, error) {
	f, err := os.OpenFile(fmt.Sprintf("/dev/i2c-%d", bus), os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	if err := ioctl(f.Fd(), I2C_SLAVE, uintptr(addr)); err != nil {
		return nil, err
	}
	v := &I2C{rc: f, bus: bus, addr: addr}
	return v, nil
}

// GetBus 返回访问的i2c接口id
// GetBus return bus line, where I2C-device is allocated.
func (v *I2C) GetBus() int {
	return v.bus
}

// GetAddr 返回访问的i2c接口上的器件地址
// GetAddr return device occupied address in the bus.
func (v *I2C) GetAddr() uint8 {
	return v.addr
}

func (v *I2C) write(buf []byte) (int, error) {
	return v.rc.Write(buf)
}

// WriteBytes 向器件发送数据
// WriteBytes send bytes to the remote I2C-device. The interpretation of
// the message is implementation-dependent.
func (v *I2C) WriteBytes(buf []byte) (int, error) {
	log.WithField("op", "i2c-write").
		Tracef("Write %d hex bytes: [%+v]", len(buf), hex.EncodeToString(buf))
	return v.write(buf)
}

func (v *I2C) read(buf []byte) (int, error) {
	return v.rc.Read(buf)
}

// ReadBytes 从器件读取数据,返回读取的数据长度
// ReadBytes read bytes from I2C-device.
// Number of bytes read correspond to buf parameter length.
func (v *I2C) ReadBytes(buf []byte) (int, error) {
	n, err := v.read(buf)
	if err != nil {
		return n, err
	}
	log.WithField("op", "i2c-read").
		Tracef("Read %d hex bytes: [%+v]", len(buf), hex.EncodeToString(buf))
	return n, nil
}

// Close 关闭i2c连接
// Close I2C-connection.
func (v *I2C) Close() error {
	return v.rc.Close()
}

// ReadRegBytes 从器件读取n字节数据
// ReadRegBytes read count of n byte's sequence from I2C-device
// starting from reg address.
// SMBus (System Management Bus) protocol over I2C.
func (v *I2C) ReadRegBytes(reg byte, n int) ([]byte, int, error) {

	return v.ReadRegBytesWithDelay(reg, n, DefaultReadDelay)

}

// ReadRegBytesWithDelay 从器件读取n字节数据，在发送读取指令和读取之间附带延迟
// ReadRegBytesWithDelay read count of n byte's sequence from I2C-device
// starting from reg address with delay between send and read.
// SMBus (System Management Bus) protocol over I2C.
func (v *I2C) ReadRegBytesWithDelay(reg byte, n int, delay time.Duration) ([]byte, int, error) {
	log.WithField("op", "i2c-read").WithField("delay", delay).
		Tracef("Read %d bytes starting from reg 0x%0X...", n, reg)
	_, err := v.WriteBytes([]byte{reg})
	if err != nil {
		return nil, 0, err
	}
	buf := make([]byte, n)
	time.Sleep(delay)
	c, err := v.ReadBytes(buf)
	if err != nil {
		return nil, 0, err
	}
	return buf, c, nil
}

// ReadRegU8 从器件读取1字节数据
// ReadRegU8 reads byte from I2C-device register specified in reg.
// SMBus (System Management Bus) protocol over I2C.
func (v *I2C) ReadRegU8(reg byte) (byte, error) {
	return v.ReadRegU8WithDelay(reg, DefaultReadDelay)
}

// ReadRegU8WithDelay 从器件读取1字节数据，在发送读取指令和读取之间附带延迟
// ReadRegU8WithDelay reads byte from I2C-device register specified in reg with delay between send and read.
// SMBus (System Management Bus) protocol over I2C.
func (v *I2C) ReadRegU8WithDelay(reg byte, delay time.Duration) (byte, error) {
	_, err := v.WriteBytes([]byte{reg})
	if err != nil {
		return 0, err
	}
	buf := make([]byte, 1)
	time.Sleep(delay)
	_, err = v.ReadBytes(buf)
	if err != nil {
		return 0, err
	}
	log.WithField("op", "i2c-read").
		Tracef("Read U8 %d from reg 0x%0X", buf[0], reg)
	return buf[0], nil
}

// WriteRegU8 向器件写入1字节数据
// WriteRegU8 writes byte to I2C-device register specified in reg.
// SMBus (System Management Bus) protocol over I2C.
func (v *I2C) WriteRegU8(reg byte, value byte) error {
	buf := []byte{reg, value}
	_, err := v.WriteBytes(buf)
	if err != nil {
		return err
	}
	log.WithField("op", "i2c-write").
		Tracef("Write U8 %d to reg 0x%0X", value, reg)
	return nil
}

// ReadRegU16BE 从器件读取2字节无符号数据（大端优先）
// ReadRegU16BE reads unsigned big endian word (16 bits)
// from I2C-device starting from address specified in reg.
// SMBus (System Management Bus) protocol over I2C.
func (v *I2C) ReadRegU16BE(reg byte) (uint16, error) {
	return v.ReadRegU16BEWithDelay(reg, DefaultReadDelay)
}

// ReadRegU16BEWithDelay 从器件读取2字节无符号数据（大端优先），在发送读取指令和读取之间附带延迟
// ReadRegU16BEWithDelay reads unsigned big endian word (16 bits)
// from I2C-device starting from address specified in reg with delay between send and read.
// SMBus (System Management Bus) protocol over I2C.
func (v *I2C) ReadRegU16BEWithDelay(reg byte, delay time.Duration) (uint16, error) {
	_, err := v.WriteBytes([]byte{reg})
	if err != nil {
		return 0, err
	}
	buf := make([]byte, 2)
	time.Sleep(delay)
	_, err = v.ReadBytes(buf)
	if err != nil {
		return 0, err
	}
	w := uint16(buf[0])<<8 + uint16(buf[1])
	log.WithField("op", "i2c-read").
		Tracef("Read U16 %d from reg 0x%0X", w, reg)
	return w, nil
}

// ReadRegU16LE 从器件读取2字节无符号数据（小端优先）
// ReadRegU16LE reads unsigned little endian word (16 bits)
// from I2C-device starting from address specified in reg.
// SMBus (System Management Bus) protocol over I2C.
func (v *I2C) ReadRegU16LE(reg byte) (uint16, error) {
	return v.ReadRegU16LEWithDelay(reg, DefaultReadDelay)
}

// ReadRegU16LEWithDelay 从器件读取2字节无符号数据（小端优先），在发送读取指令和读取之间附带延迟
// ReadRegU16LEWithDelay reads unsigned little endian word (16 bits)
// from I2C-device starting from address specified in reg with delay between send and read.
// SMBus (System Management Bus) protocol over I2C.
func (v *I2C) ReadRegU16LEWithDelay(reg byte, delay time.Duration) (uint16, error) {
	w, err := v.ReadRegU16BEWithDelay(reg, delay)
	if err != nil {
		return 0, err
	}
	// exchange bytes
	w = (w&0xFF)<<8 + w>>8
	return w, nil
}

// ReadRegS16BE 从器件读取2字节有符号数据（大端优先）
// ReadRegS16BE reads signed big endian word (16 bits)
// from I2C-device starting from address specified in reg.
// SMBus (System Management Bus) protocol over I2C.
func (v *I2C) ReadRegS16BE(reg byte) (int16, error) {
	return v.ReadRegS16BEWithDelay(reg, DefaultReadDelay)
}

// ReadRegS16BEWithDelay 从器件读取2字节有符号数据（大端优先），在发送读取指令和读取之间附带延迟
// ReadRegS16BEWithDelay reads signed big endian word (16 bits)
// from I2C-device starting from address specified in reg with delay between send and read.
// SMBus (System Management Bus) protocol over I2C.
func (v *I2C) ReadRegS16BEWithDelay(reg byte, delay time.Duration) (int16, error) {
	_, err := v.WriteBytes([]byte{reg})
	if err != nil {
		return 0, err
	}
	buf := make([]byte, 2)
	time.Sleep(delay)
	_, err = v.ReadBytes(buf)
	if err != nil {
		return 0, err
	}
	w := int16(buf[0])<<8 + int16(buf[1])
	log.WithField("op", "i2c-read").
		Tracef("Read S16 %d from reg 0x%0X", w, reg)
	return w, nil
}

// ReadRegS16LE 从器件读取2字节有符号数据（小端优先）
// ReadRegS16LE reads signed little endian word (16 bits)
// from I2C-device starting from address specified in reg.
// SMBus (System Management Bus) protocol over I2C.
func (v *I2C) ReadRegS16LE(reg byte) (int16, error) {
	return v.ReadRegS16LEWithDelay(reg, DefaultReadDelay)
}

// ReadRegS16LEWithDelay 从器件读取2字节有符号数据（小端优先），在发送读取指令和读取之间附带延迟
// ReadRegS16LEWithDelay reads signed little endian word (16 bits)
// from I2C-device starting from address specified in reg with delay between send and read.
// SMBus (System Management Bus) protocol over I2C.
func (v *I2C) ReadRegS16LEWithDelay(reg byte, delay time.Duration) (int16, error) {
	w, err := v.ReadRegS16BEWithDelay(reg, delay)
	if err != nil {
		return 0, err
	}
	// exchange bytes
	w = (w&0xFF)<<8 + w>>8
	return w, nil

}

// WriteRegU16BE 向器件写入2字节无符号数据（大端优先）
// WriteRegU16BE writes unsigned big endian word (16 bits)
// value to I2C-device starting from address specified in reg.
// SMBus (System Management Bus) protocol over I2C.
func (v *I2C) WriteRegU16BE(reg byte, value uint16) error {
	buf := []byte{reg, byte((value & 0xFF00) >> 8), byte(value & 0xFF)}
	_, err := v.WriteBytes(buf)
	if err != nil {
		return err
	}
	log.WithField("op", "i2c-write").
		Tracef("Write U16 %d to reg 0x%0X", value, reg)
	return nil
}

// WriteRegU16LE 向器件写入2字节无符号数据（小端优先）
// WriteRegU16LE writes unsigned little endian word (16 bits)
// value to I2C-device starting from address specified in reg.
// SMBus (System Management Bus) protocol over I2C.
func (v *I2C) WriteRegU16LE(reg byte, value uint16) error {
	w := (value*0xFF00)>>8 + value<<8
	return v.WriteRegU16BE(reg, w)
}

// WriteRegS16BE 向器件写入2字节有符号数据（大端优先）
// WriteRegS16BE writes signed big endian word (16 bits)
// value to I2C-device starting from address specified in reg.
// SMBus (System Management Bus) protocol over I2C.
func (v *I2C) WriteRegS16BE(reg byte, value int16) error {
	buf := []byte{reg, byte((uint16(value) & 0xFF00) >> 8), byte(value & 0xFF)}
	_, err := v.WriteBytes(buf)
	if err != nil {
		return err
	}
	log.WithField("op", "i2c-write").
		Tracef("Write S16 %d to reg 0x%0X", value, reg)
	return nil
}

// WriteRegS16LE 向器件写入2字节有符号数据（小端优先）
// WriteRegS16LE writes signed little endian word (16 bits)
// value to I2C-device starting from address specified in reg.
// SMBus (System Management Bus) protocol over I2C.
func (v *I2C) WriteRegS16LE(reg byte, value int16) error {
	w := int16((uint16(value)*0xFF00)>>8) + value<<8
	return v.WriteRegS16BE(reg, w)
}

func ioctl(fd, cmd, arg uintptr) error {
	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, fd, cmd, arg, 0, 0, 0)
	if err != 0 {
		return err
	}
	return nil
}
