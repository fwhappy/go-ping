package ping

import (
	"net"
	"time"
)

const (
	SEND_SIZE             = 32
	ECHO_REQUEST_HEAD_LEN = 8
	ECHO_REPLY_HEAD_LEN   = 20
)

// Pinger represents ICMP packet sender/receiver
type Pinger struct {
	// Interval is the wait time between each packet send. Default is 1s.
	Interval time.Duration

	// Timeout specifies a timeout before ping exits, regardless of how many
	// packets have been received. Default is 1s.
	Timeout time.Duration

	// Count tells pinger to stop after sending (and receiving) Count echo
	// packets.
	Count int

	// Debug runs in debug mode
	Debug bool

	// OnFinish is called when Pinger starting run
	OnStart func()

	// OnRecv is called when Pinger receives and processes a packet
	OnRecv func()

	// OnFinish is called when Pinger exits
	OnFinish func()

	OnTimeout func()

	// Number of packets sent
	PacketsSent int

	// Number of packets received
	PacketsRecv int

	ipaddr *net.IPAddr
	addr   string

	ipv4 bool
	// source   string
	size int
	// sequence int
	// network  string

	Stats *statistics
}

type statistics struct {
	SumT   int
	ShortT int
	LongT  int

	// last ping
	TTL         int
	Endduration int
}

// NewPinger pinger creator
func NewPinger(addr string) (*Pinger, error) {
	ipaddr, err := net.ResolveIPAddr("ip", addr)
	if err != nil {
		return nil, err
	}

	var ipv4 bool
	if isIPv4(ipaddr.IP) {
		ipv4 = true
	} else if isIPv6(ipaddr.IP) {
		ipv4 = false
	}

	pinger := &Pinger{
		Interval: time.Second,
		Timeout:  time.Second,
		Count:    4,
		Debug:    false,

		Stats: &statistics{},

		OnStart:  func() {},
		OnRecv:   func() {},
		OnFinish: func() {},

		addr:   addr,
		ipaddr: ipaddr,

		ipv4: ipv4,

		size: SEND_SIZE,
	}
	return pinger, nil
}

// Addr
func (pinger *Pinger) Addr() string {
	return pinger.addr
}

// Size
func (pinger *Pinger) Size() int {
	return pinger.size
}

// Run starting ping
func (pinger *Pinger) Run() error {
	return pinger.ping()
}

func (pinger *Pinger) ping() error {
	starttime := time.Now()
	/*
		cname, _ := net.LookupCNAME(pinger.addr)
		conn, err := net.DialTimeout("ip4:icmp", pinger.addr, pinger.Timeout)
		if err != nil {
			return err
		}
		// ip := conn.RemoteAddr()
		// fmt.Println("正在 Ping " + cname + " [" + ip.String() + "] 具有 32 字节的数据:")
	*/
	pinger.OnStart()

	var seq int16 = 1
	id0, id1 := genidentifier(pinger.addr)
	for {
		pinger.PacketsSent++
		var msg = make([]byte, pinger.size+ECHO_REQUEST_HEAD_LEN)
		msg[0] = 8                        // echo
		msg[1] = 0                        // code 0
		msg[2] = 0                        // checksum
		msg[3] = 0                        // checksum
		msg[4], msg[5] = id0, id1         //identifier[0] identifier[1]
		msg[6], msg[7] = gensequence(seq) //sequence[0], sequence[1]
		length := pinger.size + ECHO_REQUEST_HEAD_LEN
		check := checkSum(msg[0:length])
		msg[2] = byte(check >> 8)
		msg[3] = byte(check & 255)

		conn, err := net.DialTimeout("ip:icmp", pinger.addr, pinger.Timeout)
		if err != nil {
			return err
		}

		starttime = time.Now()
		conn.SetDeadline(starttime.Add(pinger.Timeout))
		_, err = conn.Write(msg[0:length])
		if err != nil {
			return err
		}

		var receive = make([]byte, ECHO_REPLY_HEAD_LEN+length)
		n, err := conn.Read(receive)
		_ = n
		if err != nil {
			// return err
		}

		var endduration = int(int64(time.Since(starttime)) / (1000 * 1000))
		pinger.Stats.Endduration = endduration
		pinger.Stats.SumT += endduration
		time.Sleep(pinger.Interval)

		if err != nil ||
			receive[ECHO_REPLY_HEAD_LEN+4] != msg[4] ||
			receive[ECHO_REPLY_HEAD_LEN+5] != msg[5] ||
			receive[ECHO_REPLY_HEAD_LEN+6] != msg[6] ||
			receive[ECHO_REPLY_HEAD_LEN+7] != msg[7] ||
			// endduration >= int(timeout) ||
			endduration >= int(int64(pinger.Timeout)/(1000*1000)) ||
			receive[ECHO_REPLY_HEAD_LEN] == 11 {
			pinger.OnTimeout()

		} else {
			if pinger.Stats.ShortT == 0 || pinger.Stats.ShortT > endduration {
				pinger.Stats.ShortT = endduration
			}
			if pinger.Stats.LongT < endduration {
				pinger.Stats.LongT = endduration
			}
			pinger.PacketsRecv++
			pinger.Stats.TTL = int(receive[8])
			pinger.OnRecv()
		}

		if pinger.Count == 0 && pinger.PacketsSent >= pinger.Count {
			break
		}
	}
	pinger.OnFinish()
	return nil
}

func checkSum(msg []byte) uint16 {
	sum := 0

	length := len(msg)
	for i := 0; i < length-1; i += 2 {
		sum += int(msg[i])*256 + int(msg[i+1])
	}
	if length%2 == 1 {
		sum += int(msg[length-1]) * 256 // notice here, why *256?
	}

	sum = (sum >> 16) + (sum & 0xffff)
	sum += (sum >> 16)
	var answer = uint16(^sum)
	return answer
}

func gensequence(v int16) (byte, byte) {
	ret1 := byte(v >> 8)
	ret2 := byte(v & 255)
	return ret1, ret2
}

func genidentifier(host string) (byte, byte) {
	return host[0], host[1]
}

func isIPv4(ip net.IP) bool {
	return len(ip.To4()) == net.IPv4len
}

func isIPv6(ip net.IP) bool {
	return len(ip) == net.IPv6len
}
