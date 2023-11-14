package main

// #include "protocol/packet.h"
import "C"
import (
	"net"
	"unsafe"
)

var conn *net.TCPConn
var err error

func initConnection(ip string, port int) {
	conn, err = net.DialTCP("tcp", nil, &net.TCPAddr{IP: net.ParseIP(ip), Port: port})
	check(err)
	conn.SetNoDelay(true)
	conn.SetKeepAlive(true)
}

// The hc_client_discord_plays_special_input_t packet, see https://github.com/hydra-emu/protocol
// uint16_t frames_before;
// uint16_t frames_during;
// uint16_t frames_after;
// uint8_t button;
// uint16_t sample_every;
// char username[32];

func pressImpl(key ButtonType, name string) {
	var packet C.hc_client_discord_plays_special_input_t
	packet.frames_before = C.uint16_t(0)
	packet.frames_during = C.uint16_t(settings.FramesSteppedPressed)
	packet.frames_after = C.uint16_t(settings.FramesSteppedReleased)
	packet.button = C.uint8_t(key)
	packet.sample_every = C.uint16_t(settings.FramesToSample)
	arr := [32]C.char{}
	for i, c := range name {
		if i >= 30 {
			arr[i] = '.'
			break
		}
		arr[i] = C.char(c)
	}
	packet.username = arr
	var buf [1 + 4 + C.HC_PACKET_SIZE_discord_plays_special_input]byte
	buf[0] = byte(C.HC_PACKET_TYPE_discord_plays_special_input)
	buf[1] = byte(C.HC_PACKET_SIZE_discord_plays_special_input)
	buf[2] = 0
	buf[3] = 0
	buf[4] = 0
	for i := 0; i < int(C.HC_PACKET_SIZE_discord_plays_special_input); i++ {
		buf[5+i] = *(*byte)(unsafe.Pointer(uintptr(unsafe.Pointer(&packet)) + uintptr(i)))
	}
	_, err = conn.Write(buf[:])
	check(err)
}
